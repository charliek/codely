package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/charliek/codely/internal/config"
	"github.com/charliek/codely/internal/domain"
	"github.com/charliek/codely/internal/pathutil"
)

// FlatSkin renders projects as a scrollable list of cards.
type FlatSkin struct {
	projects    []*domain.Project
	items       []flatCardItem
	selectedIdx int
	config      *config.Config
	keys        KeyMap
}

type flatCardItem struct {
	project *domain.Project
	session *domain.Session
}

// NewFlatSkin creates a flat card skin.
func NewFlatSkin(projects []*domain.Project, cfg *config.Config, keys KeyMap) *FlatSkin {
	s := &FlatSkin{
		projects: projects,
		config:   cfg,
		keys:     keys,
	}
	s.rebuildItems()
	return s
}

var (
	styleCardBorder = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colorSecondary).
			Padding(0, 1).
			MarginBottom(1)

	styleCardBorderSelected = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(colorPrimary).
				Padding(0, 1).
				MarginBottom(1)

	styleCardTitle = lipgloss.NewStyle().
			Bold(true)

	styleCardPath = lipgloss.NewStyle().
			Foreground(colorMuted).
			Italic(true)

	styleCardMeta = lipgloss.NewStyle().
			Foreground(colorMuted)

	styleCardActive = lipgloss.NewStyle().
			Foreground(colorSuccess)
)

func (s *FlatSkin) View(m *Model) string {
	if len(s.items) == 0 {
		return styleProjectPath.Render("\n  No projects yet. Press 'n' to create one.\n")
	}

	var b strings.Builder

	for i, item := range s.items {
		card := s.renderCard(m, item)
		if i == s.selectedIdx {
			b.WriteString(styleCardBorderSelected.Width(m.width - 4).Render(card))
		} else {
			b.WriteString(styleCardBorder.Width(m.width - 4).Render(card))
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (s *FlatSkin) renderCard(m *Model, item flatCardItem) string {
	var b strings.Builder
	proj := item.project

	if item.session == nil {
		b.WriteString(styleCardTitle.Render(proj.Name))
		b.WriteString("\n")
		b.WriteString("\n")
		b.WriteString(styleCardPath.Render(pathutil.ContractHome(proj.DisplayPath())))
		b.WriteString("\n")
		b.WriteString(styleCardMeta.Render("No terminals yet"))
		return b.String()
	}

	sess := item.session
	b.WriteString(styleCardTitle.Render(sess.Command.Name()))
	b.WriteString("\n")
	b.WriteString(styleCardMeta.Render(fmt.Sprintf("Project: %s", proj.Name)))
	b.WriteString("\n")
	b.WriteString(styleCardPath.Render(pathutil.ContractHome(proj.DisplayPath())))

	statusStr := sess.Status.Icon()
	if sess.Status == domain.StatusError && sess.ExitCode != nil {
		statusStr = fmt.Sprintf("%s (%d)", statusStr, *sess.ExitCode)
	}

	b.WriteString("\n")
	if sess.IsVisible {
		b.WriteString(styleCardActive.Render("● visible"))
	} else {
		b.WriteString(styleCardMeta.Render("○ hidden"))
	}
	b.WriteString(styleCardMeta.Render("  "))
	b.WriteString(styleStatus(sess.Status).Render(statusStr))

	return b.String()
}

func (s *FlatSkin) rebuildItems() {
	s.items = s.items[:0]
	for _, proj := range s.projects {
		if len(proj.Sessions) == 0 {
			s.items = append(s.items, flatCardItem{project: proj})
			continue
		}
		for i := range proj.Sessions {
			s.items = append(s.items, flatCardItem{
				project: proj,
				session: &proj.Sessions[i],
			})
		}
	}
}

func (s *FlatSkin) HandleKey(m *Model, msg tea.KeyMsg) (bool, tea.Cmd) {
	switch {
	case key.Matches(msg, s.keys.Up):
		if s.selectedIdx > 0 {
			s.selectedIdx--
		}
		return true, nil
	case key.Matches(msg, s.keys.Down):
		if s.selectedIdx < len(s.projects)-1 {
			s.selectedIdx++
		}
		return true, nil
	case key.Matches(msg, s.keys.Left), key.Matches(msg, s.keys.Right), key.Matches(msg, s.keys.Space):
		// No-op in flat view for these keys
		return true, nil
	}
	return false, nil
}

func (s *FlatSkin) SelectedProject() *domain.Project {
	if len(s.items) == 0 || s.selectedIdx < 0 || s.selectedIdx >= len(s.items) {
		return nil
	}
	return s.items[s.selectedIdx].project
}

func (s *FlatSkin) SelectedSession() *domain.Session {
	if len(s.items) == 0 || s.selectedIdx < 0 || s.selectedIdx >= len(s.items) {
		return nil
	}
	return s.items[s.selectedIdx].session
}

func (s *FlatSkin) IsSessionSelected() bool {
	return s.SelectedSession() != nil
}

func (s *FlatSkin) SetProjects(projects []*domain.Project) {
	s.projects = projects
	s.rebuildItems()
	if s.selectedIdx >= len(s.items) {
		s.selectedIdx = len(s.items) - 1
	}
	if s.selectedIdx < 0 {
		s.selectedIdx = 0
	}
}

func (s *FlatSkin) SelectByProjectID(id string) {
	for i, item := range s.items {
		if item.project != nil && item.project.ID == id {
			s.selectedIdx = i
			return
		}
	}
}

func (s *FlatSkin) SelectBySessionID(projectID, sessionID string) {
	for i, item := range s.items {
		if item.project != nil && item.project.ID == projectID &&
			item.session != nil && item.session.ID == sessionID {
			s.selectedIdx = i
			return
		}
	}
	s.SelectByProjectID(projectID)
}

func (s *FlatSkin) ToggleProject() {
	// No expand/collapse in flat skin
}
