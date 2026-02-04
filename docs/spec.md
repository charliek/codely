# Codely: AI Coding Session Manager

## Overview

Codely is a terminal-based project manager for orchestrating AI coding sessions across local directories and remote development containers. It provides a unified interface for launching, monitoring, and switching between multiple concurrent coding sessions running tools like Claude Code, OpenCode, Codex, or standard shells.

**Key Value Proposition:**
- Single pane of glass for all active AI coding sessions
- Seamless switching between local and remote (shed) environments
- Real-time status monitoring (idle, thinking, executing)
- Quick project provisioning via shed integration

---

## Architecture

### System Context

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              tmux session                               │
│                                                                         │
│  ┌──────────────────────┐  ┌──────────────────────────────────────────┐ │
│  │                      │  │                                          │ │
│  │  Codely TUI        │  │  Active Pane                             │ │
│  │  (Go + Bubble Tea)   │  │  (claude/opencode/codex/bash)            │ │
│  │                      │  │                                          │ │
│  │  - Project list      │  │  Managed by Codely                     │ │
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
│  │  - mini-desktop.tailnet.ts.net                                       ││
│  │  - cloud-vps.tailnet.ts.net                                          ││
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

---

## User Interface

### Main View: Project Tree

```
┌─────────────────────────────────────────┐
│ Codely                            v0.1│
├─────────────────────────────────────────┤
│                                         │
│  LOCAL                                  │
│  ▼ codelens                             │
│    ~/projects/codelens                  │
│    ● claude              🤔 thinking    │
│    ○ opencode            💤 idle        │
│                                         │
│  ▼ api-server                           │
│    ~/work/smartthings/api               │
│    ○ claude              ⚡ executing   │
│                                         │
│  ▼ frontend                             │
│    ~/work/smartthings/web               │
│    (no terminals)                       │
│                                         │
│  ▶ mobile                 (2 sessions)  │
│                                         │
│  SHEDS                                  │
│  ▼ test-shed (mini)                     │
│    charliek/test-repo                   │
│    ○ claude              💤 idle        │
│    ○ lazygit             💤 idle        │
│                                         │
│  ▼ old-project (mini)    ⏸️ stopped     │
│    charliek/old-repo                    │
│    (no terminals)                       │
│                                         │
│  ▶ stbot (cloud)          (1 session)   │
│                                         │
├─────────────────────────────────────────┤
│ [n]ew project [t]erminal [x]close [q]uit│
└─────────────────────────────────────────┘
```

Tip: Use tmux zoom (`prefix` + `z`) to toggle fullscreen for the active pane.

**Notes:**
- Projects with zero sessions show "(no terminals)" 
- Stopped sheds show ⏸️ status at project level
- Collapsed projects show session count

### Tree Navigation

- Projects can be **expanded** (▼) or **collapsed** (▶)
- When collapsed, shows session count
- Selection can be on:
  - A project row (for project-level actions)
  - A session row (for focusing/closing that session)

### Selection States

```
# Project selected (expanded)
▼ codelens                    ← SELECTED
  ~/projects/codelens
  ○ claude              💤
  ○ opencode            💤

# Session selected
▼ codelens
  ~/projects/codelens
  ● claude              🤔    ← SELECTED
  ○ opencode            💤

# Project selected (collapsed)  
▶ codelens (2 sessions)       ← SELECTED
```

### Status Icons

| Icon | Status | Meaning |
|------|--------|---------|
| 💤 | `idle` | Waiting for user input (prompt visible) |
| 🤔 | `thinking` | AI is processing (spinner/thinking indicator) |
| ⚡ | `executing` | Running user code or commands |
| ❌ | `error` | Process crashed or exited with error |
| ⏸️ | `stopped` | Shed is stopped (not running) |

### View: New Project (Local)

```
┌─────────────────────────────────────────┐
│ New Local Project                       │
├─────────────────────────────────────────┤
│                                         │
│  Select directory:                      │
│                                         │
│  ~/work/smartthings/                    │
│    ○ api/                               │
│    ○ web/                               │
│    ○ mobile/                            │
│    ○ infrastructure/                    │
│                                         │
│  ~/projects/stridelabs/                 │
│    ○ audio/                             │
│    ○ codely/                          │
│                                         │
│  [/] search  [enter] select  [esc] back │
└─────────────────────────────────────────┘
```

### View: Add Terminal to Project

