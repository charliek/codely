package tui

import (
	"testing"

	"github.com/charliek/codely/internal/domain"
	"github.com/stretchr/testify/assert"
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
