package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/charliek/codely/internal/domain"
	"github.com/charliek/codely/internal/shed"
	"github.com/charliek/codely/internal/tui/components"
)

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.statusPollCmd(),
		m.loadFoldersCmd(),
		m.syncVisibilityCmd(),
	)
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		return m, nil

	case tea.KeyMsg:
		// Clear error on any key
		if m.err != nil {
			m.err = nil
		}

		return m.handleKey(msg)

	case TickMsg:
		// Poll status and schedule next tick
		cmds = append(cmds, m.pollStatusCmd(), m.statusPollCmd())

	case StatusUpdateMsg:
		m.applyStatusUpdates(msg.Updates)

	case FoldersLoadedMsg:
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.folders = msg.Folders
		}

	case ShedsLoadedMsg:
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.sheds = msg.Sheds
		}

	case ProjectCreatedMsg:
		m.handleProjectCreated(msg.Project)
		// Immediately show command picker
		m.mode = ModeCommandPicker
		m.commandIdx = m.defaultCommandIndex()

	case PaneCreatedMsg:
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.handlePaneCreated(msg)
			// Focus the new pane
			cmds = append(cmds, m.focusPaneCmd(msg.PaneID))
		}
		m.mode = ModeNormal

	case PaneKilledMsg:
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.handlePaneKilled(msg)
		}

	case FocusPaneMsg:
		if msg.Err != nil {
			m.err = msg.Err
		}

	case PaneSwappedMsg:
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.handlePaneSwapped(msg)
			// Focus the now-visible pane
			cmds = append(cmds, m.focusPaneCmd(msg.ShownPaneID))
		}

	case VisibilitySyncedMsg:
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.handleVisibilitySynced(msg)
		}

	case ShedStartedMsg:
		if msg.Err != nil {
			m.err = msg.Err
		}
		cmds = append(cmds, m.loadShedsCmd())

	case ShedStoppedMsg:
		if msg.Err != nil {
			m.err = msg.Err
		}
		cmds = append(cmds, m.loadShedsCmd())

	case ShedCreatedMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.mode = ModeNormal
		} else {
			// Create project from the new shed
			for _, s := range m.sheds {
				if s.Name == msg.ShedName {
					cmds = append(cmds, m.createShedProjectCmd(s))
					break
				}
			}
		}
		cmds = append(cmds, m.loadShedsCmd())

	case ShedDeletedMsg:
		if msg.Err != nil {
			m.err = msg.Err
		}
		cmds = append(cmds, m.loadShedsCmd())
		m.mode = ModeNormal

	case ErrorMsg:
		m.err = msg.Err

	case ClearErrorMsg:
		m.err = nil
	}

	return m, tea.Batch(cmds...)
}

// handleKey processes key events based on current mode
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case ModeNormal:
		return m.handleNormalKey(msg)
	case ModeFolderPicker:
		return m.handleFolderPickerKey(msg)
	case ModeCommandPicker:
		return m.handleCommandPickerKey(msg)
	case ModeShedPicker:
		return m.handleShedPickerKey(msg)
	case ModeShedCreate:
		return m.handleShedCreateKey(msg)
	case ModeShedClose:
		return m.handleShedCloseKey(msg)
	case ModeConfirm:
		return m.handleConfirmKey(msg)
	case ModeHelp:
		return m.handleHelpKey(msg)
	case ModeNewProjectType:
		return m.handleNewProjectTypeKey(msg)
	}
	return m, nil
}

