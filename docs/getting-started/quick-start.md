# Quick Start

## Installation

### From Source

```bash
git clone https://github.com/charliek/codely.git
cd codely
make build
make install
```

This installs the `codely` binary to `~/.local/bin/`. Make sure this directory is in your `PATH`.

### Prerequisites

- **Go 1.24+** — required to build from source
- **tmux** — required at runtime for pane management

## First Launch

Start codely from any terminal:

```bash
codely
```

If you're not already inside a tmux session, codely creates one and attaches to it. The TUI appears in a narrow left pane.

## Create a Project

1. Press `n` to open the new project dialog
2. Select **Local Directory**
3. Browse your workspace roots and pick a directory
4. Select a command to launch (e.g., Claude Code)

Codely splits a new tmux pane, starts your command in the selected directory, and focuses it.

## Add More Sessions

With a project selected in the tree, press `t` to add another terminal session. Each session runs independently in its own tmux pane.

## Navigate the Tree

| Key | Action |
|-----|--------|
| `j` / `k` | Move selection up/down |
| `Enter` | Focus session pane or toggle project expand |
| `Space` | Toggle project expand/collapse |
| `x` | Close selected session |
| `q` | Quit codely |

Use tmux zoom (`prefix` + `z`) to toggle fullscreen on the active pane.

## What Happens on Restart

Projects persist across codely restarts. When you relaunch codely, it reconnects to any tmux panes that are still alive and cleans up dead sessions. Projects with zero sessions are kept so you can quickly add new terminals.

## Next Steps

- [Configuration](../reference/configuration.md) — customize workspace roots, commands, and UI settings
- [TUI Reference](../reference/tui.md) — full keybinding reference and status icons
