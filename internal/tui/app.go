package tui

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/charliek/codely/internal/config"
	"github.com/charliek/codely/internal/shed"
	"github.com/charliek/codely/internal/store"
	"github.com/charliek/codely/internal/tmux"
)

// Run starts the TUI application
func Run(cfg *config.Config, storePath string) error {
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

	// Create shed client (optional - may not be available)
	var shedClient shed.Client
	shedDefault := shed.NewClient()
	if shedDefault.Available() {
		shedClient = shedDefault
	}

	// Find our pane ID (the pane running codely)
	var codelyPaneID int
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
			if codelyPaneID > 0 && p.ID == codelyPaneID {
				codelyWindowID = p.WindowID
				break
			}
		}
		if codelyPaneID == 0 {
			for _, p := range panes {
				if p.Active {
					codelyPaneID = p.ID
					codelyWindowID = p.WindowID
					break
				}
			}
		}
	}

	// Create model
	model := NewModel(cfg, st, tmuxClient, shedClient, codelyPaneID, codelyWindowID)

	// Resize the manager pane if possible
	if cfg.UI.ManagerWidth > 0 && codelyPaneID > 0 {
		_ = tmuxClient.ResizePane(codelyPaneID, cfg.UI.ManagerWidth)
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
