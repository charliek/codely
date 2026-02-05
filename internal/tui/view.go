package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/charliek/codely/internal/domain"
	"github.com/charliek/codely/internal/tui/components"
)

// View renders the current state
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	switch m.mode {
	case ModeHelp:
		return m.helpView()
	case ModeFolderPicker:
		return m.folderPickerView()
	case ModeCommandPicker:
		return m.commandPickerView()
	case ModeShedPicker:
		return m.shedPickerView()
	case ModeShedCreate:
		return m.shedCreateView()
	case ModeShedClose:
		return m.shedCloseView()
	case ModeConfirm:
		return m.confirmView()
	case ModeNewProjectType:
		return m.newProjectTypeView()
	default:
		return m.normalView()
	}
}

// normalView renders the main project tree view
func (m Model) normalView() string {
	var b strings.Builder

	// Header
	header := styleHeader.Width(m.width).Render(fmt.Sprintf("Codely%s", m.versionString()))
	b.WriteString(header)
	b.WriteString("\n")

	// Content area
	content := m.treeView()
	b.WriteString(content)

	// Error banner
	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(styleError.Render(fmt.Sprintf("⚠️  %s", m.err.Error())))
	}

	body := b.String()
	footer := styleFooter.Width(m.width).Render(m.helpLine())

	// Pad body to clear old lines and keep footer at bottom.
	if m.height > 0 {
		bodyHeight := lipgloss.Height(body)
		footerHeight := lipgloss.Height(footer)
		remaining := m.height - bodyHeight - footerHeight
		if remaining > 0 {
			body += strings.Repeat("\n", remaining)
		}
	}

	return body + "\n" + footer
}

// treeView renders the project tree
func (m Model) treeView() string {
	if m.tree.IsEmpty() {
		return styleProjectPath.Render("\n  No projects yet. Press 'n' to create one.\n")
	}

	var b strings.Builder

	// Separate local and shed projects
	var localProjects, shedProjects []*domain.Project
	for _, proj := range m.tree.Projects() {
		if proj.Type == domain.ProjectTypeShed {
			shedProjects = append(shedProjects, proj)
		} else {
			localProjects = append(localProjects, proj)
		}
	}

	// Render LOCAL section
	if len(localProjects) > 0 {
		b.WriteString(styleSectionHeader.Render("LOCAL"))
		b.WriteString("\n")
	}

	// Render items
	selectedIdx := m.tree.SelectedIndex()
	for i, item := range m.tree.Items() {
		isSelected := i == selectedIdx
		line := m.renderItem(item, isSelected)
		b.WriteString(line)
		b.WriteString("\n")

		// Add SHEDS header before first shed project
		if item.Type == components.ItemTypeProject &&
			item.Project.Type == domain.ProjectTypeLocal &&
			len(shedProjects) > 0 {
			// Check if next item is a shed
			if i+1 < len(m.tree.Items()) {
				next := m.tree.Items()[i+1]
				if next.Type == components.ItemTypeProject && next.Project.Type == domain.ProjectTypeShed {
					b.WriteString("\n")
					b.WriteString(styleSectionHeader.Render("SHEDS"))
					b.WriteString("\n")
				}
			}
		}
	}

	// If we only have shed projects, add header at the start
	if len(localProjects) == 0 && len(shedProjects) > 0 {
		var sb strings.Builder
		sb.WriteString(styleSectionHeader.Render("SHEDS"))
		sb.WriteString("\n")
		sb.WriteString(b.String())
		return sb.String()
	}

	return b.String()
}

// renderItem renders a single tree item
func (m Model) renderItem(item components.TreeItem, selected bool) string {
	var line string

	if item.Type == components.ItemTypeProject {
		line = m.renderProject(item.Project)
	} else {
		line = m.renderSession(item.Session)
	}

	if selected {
		return styleSelected.Width(m.width - 2).Render(line)
	}
	return line
}