When pressing `t` on a project, or after creating a new project:

```
┌─────────────────────────────────────────┐
│ Add Terminal                            │
├─────────────────────────────────────────┤
│                                         │
│  Project: codelens                      │
│  Path: ~/projects/codelens              │
│                                         │
│  Select command:                        │
│                                         │
│  ● claude                               │
│    claude --dangerously-skip-permissions│
│                                         │
│  ○ opencode                             │
│    opencode                             │
│                                         │
│  ○ codex                                │
│    codex                                │
│                                         │
│  ○ lazygit                              │
│    lazygit                              │
│                                         │
│  ○ bash                                 │
│    bash                                 │
│                                         │
│  [enter] launch  [esc] back             │
└─────────────────────────────────────────┘
```

### View: Attach to Existing Shed

```
┌─────────────────────────────────────────┐
│ Attach to Shed                          │
├─────────────────────────────────────────┤
│                                         │
│  Available Sheds:                       │
│                                         │
│  mini-desktop                           │
│  ● codelens        running    2h ago    │
│  ○ mcp-test        stopped    3d ago    │
│  ○ scratch         running    1h ago    │
│                                         │
│  cloud-vps                              │
│  ○ stbot           running    30m ago   │
│  ○ experiments     stopped    1w ago    │
│                                         │
│  [enter] select  [s] start  [esc] back  │
└─────────────────────────────────────────┘
```

### View: Create New Shed

```
┌─────────────────────────────────────────┐
│ Create New Shed                         │
├─────────────────────────────────────────┤
│                                         │
│  Shed name: my-new-project_             │
│                                         │
│  Repository (optional):                 │
│  ○ None (scratch shed)                  │
│  ● From GitHub: charliek/_              │
│                                         │
│  Server:                                │
│  ● mini-desktop (default)               │
│  ○ cloud-vps                            │
│                                         │
│  [enter] create  [esc] back             │
│                                         │
│  ⏳ Creating shed...                    │
│     Cloning repository                  │
│     Running provisioning hooks          │
└─────────────────────────────────────────┘
```

### View: Close Shed Project

```
┌─────────────────────────────────────────┐
│ Close Shed Project                      │
├─────────────────────────────────────────┤
│                                         │
│  Project: test-shed (mini-desktop)      │
│                                         │
│  What would you like to do?             │
│                                         │
│  ● Close project only                   │
│    Shed keeps running on server.        │
│    You can re-attach later.             │
│                                         │
│  ○ Close and stop shed                  │
│    Stops the container but keeps data.  │
│    Can be restarted later.              │
│                                         │
│  ○ Close and DELETE shed                │
│    Permanently removes container and    │
│    all data. Cannot be undone.          │
│                                         │
│  [enter] confirm  [esc] cancel          │
└─────────────────────────────────────────┘
```

---

## Data Model

### Overview

The data model is hierarchical:

```
Project (workspace)
├── Session (terminal 1)
├── Session (terminal 2)
└── Session (terminal 3)
```

- **Project**: A workspace, either a local directory or a remote shed
- **Session**: A terminal pane running a specific command within that project

This allows multiple tools (claude, opencode, lazygit, bash) to run simultaneously in the same project context.

### Project

```go
// Project represents a workspace (local directory or shed)
type Project struct {
    ID          string            `json:"id"`          // UUID
    Name        string            `json:"name"`        // Display name (derived from dir/shed)
    Type        ProjectType       `json:"type"`        // local or shed
    Directory   string            `json:"directory"`   // Local path (for local projects)
    
    // Shed-specific fields
    ShedName    string            `json:"shed_name,omitempty"`
    ShedServer  string            `json:"shed_server,omitempty"`
    
    // Child sessions
    Sessions    []Session         `json:"sessions"`
    
    // UI state
    Expanded    bool              `json:"-"`           // Collapsed/expanded in tree view
}

type ProjectType string
const (
    ProjectTypeLocal ProjectType = "local"
    ProjectTypeShed  ProjectType = "shed"
)

// DisplayPath returns the path shown in UI
func (p *Project) DisplayPath() string {
    if p.Type == ProjectTypeShed {
        return fmt.Sprintf("shed:%s", p.ShedServer)
    }
    return p.Directory
}
```

### Session

