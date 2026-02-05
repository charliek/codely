// Package status provides detection of session status from pane content.
package status

import (
	"path/filepath"
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

// Prompt patterns that indicate idle state (generic/shell)
var promptPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?m)^[>$#%]\s*$`),               // Common shell prompts: >, $, #, %
	regexp.MustCompile(`(?m)^claude[>:]\s*$`),           // Claude prompt (generic)
	regexp.MustCompile(`(?m)^opencode[>:]\s*$`),         // OpenCode prompt (generic)
	regexp.MustCompile(`(?m)^\(.*\)[>$]\s*$`),           // Virtualenv prompts: (venv)$
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
	"│", // Box drawing character used in TUI
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

// DetectWithMode dispatches to tool-aware detection based on mode and command identity.
// mode: auto|generic|claude|opencode|codex|shell (empty treated as auto)
func DetectWithMode(content, commandID, commandExec, mode string) domain.Status {
	mode = strings.ToLower(strings.TrimSpace(mode))
	switch mode {
	case "", "auto":
		tool := normalizeTool(commandID)
		if tool == "" {
			tool = normalizeTool(filepath.Base(commandExec))
		}
		if tool == "" {
			return Detect(content)
		}
		return detectTool(content, tool)
	case "generic":
		return Detect(content)
	default:
		return detectTool(content, mode)
	}
}

// Detect analyzes pane content and returns the detected status using generic heuristics.
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

func detectTool(content, tool string) domain.Status {
	if content == "" || strings.TrimSpace(content) == "" {
		return domain.StatusUnknown
	}

	switch tool {
	case "claude":
		return detectClaude(content)
	case "opencode":
		return detectOpenCode(content)
	case "codex":
		return detectCodex(content)
	case "shell":
		return detectShell(content)
	default:
		return Detect(content)
	}
}

func normalizeTool(value string) string {
	v := strings.ToLower(strings.TrimSpace(value))
	switch v {
	case "claude", "opencode", "codex", "shell":
		return v
	default:
		return ""
	}
}

func detectClaude(content string) domain.Status {
	lines := strings.Split(content, "\n")
	recent := getLastNonEmptyLines(lines, 15)
	recentContent := strings.Join(recent, "\n")
	recentLower := strings.ToLower(recentContent)

	if claudeBusy(recent, recentLower) {
		return domain.StatusThinking
	}
	if claudePrompt(recent, recentLower) {
		return domain.StatusWaiting
	}
	if containsFatalError(recent) {
		return domain.StatusError
	}
	return domain.StatusExecuting
}

func detectOpenCode(content string) domain.Status {
	lines := strings.Split(content, "\n")
	recent := getLastNonEmptyLines(lines, 15)
	recentContent := strings.Join(recent, "\n")

	if openCodeBusy(recentContent) {
		return domain.StatusThinking
	}
	if openCodePrompt(recentContent) {
		return domain.StatusWaiting
	}
	if containsFatalError(recent) {
		return domain.StatusError
	}
	return domain.StatusExecuting
}

func detectCodex(content string) domain.Status {
	lines := strings.Split(content, "\n")
	recent := getLastNonEmptyLines(lines, 15)
	recentContent := strings.Join(recent, "\n")

	if codexPrompt(recentContent) {
		return domain.StatusWaiting
	}
	if containsFatalError(recent) {
		return domain.StatusError
	}
	return domain.StatusExecuting
}

func detectShell(content string) domain.Status {
	lines := strings.Split(content, "\n")
	recent := getLastNonEmptyLines(lines, 10)

	if endsWithPrompt(recent) {
		return domain.StatusIdle
	}
	if containsFatalError(recent) {
		return domain.StatusError
	}
	return domain.StatusExecuting
}

func claudeBusy(lines []string, recentLower string) bool {
	busyIndicators := []string{
		"ctrl+c to interrupt",
		"esc to interrupt",
	}
	for _, indicator := range busyIndicators {
		if strings.Contains(recentLower, indicator) {
			return true
		}
	}

	// Spinner characters (braille + asterisk variants)
	spinnerChars := []string{
		"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏",
		"✳", "✽", "✶", "✢",
	}
	checkLines := lines
	if len(checkLines) > 10 {
		checkLines = checkLines[len(checkLines)-10:]
	}
	for _, line := range checkLines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			r := []rune(trimmed)[0]
			if r == '│' || r == '├' || r == '└' || r == '─' || r == '┌' || r == '┐' || r == '┘' || r == '┤' || r == '┬' || r == '┴' || r == '┼' || r == '╭' || r == '╰' || r == '╮' || r == '╯' {
				continue
			}
		}
		for _, spinner := range spinnerChars {
			if strings.Contains(line, spinner) {
				return true
			}
		}
	}

	// Ellipsis + tokens is strong indicator of processing
	if strings.Contains(recentLower, "…") && strings.Contains(recentLower, "tokens") {
		return true
	}
	if strings.Contains(recentLower, "thinking") && strings.Contains(recentLower, "tokens") {
		return true
	}
	if strings.Contains(recentLower, "connecting") && strings.Contains(recentLower, "tokens") {
		return true
	}

	return false
}

