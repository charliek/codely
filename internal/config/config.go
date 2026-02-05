package config

import (
	"fmt"
	"os"
	"time"

	"github.com/charliek/codely/internal/constants"
	"github.com/charliek/codely/internal/domain"
	"github.com/charliek/codely/internal/pathutil"
	"gopkg.in/yaml.v3"
)

// Config represents the top-level codely configuration
type Config struct {
	WorkspaceRoots []string           `yaml:"workspace_roots"`
	Commands       map[string]Command `yaml:"commands"`
	DefaultCommand string             `yaml:"default_command"`
	UI             UIConfig           `yaml:"ui"`
	Shed           ShedConfig         `yaml:"shed"`
}

// Command represents a command configuration
type Command struct {
	DisplayName string            `yaml:"display_name"`
	Exec        string            `yaml:"exec"`
	Args        []string          `yaml:"args"`
	Env         map[string]string `yaml:"env,omitempty"`
	// StatusDetection controls tool-specific status heuristics.
	// Supported: auto, generic, claude, opencode, codex, shell
	StatusDetection string `yaml:"status_detection,omitempty"`
}

// UIConfig represents UI preferences
type UIConfig struct {
	ManagerWidth       int    `yaml:"manager_width"`
	StatusPollInterval string `yaml:"status_poll_interval"`
	ShowDirectory      bool   `yaml:"show_directory"`
	AutoExpandProjects bool   `yaml:"auto_expand_projects"`
}

// ShedConfig represents shed integration settings
type ShedConfig struct {
	Enabled       bool   `yaml:"enabled"`
	DefaultServer string `yaml:"default_server"`
}

// Load reads and parses a configuration file
func Load(path string) (*Config, error) {
	// Expand ~ in path
	path = pathutil.ExpandPath(path)

	// Check if file exists
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", domain.ErrConfigNotFound, path)
		}
		return nil, fmt.Errorf("checking config file: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	return Parse(data)
}

// Parse parses configuration from YAML bytes
func Parse(data []byte) (*Config, error) {
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing yaml: %w", err)
	}

	applyDefaults(&config)

	return &config, nil
}

// Default returns a config with default values
func Default() *Config {
	config := &Config{}
	applyDefaults(config)
	return config
}

// applyDefaults sets default values for missing fields
func applyDefaults(config *Config) {
	// Default workspace roots
	if len(config.WorkspaceRoots) == 0 {
		config.WorkspaceRoots = []string{
			"~/work",
			"~/projects",
			"~/src",
		}
	}

	// Default commands
	if config.Commands == nil {
		config.Commands = make(map[string]Command)
	}
	if _, ok := config.Commands["claude"]; !ok {
		config.Commands["claude"] = Command{
			DisplayName: "Claude Code",
			Exec:        "claude",
			Args:        []string{"--dangerously-skip-permissions"},
		}
	}
	if _, ok := config.Commands["opencode"]; !ok {
		config.Commands["opencode"] = Command{
			DisplayName: "OpenCode",
			Exec:        "opencode",
			Args:        []string{},
		}
	}
	if _, ok := config.Commands["codex"]; !ok {
		config.Commands["codex"] = Command{
			DisplayName: "Codex",
			Exec:        "codex",
			Args:        []string{},
		}
	}
	if _, ok := config.Commands["lazygit"]; !ok {
		config.Commands["lazygit"] = Command{
			DisplayName: "Lazygit",
			Exec:        "lazygit",
			Args:        []string{},
		}
	}
	if _, ok := config.Commands["bash"]; !ok {
		config.Commands["bash"] = Command{
			DisplayName: "Bash Shell",
			Exec:        "bash",
			Args:        []string{},
		}
	}

	// Default command
	if config.DefaultCommand == "" {
		config.DefaultCommand = constants.DefaultCommand
	}

	// UI defaults
	if config.UI.ManagerWidth == 0 {
		config.UI.ManagerWidth = constants.DefaultManagerWidth
	}
	if config.UI.StatusPollInterval == "" {
		config.UI.StatusPollInterval = constants.DefaultStatusPollInterval.String()
	}
	// ShowDirectory and AutoExpandProjects default to false (zero value)
	// but we want them to default to true
	// Since we can't distinguish between "not set" and "explicitly set to false"
	// we'll apply these defaults only in Default() or when the config is empty

	// Shed defaults
	if !config.Shed.Enabled {
		config.Shed.Enabled = true
	}
}

// StatusPollIntervalDuration returns the status poll interval as a time.Duration
func (c *Config) StatusPollIntervalDuration() time.Duration {
	d, err := time.ParseDuration(c.UI.StatusPollInterval)
	if err != nil {
		return constants.DefaultStatusPollInterval
	}
	return d
}

// ToDomainCommand converts a config Command to a domain Command
func (c Command) ToDomainCommand(id string) domain.Command {
	return domain.Command{
		ID:          id,
		DisplayName: c.DisplayName,
		Exec:        c.Exec,
		Args:        c.Args,
		Env:         c.Env,
	}
}
