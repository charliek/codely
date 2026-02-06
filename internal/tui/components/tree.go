// Package components provides TUI components for the application.
package components

import (
	"github.com/charliek/codely/internal/domain"
)

// ItemType represents the type of item in the tree
type ItemType int

const (
	ItemTypeProject ItemType = iota
	ItemTypeSession
)

// TreeItem represents a row in the flattened tree
type TreeItem struct {
	Type    ItemType
	Project *domain.Project
	Session *domain.Session // nil if Type is ItemTypeProject
}

// Tree manages a hierarchical tree view of projects and sessions
type Tree struct {
	projects    []*domain.Project
	items       []TreeItem
	selectedIdx int
}

// NewTree creates a new tree from a slice of projects
func NewTree(projects []*domain.Project) *Tree {
	t := &Tree{
		projects: projects,
	}
	t.Flatten()
	return t
}

// Flatten rebuilds the flat item list from projects
func (t *Tree) Flatten() {
	t.items = nil

	for _, proj := range t.projects {
		// Add project row
		t.items = append(t.items, TreeItem{
			Type:    ItemTypeProject,
			Project: proj,
		})

		// Add session rows if expanded
		if proj.Expanded {
			for i := range proj.Sessions {
				t.items = append(t.items, TreeItem{
					Type:    ItemTypeSession,
					Project: proj,
					Session: &proj.Sessions[i],
				})
			}
		}
	}

	// Clamp selected index
	if t.selectedIdx >= len(t.items) {
		t.selectedIdx = len(t.items) - 1
	}
	if t.selectedIdx < 0 {
		t.selectedIdx = 0
	}
}

// Items returns the flattened items
func (t *Tree) Items() []TreeItem {
	return t.items
}

// SelectedIndex returns the current selected index
func (t *Tree) SelectedIndex() int {
	return t.selectedIdx
}

// Selected returns the currently selected item
func (t *Tree) Selected() *TreeItem {
	if len(t.items) == 0 {
		return nil
	}
	if t.selectedIdx < 0 || t.selectedIdx >= len(t.items) {
		return nil
	}
	return &t.items[t.selectedIdx]
}

// MoveUp moves selection up
func (t *Tree) MoveUp() {
	if t.selectedIdx > 0 {
		t.selectedIdx--
	}
}

// MoveDown moves selection down
func (t *Tree) MoveDown() {
	if t.selectedIdx < len(t.items)-1 {
		t.selectedIdx++
	}
}

// Toggle expands or collapses the current project
func (t *Tree) Toggle() {
	item := t.Selected()
	if item == nil {
		return
	}

	// Find the project to toggle
	var proj *domain.Project
	if item.Type == ItemTypeProject {
		proj = item.Project
	} else {
		proj = item.Project
	}

	proj.Expanded = !proj.Expanded
	t.Flatten()
}

// Expand expands the current project
func (t *Tree) Expand() {
	item := t.Selected()
	if item == nil {
		return
	}

	if item.Type == ItemTypeProject && !item.Project.Expanded {
		item.Project.Expanded = true
		t.Flatten()
	}
}

// Collapse collapses the current project or moves to parent
func (t *Tree) Collapse() {
	item := t.Selected()
	if item == nil {
		return
	}

	if item.Type == ItemTypeProject {
		if item.Project.Expanded {
			item.Project.Expanded = false
			t.Flatten()
		}
	} else {
		// Move to parent project
		for i, it := range t.items {
			if it.Type == ItemTypeProject && it.Project.ID == item.Project.ID {
				t.selectedIdx = i
				break
			}
		}
	}
}

// SelectByProjectID selects a project by its ID
func (t *Tree) SelectByProjectID(id string) {
	for i, item := range t.items {
		if item.Type == ItemTypeProject && item.Project.ID == id {
			t.selectedIdx = i
			return
		}
	}
}

// SelectBySessionID selects a session by its ID
func (t *Tree) SelectBySessionID(projectID, sessionID string) {
	for i, item := range t.items {
		if item.Type == ItemTypeSession &&
			item.Project.ID == projectID &&
			item.Session.ID == sessionID {
			t.selectedIdx = i
			return
		}
	}
}

// SetProjects updates the projects and rebuilds the tree
func (t *Tree) SetProjects(projects []*domain.Project) {
	t.projects = projects
	t.Flatten()
}

// Projects returns the underlying projects
func (t *Tree) Projects() []*domain.Project {
	return t.projects
}

// IsEmpty returns true if the tree has no items
func (t *Tree) IsEmpty() bool {
	return len(t.items) == 0
}

// Count returns the number of visible items
func (t *Tree) Count() int {
	return len(t.items)
}

// ProjectCount returns the number of projects
func (t *Tree) ProjectCount() int {
	return len(t.projects)
}
