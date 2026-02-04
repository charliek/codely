package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all key bindings for the application
type KeyMap struct {
	// Navigation
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Enter  key.Binding
	Space  key.Binding

	// Actions
	NewProject  key.Binding
	AddTerminal key.Binding
	Close       key.Binding
	CloseAll    key.Binding
	StartShed   key.Binding
	StopShed    key.Binding
	Refresh     key.Binding

	// Dialog
	Confirm key.Binding
	Cancel  key.Binding

	// Search
	Search key.Binding

	// Global
	Help key.Binding
	Quit key.Binding
}

// DefaultKeyMap returns the default key bindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "collapse"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "expand"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select/focus"),
		),
		Space: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "toggle"),
		),
		NewProject: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new project"),
		),
		AddTerminal: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "add terminal"),
		),
		Close: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "close"),
		),
		CloseAll: key.NewBinding(
			key.WithKeys("X"),
			key.WithHelp("X", "close project"),
		),
		StartShed: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("S", "start shed"),
		),
		StopShed: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "stop shed"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "yes"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("n", "esc"),
			key.WithHelp("n/esc", "cancel"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

// ShortHelp returns help for common keys
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.NewProject, k.AddTerminal, k.Close, k.Quit}
}

// FullHelp returns help for all keys
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Enter, k.Space, k.NewProject, k.AddTerminal},
		{k.Close, k.CloseAll, k.Refresh, k.Quit},
	}
}
