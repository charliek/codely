package tui

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"

	"github.com/charliek/codely/internal/domain"
	"github.com/charliek/codely/internal/shed"
	"github.com/charliek/codely/internal/status"
)

// statusPollCmd returns a command that polls status after the configured interval
func (m *Model) statusPollCmd() tea.Cmd {
	return tea.Tick(m.config.StatusPollIntervalDuration(), func(t time.Time) tea.Msg {
		return TickMsg{}
	})
}

// pollStatusCmd captures pane content and detects status for all sessions
func (m *Model) pollStatusCmd() tea.Cmd {
	return func() tea.Msg {
		updates := make(map[string]domain.Status)

		panes, listErr := m.tmux.ListPanes()
		paneMap := make(map[int]bool)
		if listErr == nil {
			for _, p := range panes {
				paneMap[p.ID] = true
			}
		}

		for _, proj := range m.store.Projects() {
			for _, sess := range proj.Sessions {
				if sess.PaneID == 0 {
					continue
				}

				if listErr == nil && !paneMap[sess.PaneID] {
					updates[sess.ID] = domain.StatusExited
					continue
				}

				content, capErr := m.tmux.CapturePane(sess.PaneID, 15)
				if capErr != nil {
					if listErr == nil && !paneMap[sess.PaneID] {
						updates[sess.ID] = domain.StatusExited
					} else {
						updates[sess.ID] = domain.StatusError
					}
					continue
				}

				updates[sess.ID] = status.Detect(content)
			}
		}

		return StatusUpdateMsg{Updates: updates}
	}
}

// loadFoldersCmd loads available folders from workspace roots
func (m *Model) loadFoldersCmd() tea.Cmd {
	return func() tea.Msg {
		var folders []string

		for _, root := range m.config.WorkspaceRoots {
			// Expand ~
			if strings.HasPrefix(root, "~") {
				home, err := os.UserHomeDir()
				if err != nil {
					continue
				}
				root = home + root[1:]
			}

			// Read directory entries
			entries, err := os.ReadDir(root)
			if err != nil {
				continue
			}

			for _, entry := range entries {
				if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
					folders = append(folders, filepath.Join(root, entry.Name()))
				}
			}
		}

		return FoldersLoadedMsg{Folders: folders}
	}
}

// loadShedsCmd loads available sheds
func (m *Model) loadShedsCmd() tea.Cmd {
	return func() tea.Msg {
		if m.shed == nil || !m.shed.Available() {
			return ShedsLoadedMsg{Sheds: nil, Err: nil}
		}

		sheds, err := m.shed.ListSheds()
		return ShedsLoadedMsg{Sheds: sheds, Err: err}
	}
}

// createPaneCmd creates a new tmux pane for a session.
// Codely always keeps a single visible terminal pane in the main window.
func (m *Model) createPaneCmd(project *domain.Project, session *domain.Session) tea.Cmd {
	// Capture only IDs and immutable data, not the full objects
	// This avoids stale data issues when the closure executes
	projectID := project.ID
	sessionID := session.ID
	projectType := project.Type
	projectDir := project.Directory
	shedName := project.ShedName
	cmdExec := session.Command.Exec
	cmdArgs := session.Command.Args

	return func() tea.Msg {
		var paneID int
		var err error
		var hiddenProjectID string
		var hiddenSessionID string
		var hiddenPaneID int

		// Determine the directory and command
		var dir string
		var execCmd string
		var execArgs []string

		if projectType == domain.ProjectTypeLocal {
			dir = projectDir
			execCmd = cmdExec
			execArgs = cmdArgs
		} else {
			// Shed project: use shed exec
			if m.shed == nil {
				return PaneCreatedMsg{
					ProjectID: projectID,
					SessionID: sessionID,
					Err:       domain.ErrShedNotFound,
				}
			}

			// Build the command string for shed exec
			cmd := m.shed.ExecCommand(shedName, cmdExec, cmdArgs...)
			execCmd = strings.Join(cmd.Args, " ")
			execArgs = nil
			dir = ""
		}

		// Find existing visible terminal pane in Codely's window (excluding Codely's pane).
		var existingTerminalPaneID int
		var existingSessionID string
		panes, listErr := m.tmux.ListPanes()
		if listErr == nil {
			codelyWindowID := m.codelyWindowID
			if codelyWindowID == "" && m.codelyPaneID > 0 {
				for _, p := range panes {
					if p.ID == m.codelyPaneID {
						codelyWindowID = p.WindowID
						break
					}
				}
			}

			if codelyWindowID != "" {
				for _, p := range panes {
					if p.WindowID == codelyWindowID && p.ID != m.codelyPaneID {
						existingTerminalPaneID = p.ID
						break
					}
				}
			}
		}

		if existingTerminalPaneID > 0 {
			for _, proj := range m.store.Projects() {
				for _, sess := range proj.Sessions {
					if sess.PaneID == existingTerminalPaneID {
						existingSessionID = sess.ID
						hiddenProjectID = proj.ID
						break
					}
				}
				if existingSessionID != "" {
					break
				}
			}
		}

		if existingTerminalPaneID > 0 {
			// Hide the currently visible terminal pane using break-pane
			newPaneID, breakErr := m.tmux.BreakPane(existingTerminalPaneID)
			if breakErr != nil {
				// If the window is zoomed, break-pane can fail. Try unzooming once.
				if m.codelyPaneID > 0 {
					_ = m.tmux.ToggleZoom(m.codelyPaneID)
					newPaneID, breakErr = m.tmux.BreakPane(existingTerminalPaneID)
				}
			}
			if breakErr != nil {
				return PaneCreatedMsg{
					ProjectID: projectID,
					SessionID: sessionID,
					Err:       breakErr,
				}
			}
			hiddenSessionID = existingSessionID
			hiddenPaneID = newPaneID
		}

		// Split from Codely's pane (horizontal split to the right)
		if m.codelyPaneID > 0 {
			paneID, err = m.tmux.SplitPane(m.codelyPaneID, false, dir, execCmd, execArgs...)
		} else {
			paneID, err = m.tmux.SplitWindow(dir, execCmd, execArgs...)
		}

		return PaneCreatedMsg{
			ProjectID:       projectID,
			SessionID:       sessionID,
			PaneID:          paneID,
			HiddenProjectID: hiddenProjectID,
			HiddenSessionID: hiddenSessionID,
			HiddenPaneID:    hiddenPaneID,
			Err:             err,
		}
	}
}