// handleNormalKey handles keys in normal mode
func (m Model) handleNormalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		// Save state before quitting
		_ = m.store.Save()
		return m, tea.Quit

	case key.Matches(msg, m.keys.Help):
		m.showHelp = !m.showHelp
		if m.showHelp {
			m.mode = ModeHelp
		}
		return m, nil

	case key.Matches(msg, m.keys.Up):
		m.tree.MoveUp()
		return m, nil

	case key.Matches(msg, m.keys.Down):
		m.tree.MoveDown()
		return m, nil

	case key.Matches(msg, m.keys.Left):
		m.tree.Collapse()
		return m, nil

	case key.Matches(msg, m.keys.Right):
		m.tree.Expand()
		return m, nil

	case key.Matches(msg, m.keys.Enter):
		return m.handleEnter()

	case key.Matches(msg, m.keys.Space):
		m.tree.Toggle()
		return m, nil

	case key.Matches(msg, m.keys.NewProject):
		// Show project type selector if shed is available
		if m.shed != nil && m.shed.Available() {
			m.mode = ModeNewProjectType
			m.newProjectTypeIdx = 0
			return m, m.loadShedsCmd()
		}
		// Otherwise go straight to folder picker
		m.mode = ModeFolderPicker
		m.folderIdx = 0
		return m, nil

	case key.Matches(msg, m.keys.AddTerminal):
		proj := m.SelectedProject()
		if proj != nil {
			m.pendingProject = proj
			m.mode = ModeCommandPicker
			m.commandIdx = m.defaultCommandIndex()
		}
		return m, nil

	case key.Matches(msg, m.keys.Close):
		return m.handleClose()

	case key.Matches(msg, m.keys.CloseAll):
		return m.handleCloseAll()

	case key.Matches(msg, m.keys.Refresh):
		return m, tea.Batch(m.pollStatusCmd(), m.loadShedsCmd())

	case key.Matches(msg, m.keys.StartShed):
		proj := m.SelectedProject()
		if proj != nil && proj.Type == domain.ProjectTypeShed {
			return m, m.startShedCmd(proj.ShedName)
		}
		return m, nil

	case key.Matches(msg, m.keys.StopShed):
		proj := m.SelectedProject()
		if proj != nil && proj.Type == domain.ProjectTypeShed {
			return m, m.stopShedCmd(proj.ShedName)
		}
		return m, nil
	}

	return m, nil
}

// handleEnter processes Enter key in normal mode
func (m Model) handleEnter() (tea.Model, tea.Cmd) {
	item := m.tree.Selected()
	if item == nil {
		return m, nil
	}

	if item.Type == components.ItemTypeSession {
		// Check if session has a pane
		if item.Session.PaneID == 0 {
			return m, nil
		}

		// If session is already visible, just focus it
		if item.Session.IsVisible {
			return m, m.focusPaneCmd(item.Session.PaneID)
		}

		// Session is hidden - need to swap panes
		// Find the currently visible session globally
		var visibleSession *domain.Session
		var visibleProject *domain.Project
		for _, proj := range m.store.Projects() {
			for i := range proj.Sessions {
				sess := &proj.Sessions[i]
				if sess.ID != item.Session.ID && sess.PaneID > 0 && sess.IsVisible {
					visibleSession = sess
					visibleProject = proj
					break
				}
			}
			if visibleSession != nil {
				break
			}
		}

		if visibleSession != nil {
			// Swap panes: hide visible, show hidden
			return m, m.swapPanesCmd(item.Project, item.Session, visibleProject, visibleSession)
		}

		// No visible session to hide; just show this pane in the main window
		return m, m.showPaneCmd(item.Project, item.Session)
	}

	// Toggle project expand/collapse
	m.tree.Toggle()
	return m, nil
}

// handleClose processes close action
func (m Model) handleClose() (tea.Model, tea.Cmd) {
	item := m.tree.Selected()
	if item == nil {
		return m, nil
	}

	if item.Type == components.ItemTypeSession {
		// Confirm closing session
		m.confirmAction = ConfirmCloseSession
		m.confirmProject = item.Project
		m.confirmSession = item.Session
		m.mode = ModeConfirm
	} else {
		// If project has sessions, confirm first
		if len(item.Project.Sessions) > 0 {
			m.confirmAction = ConfirmCloseProject
			m.confirmProject = item.Project
			m.mode = ModeConfirm
		} else {
			// No sessions, just remove project
			_ = m.store.RemoveProject(item.Project.ID)
			_ = m.store.Save()
			m.tree.SetProjects(m.store.Projects())
		}
	}

	return m, nil
}

