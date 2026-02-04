package domain

// Status represents the current state of a session
type Status string

const (
	// StatusIdle indicates the session is waiting for user input (prompt visible)
	StatusIdle Status = "idle"

	// StatusThinking indicates the AI is processing (spinner visible)
	StatusThinking Status = "thinking"

	// StatusExecuting indicates the session is running code or commands
	StatusExecuting Status = "executing"

	// StatusError indicates the process crashed or exited with error
	StatusError Status = "error"

	// StatusStopped indicates the shed is not running
	StatusStopped Status = "stopped"

	// StatusUnknown indicates the status cannot be determined
	StatusUnknown Status = "unknown"
)

// String returns the string representation of the status
func (s Status) String() string {
	return string(s)
}

// Icon returns the emoji icon for the status
func (s Status) Icon() string {
	switch s {
	case StatusIdle:
		return "💤"
	case StatusThinking:
		return "🤔"
	case StatusExecuting:
		return "⚡"
	case StatusError:
		return "❌"
	case StatusStopped:
		return "⏸️"
	default:
		return "❓"
	}
}
