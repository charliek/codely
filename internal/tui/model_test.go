package tui

import (
	"testing"

	"github.com/charliek/codely/internal/config"
	"github.com/charliek/codely/internal/domain"
	"github.com/charliek/codely/internal/shed"
	"github.com/charliek/codely/internal/store"
	"github.com/charliek/codely/internal/tmux"
	"github.com/stretchr/testify/assert"
)

func TestNewModel(t *testing.T) {
	cfg := config.Default()
	st := store.New("/tmp/test-state.json")
	tmuxClient := tmux.NewMockClient()
	shedClient := shed.NewMockClient()

	model := NewModel(cfg, st, tmuxClient, shedClient, 0, "")

	assert.NotNil(t, model)
	assert.Equal(t, ModeNormal, model.mode)
	assert.NotNil(t, model.tree)
}

func TestModelWithProjects(t *testing.T) {
	cfg := config.Default()
	st := store.New("/tmp/test-state.json")

	// Add a project
	proj := &domain.Project{
		ID:        "proj-1",
		Name:      "test-project",
		Type:      domain.ProjectTypeLocal,
		Directory: "/tmp/test",
		Expanded:  true,
		Sessions: []domain.Session{
			{
				ID:        "sess-1",
				ProjectID: "proj-1",
				Command: domain.Command{
					ID:          "claude",
					DisplayName: "Claude Code",
					Exec:        "claude",
				},
			},
		},
	}
	_ = st.AddProject(proj)

	tmuxClient := tmux.NewMockClient()
	shedClient := shed.NewMockClient()

	model := NewModel(cfg, st, tmuxClient, shedClient, 0, "")

	// Verify project is in tree
	assert.Equal(t, 1, model.tree.ProjectCount())
	assert.Equal(t, 2, model.tree.Count()) // project + session

	// Verify selection methods
	assert.NotNil(t, model.SelectedProject())
	assert.Equal(t, "proj-1", model.SelectedProject().ID)

	// Move to session
	model.tree.MoveDown()
	assert.True(t, model.IsSessionSelected())
	assert.NotNil(t, model.SelectedSession())
	assert.Equal(t, "sess-1", model.SelectedSession().ID)
}

func TestModelInit(t *testing.T) {
	cfg := config.Default()
	st := store.New("/tmp/test-state.json")
	tmuxClient := tmux.NewMockClient()
	shedClient := shed.NewMockClient()

	model := NewModel(cfg, st, tmuxClient, shedClient, 0, "")

	// Init should return commands for polling and loading
	cmd := model.Init()
	assert.NotNil(t, cmd)
}

func TestModelModes(t *testing.T) {
	cfg := config.Default()
	st := store.New("/tmp/test-state.json")
	tmuxClient := tmux.NewMockClient()
	shedClient := shed.NewMockClient()

	model := NewModel(cfg, st, tmuxClient, shedClient, 0, "")

	// Test mode transitions
	assert.Equal(t, ModeNormal, model.mode)

	model.mode = ModeFolderPicker
	assert.Equal(t, ModeFolderPicker, model.mode)

	model.mode = ModeCommandPicker
	assert.Equal(t, ModeCommandPicker, model.mode)

	model.mode = ModeConfirm
	assert.Equal(t, ModeConfirm, model.mode)
}

func TestStatusUpdateMsg(t *testing.T) {
	cfg := config.Default()
	st := store.New("/tmp/test-state.json")

	proj := &domain.Project{
		ID:   "proj-1",
		Name: "test",
		Sessions: []domain.Session{
			{ID: "sess-1", ProjectID: "proj-1"},
		},
	}
	_ = st.AddProject(proj)

	tmuxClient := tmux.NewMockClient()
	shedClient := shed.NewMockClient()

	model := NewModel(cfg, st, tmuxClient, shedClient, 0, "")

	// Apply status update
	updates := map[string]domain.Status{
		"sess-1": domain.StatusThinking,
	}
	model.applyStatusUpdates(updates)

	// Verify status was updated
	p, _ := st.GetProject("proj-1")
	assert.Equal(t, domain.StatusThinking, p.Sessions[0].Status)
}