```go
// Session represents a terminal pane running within a project
type Session struct {
    ID          string            `json:"id"`          // UUID
    ProjectID   string            `json:"project_id"`  // Parent project
    Command     Command           `json:"command"`     // What's running
    
    // Runtime state (not persisted)
    PaneID      int               `json:"-"`           // tmux pane ID
    Status      Status            `json:"-"`           // Current status
    StartedAt   time.Time         `json:"-"`
}
```

### Command

```go
// Command defines what runs in a session
type Command struct {
    ID          string            `json:"id"`          // e.g., "claude", "lazygit"
    DisplayName string            `json:"display_name"`
    Exec        string            `json:"exec"`        // Binary to run
    Args        []string          `json:"args"`        // Arguments
    Env         map[string]string `json:"env"`         // Environment variables
}
```

### Status

```go
type Status string
const (
    StatusIdle      Status = "idle"       // Prompt visible, waiting for input
    StatusThinking  Status = "thinking"   // AI processing (spinner visible)
    StatusExecuting Status = "executing"  // Running code/commands
    StatusError     Status = "error"      // Crashed or error state
    StatusStopped   Status = "stopped"    // Shed not running
    StatusUnknown   Status = "unknown"    // Cannot determine
)
```

### Shed (from shed CLI)

```go
// Shed represents a remote development container
type Shed struct {
    Name      string    `json:"name"`
    Server    string    `json:"server"`
    Status    string    `json:"status"`    // "running" or "stopped"
    CreatedAt time.Time `json:"created_at"`
    Repo      string    `json:"repo,omitempty"`
}
```

### Relationships

```
┌─────────────────────────────────────────────────────────────┐
│                         Project                             │
│  id: "proj-123"                                            │
│  name: "codelens"                                          │
│  type: local                                               │
│  directory: ~/projects/codelens                            │
│                                                            │
│  ┌─────────────────────┐  ┌─────────────────────┐         │
│  │      Session        │  │      Session        │         │
│  │  id: "sess-abc"     │  │  id: "sess-def"     │         │
│  │  command: claude    │  │  command: lazygit   │         │
│  │  pane_id: 5         │  │  pane_id: 7         │         │
│  │  status: thinking   │  │  status: idle       │         │
│  └─────────────────────┘  └─────────────────────┘         │
└─────────────────────────────────────────────────────────────┘
```

---

## Configuration

### Config File Location

```
~/.config/codely/config.yaml
```

### Config Schema

```yaml
# Codely Configuration

# Directories to show in folder picker
workspace_roots:
  - ~/work/smartthings
  - ~/projects/stridelabs
  - ~/src

# Available commands
commands:
  claude:
    display_name: "Claude Code"
    exec: claude
    args:
      - "--dangerously-skip-permissions"
    
  opencode:
    display_name: "OpenCode"
    exec: opencode
    args: []
    
  codex:
    display_name: "Codex"
    exec: codex
    args: []
    
  lazygit:
    display_name: "Lazygit"
    exec: lazygit
    args: []
    
  bash:
    display_name: "Bash Shell"
    exec: bash
    args: []

# Default command when adding first terminal to a project
default_command: claude

# UI preferences
ui:
  manager_width: 30          # Width of left panel in characters
  status_poll_interval: 1s   # How often to check pane status
  show_directory: true       # Show full path in project list
  auto_expand_projects: true # Expand projects by default

# Shed integration
shed:
  enabled: true
  default_server: mini-desktop
```

### Session State (Runtime)

Active projects and sessions are stored in:
```
~/.local/state/codely/session.json
```

**Persistence behavior:**
- **Projects persist** across Codely restarts (even with zero sessions)
- **Sessions are reconnected** if their tmux panes still exist
- **Dead sessions are cleaned up** on startup (pane no longer exists)
- Projects are only removed when explicitly closed by user

This means:
1. Close Codely → Reopen Codely → Projects still there
2. tmux panes continue running even if Codely exits
3. On restart, Codely reconnects to existing panes