// renderProject renders a project row
func (m Model) renderProject(proj *domain.Project) string {
	// Expand/collapse indicator
	indicator := "▶"
	if proj.Expanded {
		indicator = "▼"
	}

	// Session count for collapsed
	countStr := ""
	if !proj.Expanded && len(proj.Sessions) > 0 {
		countStr = fmt.Sprintf(" (%d sessions)", len(proj.Sessions))
	}

	// Stopped indicator for shed projects
	stoppedStr := ""
	if proj.Type == domain.ProjectTypeShed {
		// Check if shed is stopped
		for _, s := range m.sheds {
			if s.Name == proj.ShedName && s.Status == "stopped" {
				stoppedStr = " ⏸️ stopped"
				break
			}
		}
	}

	name := styleProjectName.Render(proj.Name)
	line := fmt.Sprintf("%s %s%s%s", indicator, name, countStr, stoppedStr)

	// Add path on next line if expanded and showing directory
	if proj.Expanded && m.config.UI.ShowDirectory {
		path := proj.DisplayPath()
		// Shorten home directory
		home := homeDir()
		if strings.HasPrefix(path, home) {
			path = "~" + path[len(home):]
		}
		line = fmt.Sprintf("%s\n    %s", line, styleProjectPath.Render(path))
	}

	return line
}

// renderSession renders a session row
func (m Model) renderSession(sess *domain.Session) string {
	// Focus indicator
	focusIndicator := "○"
	if sess.IsVisible {
		focusIndicator = "●"
	}

	// Status with icon
	statusLabel := string(sess.Status)
	if sess.Status == domain.StatusError && sess.ExitCode != nil {
		statusLabel = fmt.Sprintf("%s (%d)", statusLabel, *sess.ExitCode)
	}
	statusStr := sess.Status.Icon() + " " + statusLabel
	statusStyled := m.styleStatus(sess.Status).Render(statusStr)

	name := sess.Command.DisplayName
	if name == "" {
		name = sess.Command.ID
	}

	return fmt.Sprintf("    %s %s%s%s",
		focusIndicator,
		styleSessionName.Render(name),
		strings.Repeat(" ", 14-len(name)), // Padding for alignment
		statusStyled)
}

// styleStatus returns the appropriate style for a status
func (m Model) styleStatus(status domain.Status) lipgloss.Style {
	switch status {
	case domain.StatusIdle:
		return styleStatusIdle
	case domain.StatusThinking:
		return styleStatusThinking
	case domain.StatusExecuting:
		return styleStatusExecuting
	case domain.StatusError:
		return styleStatusError
	case domain.StatusExited:
		return styleStatusExited
	case domain.StatusStopped:
		return styleStatusStopped
	default:
		return styleStatusIdle
	}
}

// folderPickerView renders the folder picker
func (m Model) folderPickerView() string {
	var b strings.Builder

	b.WriteString(styleDialogTitle.Render("New Local Project"))
	b.WriteString("\n\n")
	b.WriteString("Select directory:\n\n")

	folders := m.filteredFolders()

	// Group by parent directory
	groups := make(map[string][]string)
	var groupOrder []string

	for _, f := range folders {
		parent := filepath.Dir(f)
		if _, ok := groups[parent]; !ok {
			groupOrder = append(groupOrder, parent)
		}
		groups[parent] = append(groups[parent], f)
	}

	idx := 0
	for _, parent := range groupOrder {
		// Shorten home directory
		displayParent := parent
		home := homeDir()
		if strings.HasPrefix(displayParent, home) {
			displayParent = "~" + displayParent[len(home):]
		}

		b.WriteString(styleProjectPath.Render(displayParent + "/"))
		b.WriteString("\n")

		for _, f := range groups[parent] {
			prefix := "  ○ "
			name := filepath.Base(f) + "/"

			if idx == m.folderIdx {
				b.WriteString(styleDialogOptionSelected.Render(prefix + name))
			} else {
				b.WriteString(styleDialogOption.Render(prefix + name))
			}
			b.WriteString("\n")
			idx++
		}
		b.WriteString("\n")
	}

	// Search box
	if m.folderSearching {
		b.WriteString("\n")
		b.WriteString(m.folderSearch.View())
	}

	b.WriteString("\n")
	b.WriteString(styleHelp.Render("[/] search  [enter] select  [esc] clear/back"))

	return styleDialog.Render(b.String())
}

