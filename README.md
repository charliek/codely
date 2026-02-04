# Codely

AI Coding Session Manager - A terminal-based project manager for orchestrating AI coding sessions across local directories and remote development containers.

## Features

- Single pane of glass for all active AI coding sessions
- Seamless switching between local and remote (shed) environments
- Real-time status monitoring (idle, thinking, executing)
- Quick project provisioning via shed integration

## Installation

```bash
# Build from source
make build

# Install to ~/.local/bin
make install
```

## Usage

```bash
# Start codely with default config
codely

# Use a custom config file
codely -c ~/my-config.yaml

# Show version
codely --version

# Enable verbose output
codely -v
```

Tip: Use tmux zoom (`prefix` + `z`) to toggle fullscreen for the active pane.

## Configuration

Configuration file location: `~/.config/codely/config.yaml`

See `docs/spec.md` for the full configuration schema.

## Development

```bash
# Build
make build

# Run tests
make test

# Run linter
make lint

# Clean build artifacts
make clean
```

## Requirements

- Go 1.24+
- tmux (required)
- shed CLI (optional, for remote container support)
