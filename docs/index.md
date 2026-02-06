# codely

A terminal-based project manager for orchestrating AI coding sessions across local directories and remote development containers.

Codely runs inside tmux and provides a unified interface for launching, monitoring, and switching between multiple concurrent coding sessions.

- **Session management** — run Claude Code, OpenCode, Codex, lazygit, or shell sessions side by side
- **Project tree** — organize sessions under projects with expand/collapse navigation
- **Status monitoring** — real-time detection of idle, thinking, executing, and error states
- **tmux integration** — splits panes automatically, tracks pane lifecycle, reconnects on restart
- **Persistence** — projects and sessions survive codely restarts; dead panes are cleaned up automatically

## Quick Example

```yaml
# ~/.config/codely/config.yaml
workspace_roots:
  - ~/work
  - ~/projects

commands:
  claude:
    display_name: Claude Code
    exec: claude
    args: ["--dangerously-skip-permissions"]

default_command: claude
```

```bash
codely
```

This launches the TUI inside tmux. From there, press `n` to create a project and `t` to add terminal sessions.

## Current Scope

Codely manages local projects with full tmux integration. Remote container support via [shed](https://github.com/charliek/shed) is a planned enhancement and not yet stable.

## Next Steps

- [Quick Start](getting-started/quick-start.md) — install and run your first session
- [CLI Reference](reference/cli.md) — command-line options
- [Configuration](reference/configuration.md) — config file schema
- [TUI Reference](reference/tui.md) — keybindings and status icons
- [Development Setup](development/setup.md) — build from source
