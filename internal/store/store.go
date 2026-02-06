// Package store provides persistence for projects and sessions.
package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/charliek/codely/internal/debug"
	"github.com/charliek/codely/internal/domain"
	"github.com/charliek/codely/internal/pathutil"
	"github.com/charliek/codely/internal/tmux"
)

// State represents the persisted session state
type State struct {
	Projects    []*domain.Project `json:"projects"`
	TmuxSession string            `json:"tmux_session"`
}

// Store handles persistence of projects and sessions
type Store struct {
	path  string
	state State
	mu    sync.RWMutex
}

// New creates a new store with the given path
func New(path string) *Store {
	return &Store{
		path: pathutil.ExpandPath(path),
		state: State{
			Projects:    []*domain.Project{},
			TmuxSession: "codely",
		},
	}
}

// Load reads the state from disk
func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Ensure directory exists with restricted permissions (owner only)
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating state directory: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(s.path); os.IsNotExist(err) {
		// No existing state, use defaults
		return nil
	}

	data, err := os.ReadFile(s.path)
	if err != nil {
		return fmt.Errorf("reading state file: %w", err)
	}

	if len(data) == 0 {
		return nil
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("parsing state file: %w", err)
	}

	s.state = state
	return nil
}

// Save writes the state to disk
func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Ensure directory exists with restricted permissions (owner only)
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating state directory: %w", err)
	}

	data, err := json.MarshalIndent(s.state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}

	// Write with restricted permissions (owner read/write only)
	if err := os.WriteFile(s.path, data, 0600); err != nil {
		return fmt.Errorf("writing state file: %w", err)
	}

	return nil
}

// Projects returns all projects
func (s *Store) Projects() []*domain.Project {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Projects
}

// AddProject adds a new project to the store
func (s *Store) AddProject(p *domain.Project) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for duplicate ID
	for _, existing := range s.state.Projects {
		if existing.ID == p.ID {
			return fmt.Errorf("project with ID %s already exists", p.ID)
		}
	}

	s.state.Projects = append(s.state.Projects, p)
	return nil
}

// RemoveProject removes a project by ID
func (s *Store) RemoveProject(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, p := range s.state.Projects {
		if p.ID == id {
			s.state.Projects = append(s.state.Projects[:i], s.state.Projects[i+1:]...)
			return nil
		}
	}

	return domain.ErrProjectNotFound
}

// GetProject returns a project by ID
func (s *Store) GetProject(id string) (*domain.Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, p := range s.state.Projects {
		if p.ID == id {
			return p, nil
		}
	}

	return nil, domain.ErrProjectNotFound
}

// UpdateProject updates a project in the store
func (s *Store) UpdateProject(p *domain.Project) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.state.Projects {
		if existing.ID == p.ID {
			s.state.Projects[i] = p
			return nil
		}
	}

	return domain.ErrProjectNotFound
}

// AddSession adds a session to a project
func (s *Store) AddSession(projectID string, session *domain.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, p := range s.state.Projects {
		if p.ID == projectID {
			p.Sessions = append(p.Sessions, *session)
			return nil
		}
	}

	return domain.ErrProjectNotFound
}

// RemoveSession removes a session from a project
func (s *Store) RemoveSession(projectID, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, p := range s.state.Projects {
		if p.ID == projectID {
			for i, sess := range p.Sessions {
				if sess.ID == sessionID {
					p.Sessions = append(p.Sessions[:i], p.Sessions[i+1:]...)
					return nil
				}
			}
			return domain.ErrSessionNotFound
		}
	}

	return domain.ErrProjectNotFound
}

// GetSession returns a session by project and session ID
func (s *Store) GetSession(projectID, sessionID string) (*domain.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, p := range s.state.Projects {
		if p.ID == projectID {
			for i := range p.Sessions {
				if p.Sessions[i].ID == sessionID {
					return &p.Sessions[i], nil
				}
			}
			return nil, domain.ErrSessionNotFound
		}
	}

	return nil, domain.ErrProjectNotFound
}

// CleanupDeadSessions removes sessions whose panes no longer exist
func (s *Store) CleanupDeadSessions(tmuxClient tmux.Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, p := range s.state.Projects {
		var liveSessions []domain.Session
		for _, sess := range p.Sessions {
			if sess.PaneID > 0 && tmuxClient.PaneExists(sess.PaneID) {
				liveSessions = append(liveSessions, sess)
			}
		}
		p.Sessions = liveSessions
	}
}

// ReconnectSessions attempts to reconnect sessions to their existing panes
// by matching pane IDs. Sessions with pane ID 0 or non-existent panes are removed.
func (s *Store) ReconnectSessions(tmuxClient tmux.Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	panes, err := tmuxClient.ListPanes()
	if err != nil {
		return
	}

	paneMap := make(map[int]bool)
	for _, p := range panes {
		paneMap[p.ID] = true
	}
	debug.Log("reconnectSessions: panes=%d", len(paneMap))

	for _, p := range s.state.Projects {
		before := len(p.Sessions)
		var liveSessions []domain.Session
		for _, sess := range p.Sessions {
			if sess.PaneID > 0 && paneMap[sess.PaneID] {
				liveSessions = append(liveSessions, sess)
			}
		}
		p.Sessions = liveSessions
		debug.Log("reconnectSessions: project=%s before=%d after=%d", p.Name, before, len(liveSessions))
	}
}

// TmuxSession returns the tmux session name
func (s *Store) TmuxSession() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.TmuxSession
}

// SetTmuxSession sets the tmux session name
func (s *Store) SetTmuxSession(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.TmuxSession = name
}
