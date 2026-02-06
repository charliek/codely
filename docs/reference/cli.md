# CLI

Codely is a single-command binary. Running `codely` starts the TUI inside tmux.

## Usage

```bash
codely [flags]
```

## Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--config` | `-c` | `~/.config/codely/config.yaml` | Path to configuration file |
| `--debug` | `-d` | `false` | Enable debug logging to file |
| `--debug-file` | | `~/.local/state/codely/debug.log` | Debug log file path |
| `--version` | `-v` | | Print version and exit |
| `--help` | `-h` | | Print help and exit |

## Examples

```bash
# Start with default config
codely

# Use a custom config file
codely -c ~/my-config.yaml

# Enable debug logging
codely --debug

# Debug logging to a specific file
codely --debug --debug-file /tmp/codely-debug.log

# Print version
codely --version
```

## Behavior

If codely is launched outside of a tmux session, it creates a new tmux session named `codely` and attaches to it. If already inside tmux, it runs directly in the current session.

On startup, codely loads saved state from `~/.local/state/codely/session.json` and reconnects to any tmux panes that still exist.
