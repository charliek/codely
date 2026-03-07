# Status Detection

Codely detects session status by inspecting recent tmux pane output. Detection is tool-aware for common AI CLIs and falls back to a generic heuristic.

## Status Values

- `thinking`: Busy indicator visible (spinners, tool "busy" banners, Claude timing lines).
- `executing`: Output changing or unknown activity with no explicit busy indicator.
- `waiting`: Prompt or confirmation UI visible (permission dialogs, `>`/`❯` prompts).
- `idle`: Shell prompt or known idle UI (like lazygit) visible, not waiting for input.
- `error`: Crash/panic indicators or exit with error.
- `exited`: tmux pane no longer exists.
- `stopped`: shed project is stopped (remote only).

## Auto Detection

By default, Codely uses **auto detection**:

1. Check `command.ID` for a known tool (`claude`, `opencode`, `codex`, `shell`).
2. If not found, check the exec basename (e.g. `claude`, `codex`).
3. If still unknown, fall back to the generic heuristic.

## Config Override

You can force detection for a command via `status_detection`:

```yaml
commands:
  claude:
    display_name: Claude Code
    exec: claude
    args: ["--dangerously-skip-permissions"]
    status_detection: auto   # auto | generic | claude | opencode | codex | shell
```

### Values

- `auto`: Use command ID + exec basename, then generic fallback.
- `generic`: Use the legacy generic detector only.
- `claude|opencode|codex|shell`: Force that detector regardless of ID/exec.

### Example: Alias with override

```yaml
commands:
  my-claude:
    display_name: Claude Code (alias)
    exec: claude
    args: ["--dangerously-skip-permissions"]
    status_detection: claude
```

## Detection Patterns

The status detector checks the last several lines of terminal output. Each tool-specific detector applies its own patterns, but the generic detector uses the following heuristics:

**Spinner characters (thinking):**

```go
var spinnerChars = []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}
```

**Thinking text patterns:**

```go
var thinkingPatterns = []string{
    "thinking",
    "analyzing",
    "reading",
    "processing",
    "Generating",
}
```

**Prompt patterns (idle/waiting):**

```go
var promptPatterns = []*regexp.Regexp{
    regexp.MustCompile(`(?m)^[>$#%] ?$`),           // Common shell prompts
    regexp.MustCompile(`(?m)^claude[>:] ?$`),       // Claude prompt
    regexp.MustCompile(`(?m)^\(.*\)[>$] ?$`),       // Virtualenv prompts
}
```

**Error patterns:**

```go
var errorPatterns = []string{
    "error:",
    "Error:",
    "ERROR",
    "panic:",
    "Traceback",
    "Exception:",
}
```

## Test Cases

| Pane Content (last lines) | Expected Status |
|---------------------------|-----------------|
| `⠋ Thinking...` | thinking |
| `> ` | idle |
| `$ ` | idle |
| `claude> ` | idle |
| `Running tests...` | executing |
| `error: command not found` | executing (output, not a crash) |
| `panic: runtime error` | error |
| (empty) | unknown |

## tmux Notifications

Codely updates tmux `status-right` with a segment like:

```text
Codely: [1] api/claude [2] web/opencode ! db/codex
```

- Shows sessions in `waiting` or `error`.
- `!` indicates error.
- `prefix+1..6` jumps to the corresponding pane while Codely is running.
- The keybinding overrides tmux default window selection during runtime and is restored on exit.
