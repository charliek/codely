package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charliek/codely/internal/domain"
)

const statusBarPrefix = "Codely:"

func (m *Model) updateTmuxNotifications() {
	items := m.collectNotificationItems()
	segment, keyMap := formatStatusSegment(items)

	m.updateStatusRight(segment)
	m.updateJumpKeys(keyMap)
}

func (m *Model) clearTmuxNotifications() {
	m.updateStatusRight("")
	m.updateJumpKeys(map[string]int{})
}

func (m *Model) collectNotificationItems() []notificationItem {
	var items []notificationItem
	for _, proj := range m.store.Projects() {
		for _, sess := range proj.Sessions {
			if sess.PaneID == 0 {
				continue
			}
			switch sess.Status {
			case domain.StatusWaiting, domain.StatusError:
				label := fmt.Sprintf("%s/%s", proj.Name, sess.Command.ID)
				items = append(items, notificationItem{
					label:  label,
					paneID: sess.PaneID,
					status: sess.Status,
				})
			}
		}
	}
	return items
}

type notificationItem struct {
	label  string
	paneID int
	status domain.Status
}

func formatStatusSegment(items []notificationItem) (string, map[string]int) {
	if len(items) == 0 {
		return "", map[string]int{}
	}

	const maxItems = 6
	total := len(items)
	if total > maxItems {
		items = items[:maxItems]
	}

	parts := make([]string, 0, len(items)+1)
	keyMap := make(map[string]int, len(items))
	for i, item := range items {
		key := strconv.Itoa(i + 1)
		keyMap[key] = item.paneID
		label := item.label
		if item.status == domain.StatusError {
			label = "! " + label
		}
		parts = append(parts, fmt.Sprintf("[%s] %s", key, label))
	}

	if total > maxItems {
		parts = append(parts, fmt.Sprintf("+%d", total-maxItems))
	}
	segment := statusBarPrefix + " " + strings.Join(parts, " ")
	return segment, keyMap
}

func (m *Model) updateStatusRight(segment string) {
	current, err := m.tmux.GetStatusRight()
	if err != nil {
		return
	}
	base := stripStatusSegment(current)
	newStatus := base
	if segment != "" {
		if base != "" {
			newStatus = base + " | " + segment
		} else {
			newStatus = segment
		}
	}

	if newStatus == m.statusBarLast {
		return
	}
	if err := m.tmux.SetStatusRight(newStatus); err != nil {
		return
	}
	m.statusBarLast = newStatus
}

func (m *Model) updateJumpKeys(newKeys map[string]int) {
	for key, paneID := range newKeys {
		if existing, ok := m.statusBarKeys[key]; ok && existing == paneID {
			continue
		}
		_ = m.tmux.BindJumpKey(key, paneID)
	}
	for key := range m.statusBarKeys {
		if _, ok := newKeys[key]; !ok {
			_ = m.tmux.UnbindJumpKey(key)
		}
	}
	m.statusBarKeys = newKeys
}

func stripStatusSegment(statusRight string) string {
	idx := strings.Index(statusRight, statusBarPrefix)
	if idx == -1 {
		return strings.TrimSpace(statusRight)
	}
	if idx >= 3 && statusRight[idx-3:idx] == " | " {
		return strings.TrimSpace(statusRight[:idx-3])
	}
	return strings.TrimSpace(statusRight[:idx])
}