// handleCloseAll closes project and all its sessions
func (m Model) handleCloseAll() (tea.Model, tea.Cmd) {
	proj := m.SelectedProject()
	if proj == nil {
		return m, nil
	}

	// For shed projects, show close options dialog
	if proj.Type == domain.ProjectTypeShed {
		m.confirmProject = proj
		m.shedCloseOption = 0
		m.mode = ModeShedClose
		return m, nil
	}

	// For local projects, confirm if has sessions
	if len(proj.Sessions) > 0 {
		m.confirmAction = ConfirmCloseProject
		m.confirmProject = proj
		m.mode = ModeConfirm
	} else {
		_ = m.store.RemoveProject(proj.ID)
		_ = m.store.Save()
		m.tree.SetProjects(m.store.Projects())
	}

	return m, nil
}

// handleFolderPickerKey handles keys in folder picker mode
func (m Model) handleFolderPickerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.folderSearching {
		switch msg.Type {
		case tea.KeyEsc:
			m.folderSearching = false
			m.folderSearch.Blur()
			return m, nil
		case tea.KeyEnter:
			m.folderSearching = false
			m.folderSearch.Blur()
			return m, nil
		default:
			var cmd tea.Cmd
			m.folderSearch, cmd = m.folderSearch.Update(msg)
			return m, cmd
		}
	}

	switch {
	case key.Matches(msg, m.keys.Cancel):
		m.mode = ModeNormal
		return m, nil

	case key.Matches(msg, m.keys.Up):
		if m.folderIdx > 0 {
			m.folderIdx--
		}
		return m, nil

	case key.Matches(msg, m.keys.Down):
		if m.folderIdx < len(m.filteredFolders())-1 {
			m.folderIdx++
		}
		return m, nil

	case key.Matches(msg, m.keys.Enter):
		folders := m.filteredFolders()
		if m.folderIdx < len(folders) {
			return m, m.createProjectCmd(folders[m.folderIdx])
		}
		return m, nil

	case key.Matches(msg, m.keys.Search):
		m.folderSearching = true
		m.folderSearch.Focus()
		return m, nil
	}

	return m, nil
}

// handleCommandPickerKey handles keys in command picker mode
func (m Model) handleCommandPickerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Cancel):
		m.mode = ModeNormal
		m.pendingProject = nil
		return m, nil

	case key.Matches(msg, m.keys.Up):
		if m.commandIdx > 0 {
			m.commandIdx--
		}
		return m, nil

	case key.Matches(msg, m.keys.Down):
		if m.commandIdx < len(m.commands)-1 {
			m.commandIdx++
		}
		return m, nil

	case key.Matches(msg, m.keys.Enter):
		proj := m.pendingProject
		if proj == nil {
			m.mode = ModeNormal
			return m, nil
		}

		// Create session with selected command
		cmdID := m.commandKeys[m.commandIdx]
		cmd := m.config.Commands[cmdID].ToDomainCommand(cmdID)
		session := newSession(proj.ID, cmdID, cmd)

		// Add to store
		_ = m.store.AddSession(proj.ID, session)
		_ = m.store.Save()
		m.tree.SetProjects(m.store.Projects())

		// Select the new session
		proj.Expanded = true
		m.tree.Flatten()
		m.tree.SelectBySessionID(proj.ID, session.ID)

		m.pendingProject = nil
		return m, m.createPaneCmd(proj, session)
	}

	return m, nil
}

// handleShedPickerKey handles keys in shed picker mode
func (m Model) handleShedPickerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Cancel):
		m.mode = ModeNormal
		return m, nil

	case key.Matches(msg, m.keys.Up):
		if m.shedIdx > 0 {
			m.shedIdx--
		}
		return m, nil

	case key.Matches(msg, m.keys.Down):
		if m.shedIdx < len(m.sheds)-1 {
			m.shedIdx++
		}
		return m, nil

	case key.Matches(msg, m.keys.Enter):
		if m.shedIdx < len(m.sheds) {
			shed := m.sheds[m.shedIdx]
			if shed.Status == "stopped" {
				// Start shed first
				return m, m.startShedCmd(shed.Name)
			}
			return m, m.createShedProjectCmd(shed)
		}
		return m, nil

	case key.Matches(msg, m.keys.StartShed):
		if m.shedIdx < len(m.sheds) {
			return m, m.startShedCmd(m.sheds[m.shedIdx].Name)
		}
		return m, nil
	}

	return m, nil
}

