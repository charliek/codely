package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/charliek/codely/internal/config"
	"github.com/charliek/codely/internal/domain"
)

// SkinName identifies a UI skin
type SkinName string

const (
	SkinTree SkinName = "tree"
	SkinFlat SkinName = "flat"
)

// Skin controls how the main panel renders projects/sessions
// and how navigation keys work in ModeNormal.
type Skin interface {
	// View renders the main content area (between header and footer).
	View(m *Model) string

	// HandleKey handles navigation keys (up/down/left/right/space) in ModeNormal.
	// Returns true if the key was consumed.
	HandleKey(m *Model, msg tea.KeyMsg) (handled bool, cmd tea.Cmd)

	// Selection queries
	SelectedProject() *domain.Project
	SelectedSession() *domain.Session
	IsSessionSelected() bool

	// Actions
	ToggleProject()

	// Data sync
	SetProjects(projects []*domain.Project)
	SelectByProjectID(id string)
	SelectBySessionID(projectID, sessionID string)
}

// NewSkin creates a skin by name
func NewSkin(name SkinName, projects []*domain.Project, cfg *config.Config, keys KeyMap) Skin {
	switch name {
	case SkinFlat:
		return NewFlatSkin(projects, cfg, keys)
	default:
		return NewTreeSkin(projects, cfg, keys)
	}
}