```json
{
  "projects": [
    {
      "id": "proj-abc123",
      "name": "codelens",
      "type": "local",
      "directory": "/Users/charlie/projects/codelens",
      "sessions": [
        {
          "id": "sess-def456",
          "project_id": "proj-abc123",
          "command": {"id": "claude"},
          "pane_id": 5,
          "started_at": "2025-02-03T10:30:00Z"
        },
        {
          "id": "sess-ghi789",
          "project_id": "proj-abc123",
          "command": {"id": "lazygit"},
          "pane_id": 7,
          "started_at": "2025-02-03T10:35:00Z"
        }
      ]
    },
    {
      "id": "proj-xyz999",
      "name": "test-shed",
      "type": "shed",
      "shed_name": "test-shed",
      "shed_server": "mini-desktop",
      "sessions": [
        {
          "id": "sess-aaa111",
          "project_id": "proj-xyz999",
          "command": {"id": "claude"},
          "pane_id": 9,
          "started_at": "2025-02-03T11:00:00Z"
        }
      ]
    },
    {
      "id": "proj-empty",
      "name": "frontend",
      "type": "local",
      "directory": "/Users/charlie/work/frontend",
      "sessions": []
    }
  ],
  "tmux_session": "codely"
}
```

Note: The `"frontend"` project above has zero sessions but still persists, allowing the user to quickly add terminals to it later.

---

## Core Workflows

### Workflow 1: Launch Codely

```
User runs: codely

1. Check if inside tmux session
   ├─ No  → Create new tmux session "codely", exec self inside
   └─ Yes → Continue

2. Load config from ~/.config/codely/config.yaml

3. Load session state from ~/.local/state/codely/session.json
   ├─ For each saved project:
   │   ├─ Project is ALWAYS restored (persists across restarts)
   │   ├─ For each session in project:
   │   │   ├─ Check if pane still exists (tmux list-panes)
   │   │   ├─ If exists → restore session, reconnect to pane
   │   │   └─ If not → remove session from project (pane died)
   │   └─ Project kept even if all sessions were removed
   └─ Save cleaned session state

4. For SHED projects: check shed status
   ├─ Run: shed list --all --json
   ├─ Mark shed projects as "stopped" if shed not running
   └─ This affects status display, not project persistence

5. Set up tmux layout
   - Resize current pane to manager_width
   - This becomes the Codely UI pane

6. Render project tree view
   - Show all persisted projects (even those with 0 sessions)
   - Start status polling loop for active sessions
```

### Workflow 2: Create Local Project

```
User presses: n (new project)

1. Show workspace type picker
   - "Local Directory"
   - "Attach to Shed"
   - "Create New Shed"
   
   User selects "Local Directory"

2. Show folder picker view
   - List directories from workspace_roots
   - Support fuzzy search with /
   
3. User selects directory

4. Create project entry
   - Generate UUID for project
   - Name derived from directory basename
   - Type = "local"
   - Sessions = [] (empty)
   - Expanded = true

5. Add project to tree, save session state

6. Immediately show "Add Terminal" view
   - User selects command (claude, opencode, lazygit, etc.)

7. Create first session (see Workflow 5)
```

### Workflow 3: Attach to Existing Shed

```
User presses: n (new project) → selects "Attach to Shed"

1. Fetch shed list
   - Run: shed list --all --json
   - Parse JSON output

2. Show shed picker view
   - Group by server
   - Show status (running/stopped)

3. User selects shed

4. If shed is stopped:
   - Prompt: "Shed is stopped. Start it? [y/n]"
   - If yes: shed start <n>
   - Show spinner while waiting

5. Create project entry
   - Generate UUID
   - Name = shed name
   - Type = "shed"
   - ShedName = shed name
   - ShedServer = server name
   - Sessions = []
   - Expanded = true

6. Add project to tree, save session state

7. Immediately show "Add Terminal" view
   - User selects command

8. Create first session (see Workflow 5)
```

### Workflow 4: Create New Shed

```
User presses: n (new project) → selects "Create New Shed"

1. Show shed creation form
   - Shed name (required)
   - Repository (optional, supports autocomplete)
   - Server selection

2. User fills form and confirms

3. Create shed
   - Run: shed create <n> --repo <repo> --server <server>
   - Show progress spinner
   - This may take several minutes for large repos

4. On success:
   - Create project entry (as in Workflow 3 step 5)
   - Continue to "Add Terminal" view

5. On failure:
   - Show error message
   - Return to form
```

### Workflow 5: Add Terminal to Project