// handleShedCreateKey handles keys in shed create mode
func (m Model) handleShedCreateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.mode = ModeNormal
		return m, nil

	case tea.KeyTab, tea.KeyDown:
		m.shedCreateFocus = (m.shedCreateFocus + 1) % 3
		m.updateShedCreateFocus()
		return m, nil

	case tea.KeyShiftTab, tea.KeyUp:
		m.shedCreateFocus = (m.shedCreateFocus + 2) % 3
		m.updateShedCreateFocus()
		return m, nil

	case tea.KeyEnter:
		if m.shedCreateFocus == 2 {
			// Submit
			name := m.shedCreateName.Value()
			if name == "" {
				return m, nil
			}

			// Get server (would need server list - for now use default)
			server := m.config.Shed.DefaultServer

			return m, m.createShedCmd(name, shed.CreateOpts{
				Repo:   m.shedCreateRepo.Value(),
				Server: server,
			})
		}
		// Move to next field
		m.shedCreateFocus = (m.shedCreateFocus + 1) % 3
		m.updateShedCreateFocus()
		return m, nil
	}

	// Update the focused input
	var cmd tea.Cmd
	switch m.shedCreateFocus {
	case 0:
		m.shedCreateName, cmd = m.shedCreateName.Update(msg)
	case 1:
		m.shedCreateRepo, cmd = m.shedCreateRepo.Update(msg)
	}

	return m, cmd
}

// handleShedCloseKey handles keys in shed close dialog
func (m Model) handleShedCloseKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Cancel):
		m.mode = ModeNormal
		return m, nil

	case key.Matches(msg, m.keys.Up):
		if m.shedCloseOption > 0 {
			m.shedCloseOption--
		}
		return m, nil

	case key.Matches(msg, m.keys.Down):
		if m.shedCloseOption < 2 {
			m.shedCloseOption++
		}
		return m, nil

	case key.Matches(msg, m.keys.Enter):
		proj := m.confirmProject
		if proj == nil {
			m.mode = ModeNormal
			return m, nil
		}

		var cmds []tea.Cmd

		// Kill all sessions
		for i := range proj.Sessions {
			if proj.Sessions[i].PaneID > 0 {
				cmds = append(cmds, m.killPaneCmd(proj, &proj.Sessions[i]))
			}
		}

		switch m.shedCloseOption {
		case 0:
			// Close project only
			_ = m.store.RemoveProject(proj.ID)
		case 1:
			// Close and stop shed
			_ = m.store.RemoveProject(proj.ID)
			cmds = append(cmds, m.stopShedCmd(proj.ShedName))
		case 2:
			// Close and delete shed - need confirmation
			m.confirmAction = ConfirmDeleteShed
			m.mode = ModeConfirm
			return m, nil
		}

		_ = m.store.Save()
		m.tree.SetProjects(m.store.Projects())
		m.mode = ModeNormal
		m.confirmProject = nil

		return m, tea.Batch(cmds...)
	}

	return m, nil
}

// handleConfirmKey handles keys in confirm dialog
func (m Model) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Confirm):
		return m.executeConfirmedAction()

	case key.Matches(msg, m.keys.Cancel):
		m.mode = ModeNormal
		m.confirmAction = ConfirmNone
		m.confirmProject = nil
		m.confirmSession = nil
		return m, nil
	}

	return m, nil
}

// handleHelpKey handles keys in help mode
func (m Model) handleHelpKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Help), key.Matches(msg, m.keys.Cancel), key.Matches(msg, m.keys.Quit):
		m.mode = ModeNormal
		m.showHelp = false
		return m, nil
	}
	return m, nil
}

