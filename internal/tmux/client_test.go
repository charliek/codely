package tmux

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockClient(t *testing.T) {
	m := NewMockClient()

	assert.True(t, m.InTmux())
	assert.Len(t, m.Calls, 1)
	assert.Equal(t, "InTmux", m.Calls[0].Method)
}

func TestMockClientSplitWindow(t *testing.T) {
	m := NewMockClient()
	m.SplitWindowPaneID = 42

	paneID, err := m.SplitWindow("/tmp", "bash", "-c", "echo hello")

	assert.NoError(t, err)
	assert.Equal(t, 42, paneID)
	assert.Len(t, m.Calls, 1)
	assert.Equal(t, "SplitWindow", m.Calls[0].Method)
}

func TestMockClientFocusPane(t *testing.T) {
	m := NewMockClient()

	err := m.FocusPane(5)

	assert.NoError(t, err)
	assert.Len(t, m.Calls, 1)
	assert.Equal(t, "FocusPane", m.Calls[0].Method)
	assert.Equal(t, 5, m.Calls[0].Args[0])
}

func TestMockClientCapturePane(t *testing.T) {
	m := NewMockClient()
	m.CapturePaneResult = "$ echo hello\nhello\n$ "

	content, err := m.CapturePane(5, 10)

	assert.NoError(t, err)
	assert.Equal(t, "$ echo hello\nhello\n$ ", content)
}

func TestMockClientListPanes(t *testing.T) {
	m := NewMockClient()
	m.ListPanesResult = []PaneInfo{
		{ID: 1, Command: "bash", Active: true},
		{ID: 2, Command: "vim", Active: false},
	}

	panes, err := m.ListPanes()

	assert.NoError(t, err)
	assert.Len(t, panes, 2)
	assert.Equal(t, 1, panes[0].ID)
	assert.Equal(t, "bash", panes[0].Command)
}

func TestMockClientPaneExists(t *testing.T) {
	m := NewMockClient()
	m.PaneExistsResult = true

	assert.True(t, m.PaneExists(5))

	m.PaneExistsResult = false
	assert.False(t, m.PaneExists(5))
}
