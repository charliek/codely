// Package tmux provides a client for interacting with tmux.
package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// PaneInfo contains information about a tmux pane
type PaneInfo struct {
	ID       int
	Command  string
	Active   bool
	WindowID string
}

// Client defines the interface for tmux operations
type Client interface {
	// Session management
	InTmux() bool
	CreateSession(name string) error
	AttachSession(name string) error

	// Pane management
	SplitWindow(dir, command string, args ...string) (paneID int, err error)
	SplitPane(targetPaneID int, vertical bool, dir, command string, args ...string) (paneID int, err error)
	FocusPane(paneID int) error
	KillPane(paneID int) error
	ResizePane(paneID int, width int) error
	ToggleZoom(paneID int) error

	// Pane visibility management (for single visible pane mode)
	BreakPane(paneID int) (newPaneID int, err error)                  // Move pane to background window
	JoinPane(paneID int, targetPaneID int) (newPaneID int, err error) // Bring pane back to main window

	// Content capture
	CapturePane(paneID int, lines int) (string, error)

	// Information
	ListPanes() ([]PaneInfo, error)
	PaneExists(paneID int) bool
}

// DefaultClient implements the Client interface using tmux commands
type DefaultClient struct{}

// NewClient creates a new default tmux client
func NewClient() *DefaultClient {
	return &DefaultClient{}
}

// InTmux returns true if currently running inside a tmux session
func (c *DefaultClient) InTmux() bool {
	return os.Getenv("TMUX") != ""
}

// CreateSession creates a new tmux session with the given name
func (c *DefaultClient) CreateSession(name string) error {
	cmd := exec.Command("tmux", "new-session", "-d", "-s", name)
	return cmd.Run()
}

// AttachSession attaches to an existing tmux session
func (c *DefaultClient) AttachSession(name string) error {
	cmd := exec.Command("tmux", "attach-session", "-t", name)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// SplitWindow creates a new pane by splitting the current window horizontally
// It runs the specified command with args in the given directory
// Returns the pane ID of the newly created pane
func (c *DefaultClient) SplitWindow(dir, command string, args ...string) (int, error) {
	// Build the full command string
	cmdParts := []string{command}
	cmdParts = append(cmdParts, args...)
	fullCmd := strings.Join(cmdParts, " ")

	// Build tmux split-window command
	tmuxArgs := []string{
		"split-window",
		"-h",               // horizontal split (new pane to the right)
		"-P",               // print pane info
		"-F", "#{pane_id}", // format: just the pane id
	}

	if dir != "" {
		tmuxArgs = append(tmuxArgs, "-c", dir)
	}

	tmuxArgs = append(tmuxArgs, fullCmd)

	cmd := exec.Command("tmux", tmuxArgs...)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("split-window failed: %w", err)
	}

	// Parse pane ID from output (format: %N)
	paneStr := strings.TrimSpace(string(output))
	if len(paneStr) > 0 && paneStr[0] == '%' {
		paneStr = paneStr[1:]
	}

	paneID, err := strconv.Atoi(paneStr)
	if err != nil {
		return 0, fmt.Errorf("parsing pane id %q: %w", paneStr, err)
	}

	return paneID, nil
}

// SplitPane creates a new pane by splitting a specific target pane
// If vertical is true, splits vertically (new pane below); otherwise horizontally (new pane to right)
// Returns the pane ID of the newly created pane
func (c *DefaultClient) SplitPane(targetPaneID int, vertical bool, dir, command string, args ...string) (int, error) {
	// Build the full command string
	cmdParts := []string{command}
	cmdParts = append(cmdParts, args...)
	fullCmd := strings.Join(cmdParts, " ")

	// Build tmux split-window command
	splitFlag := "-h" // horizontal split (new pane to the right)
	if vertical {
		splitFlag = "-v" // vertical split (new pane below)
	}

	tmuxArgs := []string{
		"split-window",
		splitFlag,
		"-t", fmt.Sprintf("%%%d", targetPaneID), // target pane
		"-P",               // print pane info
		"-F", "#{pane_id}", // format: just the pane id
	}

	if dir != "" {
		tmuxArgs = append(tmuxArgs, "-c", dir)
	}

	tmuxArgs = append(tmuxArgs, fullCmd)

	cmd := exec.Command("tmux", tmuxArgs...)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("split-window failed: %w", err)
	}

	// Parse pane ID from output (format: %N)
	paneStr := strings.TrimSpace(string(output))
	if len(paneStr) > 0 && paneStr[0] == '%' {
		paneStr = paneStr[1:]
	}

	paneID, err := strconv.Atoi(paneStr)
	if err != nil {
		return 0, fmt.Errorf("parsing pane id %q: %w", paneStr, err)
	}

	return paneID, nil
}

// FocusPane switches focus to the specified pane
func (c *DefaultClient) FocusPane(paneID int) error {
	cmd := exec.Command("tmux", "select-pane", "-t", fmt.Sprintf("%%%d", paneID))
	return cmd.Run()
}

