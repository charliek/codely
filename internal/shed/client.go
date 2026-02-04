// Package shed provides a client for interacting with the shed CLI.
package shed

import (
	"encoding/json"
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
		return nil, fmt.Errorf("shed list failed: %w", err)
	}

	var sheds []Shed
	if err := json.Unmarshal(output, &sheds); err != nil {
		// Try parsing as a different format (shed might return objects keyed by server)
		var shedMap map[string][]Shed
		if err2 := json.Unmarshal(output, &shedMap); err2 != nil {
			return nil, fmt.Errorf("parsing shed list: %w (original: %w)", err2, err)
		}

		// Flatten the map
		for server, serverSheds := range shedMap {
			for i := range serverSheds {
				serverSheds[i].Server = server
				sheds = append(sheds, serverSheds[i])
			}
		}
	}

	return sheds, nil
}

// CreateShed creates a new shed with the given options
func (c *DefaultClient) CreateShed(name string, opts CreateOpts) error {
	args := []string{"create", name}

	if opts.Repo != "" {
		args = append(args, "--repo", opts.Repo)
	}
	if opts.Server != "" {
		args = append(args, "--server", opts.Server)
	}
	if opts.Image != "" {
		args = append(args, "--image", opts.Image)
	}

	cmd := exec.Command("shed", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("shed create failed: %s: %w", strings.TrimSpace(string(output)), err)
	}

	return nil
}

// StartShed starts a stopped shed
func (c *DefaultClient) StartShed(name string) error {
	cmd := exec.Command("shed", "start", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("shed start failed: %s: %w", strings.TrimSpace(string(output)), err)
	}
	return nil
}

// StopShed stops a running shed
func (c *DefaultClient) StopShed(name string) error {
	cmd := exec.Command("shed", "stop", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("shed stop failed: %s: %w", strings.TrimSpace(string(output)), err)
	}
	return nil
}

// DeleteShed deletes a shed permanently
func (c *DefaultClient) DeleteShed(name string, force bool) error {
	args := []string{"delete", name}
	if force {
		args = append(args, "--force")
	}

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
