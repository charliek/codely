# TUI

Codely runs as a Bubble Tea terminal application inside tmux. The TUI occupies a narrow left pane and manages the remaining space for your coding sessions.

## Skins

Codely supports multiple UI skins that change how the manager panel renders projects and sessions. The tmux pane layout and all project/session management actions remain the same regardless of skin.

Select a skin with the `--skin` CLI flag or the `ui.skin` config option:

```bash
codely --skin tree   # hierarchical tree (default)
codely --skin flat   # flat card list
```

### Tree Skin (default)

The tree skin shows projects in a hierarchical view with expand/collapse navigation. Sessions are nested under their parent project.

```text
┌──────────────────────┐
│ LOCAL                │
│ ▼ my-project         │
│   ~/projects/my-proj │
│   ● claude   🤔      │
│   ○ bash     💤      │
│                      │
│ ▶ other-proj (1)     │
└──────────────────────┘
```

### Flat Skin

The flat skin shows projects as a scrollable list of cards. Each card displays the project name, path, session count, and per-session status.

```text
┌──────────────────────────┐
│ ╭────────────────────────╮│
│ │ my-project             ││
│ │ ~/projects/my-proj     ││
│ │ 2 sessions  ● 1 active ││
│ │ Claude Code 🤔  Bash 💤││
│ ╰────────────────────────╯│
│ ╭────────────────────────╮│
│ │ other-proj             ││
│ │ ~/work/other-proj      ││
│ │ 1 session              ││
│ │ Claude Code 💤         ││
│ ╰────────────────────────╯│
└──────────────────────────┘
```

Navigation in the flat skin uses up/down only (no expand/collapse). Left, right, and space are no-ops. Enter on a project toggles its expanded state; all other actions (new project, terminal, close) work the same.

### Selection States

The tree skin supports three selection states:

```text
# Project selected (expanded)
▼ codelens                    <-- SELECTED
  ~/projects/codelens
  ○ claude              💤
  ○ opencode            💤

# Session selected
▼ codelens
  ~/projects/codelens
  ● claude              🤔    <-- SELECTED
  ○ opencode            💤

# Project selected (collapsed)
▶ codelens (2 sessions)       <-- SELECTED
```

## Layout

```text
┌──────────────────────────────────────────────────────────────────────┐
│                            tmux session                              │
│                                                                      │
│  ┌─────────────────────┐  ┌────────────────────────────────────────┐ │
│  │ Codely TUI          │  │ Active Pane                            │ │
│  │                      │  │ (claude / opencode / bash / etc.)      │ │
│  │  (skin renders here) │  │                                        │ │
│  │                      │  │                                        │ │
│  │                      │  │                                        │ │
│  │                      │  │                                        │ │
│  │                      │  │                                        │ │
│  ├──────────────────────┤  │                                        │ │
│  │ [n]ew [t]erm [x]close│  │                                        │ │
│  └──────────────────────┘  └────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────────────┘
```

Use tmux zoom (`prefix` + `z`) to toggle fullscreen on the active pane.

### Views

#### New Project

```text
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
│                                         │
│  ~/projects/stridelabs/                 │
│    ○ audio/                             │
│    ○ codely/                            │
│                                         │
│  [/] search  [enter] select  [esc] back │
└─────────────────────────────────────────┘
```

#### Add Terminal

```text
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
│    claude --dangerously-skip-permissions │
│                                         │
│  ○ opencode                             │
│    opencode                             │
│                                         │
│  ○ bash                                 │
│    bash                                 │
│                                         │
│  [enter] launch  [esc] back             │
└─────────────────────────────────────────┘
```

#### Attach to Shed

```text
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
│                                         │
│  [enter] select  [s] start  [esc] back  │
└─────────────────────────────────────────┘
```

#### Create New Shed

```text
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
└─────────────────────────────────────────┘
```

#### Close Shed Project

```text
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
│                                         │
│  ○ Close and stop shed                  │
│    Can be restarted later.              │
│                                         │
│  ○ Close and DELETE shed                │
│    Permanently removes container.       │
│                                         │
│  [enter] confirm  [esc] cancel          │
└─────────────────────────────────────────┘
```

## Status Icons

| Icon | Status | Meaning |
|------|--------|---------|
| 💤 | idle | Shell prompt visible, waiting for input |
| 🤔 | thinking | AI is processing (spinner or thinking indicator) |
| ⚡ | executing | Running code or commands |
| ❌ | error | Process crashed or exited with error |
| ⏸️ | stopped | Shed container is stopped |

## Keybindings

### Global

| Key | Action |
|-----|--------|
| `q` / `Ctrl+c` | Quit codely |
| `?` | Toggle help overlay |
| `r` | Refresh status and shed list |

### Navigation (skin-specific)

These keys are handled by the active skin:

| Key | Tree Skin | Flat Skin |
|-----|-----------|-----------|
| `j` / `↓` | Move selection down | Move selection down |
| `k` / `↑` | Move selection up | Move selection up |
| `h` / `←` | Collapse project or move to parent | No-op |
| `l` / `→` | Expand project or move to first child | No-op |
| `Space` | Toggle project expand/collapse | No-op |

### Project Actions (all skins)

| Key | Action |
|-----|--------|
| `Enter` | Focus session pane (session) / toggle expand (project, tree skin only) |
| `n` | New project |
| `t` | Add terminal to selected project |
| `x` | Close selected session |
| `X` | Close selected project and all sessions |
| `s` | Stop shed (shed projects) |
| `S` | Start shed (stopped shed projects) |

### Folder Picker

| Key | Action |
|-----|--------|
| `j` / `↓` | Move selection down |
| `k` / `↑` | Move selection up |
| `/` | Start search |
| `Enter` | Select directory |
| `Esc` | Cancel |

### Command Picker

| Key | Action |
|-----|--------|
| `j` / `↓` | Move selection down |
| `k` / `↑` | Move selection up |
| `Enter` | Launch with selected command |
| `Esc` | Cancel |

### Confirmation Dialogs

| Key | Action |
|-----|--------|
| `y` | Confirm |
| `n` / `Esc` | Cancel |

## tmux Notifications

Codely updates the tmux status bar with a segment showing sessions that need attention:

```text
Codely: [1] api/claude [2] web/opencode ! db/codex
```

Sessions in `waiting` or `error` state appear in the status line. The `!` prefix indicates an error. While codely is running, `prefix+1..6` jumps to the corresponding pane.
