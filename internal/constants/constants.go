// Package constants provides shared configuration values used across the codely application.
package constants

import "time"

// Configuration file defaults
const (
	// DefaultConfigPath is the default configuration file path
	DefaultConfigPath = "~/.config/codely/config.yaml"

	// DefaultStatePath is the default state file path
	DefaultStatePath = "~/.local/state/codely/session.json"
)

// UI defaults
const (
	// DefaultManagerWidth is the default width of the manager panel in characters
	DefaultManagerWidth = 38

	// DefaultStatusPollInterval is the default interval for polling session status
	DefaultStatusPollInterval = 1 * time.Second
)

// Default command
const (
	// DefaultCommand is the default command when adding a terminal to a project
	DefaultCommand = "claude"
)