```
User presses: t (terminal) with a project selected
-- OR --
After creating a new project (Workflows 2-4)

1. Validate selection is a project (not a session)

2. For shed projects: verify shed is running
   - If stopped, prompt to start

3. Show command picker view
   - List commands from config
   - Highlight default_command

4. User selects command

5. Create tmux pane
   
   For LOCAL projects:
   - tmux split-window -h -c <directory> -P -F "#{pane_id}" "<command> <args>"
   
   For SHED projects:
   - tmux split-window -h -P -F "#{pane_id}" "shed exec <shed_name> <command> <args>"
   - For bash: "shed console <shed_name>"

6. Capture returned pane ID

7. Create session entry
   - Generate UUID
   - ProjectID = parent project ID
   - Command = selected command
   - PaneID = captured ID
   - Status = "unknown"

8. Add session to project's Sessions array

9. Save session state

10. Focus the new pane
    - tmux select-pane -t %<pane_id>

11. Expand project if collapsed
```

### Workflow 6: Navigate and Focus

```
User navigates with j/k or arrows

Tree structure allows two types of selection:
- PROJECT row: can add terminal, collapse/expand, close project
- SESSION row: can focus pane, close session

Navigation rules:
- j/k moves through visible rows (respecting collapsed state)
- Enter on SESSION → focus that pane
- Enter on PROJECT → toggle expand/collapse
- Space on PROJECT → toggle expand/collapse
- Left arrow → collapse current project (or go to parent)
- Right arrow → expand current project (or go to first child)
```

### Workflow 7: Focus Session Pane

```
User presses: Enter on a session row

1. Get session's pane ID

2. For shed sessions, verify shed is running
   - If stopped, prompt to start, recreate pane

3. Focus the pane
   - tmux select-pane -t %<pane_id>

4. Update UI to show this session as selected (●)
```

### Workflow 8: Close Session

```
User presses: x with a session selected

1. Confirm: "Close <command> in <project>? [y/n]"

2. Kill the tmux pane
   - tmux kill-pane -t %<pane_id>

3. Remove session from project's Sessions array

4. Save session state

5. Note: Project remains even with zero sessions
```

### Workflow 9: Close Project

```
User presses: X (shift-x) with a project selected
-- OR --
User presses: x with a project selected (not a session)

1. For LOCAL projects:
   - Confirm: "Close project <name> and all sessions? [y/n]"
   - If yes: kill all session panes, remove project

2. For SHED projects:
   - Show options dialog:
     ┌─────────────────────────────────────┐
     │ Close shed project: test-shed       │
     ├─────────────────────────────────────┤
     │                                     │
     │  What would you like to do?         │
     │                                     │
     │  ● Close project only               │
     │    (shed keeps running)             │
     │                                     │
     │  ○ Close and stop shed              │
     │    (can restart later)              │
     │                                     │
     │  ○ Close and delete shed            │
     │    (removes container & data)       │
     │                                     │
     │  [enter] confirm  [esc] cancel      │
     └─────────────────────────────────────┘

3. Based on selection:
   - "Close project only": 
     - Kill all session panes
     - Remove project from Codely
     - Shed continues running on server
   
   - "Close and stop shed":
     - Kill all session panes
     - Run: shed stop <name>
     - Remove project from Codely
   
   - "Close and delete shed":
     - Confirm again: "This will permanently delete the shed. Continue? [y/n]"
     - Kill all session panes
     - Run: shed delete <name> --force
     - Remove project from Codely

4. Save session state
```

### Workflow 10: Collapse/Expand Project

```
User presses: Enter or Space on a project row
-- OR --
User presses: Left/Right arrow

1. Toggle project.Expanded

2. If collapsing:
   - Hide all child session rows in UI
   - Show session count: "▶ codelens (2 sessions)"

3. If expanding:
   - Show all child session rows
   - Show "▼ codelens" with sessions listed below
```

---

## Status Detection

### Detection Strategy

Status is determined by capturing and analyzing the last N lines of terminal output from each pane.

```go
func DetectStatus(paneContent string) Status {
    lines := strings.Split(paneContent, "\n")
    recent := getLastNLines(lines, 10)
    
    // Check for spinner characters (Claude/AI thinking)
    if containsSpinner(recent) {
        return StatusThinking
    }
    
    // Check for explicit thinking indicators
    if containsThinkingText(recent) {
        return StatusThinking
    }
    
    // Check for prompt (idle state)
    if endsWithPrompt(recent) {
        return StatusIdle
    }
    
    // Check for error indicators
    if containsError(recent) {
        return StatusError
    }
    
    // Default: probably executing
    return StatusExecuting
}
```

### Detection Patterns

