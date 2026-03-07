package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/charliek/codely/internal/config"
	"github.com/charliek/codely/internal/domain"
)

// FlatSkin renders projects as a scrollable list of cards.
type FlatSkin struct {
	projects    []*domain.Project
	selectedIdx int
	config      *config.Config
	keys        KeyMap
}

// NewFlatSkin creates a flat card skin.
func NewFlatSkin(projects []*domain.Project, cfg *config.Config, keys KeyMap) *FlatSkin {
	return &FlatSkin{
		projects: projects,
		config:   cfg,
		keys:     keys,
	}
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
)

func (s *FlatSkin) View(m *Model) string {
	if len(s.projects) == 0 {
		return styleProjectPath.Render("\n  No projects yet. Press 'n' to create one.\n")
	}

	var b strings.Builder

	for i, proj := range s.projects {
		card := s.renderCard(m, proj)
		if i == s.selectedIdx {
			b.WriteString(styleCardBorderSelected.Width(m.width - 4).Render(card))
		} else {
			b.WriteString(styleCardBorder.Width(m.width - 4).Render(card))
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (s *FlatSkin) renderCard(m *Model, proj *domain.Project) string {
	var b strings.Builder

	// Title line
	b.WriteString(styleCardTitle.Render(proj.Name))

	// Path
	path := proj.DisplayPath()
	home := homeDir()
	if strings.HasPrefix(path, home) {
		path = "~" + path[len(home):]
	}
	b.WriteString("\n")
	b.WriteString(styleCardPath.Render(path))

	// Session summary
	total := len(proj.Sessions)
	active := 0
	for _, sess := range proj.Sessions {
		if sess.Status == domain.StatusThinking || sess.Status == domain.StatusExecuting || sess.Status == domain.StatusWaiting {
			active++
		}
	}

	if total > 0 {
		b.WriteString("\n")
		b.WriteString(styleCardMeta.Render(fmt.Sprintf("%d sessions", total)))
		if active > 0 {
			b.WriteString(styleCardMeta.Render("  "))
			b.WriteString(lipgloss.NewStyle().Foreground(colorSuccess).Render(fmt.Sprintf("● %d active", active)))
		}
	}

	// Session detail line
	if total > 0 {
		b.WriteString("\n")
		var parts []string
		for _, sess := range proj.Sessions {
			name := sess.Command.DisplayName
			if name == "" {
				name = sess.Command.ID
			}
			icon := sess.Status.Icon()
			styled := styleStatus(sess.Status).Render(icon)
			parts = append(parts, fmt.Sprintf("%s %s", name, styled))
		}
		b.WriteString(strings.Join(parts, "  "))
	}

	return b.String()
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
	if len(s.projects) == 0 || s.selectedIdx < 0 || s.selectedIdx >= len(s.projects) {
		return nil
	}
	return s.projects[s.selectedIdx]
}

func (s *FlatSkin) SelectedSession() *domain.Session {
	// Flat skin selects at project level, not session level
	return nil
}

func (s *FlatSkin) IsSessionSelected() bool {
	return false
}

func (s *FlatSkin) SetProjects(projects []*domain.Project) {
	s.projects = projects
	if s.selectedIdx >= len(s.projects) {
		s.selectedIdx = len(s.projects) - 1
	}
	if s.selectedIdx < 0 {
		s.selectedIdx = 0
	}
}

func (s *FlatSkin) SelectByProjectID(id string) {
	for i, p := range s.projects {
		if p.ID == id {
			s.selectedIdx = i
			return
		}
	}
}

func (s *FlatSkin) SelectBySessionID(projectID, sessionID string) {
	// Select the project containing this session
	s.SelectByProjectID(projectID)
}

func (s *FlatSkin) Flatten() {
	// No-op for flat skin — no expand/collapse semantics
}
