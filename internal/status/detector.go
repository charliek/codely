// Package status provides detection of session status from pane content.
package status

import (
	"regexp"
	"strings"

	"github.com/charliek/codely/internal/domain"
)

// Spinner characters used by various CLI tools
var spinnerChars = []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}

// Thinking text patterns (case-insensitive)
var thinkingPatterns = []string{
	"thinking",
	"analyzing",
	"reading",
	"processing",
	"generating",
	"working",
	"loading",
}

// Prompt patterns that indicate idle state
var promptPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?m)^[>$#%]\s*$`),         // Common shell prompts: >, $, #, %
	regexp.MustCompile(`(?m)^claude[>:]\s*$`),     // Claude prompt
	regexp.MustCompile(`(?m)^opencode[>:]\s*$`),   // OpenCode prompt
	regexp.MustCompile(`(?m)^\(.*\)[>$]\s*$`),     // Virtualenv prompts: (venv)$
	regexp.MustCompile(`(?m)^\w+@[\w-]+[:#].*[$#]\s*$`), // SSH/system prompts
}

// Error patterns that indicate error state (only for crashes/panics)
var errorPatterns = []string{
	"panic:",
	"Traceback",
	"Segmentation fault",
	"core dumped",
	"fatal error:",
	"FATAL:",
}

// Lazygit detection - if we see lazygit UI elements, it's idle (interactive)
var lazygitPatterns = []string{
	"│",  // Box drawing character used in TUI
	"┌",
	"└",
	"┐",
	"┘",
	"─",
	"Status",
	"Files",
	"Branches",
	"Commits",
}

// Detect analyzes pane content and returns the detected status
func Detect(content string) domain.Status {
	if content == "" || strings.TrimSpace(content) == "" {
		return domain.StatusUnknown
	}

	lines := strings.Split(content, "\n")
	recent := getLastNonEmptyLines(lines, 15)

	// Check for spinner characters (thinking)
	if containsSpinner(recent) {
		return domain.StatusThinking
	}

	// Check for explicit thinking indicators
	if containsThinkingText(recent) {
		return domain.StatusThinking
	}

	// Check for lazygit (interactive TUI = idle)
	if isLazygitUI(recent) {
		return domain.StatusIdle
	}

	// Check for prompt (idle state)
	if endsWithPrompt(recent) {
		return domain.StatusIdle
	}

	// Check for fatal errors/crashes
	if containsFatalError(recent) {
		return domain.StatusError
	}

	// Default: probably executing
	return domain.StatusExecuting
}

// getLastNonEmptyLines returns the last N non-empty lines
func getLastNonEmptyLines(lines []string, n int) []string {
	var result []string
	for i := len(lines) - 1; i >= 0 && len(result) < n; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" {
			result = append([]string{line}, result...)
		}
	}
	return result
}

// containsSpinner checks if any line contains spinner characters
func containsSpinner(lines []string) bool {
	for _, line := range lines {
		for _, r := range line {
			for _, s := range spinnerChars {
				if r == s {
					return true
				}
			}
		}
	}
	return false
}

// containsThinkingText checks for thinking indicator text
func containsThinkingText(lines []string) bool {
	for _, line := range lines {
		lower := strings.ToLower(line)
		for _, pattern := range thinkingPatterns {
			// Must be followed by "..." or start with the pattern (status message)
			if strings.Contains(lower, pattern+"...") ||
				strings.HasPrefix(lower, pattern+" ") ||
				strings.HasPrefix(lower, pattern) {
				return true
			}
		}
	}
	return false
}

// endsWithPrompt checks if the content ends with a command prompt
func endsWithPrompt(lines []string) bool {
	if len(lines) == 0 {
		return false
	}

	// Check last few lines for prompts
	checkLines := lines
	if len(checkLines) > 3 {
		checkLines = checkLines[len(checkLines)-3:]
	}

	for _, line := range checkLines {
		for _, pattern := range promptPatterns {
			if pattern.MatchString(line) {
				return true
			}
		}
	}
	return false
}

// containsFatalError checks for crash/panic indicators
func containsFatalError(lines []string) bool {
	for _, line := range lines {
		for _, pattern := range errorPatterns {
			if strings.Contains(line, pattern) {
				return true
			}
		}
	}
	return false
}

// isLazygitUI checks if the content looks like lazygit TUI
func isLazygitUI(lines []string) bool {
	boxDrawingCount := 0
	hasLazygitKeyword := false

	for _, line := range lines {
		// Count box drawing characters
		for _, pattern := range lazygitPatterns[:6] { // Box drawing chars
			if strings.Contains(line, pattern) {
				boxDrawingCount++
				break
			}
		}

		// Check for lazygit-specific keywords
		for _, pattern := range lazygitPatterns[6:] { // Keywords
			if strings.Contains(line, pattern) {
				hasLazygitKeyword = true
				break
			}
		}
	}

	// If we have many box drawing characters and lazygit keywords, it's likely lazygit
	return boxDrawingCount >= 3 && hasLazygitKeyword
}
