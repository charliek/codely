// Package shed provides a client for interacting with the shed CLI.
package shed

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Shed represents a remote development container
type Shed struct {
	Name      string    `json:"name"`
	Server    string    `json:"server"`
	Status    string    `json:"status"` // "running" or "stopped"
	CreatedAt time.Time `json:"created_at"`
	Repo      string    `json:"repo,omitempty"`
	Backend   string    `json:"backend,omitempty"`
}

// CreateOpts contains options for creating a new shed
type CreateOpts struct {
	Repo   string
	Server string
	Image  string
}

// Client defines the interface for shed operations
type Client interface {
	// Available returns true if the shed CLI is installed
	Available() bool

	// Listing
	ListSheds() ([]Shed, error)

	// Lifecycle
	CreateShed(name string, opts CreateOpts) error
	StartShed(name string) error
	StopShed(name string) error
	DeleteShed(name string, force bool) error

	// Execution - returns *exec.Cmd so caller can set up terminal
	ExecCommand(shedName, command string, args ...string) *exec.Cmd
	Console(shedName string) *exec.Cmd
}

// actionResult is the JSON envelope returned by shed mutation commands with --json.
// Only Status is currently inspected; the remaining fields match the shed CLI's
// response schema and are decoded for forward-compatibility.
type actionResult struct {
	Status  string          `json:"status"`
	Action  string          `json:"action"`
	Name    string          `json:"name,omitempty"`
	Server  string          `json:"server,omitempty"`
	Details json.RawMessage `json:"details,omitempty"`
}

// jsonError is the JSON error format written to stderr by the shed CLI.
type jsonError struct {
	Error string `json:"error"`
}

// DefaultClient implements the Client interface using the shed CLI
type DefaultClient struct{}

// NewClient creates a new default shed client
func NewClient() *DefaultClient {
	return &DefaultClient{}
}

// Available checks if the shed CLI is installed and accessible
func (c *DefaultClient) Available() bool {
	_, err := exec.LookPath("shed")
	return err == nil
}

// ListSheds returns all available sheds from all servers
func (c *DefaultClient) ListSheds() ([]Shed, error) {
	cmd := exec.Command("shed", "list", "--all", "--json")
	output, err := cmd.Output()
	if err != nil {
		return nil, parseExecError("list", err)
	}

	var sheds []Shed
	if err := json.Unmarshal(output, &sheds); err != nil {
		return nil, fmt.Errorf("shed list: parsing response: %w", err)
	}

	return sheds, nil
}

// runJSONAction executes a shed CLI command with --json and validates the response.
func runJSONAction(action string, args ...string) error {
	cmd := exec.Command("shed", args...)
	output, err := cmd.Output()
	if err != nil {
		return parseExecError(action, err)
	}

	var result actionResult
	if jsonErr := json.Unmarshal(output, &result); jsonErr != nil {
		return fmt.Errorf("shed %s: unexpected response: %w", action, jsonErr)
	}
	if result.Status != "ok" {
		return fmt.Errorf("shed %s: unexpected status %q", action, result.Status)
	}

	return nil
}

// parseExecError extracts a structured error message from a shed CLI command failure.
// When --json is used, the shed CLI writes {"error": "..."} to stderr on failure.
func parseExecError(action string, err error) error {
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return fmt.Errorf("shed %s failed: %w", action, err)
	}

	stderr := bytes.TrimSpace(exitErr.Stderr)
	if len(stderr) == 0 {
		return fmt.Errorf("shed %s failed: %w", action, err)
	}

	var jsonErr jsonError
	if parseErr := json.Unmarshal(stderr, &jsonErr); parseErr == nil && jsonErr.Error != "" {
		return fmt.Errorf("shed %s: %s: %w", action, jsonErr.Error, err)
	}

	return fmt.Errorf("shed %s failed: %s: %w", action, string(stderr), err)
}

// CreateShed creates a new shed with the given options
func (c *DefaultClient) CreateShed(name string, opts CreateOpts) error {
	args := []string{"create", name, "--json"}

	if opts.Repo != "" {
		args = append(args, "--repo", opts.Repo)
	}
	if opts.Server != "" {
		args = append(args, "--server", opts.Server)
	}
	if opts.Image != "" {
		args = append(args, "--image", opts.Image)
	}

	return runJSONAction("create", args...)
}

// StartShed starts a stopped shed
func (c *DefaultClient) StartShed(name string) error {
	args := []string{"start", name, "--json"}
	return runJSONAction("start", args...)
}

// StopShed stops a running shed
func (c *DefaultClient) StopShed(name string) error {
	args := []string{"stop", name, "--json"}
	return runJSONAction("stop", args...)
}

// DeleteShed deletes a shed permanently
func (c *DefaultClient) DeleteShed(name string, force bool) error {
	args := []string{"delete", name}
	if force {
		// shed CLI requires --force when --json is used for delete
		args = append(args, "--force", "--json")
		return runJSONAction("delete", args...)
	}

	// Non-force path: no --json (interactive confirmation required by shed CLI)
	cmd := exec.Command("shed", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("shed delete failed: %s: %w", strings.TrimSpace(string(output)), err)
	}
	return nil
}

// ExecCommand returns a command that will run in the shed
// The caller should set up stdin/stdout/stderr and Run() the command
func (c *DefaultClient) ExecCommand(shedName, command string, args ...string) *exec.Cmd {
	cmdArgs := []string{"exec", shedName, command}
	cmdArgs = append(cmdArgs, args...)
	return exec.Command("shed", cmdArgs...)
}

// Console returns a command that will open an interactive shell in the shed
func (c *DefaultClient) Console(shedName string) *exec.Cmd {
	return exec.Command("shed", "console", shedName)
}
