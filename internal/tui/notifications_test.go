package tui

import (
	"testing"

	"github.com/charliek/codely/internal/config"
	"github.com/charliek/codely/internal/domain"
	"github.com/charliek/codely/internal/shed"
	"github.com/charliek/codely/internal/store"
	"github.com/charliek/codely/internal/tmux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStripStatusSegment(t *testing.T) {
	assert.Equal(t, "foo", stripStatusSegment("foo | Codely: [1] a"))
	assert.Equal(t, "", stripStatusSegment("Codely: [1] a"))
	assert.Equal(t, "bar", stripStatusSegment("bar"))
}

func TestFormatStatusSegment(t *testing.T) {
	items := []notificationItem{
		{label: "api/claude", paneID: 1, status: domain.StatusWaiting},
		{label: "web/opencode", paneID: 2, status: domain.StatusError},
	}
	segment, keyMap := formatStatusSegment(items)
	assert.Equal(t, "Codely: [1] api/claude [2] ! web/opencode", segment)
	assert.Equal(t, 1, keyMap["1"])
	assert.Equal(t, 2, keyMap["2"])
}

func TestCollectNotificationItemsUsesSessionDisplayName(t *testing.T) {
	st := store.New(t.TempDir() + "/state.json")
	require.NoError(t, st.AddProject(&domain.Project{
		ID:   "proj-1",
		Name: "api",
		Sessions: []domain.Session{
			{
				ID:        "sess-1",
				ProjectID: "proj-1",
				PaneID:    7,
				Status:    domain.StatusWaiting,
				Command: domain.Command{
					ID:          "claude",
					DisplayName: "feature x",
				},
			},
		},
	}))

	model := NewModel(config.Default(), st, tmux.NewMockClient(), shed.NewMockClient(), 0, "", SkinTree)
	items := model.collectNotificationItems()

	require.Len(t, items, 1)
	assert.Equal(t, "api/feature x", items[0].label)
}
