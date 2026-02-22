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
See `docs/status-detection.md` for status values, auto-detection, and overrides.

## Development

### Prerequisites

This project uses [mise](https://mise.jdx.dev/) to manage tool versions. With mise installed, all dependencies are set up automatically:

```bash
mise install
```

This installs the correct versions of Go and golangci-lint as defined in `.mise.toml`.

Alternatively, install manually:
- Go 1.24+
- golangci-lint v2 (`brew install golangci-lint` on macOS, or see [install docs](https://golangci-lint.run/docs/welcome/install/))

```bash
make build    # Build the binary
make test     # Run tests
make lint     # Run linters
make clean    # Remove build artifacts
```

## Requirements

- tmux (required)
- shed CLI (optional, for remote container support)
