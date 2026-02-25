package tui

import (
	"github.com/charliek/codely/internal/domain"
	"github.com/charliek/codely/internal/shed"
)

// TickMsg is sent periodically for status polling
type TickMsg struct{}

// StatusUpdateMsg contains updated status for sessions
type StatusUpdateMsg struct {
	Updates   map[string]domain.Status // session ID -> status
	ExitCodes map[string]*int          // session ID -> exit code (if any)
}

// PaneCreatedMsg is sent when a new tmux pane is created
type PaneCreatedMsg struct {
	ProjectID        string
	SessionID        string
	PaneID           int
	ReplacedSessions []string // Session IDs that were killed (in replace mode) - deprecated
	HiddenProjectID  string   // Project that was hidden (break-pane) in single visible pane mode
	HiddenSessionID  string   // Session that was hidden (break-pane) in single visible pane mode
	HiddenPaneID     int      // New pane ID of the hidden session (pane ID changes after break)
	DetectedWidth    int      // Width of codely pane before break (to restore after split)
	Err              error
}

// PaneKilledMsg is sent when a tmux pane is killed
type PaneKilledMsg struct {
	ProjectID string
	SessionID string
	Err       error
}

// ShedsLoadedMsg is sent when sheds are loaded
type ShedsLoadedMsg struct {
	Sheds []shed.Shed
	Err   error
}

// ShedStartedMsg is sent when a shed is started
type ShedStartedMsg struct {
	ShedName string
	Err      error
}

// ShedStoppedMsg is sent when a shed is stopped
type ShedStoppedMsg struct {
	ShedName string
	Err      error
}

// ShedCreatedMsg is sent when a shed is created
type ShedCreatedMsg struct {
	ShedName string
	Err      error
}

// ShedDeletedMsg is sent when a shed is deleted
type ShedDeletedMsg struct {
	ShedName string
	Err      error
}

// shedCreateStartedMsg carries channels from the streaming create process.
type shedCreateStartedMsg struct {
	name     string
	cmdLine  string
	outputCh <-chan string
	doneCh   <-chan error
}

// shedCreateOutputMsg carries one stderr line and channels for chaining.
type shedCreateOutputMsg struct {
	line     string
	name     string
	outputCh <-chan string
	doneCh   <-chan error
}

// ErrorMsg represents an error to display
type ErrorMsg struct {
	Err error
}

// ClearErrorMsg clears the current error
type ClearErrorMsg struct{}

// FocusPaneMsg is sent to focus a pane
type FocusPaneMsg struct {
	PaneID int
	Err    error
}

// ProjectCreatedMsg is sent when a project is created
type ProjectCreatedMsg struct {
	Project *domain.Project
}

// FoldersLoadedMsg is sent when folders are loaded for picker
type FoldersLoadedMsg struct {
	Folders []string
	Err     error
}

// PaneSwappedMsg is sent when panes are swapped (hidden session brought to front)
type PaneSwappedMsg struct {
	ShownProjectID  string
	HiddenProjectID string
	ShownSessionID  string // Session that was brought to front
	ShownPaneID     int    // New pane ID of the shown session
	HiddenSessionID string // Session that was hidden
	HiddenPaneID    int    // New pane ID of the hidden session
	DetectedWidth   int    // Width of codely pane before swap (to restore after join)
	Err             error
}

// VisibilitySyncedMsg is sent when session visibility has been reconciled to tmux state
type VisibilitySyncedMsg struct {
	VisibleSessionID string // Session ID that is visible in the main window (if any)
	Err              error
}