// KillPane terminates the specified pane
func (c *DefaultClient) KillPane(paneID int) error {
	cmd := exec.Command("tmux", "kill-pane", "-t", fmt.Sprintf("%%%d", paneID))
	return cmd.Run()
}

// ResizePane sets the width of the specified pane
func (c *DefaultClient) ResizePane(paneID int, width int) error {
	cmd := exec.Command("tmux", "resize-pane", "-t", fmt.Sprintf("%%%d", paneID), "-x", strconv.Itoa(width))
	return cmd.Run()
}

// ToggleZoom toggles zoom for the pane's window.
func (c *DefaultClient) ToggleZoom(paneID int) error {
	cmd := exec.Command("tmux", "resize-pane", "-Z", "-t", fmt.Sprintf("%%%d", paneID))
	return cmd.Run()
}

// CapturePane captures the last N lines of content from the specified pane
func (c *DefaultClient) CapturePane(paneID int, lines int) (string, error) {
	cmd := exec.Command("tmux", "capture-pane",
		"-t", fmt.Sprintf("%%%d", paneID),
		"-p",                            // print to stdout
		"-S", fmt.Sprintf("-%d", lines), // start from -N lines
	)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("capture-pane failed: %w", err)
	}
	return string(output), nil
}

// ListPanes returns information about all panes in the current session
func (c *DefaultClient) ListPanes() ([]PaneInfo, error) {
	cmd := exec.Command("tmux", "list-panes",
		"-a", // all panes across all sessions
		"-F", "#{pane_id}:#{pane_current_command}:#{pane_active}:#{window_id}",
	)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("list-panes failed: %w", err)
	}

	var panes []PaneInfo
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 4)
		if len(parts) < 4 {
			continue
		}

		// Parse pane ID (format: %N)
		idStr := parts[0]
		if len(idStr) > 0 && idStr[0] == '%' {
			idStr = idStr[1:]
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}

		panes = append(panes, PaneInfo{
			ID:       id,
			Command:  parts[1],
			Active:   parts[2] == "1",
			WindowID: parts[3],
		})
	}

	return panes, nil
}

// PaneExists checks if a pane with the given ID exists
func (c *DefaultClient) PaneExists(paneID int) bool {
	panes, err := c.ListPanes()
	if err != nil {
		return false
	}

	for _, p := range panes {
		if p.ID == paneID {
			return true
		}
	}
	return false
}

// BreakPane moves a pane to its own background window
// Returns the new pane ID (pane ID changes after break-pane)
func (c *DefaultClient) BreakPane(paneID int) (int, error) {
	selectCmd := exec.Command("tmux", "select-pane", "-t", fmt.Sprintf("%%%d", paneID))
	if err := selectCmd.Run(); err != nil {
		return 0, fmt.Errorf("select-pane failed: %w", err)
	}

	cmd := exec.Command("tmux", "break-pane",
		"-d", // detach (stay in current window)
		"-P", // print pane info
		"-F", "#{pane_id}",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("break-pane failed: %w: %s", err, strings.TrimSpace(string(output)))
	}

	// Parse new pane ID from output (format: %N)
	paneStr := strings.TrimSpace(string(output))
	if len(paneStr) > 0 && paneStr[0] == '%' {
		paneStr = paneStr[1:]
	}

	newPaneID, err := strconv.Atoi(paneStr)
	if err != nil {
		return 0, fmt.Errorf("parsing pane id %q: %w", paneStr, err)
	}

	return newPaneID, nil
}

// JoinPane brings a pane from a background window back to the main window
// The pane is joined as a horizontal split next to targetPaneID
// Returns the new pane ID (pane ID may change after join-pane)
func (c *DefaultClient) JoinPane(paneID int, targetPaneID int) (int, error) {
	cmd := exec.Command("tmux", "join-pane",
		"-s", fmt.Sprintf("%%%d", paneID), // source pane
		"-t", fmt.Sprintf("%%%d", targetPaneID), // target pane
		"-h", // horizontal split (join to the right)
		"-P", // print pane info
		"-F", "#{pane_id}",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(output))
		// tmux 3.4 doesn't support -P on join-pane
		if strings.Contains(msg, "unknown flag -P") {
			cmd = exec.Command("tmux", "join-pane",
				"-s", fmt.Sprintf("%%%d", paneID),
				"-t", fmt.Sprintf("%%%d", targetPaneID),
				"-h",
			)
			if err := cmd.Run(); err != nil {
				return 0, fmt.Errorf("join-pane failed: %w", err)
			}
			// Assume pane ID stays the same; caller can verify if needed.
			return paneID, nil
		}
		return 0, fmt.Errorf("join-pane failed: %w: %s", err, msg)
	}

	// Parse new pane ID from output (format: %N)
	paneStr := strings.TrimSpace(string(output))
	if len(paneStr) > 0 && paneStr[0] == '%' {
		paneStr = paneStr[1:]
	}

	newPaneID, err := strconv.Atoi(paneStr)
	if err != nil {
		return 0, fmt.Errorf("parsing pane id %q: %w", paneStr, err)
	}

	return newPaneID, nil
}