// handleNewProjectTypeKey handles keys in new project type selector
func (m Model) handleNewProjectTypeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Cancel):
		m.mode = ModeNormal
		return m, nil

	case key.Matches(msg, m.keys.Up):
		if m.newProjectTypeIdx > 0 {
			m.newProjectTypeIdx--
		}
		return m, nil

	case key.Matches(msg, m.keys.Down):
		if m.newProjectTypeIdx < 2 {
			m.newProjectTypeIdx++
		}
		return m, nil

	case key.Matches(msg, m.keys.Enter):
		switch m.newProjectTypeIdx {
		case 0:
			m.mode = ModeFolderPicker
			m.folderIdx = 0
		case 1:
			m.mode = ModeShedPicker
			m.shedIdx = 0
		case 2:
			m.mode = ModeShedCreate
			m.shedCreateFocus = 0
			m.shedCreateName.SetValue("")
			m.shedCreateRepo.SetValue("")
			m.updateShedCreateFocus()
		}
		return m, nil
	}

	return m, nil
}

// executeConfirmedAction executes the confirmed action
func (m Model) executeConfirmedAction() (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch m.confirmAction {
	case ConfirmCloseSession:
		if m.confirmSession != nil && m.confirmProject != nil {
			cmds = append(cmds, m.killPaneCmd(m.confirmProject, m.confirmSession))
			_ = m.store.RemoveSession(m.confirmProject.ID, m.confirmSession.ID)
			_ = m.store.Save()
			m.tree.SetProjects(m.store.Projects())
		}

	case ConfirmCloseProject:
		if m.confirmProject != nil {
			// Kill all sessions
			for i := range m.confirmProject.Sessions {
				if m.confirmProject.Sessions[i].PaneID > 0 {
					cmds = append(cmds, m.killPaneCmd(m.confirmProject, &m.confirmProject.Sessions[i]))
				}
			}
			_ = m.store.RemoveProject(m.confirmProject.ID)
			_ = m.store.Save()
			m.tree.SetProjects(m.store.Projects())
		}

	case ConfirmDeleteShed:
		if m.confirmProject != nil {
			// Kill all sessions
			for i := range m.confirmProject.Sessions {
				if m.confirmProject.Sessions[i].PaneID > 0 {
					cmds = append(cmds, m.killPaneCmd(m.confirmProject, &m.confirmProject.Sessions[i]))
				}
			}
			_ = m.store.RemoveProject(m.confirmProject.ID)
			_ = m.store.Save()
			m.tree.SetProjects(m.store.Projects())
			cmds = append(cmds, m.deleteShedCmd(m.confirmProject.ShedName, true))
		}
	}

	m.mode = ModeNormal
	m.confirmAction = ConfirmNone
	m.confirmProject = nil
	m.confirmSession = nil

	return m, tea.Batch(cmds...)
}

// Helper methods

func (m *Model) applyStatusUpdates(updates map[string]domain.Status) {
	for _, proj := range m.store.Projects() {
		for i := range proj.Sessions {
			if status, ok := updates[proj.Sessions[i].ID]; ok {
				proj.Sessions[i].Status = status
			}
		}
	}
}

func (m *Model) handleProjectCreated(proj *domain.Project) {
	_ = m.store.AddProject(proj)
	_ = m.store.Save()
	m.tree.SetProjects(m.store.Projects())
	m.tree.SelectByProjectID(proj.ID)
	m.pendingProject = proj
}

