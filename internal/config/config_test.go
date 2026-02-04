package config

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/charliek/codely/internal/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_DefaultConfig(t *testing.T) {
	cfg, err := Load(filepath.Join("..", "..", "testdata", "configs", "default.yaml"))
	require.NoError(t, err)

	// Check workspace roots
	assert.Len(t, cfg.WorkspaceRoots, 3)
	assert.Contains(t, cfg.WorkspaceRoots, "~/work")

	// Check commands
	assert.Contains(t, cfg.Commands, "claude")
	assert.Equal(t, "Claude Code", cfg.Commands["claude"].DisplayName)
	assert.Equal(t, "claude", cfg.Commands["claude"].Exec)

	// Check UI config
	assert.Equal(t, 30, cfg.UI.ManagerWidth)
	assert.Equal(t, "1s", cfg.UI.StatusPollInterval)

	// Check shed config
	assert.True(t, cfg.Shed.Enabled)
}

func TestParse_AppliesDefaults(t *testing.T) {
	// Parse an empty config
	cfg, err := Parse([]byte("{}"))
	require.NoError(t, err)

	// Should have default workspace roots
	assert.NotEmpty(t, cfg.WorkspaceRoots)

	// Should have default commands
	assert.Contains(t, cfg.Commands, "claude")
	assert.Contains(t, cfg.Commands, "opencode")
	assert.Contains(t, cfg.Commands, "codex")
	assert.Contains(t, cfg.Commands, "lazygit")
	assert.Contains(t, cfg.Commands, "bash")

	// Should have default UI settings
	assert.Equal(t, constants.DefaultManagerWidth, cfg.UI.ManagerWidth)

	// Should have default command
	assert.Equal(t, constants.DefaultCommand, cfg.DefaultCommand)
}

func TestDefault_ReturnsValidConfig(t *testing.T) {
	cfg := Default()

	// Should have all required fields set
	assert.NotEmpty(t, cfg.WorkspaceRoots)
	assert.NotEmpty(t, cfg.Commands)
	assert.NotEmpty(t, cfg.DefaultCommand)
	assert.Greater(t, cfg.UI.ManagerWidth, 0)
	assert.NotEmpty(t, cfg.UI.StatusPollInterval)

	// Default command should exist in commands
	_, ok := cfg.Commands[cfg.DefaultCommand]
	assert.True(t, ok, "default command should exist in commands map")
}

func TestConfig_StatusPollIntervalDuration(t *testing.T) {
	t.Run("parses valid duration", func(t *testing.T) {
		cfg := &Config{
			UI: UIConfig{StatusPollInterval: "2s"},
		}
		assert.Equal(t, 2*time.Second, cfg.StatusPollIntervalDuration())
	})

	t.Run("returns default for invalid duration", func(t *testing.T) {
		cfg := &Config{
			UI: UIConfig{StatusPollInterval: "invalid"},
		}
		assert.Equal(t, constants.DefaultStatusPollInterval, cfg.StatusPollIntervalDuration())
	})

	t.Run("returns default for empty duration", func(t *testing.T) {
		cfg := &Config{
			UI: UIConfig{StatusPollInterval: ""},
		}
		assert.Equal(t, constants.DefaultStatusPollInterval, cfg.StatusPollIntervalDuration())
	})
}

func TestCommand_ToDomainCommand(t *testing.T) {
	cmd := Command{
		DisplayName: "Claude Code",
		Exec:        "claude",
		Args:        []string{"--dangerously-skip-permissions"},
		Env:         map[string]string{"DEBUG": "true"},
	}

	domainCmd := cmd.ToDomainCommand("claude")

	assert.Equal(t, "claude", domainCmd.ID)
	assert.Equal(t, "Claude Code", domainCmd.DisplayName)
	assert.Equal(t, "claude", domainCmd.Exec)
	assert.Equal(t, []string{"--dangerously-skip-permissions"}, domainCmd.Args)
	assert.Equal(t, map[string]string{"DEBUG": "true"}, domainCmd.Env)
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestParse_InvalidYAML(t *testing.T) {
	_, err := Parse([]byte("invalid: yaml: content:"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing yaml")
}
