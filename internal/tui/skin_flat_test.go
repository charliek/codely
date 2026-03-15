package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/charliek/codely/internal/config"
	"github.com/charliek/codely/internal/domain"
)

func TestFlatSkinBuildsCardPerSession(t *testing.T) {
	projects := []*domain.Project{
		{
			ID:        "proj-1",
			Name:      "project-1",
			Type:      domain.ProjectTypeLocal,
			Directory: "/tmp/project-1",
			Sessions: []domain.Session{
				{ID: "sess-1", ProjectID: "proj-1"},
				{ID: "sess-2", ProjectID: "proj-1"},
			},
		},
		{
			ID:        "proj-2",
			Name:      "project-2",
			Type:      domain.ProjectTypeLocal,
			Directory: "/tmp/project-2",
		},
	}

	skin := NewFlatSkin(projects, config.Default(), DefaultKeyMap())

	require.Len(t, skin.items, 3)
	assert.Equal(t, "sess-1", skin.items[0].session.ID)
	assert.Equal(t, "sess-2", skin.items[1].session.ID)
	assert.Nil(t, skin.items[2].session)
	assert.Equal(t, "proj-2", skin.items[2].project.ID)
}

func TestFlatSkinSelectBySessionID(t *testing.T) {
	projects := []*domain.Project{
		{
			ID:        "proj-1",
			Name:      "project-1",
			Type:      domain.ProjectTypeLocal,
			Directory: "/tmp/project-1",
			Sessions: []domain.Session{
				{ID: "sess-1", ProjectID: "proj-1"},
				{ID: "sess-2", ProjectID: "proj-1"},
			},
		},
	}

	skin := NewFlatSkin(projects, config.Default(), DefaultKeyMap())
	skin.SelectBySessionID("proj-1", "sess-2")

	require.NotNil(t, skin.SelectedProject())
	require.NotNil(t, skin.SelectedSession())
	assert.Equal(t, "proj-1", skin.SelectedProject().ID)
	assert.Equal(t, "sess-2", skin.SelectedSession().ID)
	assert.True(t, skin.IsSessionSelected())
}

func TestFlatSkinSetProjectsPreservesSelectionByID(t *testing.T) {
	projects := []*domain.Project{
		{
			ID:   "proj-a",
			Name: "alpha",
			Sessions: []domain.Session{
				{ID: "sess-1", ProjectID: "proj-a"},
				{ID: "sess-2", ProjectID: "proj-a"},
			},
		},
		{
			ID:   "proj-b",
			Name: "beta",
			Sessions: []domain.Session{
				{ID: "sess-3", ProjectID: "proj-b"},
			},
		},
	}

	skin := NewFlatSkin(projects, config.Default(), DefaultKeyMap())

	// Select sess-3 (index 2)
	skin.SelectBySessionID("proj-b", "sess-3")
	require.Equal(t, "sess-3", skin.SelectedSession().ID)

	// Reorder: put proj-b first
	reordered := []*domain.Project{projects[1], projects[0]}
	skin.SetProjects(reordered)

	// Selection should still be sess-3 even though it moved to index 0
	require.NotNil(t, skin.SelectedSession())
	assert.Equal(t, "sess-3", skin.SelectedSession().ID)
	assert.Equal(t, "proj-b", skin.SelectedProject().ID)
	assert.Equal(t, 0, skin.selectedIdx)
}

func TestFlatSkinSetProjectsPreservesProjectSelection(t *testing.T) {
	projects := []*domain.Project{
		{ID: "proj-a", Name: "alpha"},
		{ID: "proj-b", Name: "beta"},
	}

	skin := NewFlatSkin(projects, config.Default(), DefaultKeyMap())

	// Select proj-b (index 1)
	skin.SelectByProjectID("proj-b")
	require.Equal(t, "proj-b", skin.SelectedProject().ID)

	// Reorder
	reordered := []*domain.Project{projects[1], projects[0]}
	skin.SetProjects(reordered)

	assert.Equal(t, "proj-b", skin.SelectedProject().ID)
	assert.Equal(t, 0, skin.selectedIdx)
}
