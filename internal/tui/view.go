package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/charliek/codely/internal/domain"
	"github.com/charliek/codely/internal/pathutil"
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
	case ModeRename:
		return m.renameView()
	case ModeShedPicker:
		return m.shedPickerView()
	case ModeShedCreate:
		return m.shedCreateView()
	case ModeShedCreating:
		return m.shedCreatingView()
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
	content := m.skin.View(&m)
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

// styleStatus returns the appropriate style for a status (package-level for use by all skins)
func styleStatus(status domain.Status) lipgloss.Style {
	switch status {
	case domain.StatusIdle:
		return styleStatusIdle
	case domain.StatusWaiting:
		return styleStatusWaiting
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
		displayParent := pathutil.ContractHome(parent)

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
	fmt.Fprintf(&b, "Project: %s\n", styleProjectName.Render(projName))
	fmt.Fprintf(&b, "Path: %s\n\n", styleProjectPath.Render(projPath))
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

// renameView renders the session rename dialog.
func (m Model) renameView() string {
	var b strings.Builder

	proj, sess := m.renameTarget()
	sessionName := "Unknown"
	projectName := "Unknown"
	projectPath := ""
	if sess != nil {
		sessionName = sess.Command.Name()
	}
	if proj != nil {
		projectName = proj.Name
		projectPath = proj.DisplayPath()
	}

	b.WriteString(styleDialogTitle.Render("Rename Session"))
	b.WriteString("\n\n")
	fmt.Fprintf(&b, "Session: %s\n", styleSessionName.Render(sessionName))
	fmt.Fprintf(&b, "Project: %s\n", styleProjectName.Render(projectName))
	if projectPath != "" {
		fmt.Fprintf(&b, "Path: %s\n", styleProjectPath.Render(projectPath))
	}
	b.WriteString("\n")
	b.WriteString("New name:\n")
	b.WriteString(m.renameInput.View())
	b.WriteString("\n\n")
	b.WriteString(styleHelp.Render("[enter] save  [esc] cancel  [blank] reset default name"))

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

				fmt.Fprintf(&b, "%s    %s\n", line, statusStyled)
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

	// Backend selector
	backendLabel := "Backend: "
	if m.shedCreateFocus == 2 {
		backendLabel = styleDialogOptionSelected.Render(backendLabel)
	}
	backendOptions := []string{"(server default)", "docker", "firecracker"}
	b.WriteString(backendLabel)
	fmt.Fprintf(&b, "< %s >", backendOptions[m.shedCreateBackend])
	b.WriteString("\n\n")

	// Server field — dynamic based on loaded servers
	serverIdx := m.shedCreateServerIdx()
	serverLabel := "Server: "
	if serverIdx >= 0 && m.shedCreateFocus == serverIdx {
		serverLabel = styleDialogOptionSelected.Render(serverLabel)
	}
	b.WriteString(serverLabel)
	switch {
	case len(m.shedCreateServers) > 1:
		// Focusable cycle selector
		name := m.shedCreateServers[m.shedCreateServer].Name
		fmt.Fprintf(&b, "< %s >", name)
	case len(m.shedCreateServers) == 1:
		// Single server, display as static text
		b.WriteString(m.shedCreateServers[0].Name)
	default:
		// Still loading or no servers
		server := m.config.Shed.DefaultServer
		if server == "" {
			server = "(default)"
		}
		b.WriteString(server)
	}
	b.WriteString("\n\n")

	// Submit button
	submitIdx := m.shedCreateSubmitIdx()
	createBtn := "[ Create ]"
	if m.shedCreateFocus == submitIdx {
		createBtn = styleDialogOptionSelected.Render(createBtn)
	}
	b.WriteString(createBtn)
	b.WriteString("\n\n")

	b.WriteString(styleHelp.Render("[tab] next field  [enter] create  [esc] cancel"))

	return styleDialog.Render(b.String())
}

// shedCreatingView renders the creating-in-progress screen
func (m Model) shedCreatingView() string {
	var b strings.Builder

	b.WriteString(styleDialogTitle.Render("Create New Shed"))
	b.WriteString("\n\n")

	fmt.Fprintf(&b, "Creating shed '%s'...\n\n", m.shedCreatingName)

	if m.shedCreatingCmd != "" {
		b.WriteString(styleProjectPath.Render("$ " + m.shedCreatingCmd))
		b.WriteString("\n\n")
	}

	// Show last ~10 lines of output
	lines := m.shedCreateOutput
	if len(lines) > 10 {
		lines = lines[len(lines)-10:]
	}
	for _, line := range lines {
		b.WriteString(styleProjectPath.Render("> " + line))
		b.WriteString("\n")
	}
	if len(lines) > 0 {
		b.WriteString("\n")
	}

	b.WriteString(styleHelp.Render("Please wait."))

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
	fmt.Fprintf(&b, "Project: %s\n\n", styleProjectName.Render(projName))
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
			sessionName = m.confirmSession.Command.Name()
		}
		if m.confirmProject != nil {
			projName = m.confirmProject.Name
		}
		fmt.Fprintf(&b, "Close %s in %s?", sessionName, projName)

	case ConfirmCloseProject:
		projName := "unknown"
		sessionCount := 0
		if m.confirmProject != nil {
			projName = m.confirmProject.Name
			sessionCount = len(m.confirmProject.Sessions)
		}
		fmt.Fprintf(&b, "Close project %s and all %d sessions?", projName, sessionCount)

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
	return "[n]ew project [t]erminal [r]ename [x]close [q]uit"
}

// versionString returns the version string for the header
func (m Model) versionString() string {
	return "  v0.1"
}
