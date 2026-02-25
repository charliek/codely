package tui

import (
	"github.com/charliek/codely/internal/config"
	"github.com/charliek/codely/internal/domain"
	"github.com/charliek/codely/internal/shed"
	"github.com/charliek/codely/internal/store"
	"github.com/charliek/codely/internal/tmux"
	"github.com/charliek/codely/internal/tui/components"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/textinput"
)

// Mode represents the current UI mode
type Mode int

const (
	ModeNormal Mode = iota
	ModeFolderPicker
	ModeCommandPicker
	ModeShedPicker
	ModeShedCreate
	ModeShedClose
	ModeConfirm
	ModeHelp
	ModeNewProjectType // Choosing between local/shed
)

// ConfirmAction represents what action is being confirmed
type ConfirmAction int

const (
	ConfirmNone ConfirmAction = iota
	ConfirmCloseSession
	ConfirmCloseProject
	ConfirmDeleteShed
)

// Model is the main application model
type Model struct {
	// Dependencies
	config *config.Config
	store  *store.Store
	tmux   tmux.Client
	shed   shed.Client

	// UI state
	mode     Mode
	tree     *components.Tree
	keys     KeyMap
	help     help.Model
	width    int
	height   int
	err      error
	showHelp bool

	// Picker state
	folders         []string // Available folders for folder picker
	folderIdx       int      // Selected folder index
	folderSearch    textinput.Model
	folderSearching bool

	commands    []config.Command // Available commands
	commandKeys []string         // Command IDs in order
	commandIdx  int              // Selected command index

	sheds   []shed.Shed // Available sheds
	shedIdx int         // Selected shed index

	// Shed create state
	shedCreateName    textinput.Model
	shedCreateRepo    textinput.Model
	shedCreateBackend int // 0="(server default)", 1="docker", 2="firecracker"
	shedCreateFocus   int // 0=name, 1=repo, 2=backend, 3=submit

	// Shed close state
	shedCloseOption int // 0=close only, 1=stop, 2=delete

	// New project type state
	newProjectTypeIdx int // 0=local, 1=attach shed, 2=create shed

	// Confirm state
	confirmAction  ConfirmAction
	confirmProject *domain.Project
	confirmSession *domain.Session

	// Pending state for multi-step workflows
	pendingProject *domain.Project

	// Pane tracking for layout management
	codelyPaneID   int    // The pane running Codely (used for splitting)
	codelyWindowID string // The tmux window containing Codely
	managerWidth   int    // Current width of the codely pane (tracks manual resizes)

	// tmux status bar notifications
	statusBarLast string
	statusBarKeys map[string]int
}

// NewModel creates a new application model
func NewModel(cfg *config.Config, store *store.Store, tmuxClient tmux.Client, shedClient shed.Client, codelyPaneID int, codelyWindowID string) *Model {
	// Build tree from stored projects
	tree := components.NewTree(store.Projects())

	// Expand all projects by default if configured
	if cfg.UI.AutoExpandProjects {
		for _, p := range tree.Projects() {
			p.Expanded = true
		}
		tree.Flatten()
	}

	// Set up folder search input
	folderSearch := textinput.New()
	folderSearch.Placeholder = "Search folders..."
	folderSearch.CharLimit = 100

	// Set up shed create inputs
	shedCreateName := textinput.New()
	shedCreateName.Placeholder = "shed-name"
	shedCreateName.CharLimit = 50
	shedCreateName.Focus()

	shedCreateRepo := textinput.New()
	shedCreateRepo.Placeholder = "user/repo (optional)"
	shedCreateRepo.CharLimit = 100

	// Build commands list
	var commands []config.Command
	var commandKeys []string
	for id, cmd := range cfg.Commands {
		commandKeys = append(commandKeys, id)
		commands = append(commands, cmd)
	}

	return &Model{
		config:         cfg,
		store:          store,
		tmux:           tmuxClient,
		shed:           shedClient,
		mode:           ModeNormal,
		tree:           tree,
		keys:           DefaultKeyMap(),
		help:           help.New(),
		commands:       commands,
		commandKeys:    commandKeys,
		folderSearch:   folderSearch,
		shedCreateName: shedCreateName,
		shedCreateRepo: shedCreateRepo,
		codelyPaneID:   codelyPaneID,
		codelyWindowID: codelyWindowID,
		managerWidth:   cfg.UI.ManagerWidth,
		statusBarKeys:  make(map[string]int),
	}
}

// SelectedProject returns the currently selected project (if any)
func (m *Model) SelectedProject() *domain.Project {
	item := m.tree.Selected()
	if item == nil {
		return nil
	}
	return item.Project
}

// SelectedSession returns the currently selected session (if any)
func (m *Model) SelectedSession() *domain.Session {
	item := m.tree.Selected()
	if item == nil || item.Type != components.ItemTypeSession {
		return nil
	}
	return item.Session
}

// IsSessionSelected returns true if a session is currently selected
func (m *Model) IsSessionSelected() bool {
	item := m.tree.Selected()
	return item != nil && item.Type == components.ItemTypeSession
}