func claudePrompt(lines []string, recentLower string) bool {
	recentContent := strings.Join(lines, "\n")

	permissionPrompts := []string{
		"No, and tell Claude what to do differently",
		"Yes, allow once",
		"Yes, allow always",
		"Allow once",
		"Allow always",
		"│ Do you want",
		"│ Would you like",
		"│ Allow",
		"❯ Yes",
		"❯ No",
		"❯ Allow",
		"Do you trust the files in this folder?",
		"Allow this MCP server",
		"Run this command?",
		"Execute this?",
		"Action Required",
		"Waiting for user confirmation",
		"Allow execution of",
		"Use arrow keys to navigate",
		"Press Enter to select",
	}
	for _, prompt := range permissionPrompts {
		if strings.Contains(recentContent, prompt) {
			return true
		}
	}

	// Last non-empty line prompt detection
	if len(lines) > 0 {
		lastLine := strings.TrimSpace(lines[len(lines)-1])
		clean := strings.TrimSpace(StripANSI(lastLine))
		clean = strings.ReplaceAll(clean, "\u00a0", " ")
		if clean == ">" || clean == "❯" || clean == "> " || clean == "❯ " {
			return true
		}
		if (strings.HasPrefix(clean, "> ") || strings.HasPrefix(clean, "❯ ")) && !strings.Contains(clean, "esc") {
			if len(clean) < 100 {
				return true
			}
		}
	}

	checkLines := lines
	if len(checkLines) > 5 {
		checkLines = checkLines[len(checkLines)-5:]
	}
	for _, line := range checkLines {
		cleanLine := strings.TrimSpace(StripANSI(line))
		cleanLine = strings.ReplaceAll(cleanLine, "\u00a0", " ")
		if cleanLine == ">" || cleanLine == "❯" || cleanLine == "> " || cleanLine == "❯ " {
			return true
		}
		if strings.HasPrefix(cleanLine, "❯ Try ") || strings.HasPrefix(cleanLine, "> Try ") {
			return true
		}
	}

	questionPrompts := []string{
		"Continue?",
		"Proceed?",
		"(Y/n)",
		"(y/N)",
		"[Y/n]",
		"[y/N]",
		"(yes/no)",
		"[yes/no]",
		"Approve this plan?",
		"Execute plan?",
	}
	for _, prompt := range questionPrompts {
		if strings.Contains(recentContent, prompt) {
			return true
		}
	}

	completionIndicators := []string{
		"Task completed",
		"Done!",
		"Finished",
		"What would you like",
		"What else",
		"Anything else",
		"Let me know if",
	}
	hasCompletion := false
	for _, indicator := range completionIndicators {
		if strings.Contains(recentLower, strings.ToLower(indicator)) {
			hasCompletion = true
			break
		}
	}
	if hasCompletion {
		completionLines := lines
		if len(completionLines) > 3 {
			completionLines = completionLines[len(completionLines)-3:]
		}
		for _, line := range completionLines {
			cleanLine := strings.TrimSpace(StripANSI(line))
			if cleanLine == ">" || cleanLine == "> " || cleanLine == "❯" || cleanLine == "❯ " {
				return true
			}
		}
	}

	return false
}

func openCodeBusy(content string) bool {
	if strings.Contains(content, "esc interrupt") || strings.Contains(content, "esc to exit") {
		return true
	}
	pulseChars := []string{"█", "▓", "▒", "░"}
	for _, ch := range pulseChars {
		if strings.Contains(content, ch) {
			return true
		}
	}
	busyStrings := []string{
		"Thinking...",
		"Generating...",
		"Building tool call...",
		"Waiting for tool response...",
		"Loading...",
	}
	for _, s := range busyStrings {
		if strings.Contains(content, s) {
			return true
		}
	}
	return false
}

func openCodePrompt(content string) bool {
	if strings.Contains(content, "press enter to send") ||
		strings.Contains(content, "Ask anything") ||
		strings.Contains(content, "open code") {
		return true
	}
	return hasLineEndingWith(content, ">")
}

func codexPrompt(content string) bool {
	return strings.Contains(content, "codex>") ||
		strings.Contains(content, "Continue?") ||
		hasLineEndingWith(content, ">")
}

func hasLineEndingWith(content string, suffix string) bool {
	lines := strings.Split(content, "\n")
	start := len(lines) - 5
	if start < 0 {
		start = 0
	}
	for i := len(lines) - 1; i >= start; i-- {
		line := strings.TrimSpace(lines[i])
		if line == suffix || strings.HasSuffix(line+" ", suffix+" ") {
			return true
		}
	}
	return false
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

// StripANSI removes ANSI escape codes from content.
func StripANSI(content string) string {
	if strings.IndexByte(content, '\x1b') < 0 && strings.IndexByte(content, '\x9b') < 0 {
		return content
	}

	var b strings.Builder
	b.Grow(len(content))

	i := 0
	for i < len(content) {
		if content[i] == '\x1b' {
			if i+1 < len(content) && content[i+1] == '[' {
				j := i + 2
				for j < len(content) {
					c := content[j]
					if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
						j++
						break
					}
					j++
				}
				i = j
				continue
			}
			if i+1 < len(content) && content[i+1] == ']' {
				bellPos := strings.Index(content[i:], "\x07")
				if bellPos != -1 {
					i += bellPos + 1
					continue
				}
				stPos := strings.Index(content[i:], "\x1b\\")
				if stPos != -1 {
					i += stPos + 2
					continue
				}
			}
			if i+1 < len(content) {
				i += 2
				continue
			}
		}
		if content[i] == '\x9b' {
			j := i + 1
			for j < len(content) {
				c := content[j]
				if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
					j++
					break
				}
				j++
			}
			i = j
			continue
		}
		b.WriteByte(content[i])
		i++
	}

	return b.String()
}