func (m *Model) handlePaneCreated(msg PaneCreatedMsg) {
	proj, err := m.store.GetProject(msg.ProjectID)
	if err != nil {
		return
	}

	// Clear visibility for all sessions (single visible pane)
	for _, p := range m.store.Projects() {
		for i := range p.Sessions {
			p.Sessions[i].IsVisible = false
		}
	}

	// Update the hidden session's pane ID and visibility (if any was hidden)
	if msg.HiddenSessionID != "" {
		if msg.HiddenProjectID != "" {
			if hiddenProj, err := m.store.GetProject(msg.HiddenProjectID); err == nil {
				for i := range hiddenProj.Sessions {
					if hiddenProj.Sessions[i].ID == msg.HiddenSessionID {
						hiddenProj.Sessions[i].PaneID = msg.HiddenPaneID
						hiddenProj.Sessions[i].IsVisible = false
						break
					}
				}
			}
		} else {
			for _, p := range m.store.Projects() {
				for i := range p.Sessions {
					if p.Sessions[i].ID == msg.HiddenSessionID {
						p.Sessions[i].PaneID = msg.HiddenPaneID
						p.Sessions[i].IsVisible = false
						break
					}
				}
			}
		}
	}

	// Update the new session's pane ID and mark it as visible
	for i := range proj.Sessions {
		if proj.Sessions[i].ID == msg.SessionID {
			proj.Sessions[i].PaneID = msg.PaneID
			proj.Sessions[i].IsVisible = true
			break
		}
	}
	_ = m.store.Save()
	m.tree.SetProjects(m.store.Projects())
}

func (m *Model) handlePaneKilled(msg PaneKilledMsg) {
	// Session already removed from store in the confirm handler
}

func (m *Model) handlePaneSwapped(msg PaneSwappedMsg) {
	// Clear visibility for all sessions (single visible pane)
	for _, p := range m.store.Projects() {
		for i := range p.Sessions {
			p.Sessions[i].IsVisible = false
		}
	}

	if msg.ShownProjectID != "" {
		if shownProj, err := m.store.GetProject(msg.ShownProjectID); err == nil {
			for i := range shownProj.Sessions {
				if shownProj.Sessions[i].ID == msg.ShownSessionID {
					shownProj.Sessions[i].PaneID = msg.ShownPaneID
					shownProj.Sessions[i].IsVisible = true
					break
				}
			}
		}
	}

	if msg.HiddenSessionID != "" && msg.HiddenProjectID != "" {
		if hiddenProj, err := m.store.GetProject(msg.HiddenProjectID); err == nil {
			for i := range hiddenProj.Sessions {
				if hiddenProj.Sessions[i].ID == msg.HiddenSessionID {
					hiddenProj.Sessions[i].PaneID = msg.HiddenPaneID
					hiddenProj.Sessions[i].IsVisible = false
					break
				}
			}
		}
	}
	m.tree.SetProjects(m.store.Projects())
}

func (m *Model) handleVisibilitySynced(msg VisibilitySyncedMsg) {
	for _, p := range m.store.Projects() {
		for i := range p.Sessions {
			p.Sessions[i].IsVisible = p.Sessions[i].ID == msg.VisibleSessionID
		}
	}
	m.tree.SetProjects(m.store.Projects())
}

func (m *Model) filteredFolders() []string {
	if !m.folderSearching || m.folderSearch.Value() == "" {
		return m.folders
	}

	search := m.folderSearch.Value()
	var filtered []string
	for _, f := range m.folders {
		if containsIgnoreCase(f, search) {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

func (m *Model) defaultCommandIndex() int {
	for i, id := range m.commandKeys {
		if id == m.config.DefaultCommand {
			return i
		}
	}
	return 0
}

func (m *Model) updateShedCreateFocus() {
	m.shedCreateName.Blur()
	m.shedCreateRepo.Blur()

	switch m.shedCreateFocus {
	case 0:
		m.shedCreateName.Focus()
	case 1:
		m.shedCreateRepo.Focus()
	}
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(substr) == 0 ||
			findIgnoreCase(s, substr) >= 0)
}

func findIgnoreCase(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(s) < len(substr) {
		return -1
	}

	// Simple case-insensitive search
	lowerS := toLower(s)
	lowerSubstr := toLower(substr)

	for i := 0; i <= len(lowerS)-len(lowerSubstr); i++ {
		if lowerS[i:i+len(lowerSubstr)] == lowerSubstr {
			return i
		}
	}
	return -1
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}
