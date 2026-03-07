# Architecture

## System Context

Codely runs inside a tmux session. The TUI occupies one pane and manages additional panes for AI tools, shells, and other commands. For remote development, it delegates to the `shed` CLI.

```text
┌─────────────────────────────────────────────────────────────────────────┐
│                              tmux session                               │
│                                                                         │
│  ┌──────────────────────┐  ┌──────────────────────────────────────────┐ │
│  │                      │  │                                          │ │
│  │  Codely TUI          │  │  Active Pane                             │ │
│  │  (Go + Bubble Tea)   │  │  (claude/opencode/codex/bash)            │ │
│  │                      │  │                                          │ │
│  │  - Project list      │  │  Managed by Codely                       │ │
│  │  - Status indicators │  │  - Created via tmux split                │ │
│  │  - Quick actions     │  │  - Or via shed exec/attach               │ │
│  │                      │  │                                          │ │
│  └──────────────────────┘  └──────────────────────────────────────────┘ │
│           │                              ▲                               │
│           │ tmux commands                │                               │
│           └──────────────────────────────┘                               │
│                                                                         │
│           │ shed CLI (for remote containers)                            │
│           ▼                                                              │
│  ┌──────────────────────────────────────────────────────────────────────┐│
│  │  Shed Servers (remote dev containers)                                ││
│  └──────────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────────┘
```

### Component Overview

| Component | Responsibility |
|-----------|----------------|
| **TUI (Bubble Tea)** | User interface, keyboard handling, view rendering |
| **tmux Client** | Local pane creation, focus management, content capture |
| **shed Client** | Remote shed listing, creation, attachment |
| **Status Detector** | Parses pane output to determine session state |
| **Config Manager** | Loads/saves workspace roots, commands, preferences |
| **Project Store** | Tracks active projects and their associated panes |

## Data Model

The data model is hierarchical: projects contain sessions.

```text
Project (workspace)
├── Session (terminal 1)
├── Session (terminal 2)
└── Session (terminal 3)
```

### Core Types

```go
// Project represents a workspace (local directory or shed)
type Project struct {
	ID        string      `json:"id"`        // UUID
	Name      string      `json:"name"`      // Display name (derived from dir/shed)
	Type      ProjectType `json:"type"`      // local or shed
	Directory string      `json:"directory"` // Local path (for local projects)

	// Shed-specific fields
	ShedName   string `json:"shed_name,omitempty"`
	ShedServer string `json:"shed_server,omitempty"`

	// Child sessions
	Sessions []Session `json:"sessions"`

	// UI state (not persisted)
	Expanded bool `json:"-"` // Collapsed/expanded in tree view
}
```

```go
// Session represents a terminal pane running within a project
type Session struct {
	ID        string  `json:"id"`         // UUID
	ProjectID string  `json:"project_id"` // Parent project
	Command   Command `json:"command"`    // What's running

	// Runtime state (not persisted)
	PaneID    int       `json:"-"` // tmux pane ID
	Status    Status    `json:"-"` // Current status
	StartedAt time.Time `json:"-"`
	IsVisible bool      `json:"-"` // Currently visible in main window?
	ExitCode  *int      `json:"-"` // Exit code if process exited
}
```

```go
// Command defines what runs in a session
type Command struct {
	ID          string            `json:"id"`           // e.g., "claude", "lazygit"
	DisplayName string            `json:"display_name"` // Human-readable name
	Exec        string            `json:"exec"`         // Binary to run
	Args        []string          `json:"args"`         // Arguments
	Env         map[string]string `json:"env"`          // Environment variables
}
```

```go
type Status string
const (
	StatusIdle      Status = "idle"
	StatusWaiting   Status = "waiting"
	StatusThinking  Status = "thinking"
	StatusExecuting Status = "executing"
	StatusError     Status = "error"
	StatusExited    Status = "exited"
	StatusStopped   Status = "stopped"
	StatusUnknown   Status = "unknown"
)
```

### Relationships

```text
┌──────────────────────────────────┐
│ Project                          │
│   ID: "abc-123"                  │
│   Name: "my-service"             │
│   Type: local                    │
│   Directory: "/home/user/src"    │
│                                  │
│   Sessions:                      │
│   ┌────────────────────────────┐ │
│   │ Session                    │ │
│   │   ID: "def-456"           │ │
│   │   Command: claude          │ │
│   │   PaneID: 3               │ │
│   │   Status: thinking         │ │
│   └────────────────────────────┘ │
│   ┌────────────────────────────┐ │
│   │ Session                    │ │
│   │   ID: "ghi-789"           │ │
│   │   Command: bash            │ │
│   │   PaneID: 5               │ │
│   │   Status: idle             │ │
│   └────────────────────────────┘ │
└──────────────────────────────────┘
```

## Core Workflows

### Startup

Launch flow: check tmux -> load config -> load session state -> reconnect existing panes -> clean dead sessions -> check shed status -> set up tmux layout -> render.

### Project Creation

Three paths:

- **Local**: folder picker -> command picker -> launch in tmux pane.
- **Attach Shed**: shed picker -> optional start -> command picker -> launch.
- **Create Shed**: form (name, repo, server) -> `shed create` -> command picker -> launch.

### Session Management