// commandPickerView renders the command picker
func (m Model) commandPickerView() string {
	var b strings.Builder

	projName := "Unknown"
	projPath := ""
	if m.pendingProject != nil {
		projName = m.pendingProject.Name
		projPath = m.pendingProject.DisplayPath()
	}

	b.WriteString(styleDialogTitle.Render("Add Terminal"))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("Project: %s\n", styleProjectName.Render(projName)))
	b.WriteString(fmt.Sprintf("Path: %s\n\n", styleProjectPath.Render(projPath)))
	b.WriteString("Select command:\n\n")

	for i, id := range m.commandKeys {
		cmd := m.config.Commands[id]
		prefix := "○"
		if id == m.config.DefaultCommand {
			prefix = "●"
		}

		name := cmd.DisplayName
		if name == "" {
			name = id
		}

		// Build exec line
		execLine := cmd.Exec
		if len(cmd.Args) > 0 {
			execLine += " " + strings.Join(cmd.Args, " ")
		}

		var line string
		if i == m.commandIdx {
			line = styleDialogOptionSelected.Render(fmt.Sprintf("%s %s", prefix, name))
		} else {
			line = styleDialogOption.Render(fmt.Sprintf("%s %s", prefix, name))
		}
		b.WriteString(line)
		b.WriteString("\n")
		b.WriteString(styleProjectPath.Render(fmt.Sprintf("    %s", execLine)))
		b.WriteString("\n\n")
	}

	b.WriteString(styleHelp.Render("[enter] launch  [esc] back"))

	return styleDialog.Render(b.String())
}

