# TUI

Codely runs as a Bubble Tea terminal application inside tmux. The TUI occupies a narrow left pane and manages the remaining space for your coding sessions.

## Layout

```
┌──────────────────────────────────────────────────────────────────────┐
│                            tmux session                              │
│                                                                      │
│  ┌─────────────────────┐  ┌────────────────────────────────────────┐ │
│  │ Codely TUI          │  │ Active Pane                            │ │
│  │                      │  │ (claude / opencode / bash / etc.)      │ │
│  │ LOCAL                │  │                                        │ │
│  │ ▼ my-project         │  │                                        │ │
│  │   ~/projects/my-proj │  │                                        │ │
│  │   ● claude   🤔      │  │                                        │ │
│  │   ○ bash     💤      │  │                                        │ │
│  │                      │  │                                        │ │
│  │ ▶ other-proj (1)     │  │                                        │ │
│  │                      │  │                                        │ │
│  ├──────────────────────┤  │                                        │ │
│  │ [n]ew [t]erm [x]close│  │                                        │ │
│  └──────────────────────┘  └────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────────────┘
```

Use tmux zoom (`prefix` + `z`) to toggle fullscreen on the active pane.

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

### Project Tree

| Key | Action |
|-----|--------|
| `j` / `↓` | Move selection down |
| `k` / `↑` | Move selection up |
| `h` / `←` | Collapse project or move to parent |
| `l` / `→` | Expand project or move to first child |
| `Enter` | Focus session pane (session) / toggle expand (project) |
| `Space` | Toggle project expand/collapse |
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

```
Codely: [1] api/claude [2] web/opencode ! db/codex
```

Sessions in `waiting` or `error` state appear in the status line. The `!` prefix indicates an error. While codely is running, `prefix+1..6` jumps to the corresponding pane.