**Spinner Characters (thinking):**
```go
var spinnerChars = []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}
```

**Thinking Text Patterns:**
```go
var thinkingPatterns = []string{
    "thinking",
    "analyzing", 
    "reading",
    "processing",
    "Generating",
}
```

**Prompt Patterns (idle):**
```go
var promptPatterns = []*regexp.Regexp{
    regexp.MustCompile(`(?m)^[>$#%] ?$`),           // Common shell prompts
    regexp.MustCompile(`(?m)^claude[>:] ?$`),       // Claude prompt
    regexp.MustCompile(`(?m)^\(.*\)[>$] ?$`),       // Virtualenv prompts
}
```

**Error Patterns:**
```go
var errorPatterns = []string{
    "error:",
    "Error:",
    "ERROR",
    "panic:",
    "Traceback",
    "Exception:",
}
```

### Polling Implementation

```go
func (m *Model) pollStatus() tea.Cmd {
    return tea.Tick(m.config.UI.StatusPollInterval, func(t time.Time) tea.Msg {
        updates := make(map[string]Status)
        
        for _, proj := range m.projects {
            if proj.PaneID == 0 {
                continue
            }
            
            content, err := m.tmux.CapturePane(proj.PaneID, 15)
            if err != nil {
                updates[proj.ID] = StatusError
                continue
            }
            
            updates[proj.ID] = DetectStatus(content)
        }
        
        return statusUpdateMsg{updates: updates}
    })
}
```

---

## tmux Integration

### Client Interface

```go
type TmuxClient interface {
    // Session management
    InTmux() bool
    CreateSession(name string) error
    
    // Pane management
    SplitWindow(dir string, command string, args ...string) (paneID int, err error)
    FocusPane(paneID int) error
    KillPane(paneID int) error
    ResizePane(paneID int, width int) error
    
    // Content capture
    CapturePane(paneID int, lines int) (string, error)
    
    // Information
    ListPanes() ([]PaneInfo, error)
    GetPaneInfo(paneID int) (*PaneInfo, error)
}
```

### Key tmux Commands

| Operation | Command |
|-----------|---------|
| Check if in tmux | `[ -n "$TMUX" ]` |
| Create session | `tmux new-session -d -s codely` |
| Split horizontally | `tmux split-window -h -c <dir> -P -F "#{pane_id}" <cmd>` |
| Focus pane | `tmux select-pane -t %<id>` |
| Kill pane | `tmux kill-pane -t %<id>` |
| Resize pane | `tmux resize-pane -t %<id> -x <width>` |
| Capture content | `tmux capture-pane -t %<id> -p -S -<lines>` |
| List panes | `tmux list-panes -F "#{pane_id}:#{pane_current_command}"` |
| Send keys | `tmux send-keys -t %<id> '<command>' Enter` |

### Pane ID Format

tmux pane IDs are returned as `%N` where N is an integer. Store as int internally, format with `%` prefix for commands.

---

## shed Integration

### Client Interface

```go
type ShedClient interface {
    // Listing
    ListSheds() ([]Shed, error)
    ListSessions(shedName string) ([]Session, error)
    
    // Lifecycle
    CreateShed(name string, opts CreateOpts) error
    StartShed(name string) error
    StopShed(name string) error
    
    // Execution
    Exec(shedName string, command string, args ...string) *exec.Cmd
    Console(shedName string) *exec.Cmd
    Attach(shedName string, session string) *exec.Cmd
}

type CreateOpts struct {
    Repo   string
    Server string
    Image  string
}
```

### Key shed Commands

| Operation | Command |
|-----------|---------|
| List all sheds | `shed list --all --json` |
| List sessions | `shed sessions <name> --json` |
| Create shed | `shed create <name> --repo <repo> --server <server>` |
| Start shed | `shed start <name>` |
| Stop shed | `shed stop <name>` |
| Run command | `shed exec <name> <command>` |
| Open shell | `shed console <name>` |
| Attach to tmux | `shed attach <name> --session <session>` |

### Running Commands in Sheds

For AI tools in sheds, we use `shed exec`:

```bash
# Claude in shed
shed exec codelens claude --dangerously-skip-permissions

# OpenCode in shed  
shed exec codelens opencode

# Codex in shed
shed exec codelens codex

