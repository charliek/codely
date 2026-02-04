// Package tui provides the terminal user interface for Codely.
package tui

import "github.com/charmbracelet/lipgloss"

// Colors
var (
	colorPrimary   = lipgloss.Color("212") // Bright blue
	colorSecondary = lipgloss.Color("240") // Gray
	colorSuccess   = lipgloss.Color("82")  // Green
	colorWarning   = lipgloss.Color("214") // Orange
	colorError     = lipgloss.Color("196") // Red
	colorMuted     = lipgloss.Color("245") // Muted gray
)

// Base styles
var (
	styleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(colorSecondary).
			PaddingLeft(1).
			PaddingRight(1)

	styleFooter = lipgloss.NewStyle().
			Foreground(colorMuted).
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(colorSecondary).
			PaddingLeft(1).
			PaddingRight(1)

	styleError = lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true)

	styleHelp = lipgloss.NewStyle().
			Foreground(colorMuted)
)

// Tree styles
var (
	styleProjectName = lipgloss.NewStyle().
				Bold(true)

	styleProjectPath = lipgloss.NewStyle().
				Foreground(colorMuted).
				Italic(true)

	styleSessionName = lipgloss.NewStyle().
				PaddingLeft(2)

	styleSelected = lipgloss.NewStyle().
			Background(lipgloss.Color("237")).
			Bold(true)

	styleSectionHeader = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorSecondary).
				MarginTop(1)
)

// Status styles
var (
	styleStatusIdle = lipgloss.NewStyle().
			Foreground(colorMuted)

	styleStatusThinking = lipgloss.NewStyle().
				Foreground(colorWarning)

	styleStatusExecuting = lipgloss.NewStyle().
				Foreground(colorSuccess)

	styleStatusError = lipgloss.NewStyle().
				Foreground(colorError)

	styleStatusStopped = lipgloss.NewStyle().
				Foreground(colorMuted)
)

// Dialog styles
var (
	styleDialog = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(1, 2).
			MarginTop(1)

	styleDialogTitle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorPrimary)

	styleDialogOption = lipgloss.NewStyle().
				PaddingLeft(2)

	styleDialogOptionSelected = lipgloss.NewStyle().
					PaddingLeft(2).
					Bold(true).
					Foreground(colorPrimary)
)
