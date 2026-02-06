package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/charliek/codely/internal/config"
	"github.com/charliek/codely/internal/constants"
	"github.com/charliek/codely/internal/domain"
	"github.com/charliek/codely/internal/tui"
	"github.com/spf13/cobra"
)

// Version is set during build
var Version = "dev"

// Global flags
var (
	configPath string
	debugMode  bool
	debugFile  string
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "codely",
	Short: "AI Coding Session Manager",
	Long: `codely is a terminal-based project manager for orchestrating AI coding
sessions across local directories and remote development containers.

It provides a unified interface for launching, monitoring, and switching
between multiple concurrent coding sessions running tools like Claude Code,
OpenCode, Codex, or standard shells.`,
	Version:       Version,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runApp,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Persistent flags available to all subcommands
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", constants.DefaultConfigPath, "Config file path")
	rootCmd.PersistentFlags().BoolVarP(&debugMode, "debug", "d", false, "Enable debug logging to file")
	rootCmd.PersistentFlags().StringVar(&debugFile, "debug-file", "~/.local/state/codely/debug.log", "Debug log file path")

	// Set version template
	rootCmd.SetVersionTemplate("codely version {{.Version}}\n")
}

// runApp launches the TUI application
func runApp(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		if errors.Is(err, domain.ErrConfigNotFound) {
			// Use defaults if no config file
			cfg = config.Default()
		} else {
			return fmt.Errorf("loading config: %w", err)
		}
	}

	// Run TUI
	return tui.Run(cfg, constants.DefaultStatePath, debugMode, debugFile)
}