// shedPickerView renders the shed picker
func (m Model) shedPickerView() string {
	var b strings.Builder

	b.WriteString(styleDialogTitle.Render("Attach to Shed"))
	b.WriteString("\n\n")

	if len(m.sheds) == 0 {
		b.WriteString(styleProjectPath.Render("No sheds available."))
		b.WriteString("\n\n")
	} else {
		b.WriteString("Available Sheds:\n\n")

		// Group by server
		groups := make(map[string][]int) // server -> indices
		var serverOrder []string

		for i, s := range m.sheds {
			if _, ok := groups[s.Server]; !ok {
				serverOrder = append(serverOrder, s.Server)
			}
			groups[s.Server] = append(groups[s.Server], i)
		}

		for _, server := range serverOrder {
			b.WriteString(styleProjectName.Render(server))
			b.WriteString("\n")

			for _, idx := range groups[server] {
				s := m.sheds[idx]
				prefix := "○"
				statusStr := s.Status

				var line string
				if idx == m.shedIdx {
					line = styleDialogOptionSelected.Render(fmt.Sprintf("%s %s", prefix, s.Name))
				} else {
					line = styleDialogOption.Render(fmt.Sprintf("%s %s", prefix, s.Name))
				}

				// Add status
				statusStyled := styleStatusIdle.Render(statusStr)
				if s.Status == "stopped" {
					statusStyled = styleStatusStopped.Render(statusStr)
				}

				b.WriteString(fmt.Sprintf("%s    %s\n", line, statusStyled))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString(styleHelp.Render("[enter] select  [s] start  [esc] back"))

	return styleDialog.Render(b.String())
}

// shedCreateView renders the shed creation form
func (m Model) shedCreateView() string {
	var b strings.Builder

	b.WriteString(styleDialogTitle.Render("Create New Shed"))
	b.WriteString("\n\n")

	// Name field
	nameLabel := "Shed name: "
	if m.shedCreateFocus == 0 {
		nameLabel = styleDialogOptionSelected.Render(nameLabel)
	}
	b.WriteString(nameLabel)
	b.WriteString(m.shedCreateName.View())
	b.WriteString("\n\n")

	// Repo field
	repoLabel := "Repository (optional): "
	if m.shedCreateFocus == 1 {
		repoLabel = styleDialogOptionSelected.Render(repoLabel)
	}
	b.WriteString(repoLabel)
	b.WriteString(m.shedCreateRepo.View())
	b.WriteString("\n\n")

	// Server selection (simplified - just show default)
	serverLabel := "Server: "
	if m.shedCreateFocus == 2 {
		serverLabel = styleDialogOptionSelected.Render(serverLabel)
	}
	server := m.config.Shed.DefaultServer
	if server == "" {
		server = "(default)"
	}
	b.WriteString(serverLabel)
	b.WriteString(server)
	b.WriteString("\n\n")

	b.WriteString(styleHelp.Render("[tab] next field  [enter] create  [esc] cancel"))

	return styleDialog.Render(b.String())
}

// shedCloseView renders the shed close options dialog
func (m Model) shedCloseView() string {
	var b strings.Builder

	projName := "Unknown"
	if m.confirmProject != nil {
		projName = m.confirmProject.Name
	}

	b.WriteString(styleDialogTitle.Render("Close Shed Project"))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("Project: %s\n\n", styleProjectName.Render(projName)))
	b.WriteString("What would you like to do?\n\n")

	options := []struct {
		label string
		desc  string
	}{
		{"Close project only", "Shed keeps running on server.\nYou can re-attach later."},
		{"Close and stop shed", "Stops the container but keeps data.\nCan be restarted later."},
		{"Close and DELETE shed", "Permanently removes container and\nall data. Cannot be undone."},
	}

	for i, opt := range options {
		prefix := "○"
		if i == m.shedCloseOption {
			prefix = "●"
			b.WriteString(styleDialogOptionSelected.Render(fmt.Sprintf("%s %s", prefix, opt.label)))
		} else {
			b.WriteString(styleDialogOption.Render(fmt.Sprintf("%s %s", prefix, opt.label)))
		}
		b.WriteString("\n")
		b.WriteString(styleProjectPath.Render("    " + strings.ReplaceAll(opt.desc, "\n", "\n    ")))
		b.WriteString("\n\n")
	}

	b.WriteString(styleHelp.Render("[enter] confirm  [esc] cancel"))

	return styleDialog.Render(b.String())
}

// confirmView renders the confirmation dialog
func (m Model) confirmView() string {
	var b strings.Builder

	b.WriteString(styleDialogTitle.Render("Confirm"))
	b.WriteString("\n\n")

	switch m.confirmAction {
	case ConfirmCloseSession:
		sessionName := "unknown"
		projName := "unknown"
		if m.confirmSession != nil {
			sessionName = m.confirmSession.Command.DisplayName
			if sessionName == "" {
				sessionName = m.confirmSession.Command.ID
			}
		}
		if m.confirmProject != nil {
			projName = m.confirmProject.Name
		}
		b.WriteString(fmt.Sprintf("Close %s in %s?", sessionName, projName))

	case ConfirmCloseProject:
		projName := "unknown"
		sessionCount := 0
		if m.confirmProject != nil {
			projName = m.confirmProject.Name
			sessionCount = len(m.confirmProject.Sessions)
		}
		b.WriteString(fmt.Sprintf("Close project %s and all %d sessions?", projName, sessionCount))

	case ConfirmDeleteShed:
		shedName := "unknown"
		if m.confirmProject != nil {
			shedName = m.confirmProject.ShedName
		}
		b.WriteString(styleError.Render(fmt.Sprintf("⚠️  This will permanently delete shed '%s'.\nAll data will be lost. Continue?", shedName)))
	}

	b.WriteString("\n\n")
	b.WriteString(styleHelp.Render("[y] yes  [n/esc] cancel"))

	return styleDialog.Render(b.String())
}

// newProjectTypeView renders the project type selector
func (m Model) newProjectTypeView() string {
	var b strings.Builder

	b.WriteString(styleDialogTitle.Render("New Project"))
	b.WriteString("\n\n")
	b.WriteString("Select project type:\n\n")

	options := []struct {
		label string
		desc  string
	}{
		{"Local Directory", "Create project from a local folder"},
		{"Attach to Shed", "Connect to an existing remote shed"},
		{"Create New Shed", "Create a new remote development container"},
	}

	for i, opt := range options {
		prefix := "○"
		if i == m.newProjectTypeIdx {
			prefix = "●"
			b.WriteString(styleDialogOptionSelected.Render(fmt.Sprintf("%s %s", prefix, opt.label)))
		} else {
			b.WriteString(styleDialogOption.Render(fmt.Sprintf("%s %s", prefix, opt.label)))
		}
		b.WriteString("\n")
		b.WriteString(styleProjectPath.Render("    " + opt.desc))
		b.WriteString("\n\n")
	}

	b.WriteString(styleHelp.Render("[enter] select  [esc] cancel"))

	return styleDialog.Render(b.String())
}

// helpView renders the help screen
func (m Model) helpView() string {
	return m.help.View(m.keys)
}

// helpLine returns the footer help text
func (m Model) helpLine() string {
	return "[n]ew project [t]erminal [x]close [?]help [q]uit | tmux: prefix+z zoom"
}

// versionString returns the version string for the header
func (m Model) versionString() string {
	return "  v0.1"
}

// homeDir returns the user's home directory
func homeDir() string {
	home, _ := os.UserHomeDir()
	return home
}
