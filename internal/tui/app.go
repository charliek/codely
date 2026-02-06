package tui

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/charliek/codely/internal/config"
	"github.com/charliek/codely/internal/debug"
	"github.com/charliek/codely/internal/shed"
	"github.com/charliek/codely/internal/store"
	"github.com/charliek/codely/internal/tmux"
)

// Run starts the TUI application
func Run(cfg *config.Config, storePath string, debugMode bool, debugFile string) error {
	if debugMode {
		if err := debug.Enable(debugFile); err != nil {
			return fmt.Errorf("enabling debug log: %w", err)
		}
		defer debug.Close()
		debug.Log("codely starting")
	}

	// Create tmux client
	tmuxClient := tmux.NewClient()

	// Check if in tmux
	if !tmuxClient.InTmux() {
		return fmt.Errorf("codely must be run inside tmux. Please start tmux first with: tmux new-session -s codely")
	}

	// Create store and load state
	st := store.New(storePath)
	if err := st.Load(); err != nil {
		return fmt.Errorf("loading state: %w", err)
	}

	// Cleanup dead sessions
	st.ReconnectSessions(tmuxClient)
	if err := st.Save(); err != nil {
		return fmt.Errorf("saving state: %w", err)
	}
	debug.Log("store loaded: projects=%d", len(st.Projects()))
	debug.Log("sessions reconnected: projects=%d", len(st.Projects()))

	// Create shed client (optional - may not be available)
	var shedClient shed.Client
	shedDefault := shed.NewClient()
	if shedDefault.Available() {
		shedClient = shedDefault
	}

	// Find our pane ID (the pane running codely). Use -1 as sentinel
	// for "not found" since pane %0 is a valid tmux pane ID.
	codelyPaneID := -1
	var codelyWindowID string
	if tmuxPane := os.Getenv("TMUX_PANE"); tmuxPane != "" {
		if strings.HasPrefix(tmuxPane, "%") {
			tmuxPane = tmuxPane[1:]
		}
		if id, err := strconv.Atoi(tmuxPane); err == nil {
			codelyPaneID = id
		}
	}

	panes, err := tmuxClient.ListPanes()
	if err == nil && len(panes) > 0 {
		for _, p := range panes {
			if codelyPaneID >= 0 && p.ID == codelyPaneID {
				codelyWindowID = p.WindowID
				break
			}
		}
		if codelyPaneID < 0 {
			for _, p := range panes {
				if p.Active {
					codelyPaneID = p.ID
					codelyWindowID = p.WindowID
					break
				}
			}
		}
	}

	debug.Log("startup: TMUX_PANE=%s codelyPaneID=%d codelyWindowID=%s", os.Getenv("TMUX_PANE"), codelyPaneID, codelyWindowID)

	// Create model
	model := NewModel(cfg, st, tmuxClient, shedClient, codelyPaneID, codelyWindowID)

	// Resize the manager pane if possible
	if cfg.UI.ManagerWidth > 0 && codelyPaneID >= 0 {
		_ = tmuxClient.ResizePane(codelyPaneID, cfg.UI.ManagerWidth)
		debug.Log("initial resize: paneID=%d width=%d", codelyPaneID, cfg.UI.ManagerWidth)
	}

	// Run Bubble Tea program
	p := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("running TUI: %w", err)
	}

	// Save state on exit
	if m, ok := finalModel.(Model); ok {
		if err := m.store.Save(); err != nil {
			return fmt.Errorf("saving state on exit: %w", err)
		}
	}

	return nil
}