# Bash in shed (use console for interactive)
shed console codelens
```

The command is wrapped in tmux split-window:

```bash
tmux split-window -h "shed exec codelens claude --dangerously-skip-permissions"
```

---

## Keyboard Shortcuts

### Global

| Key | Action |
|-----|--------|
| `q` | Quit Codely |
| `?` | Show help |
| `r` | Refresh status / shed list |

### Project Tree View

| Key | Action |
|-----|--------|
| `j` / `↓` | Move selection down |
| `k` / `↑` | Move selection up |
| `Enter` | Focus session pane (if session selected) / Toggle expand (if project selected) |
| `Space` | Toggle expand/collapse (if project selected) |
| `←` | Collapse project / Move to parent |
| `→` | Expand project / Move to first child |
| `n` | New project (local, attach shed, or create shed) |
| `t` | Add terminal to selected project |
| `x` | Close selected session |
| `X` | Close selected project and all sessions |
| `s` | Stop shed (if shed project selected) |
| `S` | Start shed (if shed project selected and stopped) |

### Folder Picker View

| Key | Action |
|-----|--------|
| `j` / `↓` | Move selection down |
| `k` / `↑` | Move selection up |
| `Enter` | Select directory |
| `/` | Start fuzzy search |
| `Esc` | Cancel / Back |

### Command Picker View

| Key | Action |
|-----|--------|
| `j` / `↓` | Move selection down |
| `k` / `↑` | Move selection up |
| `Enter` | Launch with selected command |
| `Esc` | Cancel / Back |

### Shed Picker View

| Key | Action |
|-----|--------|
| `j` / `↓` | Move selection down |
| `k` / `↑` | Move selection up |
| `Enter` | Select shed |
| `s` | Start selected shed (if stopped) |
| `Esc` | Cancel / Back |

### Confirmation Dialogs

| Key | Action |
|-----|--------|
| `y` | Confirm action |
| `n` / `Esc` | Cancel action |

---

## Project Structure

```
codely/
├── cmd/
│   └── codely/
│       └── main.go                 # Entry point
├── internal/
│   ├── config/
│   │   ├── config.go               # Config loading/defaults
│   │   └── config_test.go
│   ├── tmux/
│   │   ├── client.go               # tmux command wrapper
│   │   ├── client_test.go
│   │   └── mock.go                 # Mock for testing
│   ├── shed/
│   │   ├── client.go               # shed CLI wrapper  
│   │   ├── client_test.go
│   │   └── mock.go
│   ├── status/
│   │   ├── detector.go             # Status detection logic
│   │   └── detector_test.go
│   ├── project/
│   │   ├── project.go              # Project type definitions
│   │   ├── session.go              # Session type definitions
│   │   ├── store.go                # Session state persistence
│   │   └── store_test.go
│   └── ui/
│       ├── model.go                # Main Bubble Tea model
│       ├── update.go               # Update logic
│       ├── view.go                 # View rendering
│       ├── components/
│       │   ├── tree.go             # Project/session tree component
│       │   ├── folder_picker.go    # Folder selection
│       │   ├── command_picker.go   # Command selection
│       │   ├── shed_picker.go      # Shed selection
│       │   ├── shed_create.go      # Shed creation form
│       │   ├── shed_close.go       # Shed close options dialog
│       │   └── confirm.go          # Confirmation dialog
│       ├── styles.go               # lipgloss styles
│       └── keys.go                 # Key bindings
├── configs/
│   └── default.yaml                # Default configuration
├── scripts/
│   └── install.sh                  # Installation script
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## Dependencies

### Go Modules

```go
require (
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/charmbracelet/bubbles v0.18.0
    github.com/charmbracelet/lipgloss v0.9.1
    github.com/spf13/viper v1.18.0          // Config management
    github.com/google/uuid v1.6.0           // Project IDs
    gopkg.in/yaml.v3 v3.0.1                 // YAML parsing
)
```

### External Tools

| Tool | Required | Purpose |
|------|----------|---------|
| tmux | Yes | Terminal multiplexing |
| shed | No | Remote container management |
| claude | No | Claude Code AI assistant |
| opencode | No | OpenCode AI assistant |
| codex | No | Codex AI assistant |
| lazygit | No | Git terminal UI |

---

## Error Handling

### Graceful Degradation

1. **shed not installed**: Hide shed-related options, show message in help
2. **Pane died unexpectedly**: Mark project as error, offer to restart
3. **Shed unreachable**: Show connection error, offer retry
4. **Config file missing**: Use defaults, create on first save