- **Add Terminal**: validate project -> show command picker -> `tmux split-window` -> capture pane ID -> save state -> focus pane.
- **Focus Session**: get pane ID -> `tmux select-pane` -> update UI.
- **Close Session**: confirm -> `tmux kill-pane` -> remove from project -> save state.
- **Close Project**: confirm -> kill all session panes -> remove project -> save state. Shed projects get additional options: close only, stop, or delete.

### Navigation

**Tree skin**: `j`/`k` moves through visible rows, left/right collapses/expands, Enter on a session focuses its pane, Enter on a project toggles expand.

**Flat skin**: `j`/`k` moves between projects, Enter focuses the first session, left/right/space are no-ops.

## tmux Integration

### Client Interface

The full tmux client interface from `internal/tmux/client.go`:

```go
// Client defines the interface for tmux operations
type Client interface {
	// Session management
	InTmux() bool
	CreateSession(name string) error
	AttachSession(name string) error

	// Pane management
	SplitWindow(dir, command string, args ...string) (paneID int, err error)
	SplitPane(targetPaneID int, vertical bool, dir, command string, args ...string) (paneID int, err error)
	FocusPane(paneID int) error
	KillPane(paneID int) error
	ResizePane(paneID int, width int) error
	ToggleZoom(paneID int) error
	SetRemainOnExit(paneID int, enabled bool) error

	// Pane visibility management
	BreakPane(paneID int) (newPaneID int, err error)
	JoinPane(paneID int, targetPaneID int) (newPaneID int, err error)

	// Content capture
	CapturePane(paneID int, lines int) (string, error)

	// Information
	ListPanes() ([]PaneInfo, error)
	PaneExists(paneID int) bool
	GetPaneWidth(paneID int) (int, error)

	// Status bar + key binding
	GetStatusRight() (string, error)
	SetStatusRight(value string) error
	BindJumpKey(key string, paneID int) error
	UnbindJumpKey(key string) error
}
```

### Key Commands

| Operation | Command |
|-----------|---------|
| Check if in tmux | `[ -n "$TMUX" ]` |
| Create session | `tmux new-session -d -s codely` |
| Split horizontally | `tmux split-window -h -c <dir> -P -F "#{pane_id}" <cmd>` |
| Focus pane | `tmux select-pane -t %<id>` |
| Kill pane | `tmux kill-pane -t %<id>` |
| Resize pane | `tmux resize-pane -t %<id> -x <width>` |
| Capture content | `tmux capture-pane -t %<id> -p -S -<lines>` |
| List panes | `tmux list-panes -a -F "#{pane_id}:#{pane_current_command}:..."` |
| Break pane | `tmux break-pane -d -P -F "#{pane_id}"` |
| Join pane | `tmux join-pane -s %<src> -t %<dst> -h` |

Pane IDs are returned by tmux as `%N` where `N` is an integer. They are stored as `int` internally and formatted with the `%` prefix when constructing commands.

## shed Integration

### Client Interface

The full shed client interface from `internal/shed/client.go`:

```go
// Client defines the interface for shed operations
type Client interface {
	Available() bool

	// Listing
	ListSheds() ([]Shed, error)
	ListServers() ([]Server, error)

	// Lifecycle
	CreateShed(name string, opts CreateOpts) error
	StartShed(name string) error
	StopShed(name string) error
	DeleteShed(name string, force bool) error

	// Streaming creation
	CreateShedStreaming(name string, opts CreateOpts) (cmdLine string, outputCh <-chan string, doneCh <-chan error)

	// Execution
	ExecCommand(shedName, command string, args ...string) *exec.Cmd
	Console(shedName string) *exec.Cmd
}
```

### Key Commands

| Operation | Command |
|-----------|---------|
| List all sheds | `shed list --all --json` |
| List servers | `shed server list --json` |
| Create shed | `shed create <name> --repo <repo> --server <server> --json` |
| Start shed | `shed start <name> --json` |
| Stop shed | `shed stop <name> --json` |
| Delete shed | `shed delete <name> --force --json` |
| Run command | `shed exec <name> <command>` |
| Open shell | `shed console <name>` |

### Running AI Tools in Sheds

```bash
# Claude in shed
shed exec codelens claude --dangerously-skip-permissions

# Bash in shed (use console for interactive)
shed console codelens
```

Wrapped in tmux:

```bash
tmux split-window -h "shed exec codelens claude --dangerously-skip-permissions"
```

## Error Handling

Codely applies graceful degradation:

1. **shed not installed**: Shed-related options are hidden. Local projects work normally.
2. **Pane died unexpectedly**: Session is marked as error/exited. The user is offered a restart option.
3. **Shed unreachable**: A connection error is shown with a retry option.
4. **Config file missing**: Defaults are used. Config is created on first save.

Error display example:

```text
┌─────────────────────────────────────────┐
│ Codely                                  │
├─────────────────────────────────────────┤
│  ... project list ...                   │
│                                         │
├─────────────────────────────────────────┤
│ Failed to start shed: timeout           │
│ [Enter] dismiss  [r] retry              │
└─────────────────────────────────────────┘
```

## Extension Points

- **Command plugins**: Add new AI tools or shells via config commands.
- **Custom status detectors**: Register tool-specific detection via `status_detection` config.
- **UI skins**: Implement the `Skin` interface in `internal/tui/` to create new panel renderings.
- **Theming**: Lipgloss styles in `internal/tui/styles.go` control all colors and borders.
