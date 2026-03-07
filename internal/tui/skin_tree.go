package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/charliek/codely/internal/config"
	"github.com/charliek/codely/internal/domain"
	"github.com/charliek/codely/internal/tui/components"
)

// TreeSkin renders the hierarchical tree view of projects and sessions.
type TreeSkin struct {
	tree   *components.Tree
	config *config.Config
	keys   KeyMap
}

// NewTreeSkin creates a tree skin wrapping the existing Tree component.
func NewTreeSkin(projects []*domain.Project, cfg *config.Config, keys KeyMap) *TreeSkin {
	return &TreeSkin{
		tree:   components.NewTree(projects),
		config: cfg,
		keys:   keys,
	}
}

func (s *TreeSkin) View(m *Model) string {
	if s.tree.IsEmpty() {
		return styleProjectPath.Render("\n  No projects yet. Press 'n' to create one.\n")
	}

	var b strings.Builder

	// Separate local and shed projects
	var localProjects, shedProjects []*domain.Project
	for _, proj := range s.tree.Projects() {
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
	selectedIdx := s.tree.SelectedIndex()
	for i, item := range s.tree.Items() {
		isSelected := i == selectedIdx
		line := s.renderItem(m, item, isSelected)
		b.WriteString(line)
		b.WriteString("\n")

		// Add SHEDS header before first shed project
		if item.Type == components.ItemTypeProject &&
			item.Project.Type == domain.ProjectTypeLocal &&
			len(shedProjects) > 0 {
			if i+1 < len(s.tree.Items()) {
				next := s.tree.Items()[i+1]
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

func (s *TreeSkin) renderItem(m *Model, item components.TreeItem, selected bool) string {
	var line string

	if item.Type == components.ItemTypeProject {
		line = s.renderProject(m, item.Project)
	} else {
		line = s.renderSession(m, item.Session)
	}

	if selected {
		return styleSelected.Width(m.width - 2).Render(line)
	}
	return line
}

func (s *TreeSkin) renderProject(m *Model, proj *domain.Project) string {
	indicator := "▶"
	if proj.Expanded {
		indicator = "▼"
	}

	countStr := ""
	if !proj.Expanded && len(proj.Sessions) > 0 {
		countStr = fmt.Sprintf(" (%d sessions)", len(proj.Sessions))
	}

	stoppedStr := ""
	if proj.Type == domain.ProjectTypeShed {
		for _, s := range m.sheds {
			if s.Name == proj.ShedName && s.Status == "stopped" {
				stoppedStr = " ⏸️ stopped"
				break
			}
		}
	}

	name := styleProjectName.Render(proj.Name)
	line := fmt.Sprintf("%s %s%s%s", indicator, name, countStr, stoppedStr)

	if proj.Expanded && s.config.UI.ShowDirectory {
		path := proj.DisplayPath()
		home := homeDir()
		if strings.HasPrefix(path, home) {
			path = "~" + path[len(home):]
		}
		line = fmt.Sprintf("%s\n    %s", line, styleProjectPath.Render(path))
	}

	return line
}

func (s *TreeSkin) renderSession(m *Model, sess *domain.Session) string {
	focusIndicator := "○"
	if sess.IsVisible {
		focusIndicator = "●"
	}

	statusStr := sess.Status.Icon()
	if sess.Status == domain.StatusError && sess.ExitCode != nil {
		statusStr = fmt.Sprintf("%s (%d)", statusStr, *sess.ExitCode)
	}
	statusStyled := styleStatus(sess.Status).Render(statusStr)

	name := sess.Command.DisplayName
	if name == "" {
		name = sess.Command.ID
	}

	return fmt.Sprintf("    %s %s%s%s",
		focusIndicator,
		styleSessionName.Render(name),
		strings.Repeat(" ", 14-len(name)),
		statusStyled)
}

func (s *TreeSkin) HandleKey(m *Model, msg tea.KeyMsg) (bool, tea.Cmd) {
	switch {
	case key.Matches(msg, s.keys.Up):
		s.tree.MoveUp()
		return true, nil
	case key.Matches(msg, s.keys.Down):
		s.tree.MoveDown()
		return true, nil
	case key.Matches(msg, s.keys.Left):
		s.tree.Collapse()
		return true, nil
	case key.Matches(msg, s.keys.Right):
		s.tree.Expand()
		return true, nil
	case key.Matches(msg, s.keys.Space):
		s.tree.Toggle()
		return true, nil
	}
	return false, nil
}

func (s *TreeSkin) SelectedProject() *domain.Project {
	item := s.tree.Selected()
	if item == nil {
		return nil
	}
	return item.Project
}

func (s *TreeSkin) SelectedSession() *domain.Session {
	item := s.tree.Selected()
	if item == nil || item.Type != components.ItemTypeSession {
		return nil
	}
	return item.Session
}

func (s *TreeSkin) IsSessionSelected() bool {
	item := s.tree.Selected()
	return item != nil && item.Type == components.ItemTypeSession
}

func (s *TreeSkin) SetProjects(projects []*domain.Project) {
	s.tree.SetProjects(projects)
}

func (s *TreeSkin) SelectByProjectID(id string) {
	s.tree.SelectByProjectID(id)
}

func (s *TreeSkin) SelectBySessionID(projectID, sessionID string) {
	s.tree.SelectBySessionID(projectID, sessionID)
}

func (s *TreeSkin) Flatten() {
	s.tree.Flatten()
}

// Toggle exposes the tree's Toggle for handleEnter.
func (s *TreeSkin) Toggle() {
	s.tree.Toggle()
}
