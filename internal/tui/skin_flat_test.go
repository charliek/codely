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
