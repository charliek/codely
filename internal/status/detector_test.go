package status

import (
	"testing"

	"github.com/charliek/codely/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected domain.Status
	}{
		{
			name:     "empty content",
			content:  "",
			expected: domain.StatusUnknown,
		},
		{
			name:     "whitespace only",
			content:  "   \n\n  \t  ",
			expected: domain.StatusUnknown,
		},
		{
			name:     "spinner thinking",
			content:  "⠋ Thinking...",
			expected: domain.StatusThinking,
		},
		{
			name:     "thinking text",
			content:  "Analyzing code...\nProcessing...",
			expected: domain.StatusThinking,
		},
		{
			name:     "shell prompt $",
			content:  "echo hello\nhello\n$ ",
			expected: domain.StatusIdle,
		},
		{
			name:     "shell prompt >",
			content:  "some output\n>",
			expected: domain.StatusIdle,
		},
		{
			name:     "claude prompt",
			content:  "Done editing file.\nclaude>",
			expected: domain.StatusIdle,
		},
		{
			name:     "claude prompt with colon",
			content:  "Done editing file.\nclaude:",
			expected: domain.StatusIdle,
		},
		{
			name:     "opencode prompt",
			content:  "Completed.\nopencode>",
			expected: domain.StatusIdle,
		},
		{
			name:     "virtualenv prompt",
			content:  "pip install done\n(venv)$",
			expected: domain.StatusIdle,
		},
		{
			name:     "running tests - executing",
			content:  "Running tests...\n=== RUN   TestFoo",
			expected: domain.StatusExecuting,
		},
		{
			name:     "command not found is executing not error",
			content:  "error: command not found\nbash: foo: not found",
			expected: domain.StatusExecuting,
		},
		{
			name:     "panic is error",
			content:  "panic: runtime error: invalid memory address",
			expected: domain.StatusError,
		},
		{
			name:     "traceback is error",
			content:  "Traceback (most recent call last):\n  File \"foo.py\"",
			expected: domain.StatusError,
		},
		{
			name:     "segfault is error",
			content:  "Segmentation fault (core dumped)",
			expected: domain.StatusError,
		},
		{
			name:     "lazygit UI is idle",
			content:  "┌───────────────────┐\n│ Status │ Files │ Branches │\n└───────────────────┘",
			expected: domain.StatusIdle,
		},
		{
			name:     "SSH prompt",
			content:  "user@hostname:~$",
			expected: domain.StatusIdle,
		},
		{
			name:     "generating text is thinking",
			content:  "Generating response...",
			expected: domain.StatusThinking,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Detect(tt.content)
			assert.Equal(t, tt.expected, got, "content: %q", tt.content)
		})
	}
}

func TestContainsSpinner(t *testing.T) {
	tests := []struct {
		lines    []string
		expected bool
	}{
		{[]string{"⠋ Loading..."}, true},
		{[]string{"⠙ Processing"}, true},
		{[]string{"Loading..."}, false},
		{[]string{"No spinner here"}, false},
	}

	for _, tt := range tests {
		got := containsSpinner(tt.lines)
		assert.Equal(t, tt.expected, got, "lines: %v", tt.lines)
	}
}

func TestEndsWithPrompt(t *testing.T) {
	tests := []struct {
		lines    []string
		expected bool
	}{
		{[]string{"$ "}, true},
		{[]string{"> "}, true},
		{[]string{"# "}, true},
		{[]string{"claude>"}, true},
		{[]string{"opencode:"}, true},
		{[]string{"(myenv)$"}, true},
		{[]string{"some text", "more text", "$"}, true},
		{[]string{"Running..."}, false},
		{[]string{}, false},
	}

	for _, tt := range tests {
		got := endsWithPrompt(tt.lines)
		assert.Equal(t, tt.expected, got, "lines: %v", tt.lines)
	}
}

func TestContainsThinkingText(t *testing.T) {
	tests := []struct {
		lines    []string
		expected bool
	}{
		{[]string{"thinking..."}, true},
		{[]string{"Analyzing code..."}, true},
		{[]string{"Processing files"}, true},
		{[]string{"generating"}, true},
		{[]string{"done thinking"}, false}, // "thinking" not at end or followed by ...
		{[]string{"no indicator"}, false},
	}

	for _, tt := range tests {
		got := containsThinkingText(tt.lines)
		assert.Equal(t, tt.expected, got, "lines: %v", tt.lines)
	}
}

func TestContainsFatalError(t *testing.T) {
	tests := []struct {
		lines    []string
		expected bool
	}{
		{[]string{"panic: something bad"}, true},
		{[]string{"Traceback (most recent call last):"}, true},
		{[]string{"Segmentation fault"}, true},
		{[]string{"fatal error: all goroutines are asleep"}, true},
		{[]string{"error: file not found"}, false}, // Not a fatal error
		{[]string{"normal output"}, false},
	}

	for _, tt := range tests {
		got := containsFatalError(tt.lines)
		assert.Equal(t, tt.expected, got, "lines: %v", tt.lines)
	}
}

func TestIsLazygitUI(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		expected bool
	}{
		{
			name: "lazygit UI",
			lines: []string{
				"┌───────────────────┐",
				"│ Status            │",
				"├───────────────────┤",
				"│ Files             │",
				"└───────────────────┘",
			},
			expected: true,
		},
		{
			name: "not lazygit - no keywords",
			lines: []string{
				"┌───────────────────┐",
				"│ something else    │",
				"└───────────────────┘",
			},
			expected: false,
		},
		{
			name: "plain text",
			lines: []string{
				"just some text",
				"no box drawing",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isLazygitUI(tt.lines)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestGetLastNonEmptyLines(t *testing.T) {
	tests := []struct {
		lines    []string
		n        int
		expected []string
	}{
		{
			lines:    []string{"a", "b", "c"},
			n:        2,
			expected: []string{"b", "c"},
		},
		{
			lines:    []string{"a", "", "b", "", "c", ""},
			n:        2,
			expected: []string{"b", "c"},
		},
		{
			lines:    []string{"a"},
			n:        5,
			expected: []string{"a"},
		},
		{
			lines:    []string{},
			n:        5,
			expected: nil,
		},
	}

	for _, tt := range tests {
		got := getLastNonEmptyLines(tt.lines, tt.n)
		assert.Equal(t, tt.expected, got, "lines: %v, n: %d", tt.lines, tt.n)
	}
}
