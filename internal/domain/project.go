package domain

import (
	"fmt"
	"time"
)

// ProjectType represents the type of project (local or shed)
type ProjectType string

const (
	// ProjectTypeLocal represents a local directory project
	ProjectTypeLocal ProjectType = "local"

	// ProjectTypeShed represents a remote shed project
	ProjectTypeShed ProjectType = "shed"
)

// Project represents a workspace (local directory or shed)
type Project struct {
	ID        string      `json:"id"`        // UUID
	Name      string      `json:"name"`      // Display name (derived from dir/shed)
	Type      ProjectType `json:"type"`      // local or shed
	Directory string      `json:"directory"` // Local path (for local projects)

	// Shed-specific fields
	ShedName   string `json:"shed_name,omitempty"`
	ShedServer string `json:"shed_server,omitempty"`

	// Child sessions
	Sessions []Session `json:"sessions"`

	// UI state (not persisted)
	Expanded bool `json:"-"` // Collapsed/expanded in tree view
}

// DisplayPath returns the path shown in UI
func (p *Project) DisplayPath() string {
	if p.Type == ProjectTypeShed {
		return fmt.Sprintf("shed:%s", p.ShedServer)
	}
	return p.Directory
}

// Session represents a terminal pane running within a project
type Session struct {
	ID        string  `json:"id"`         // UUID
	ProjectID string  `json:"project_id"` // Parent project
	Command   Command `json:"command"`    // What's running

	// Runtime state (not persisted)
	PaneID    int       `json:"-"` // tmux pane ID (can change after break/join)
	Status    Status    `json:"-"` // Current status
	StartedAt time.Time `json:"-"`
	IsVisible bool      `json:"-"` // Currently visible in main window?
	ExitCode  *int      `json:"-"` // Exit code if process exited
}

// Command defines what runs in a session
type Command struct {
	ID          string            `json:"id"`           // e.g., "claude", "lazygit"
	DisplayName string            `json:"display_name"` // Human-readable name
	Exec        string            `json:"exec"`         // Binary to run
	Args        []string          `json:"args"`         // Arguments
	Env         map[string]string `json:"env"`          // Environment variables
}