// killPaneCmd kills a tmux pane
func (m *Model) killPaneCmd(project *domain.Project, session *domain.Session) tea.Cmd {
	return func() tea.Msg {
		if session.PaneID > 0 {
			_ = m.tmux.KillPane(session.PaneID)
		}

		return PaneKilledMsg{
			ProjectID: project.ID,
			SessionID: session.ID,
		}
	}
}

// focusPaneCmd focuses a tmux pane
func (m *Model) focusPaneCmd(paneID int) tea.Cmd {
	return func() tea.Msg {
		err := m.tmux.FocusPane(paneID)
		return FocusPaneMsg{PaneID: paneID, Err: err}
	}
}

// swapPanesCmd swaps a hidden session to be visible and hides the currently visible one
func (m *Model) swapPanesCmd(showProject *domain.Project, showSession *domain.Session, hideProject *domain.Project, hideSession *domain.Session) tea.Cmd {
	return func() tea.Msg {
		// First, break the currently visible pane to hide it
		hiddenPaneID, err := m.tmux.BreakPane(hideSession.PaneID)
		if err != nil {
			return PaneSwappedMsg{
				ShownProjectID:  showProject.ID,
				HiddenProjectID: hideProject.ID,
				Err:             err,
			}
		}

		// Then, join the hidden pane back to the main window
		shownPaneID, err := m.tmux.JoinPane(showSession.PaneID, m.codelyPaneID)
		if err != nil {
			return PaneSwappedMsg{
				ShownProjectID:  showProject.ID,
				HiddenProjectID: hideProject.ID,
				Err:             err,
			}
		}

		if resolved, ok := m.visiblePaneInCodelyWindow(); ok {
			shownPaneID = resolved
		}

		return PaneSwappedMsg{
			ShownProjectID:  showProject.ID,
			HiddenProjectID: hideProject.ID,
			ShownSessionID:  showSession.ID,
			ShownPaneID:     shownPaneID,
			HiddenSessionID: hideSession.ID,
			HiddenPaneID:    hiddenPaneID,
		}
	}
}

// showPaneCmd brings a hidden pane back into the main window without hiding another.
func (m *Model) showPaneCmd(showProject *domain.Project, showSession *domain.Session) tea.Cmd {
	return func() tea.Msg {
		shownPaneID, err := m.tmux.JoinPane(showSession.PaneID, m.codelyPaneID)
		if err != nil {
			return PaneSwappedMsg{
				ShownProjectID: showProject.ID,
				Err:            err,
			}
		}

		if resolved, ok := m.visiblePaneInCodelyWindow(); ok {
			shownPaneID = resolved
		}

		return PaneSwappedMsg{
			ShownProjectID: showProject.ID,
			ShownSessionID: showSession.ID,
			ShownPaneID:    shownPaneID,
		}
	}
}

