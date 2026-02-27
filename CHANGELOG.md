# Changelog

## v0.0.4

- Add dynamic server selection to shed creation form
- Add streaming output and Create button to shed creation UX
- Add backend selection to shed creation form
- Add robustness fixes for shed creation flow
- Fix shed terminal creation by passing exec args separately
- Migrate shed client to structured JSON protocol
- Upgrade golangci-lint to v2 and manage via mise

## v0.0.3

- Reduce default manager panel width from 38 to 30
- Remove status text labels from terminal list, show only icons

## v0.0.2

- Fix select-pane failure after terminal exits
- Remove help hint and tmux tip from footer status bar

## v0.0.1

- Initial Codely TUI with tmux pane management
- Folder picker with search navigation
- Track exit codes and keep crash panes visible
- Show exited sessions and skip close confirmation
- Tool-aware status detection and tmux notifications
- Fix pane width handling and support pane ID 0
- Add CI/CD, release automation, and documentation site
