package components

import (
	"testing"

	"github.com/charliek/codely/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestNewTree(t *testing.T) {
	projects := []*domain.Project{
		{
			ID:       "proj-1",
			Name:     "project1",
			Expanded: true,
			Sessions: []domain.Session{
				{ID: "sess-1", ProjectID: "proj-1"},
			},
		},
		{
			ID:       "proj-2",
			Name:     "project2",
			Expanded: false,
			Sessions: []domain.Session{
				{ID: "sess-2", ProjectID: "proj-2"},
			},
		},
	}

	tree := NewTree(projects)

	// Should have 3 items: proj1, sess1, proj2 (proj2 collapsed so sess2 hidden)
	assert.Equal(t, 3, tree.Count())
}

func TestTreeNavigation(t *testing.T) {
	projects := []*domain.Project{
		{
			ID:       "proj-1",
			Name:     "project1",
			Expanded: true,
			Sessions: []domain.Session{
				{ID: "sess-1", ProjectID: "proj-1"},
				{ID: "sess-2", ProjectID: "proj-1"},
			},
		},
	}

	tree := NewTree(projects)

	// Initially at index 0 (project)
	assert.Equal(t, 0, tree.SelectedIndex())
	assert.Equal(t, ItemTypeProject, tree.Selected().Type)

	// Move down to first session
	tree.MoveDown()
	assert.Equal(t, 1, tree.SelectedIndex())
	assert.Equal(t, ItemTypeSession, tree.Selected().Type)
	assert.Equal(t, "sess-1", tree.Selected().Session.ID)

	// Move down to second session
	tree.MoveDown()
	assert.Equal(t, 2, tree.SelectedIndex())
	assert.Equal(t, "sess-2", tree.Selected().Session.ID)

	// Move down at end does nothing
	tree.MoveDown()
	assert.Equal(t, 2, tree.SelectedIndex())

	// Move up
	tree.MoveUp()
	assert.Equal(t, 1, tree.SelectedIndex())

	// Move up to start
	tree.MoveUp()
	assert.Equal(t, 0, tree.SelectedIndex())

	// Move up at start does nothing
	tree.MoveUp()
	assert.Equal(t, 0, tree.SelectedIndex())
}

func TestTreeToggle(t *testing.T) {
	projects := []*domain.Project{
		{
			ID:       "proj-1",
			Name:     "project1",
			Expanded: true,
			Sessions: []domain.Session{
				{ID: "sess-1", ProjectID: "proj-1"},
			},
		},
	}

	tree := NewTree(projects)

	// Initially 2 items (project + session)
	assert.Equal(t, 2, tree.Count())

	// Toggle collapse
	tree.Toggle()
	assert.Equal(t, 1, tree.Count()) // Only project visible

	// Toggle expand
	tree.Toggle()
	assert.Equal(t, 2, tree.Count()) // Project + session
}

func TestTreeCollapse(t *testing.T) {
	projects := []*domain.Project{
		{
			ID:       "proj-1",
			Name:     "project1",
			Expanded: true,
			Sessions: []domain.Session{
				{ID: "sess-1", ProjectID: "proj-1"},
			},
		},
	}

	tree := NewTree(projects)

	// Move to session
	tree.MoveDown()
	assert.Equal(t, ItemTypeSession, tree.Selected().Type)

	// Collapse should move to parent project
	tree.Collapse()
	assert.Equal(t, 0, tree.SelectedIndex())
	assert.Equal(t, ItemTypeProject, tree.Selected().Type)

	// Collapse project
	tree.Collapse()
	assert.False(t, projects[0].Expanded)
	assert.Equal(t, 1, tree.Count())
}

func TestTreeExpand(t *testing.T) {
	projects := []*domain.Project{
		{
			ID:       "proj-1",
			Name:     "project1",
			Expanded: false,
			Sessions: []domain.Session{
				{ID: "sess-1", ProjectID: "proj-1"},
			},
		},
	}

	tree := NewTree(projects)

	// Initially collapsed
	assert.Equal(t, 1, tree.Count())

	// Expand
	tree.Expand()
	assert.True(t, projects[0].Expanded)
	assert.Equal(t, 2, tree.Count())
}

func TestTreeSelectByID(t *testing.T) {
	projects := []*domain.Project{
		{
			ID:       "proj-1",
			Name:     "project1",
			Expanded: true,
			Sessions: []domain.Session{
				{ID: "sess-1", ProjectID: "proj-1"},
			},
		},
		{
			ID:       "proj-2",
			Name:     "project2",
			Expanded: true,
			Sessions: []domain.Session{
				{ID: "sess-2", ProjectID: "proj-2"},
			},
		},
	}

	tree := NewTree(projects)

	// Select by project ID
	tree.SelectByProjectID("proj-2")
	assert.Equal(t, "proj-2", tree.Selected().Project.ID)

	// Select by session ID
	tree.SelectBySessionID("proj-1", "sess-1")
	assert.Equal(t, ItemTypeSession, tree.Selected().Type)
	assert.Equal(t, "sess-1", tree.Selected().Session.ID)
}

func TestTreeEmpty(t *testing.T) {
	tree := NewTree([]*domain.Project{})

	assert.True(t, tree.IsEmpty())
	assert.Equal(t, 0, tree.Count())
	assert.Nil(t, tree.Selected())
}

func TestTreeSetProjects(t *testing.T) {
	tree := NewTree([]*domain.Project{})

	assert.True(t, tree.IsEmpty())

	projects := []*domain.Project{
		{ID: "proj-1", Name: "project1", Expanded: true},
	}
	tree.SetProjects(projects)

	assert.False(t, tree.IsEmpty())
	assert.Equal(t, 1, tree.ProjectCount())
}
