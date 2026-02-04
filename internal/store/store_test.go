package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/charliek/codely/internal/domain"
	"github.com/charliek/codely/internal/tmux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoreLoadSave(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "codely-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	path := filepath.Join(tmpDir, "state.json")
	s := New(path)

	// Add a project
	p := &domain.Project{
		ID:        "proj-1",
		Name:      "test-project",
		Type:      domain.ProjectTypeLocal,
		Directory: "/home/user/test",
		Sessions: []domain.Session{
			{
				ID:        "sess-1",
				ProjectID: "proj-1",
				Command: domain.Command{
					ID:   "claude",
					Exec: "claude",
				},
			},
		},
	}

	err = s.AddProject(p)
	require.NoError(t, err)

	// Save
	err = s.Save()
	require.NoError(t, err)

	// Create new store and load
	s2 := New(path)
	err = s2.Load()
	require.NoError(t, err)

	// Verify project loaded
	projects := s2.Projects()
	assert.Len(t, projects, 1)
	assert.Equal(t, "proj-1", projects[0].ID)
	assert.Equal(t, "test-project", projects[0].Name)
	assert.Len(t, projects[0].Sessions, 1)
}

func TestStoreAddRemoveProject(t *testing.T) {
	s := New("/tmp/nonexistent.json")

	p := &domain.Project{
		ID:   "proj-1",
		Name: "test",
	}

	// Add project
	err := s.AddProject(p)
	assert.NoError(t, err)
	assert.Len(t, s.Projects(), 1)

	// Add duplicate - should fail
	err = s.AddProject(p)
	assert.Error(t, err)

	// Remove project
	err = s.RemoveProject("proj-1")
	assert.NoError(t, err)
	assert.Len(t, s.Projects(), 0)

	// Remove non-existent
	err = s.RemoveProject("proj-1")
	assert.ErrorIs(t, err, domain.ErrProjectNotFound)
}

func TestStoreGetProject(t *testing.T) {
	s := New("/tmp/nonexistent.json")

	p := &domain.Project{
		ID:   "proj-1",
		Name: "test",
	}
	_ = s.AddProject(p)

	// Get existing
	got, err := s.GetProject("proj-1")
	assert.NoError(t, err)
	assert.Equal(t, "test", got.Name)

	// Get non-existent
	_, err = s.GetProject("proj-999")
	assert.ErrorIs(t, err, domain.ErrProjectNotFound)
}

func TestStoreAddRemoveSession(t *testing.T) {
	s := New("/tmp/nonexistent.json")

	p := &domain.Project{
		ID:   "proj-1",
		Name: "test",
	}
	_ = s.AddProject(p)

	sess := &domain.Session{
		ID:        "sess-1",
		ProjectID: "proj-1",
		Command: domain.Command{
			ID:   "claude",
			Exec: "claude",
		},
	}

	// Add session
	err := s.AddSession("proj-1", sess)
	assert.NoError(t, err)

	got, err := s.GetProject("proj-1")
	require.NoError(t, err)
	assert.Len(t, got.Sessions, 1)

	// Add to non-existent project
	err = s.AddSession("proj-999", sess)
	assert.ErrorIs(t, err, domain.ErrProjectNotFound)

	// Remove session
	err = s.RemoveSession("proj-1", "sess-1")
	assert.NoError(t, err)

	got, _ = s.GetProject("proj-1")
	assert.Len(t, got.Sessions, 0)

	// Remove non-existent session
	err = s.RemoveSession("proj-1", "sess-999")
	assert.ErrorIs(t, err, domain.ErrSessionNotFound)
}

func TestStoreCleanupDeadSessions(t *testing.T) {
	s := New("/tmp/nonexistent.json")

	p := &domain.Project{
		ID:   "proj-1",
		Name: "test",
		Sessions: []domain.Session{
			{ID: "sess-1", ProjectID: "proj-1", PaneID: 1},
			{ID: "sess-2", ProjectID: "proj-1", PaneID: 2},
			{ID: "sess-3", ProjectID: "proj-1", PaneID: 3},
		},
	}
	_ = s.AddProject(p)

	// Create mock that only knows about pane 2
	mock := tmux.NewMockClient()
	mock.ListPanesResult = []tmux.PaneInfo{
		{ID: 2, Command: "bash"},
	}
	// Override PaneExists behavior via list panes
	mock.PaneExistsResult = false

	// Custom PaneExists that checks the list
	mockWithCustomExists := &mockClientWithCustomPaneExists{
		MockClient: mock,
		existingPanes: map[int]bool{
			2: true,
		},
	}

	s.CleanupDeadSessions(mockWithCustomExists)

	got, _ := s.GetProject("proj-1")
	assert.Len(t, got.Sessions, 1)
	assert.Equal(t, "sess-2", got.Sessions[0].ID)
}

type mockClientWithCustomPaneExists struct {
	*tmux.MockClient
	existingPanes map[int]bool
}

func (m *mockClientWithCustomPaneExists) PaneExists(paneID int) bool {
	return m.existingPanes[paneID]
}

func TestStoreTmuxSession(t *testing.T) {
	s := New("/tmp/nonexistent.json")

	// Default
	assert.Equal(t, "codely", s.TmuxSession())

	// Set
	s.SetTmuxSession("my-session")
	assert.Equal(t, "my-session", s.TmuxSession())
}

func TestStoreLoadNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "codely-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	path := filepath.Join(tmpDir, "subdir", "state.json")
	s := New(path)

	// Load should succeed with empty state
	err = s.Load()
	assert.NoError(t, err)
	assert.Len(t, s.Projects(), 0)
}

func TestStoreUpdateProject(t *testing.T) {
	s := New("/tmp/nonexistent.json")

	p := &domain.Project{
		ID:   "proj-1",
		Name: "original",
	}
	_ = s.AddProject(p)

	// Update
	p.Name = "updated"
	err := s.UpdateProject(p)
	assert.NoError(t, err)

	got, _ := s.GetProject("proj-1")
	assert.Equal(t, "updated", got.Name)

	// Update non-existent
	err = s.UpdateProject(&domain.Project{ID: "proj-999"})
	assert.ErrorIs(t, err, domain.ErrProjectNotFound)
}
