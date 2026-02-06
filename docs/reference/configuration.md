# Configuration

Codely reads its configuration from a YAML file. If the file does not exist, built-in defaults are used.

## File Location

```
~/.config/codely/config.yaml
```

Override with `--config` / `-c` flag.

## Full Example

```yaml
workspace_roots:
  - ~/work
  - ~/projects
  - ~/src

commands:
  claude:
    display_name: Claude Code
    exec: claude
    args: ["--dangerously-skip-permissions"]
    status_detection: auto

  opencode:
    display_name: OpenCode
    exec: opencode

  codex:
    display_name: Codex
    exec: codex

  lazygit:
    display_name: Lazygit
    exec: lazygit

  bash:
    display_name: Bash Shell
    exec: bash

default_command: claude

ui:
  manager_width: 38
  status_poll_interval: 1s
  show_directory: true
  auto_expand_projects: true

shed:
  enabled: true
  default_server: ""
```

## Top-Level Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `workspace_roots` | list of strings | `~/work`, `~/projects`, `~/src` | Directories shown in the folder picker |
| `commands` | map | See below | Available commands for terminal sessions |
| `default_command` | string | `claude` | Command pre-selected when adding a terminal |

## Command Fields

Each entry under `commands` is keyed by an ID (e.g., `claude`, `bash`).

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `display_name` | string | yes | Name shown in the command picker |
| `exec` | string | yes | Binary to execute |
| `args` | list of strings | no | Arguments passed to the binary |
| `env` | map of strings | no | Environment variables set for the process |
| `status_detection` | string | no | Detection mode: `auto`, `generic`, `claude`, `opencode`, `codex`, `shell` |

When `status_detection` is `auto` (default), codely selects a detector based on the command ID and exec binary name, falling back to the generic heuristic.

## UI Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `manager_width` | int | `38` | Width of the left TUI panel in characters |
| `status_poll_interval` | duration | `1s` | How often to check pane status |
| `show_directory` | bool | `false` | Show full path in project list |
| `auto_expand_projects` | bool | `false` | Expand projects by default in the tree |

## Shed Fields

Remote container support via shed is a planned enhancement. These fields configure the integration when available.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `true` | Enable shed integration |
| `default_server` | string | `""` | Default shed server name |

## Session State

Active projects and sessions are stored separately from the config:

```
~/.local/state/codely/session.json
```

This file tracks which projects exist, their sessions, and associated tmux pane IDs. It is managed automatically by codely.
