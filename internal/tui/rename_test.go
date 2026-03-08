package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/charliek/codely/internal/config"
	"github.com/charliek/codely/internal/domain"
	"github.com/charliek/codely/internal/shed"
	"github.com/charliek/codely/internal/store"
	"github.com/charliek/codely/internal/tmux"
)

func TestHandleNormalKeyStartsRenameForSelectedSession(t *testing.T) {
	model, _ := renameTestModel(t, SkinTree, "claude", "Claude Code")
	model.skin.SelectBySessionID("proj-1", "sess-1")

	updatedTea, _ := model.handleNormalKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	updated := updatedTea.(Model)

	assert.Equal(t, ModeRename, updated.mode)
	assert.Equal(t, "proj-1", updated.renameProjectID)
	assert.Equal(t, "sess-1", updated.renameSessionID)
	assert.Equal(t, "Claude Code", updated.renameInput.Value())
}

func TestHandleRenameKeySavesCustomSessionName(t *testing.T) {
	model, st := renameTestModel(t, SkinTree, "claude", "Claude Code")
	model.skin.SelectBySessionID("proj-1", "sess-1")

	updatedTea, _ := model.handleNormalKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	updated := updatedTea.(Model)
	updated.renameInput.SetValue("feature x")

	savedTea, _ := updated.handleRenameKey(tea.KeyMsg{Type: tea.KeyEnter})
	saved := savedTea.(Model)

	proj, err := st.GetProject("proj-1")
	require.NoError(t, err)
	assert.Equal(t, "feature x", proj.Sessions[0].Command.DisplayName)
	assert.Equal(t, ModeNormal, saved.mode)
	assert.Equal(t, "sess-1", saved.SelectedSession().ID)
}

func TestHandleRenameKeyBlankResetsToDefaultCommandName(t *testing.T) {
	model, st := renameTestModel(t, SkinFlat, "bash", "feature x")
	model.skin.SelectBySessionID("proj-1", "sess-1")

	updatedTea, _ := model.handleNormalKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	updated := updatedTea.(Model)
	updated.renameInput.SetValue("   ")

	savedTea, _ := updated.handleRenameKey(tea.KeyMsg{Type: tea.KeyEnter})
	saved := savedTea.(Model)

	proj, err := st.GetProject("proj-1")
	require.NoError(t, err)
	assert.Equal(t, "Bash Shell", proj.Sessions[0].Command.DisplayName)
	assert.Equal(t, ModeNormal, saved.mode)
	assert.Equal(t, "sess-1", saved.SelectedSession().ID)
}

func renameTestModel(t *testing.T, skin SkinName, commandID, displayName string) (Model, *store.Store) {
	t.Helper()

	cfg := config.Default()
	st := store.New(t.TempDir() + "/state.json")
	project := &domain.Project{
		ID:        "proj-1",
		Name:      "project",
		Type:      domain.ProjectTypeLocal,
		Directory: "/tmp/project",
		Expanded:  true,
		Sessions: []domain.Session{
			{
				ID:        "sess-1",
				ProjectID: "proj-1",
				Command: domain.Command{
					ID:          commandID,
					DisplayName: displayName,
					Exec:        commandID,
				},
			},
		},
	}
	require.NoError(t, st.AddProject(project))

	model := NewModel(cfg, st, tmux.NewMockClient(), shed.NewMockClient(), 0, "", skin)
	return *model, st
}