### Error Display

Errors are shown in a dismissible banner at the bottom of the screen:

```
┌─────────────────────────────────────────┐
│ Codely                                │
├─────────────────────────────────────────┤
│  ... project list ...                   │
│                                         │
├─────────────────────────────────────────┤
│ ⚠️  Failed to start shed: timeout       │
│ [Enter] dismiss  [r] retry              │
└─────────────────────────────────────────┘
```

---

## Future Considerations

### Potential Enhancements (Out of Scope for v1)

1. **Project templates**: Pre-configured command + directory combinations
2. **Session naming**: Custom names for shed tmux sessions
3. **Port forwarding UI**: Manage shed tunnels from Codely
4. **File sync status**: Show shed sync state
5. **Remote Codely**: Run Codely itself in a shed
6. **Notifications**: Desktop notifications when AI completes

### Extension Points

- Command plugins (add new AI tools via config)
- Custom status detectors per command
- Theming support

---

## Acceptance Criteria

### MVP Requirements

- [ ] Launch Codely in tmux (auto-create session if needed)
- [ ] Display hierarchical tree of projects and sessions with status indicators
- [ ] Create new local project (folder picker → command picker → launch)
- [ ] Add additional terminals to existing projects
- [ ] Navigate tree with keyboard (expand/collapse, focus sessions)
- [ ] Close individual sessions
- [ ] Close projects (closes all sessions)
- [ ] Status detection working for claude, opencode, codex, lazygit, bash
- [ ] **Projects persist across Codely restarts**
- [ ] **Sessions reconnect to existing tmux panes on restart**
- [ ] **Projects with zero sessions are preserved**
- [ ] Config file support for commands and workspace roots

### shed Integration Requirements

- [ ] List available sheds from all servers
- [ ] Attach to running shed (creates project + first session)
- [ ] Start stopped shed before attaching
- [ ] Create new shed with repo
- [ ] Add multiple terminals to shed projects
- [ ] Visual distinction between local and shed projects
- [ ] **Show stopped shed status at project level**
- [ ] **Close shed project dialog with three options:**
  - [ ] Close project only (shed keeps running)
  - [ ] Close and stop shed
  - [ ] Close and delete shed (with confirmation)

### Quality Requirements

- [ ] Responsive UI (< 100ms for all interactions)
- [ ] Status polling doesn't impact performance
- [ ] Clean shutdown (no zombie panes)
- [ ] Works on macOS and Linux
- [ ] Unit tests for status detection
- [ ] Integration tests for tmux/shed clients

---

## Appendix A: Default Configuration

```yaml
# ~/.config/codely/config.yaml

workspace_roots:
  - ~/work
  - ~/projects
  - ~/src

commands:
  claude:
    display_name: "Claude Code"
    exec: claude
    args:
      - "--dangerously-skip-permissions"
    
  opencode:
    display_name: "OpenCode"
    exec: opencode
    args: []
    
  codex:
    display_name: "Codex"
    exec: codex
    args: []
    
  lazygit:
    display_name: "Lazygit"
    exec: lazygit
    args: []
    
  bash:
    display_name: "Bash Shell"
    exec: bash
    args: []

default_command: claude

ui:
  manager_width: 30
  status_poll_interval: 1s
  show_directory: true
  auto_expand_projects: true

shed:
  enabled: true
  default_server: ""  # Uses shed's default
```

---

## Appendix B: Status Detection Test Cases

| Pane Content (last lines) | Expected Status |
|---------------------------|-----------------|
| `⠋ Thinking...` | thinking |
| `> ` | idle |
| `$ ` | idle |
| `claude> ` | idle |
| `Running tests...` | executing |
| `error: command not found` | executing (not error, just output) |
| `panic: runtime error` | error |
| (empty) | unknown |
| `│ (lazygit status panel)` | idle (lazygit is interactive) |
| `opencode> ` | idle |

---

## Appendix C: Command Line Interface

```
codely - AI Coding Session Manager

USAGE:
    codely [OPTIONS]

OPTIONS:
    -c, --config <PATH>    Config file path (default: ~/.config/codely/config.yaml)
    -v, --verbose          Enable debug logging
    -h, --help             Show this help message
    --version              Show version

EXAMPLES:
    codely               Start Codely
    codely -c ~/my.yaml  Start with custom config
```