func (m *Model) visiblePaneInCodelyWindow() (int, bool) {
	panes, err := m.tmux.ListPanes()
	if err != nil || m.codelyPaneID == 0 {
		return 0, false
	}

	codelyWindowID := m.codelyWindowID
	if codelyWindowID == "" {
		for _, p := range panes {
			if p.ID == m.codelyPaneID {
				codelyWindowID = p.WindowID
				break
			}
		}
	}
	if codelyWindowID == "" {
		return 0, false
	}

	for _, p := range panes {
		if p.WindowID == codelyWindowID && p.ID != m.codelyPaneID {
			return p.ID, true
		}
	}
	return 0, false
}

// syncVisibilityCmd reconciles session visibility based on tmux window state.
func (m *Model) syncVisibilityCmd() tea.Cmd {
	return func() tea.Msg {
		panes, err := m.tmux.ListPanes()
		if err != nil {
			return VisibilitySyncedMsg{Err: err}
		}

		paneWindow := make(map[int]string)
		for _, p := range panes {
			paneWindow[p.ID] = p.WindowID
		}

		codelyWindowID := m.codelyWindowID
		if codelyWindowID == "" && m.codelyPaneID > 0 {
			for _, p := range panes {
				if p.ID == m.codelyPaneID {
					codelyWindowID = p.WindowID
					break
				}
			}
		}

		var visibleSessionID string
		for _, proj := range m.store.Projects() {
			for _, sess := range proj.Sessions {
				if sess.PaneID == 0 || codelyWindowID == "" {
					continue
				}
				if paneWindow[sess.PaneID] == codelyWindowID {
					visibleSessionID = sess.ID
					break
				}
			}
			if visibleSessionID != "" {
				break
			}
		}

		return VisibilitySyncedMsg{
			VisibleSessionID: visibleSessionID,
			Err:              nil,
		}
	}
}

// startShedCmd starts a shed
func (m *Model) startShedCmd(shedName string) tea.Cmd {
	return func() tea.Msg {
		if m.shed == nil {
			return ShedStartedMsg{ShedName: shedName, Err: domain.ErrShedNotFound}
		}
		err := m.shed.StartShed(shedName)
		return ShedStartedMsg{ShedName: shedName, Err: err}
	}
}

// stopShedCmd stops a shed
func (m *Model) stopShedCmd(shedName string) tea.Cmd {
	return func() tea.Msg {
		if m.shed == nil {
			return ShedStoppedMsg{ShedName: shedName, Err: domain.ErrShedNotFound}
		}
		err := m.shed.StopShed(shedName)
		return ShedStoppedMsg{ShedName: shedName, Err: err}
	}
}

// createShedCmd creates a new shed
func (m *Model) createShedCmd(name string, opts shed.CreateOpts) tea.Cmd {
	return func() tea.Msg {
		if m.shed == nil {
			return ShedCreatedMsg{ShedName: name, Err: domain.ErrShedNotFound}
		}
		err := m.shed.CreateShed(name, opts)
		return ShedCreatedMsg{ShedName: name, Err: err}
	}
}

// deleteShedCmd deletes a shed
func (m *Model) deleteShedCmd(name string, force bool) tea.Cmd {
	return func() tea.Msg {
		if m.shed == nil {
			return ShedDeletedMsg{ShedName: name, Err: domain.ErrShedNotFound}
		}
		err := m.shed.DeleteShed(name, force)
		return ShedDeletedMsg{ShedName: name, Err: err}
	}
}

// createProjectCmd creates a new project from a selected folder
func (m *Model) createProjectCmd(folder string) tea.Cmd {
	return func() tea.Msg {
		name := filepath.Base(folder)

		project := &domain.Project{
			ID:        uuid.New().String(),
			Name:      name,
			Type:      domain.ProjectTypeLocal,
			Directory: folder,
			Sessions:  []domain.Session{},
			Expanded:  true,
		}

		return ProjectCreatedMsg{Project: project}
	}
}

// createShedProjectCmd creates a project from a shed
func (m *Model) createShedProjectCmd(s shed.Shed) tea.Cmd {
	return func() tea.Msg {
		project := &domain.Project{
			ID:         uuid.New().String(),
			Name:       s.Name,
			Type:       domain.ProjectTypeShed,
			ShedName:   s.Name,
			ShedServer: s.Server,
			Sessions:   []domain.Session{},
			Expanded:   true,
		}

		return ProjectCreatedMsg{Project: project}
	}
}

// newSession creates a new session for a project with the given command
func newSession(projectID string, cmdID string, cmd domain.Command) *domain.Session {
	return &domain.Session{
		ID:        uuid.New().String(),
		ProjectID: projectID,
		Command: domain.Command{
			ID:          cmdID,
			DisplayName: cmd.DisplayName,
			Exec:        cmd.Exec,
			Args:        cmd.Args,
			Env:         cmd.Env,
		},
		Status:    domain.StatusUnknown,
		StartedAt: time.Now(),
	}
}
