# Development Setup

## Prerequisites

- Go 1.24+
- tmux
- golangci-lint (for linting)

## Build

```bash
make build
```

This compiles the binary to `bin/codely`.

To build and install to `~/.local/bin`:

```bash
make install
```

## Test

```bash
make test
```

Runs all Go tests with verbose output.

## Lint

```bash
make lint
```

Requires [golangci-lint](https://golangci-lint.run/). The project configuration is in `.golangci.yml`.

## Project Structure

```
codely/
├── cmd/codely/            # Entry point
│   └── main.go
├── internal/
│   ├── cli/               # Cobra CLI setup, version
│   ├── config/            # YAML configuration loading
│   ├── constants/         # Default values
│   ├── debug/             # Debug logging
│   ├── domain/            # Core data structures (Project, Session, Command, Status)
│   ├── pathutil/          # Path expansion utilities
│   ├── shed/              # Remote container client
│   ├── status/            # Tool-aware status detection
│   ├── store/             # Session/project persistence
│   ├── tmux/              # tmux client
│   └── tui/               # Bubble Tea TUI
│       └── components/    # Tree, pickers, dialogs
├── docs/                  # Documentation (this site)
├── testdata/              # Test fixtures
├── .golangci.yml          # Linter configuration
├── .goreleaser.yaml       # Release configuration
├── Makefile               # Build targets
├── mkdocs.yml             # Documentation site config
└── go.mod
```

## Documentation Site

The documentation site uses [MkDocs](https://www.mkdocs.org/) with the [Material](https://squidfun.github.io/mkdocs-material/) theme. It requires Python and [uv](https://docs.astral.sh/uv/).

```bash
# Install dependencies
uv sync --group docs

# Serve locally
uv run mkdocs serve

# Build static site
uv run mkdocs build
```

The local dev server runs at `http://127.0.0.1:7070`.
