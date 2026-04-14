package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	colorPrimary   = lipgloss.Color("#7C3AED") // violet
	colorSecondary = lipgloss.Color("#06B6D4") // cyan
	colorSuccess   = lipgloss.Color("#10B981") // green
	colorDanger    = lipgloss.Color("#EF4444") // red
	colorWarning   = lipgloss.Color("#F59E0B") // amber
	colorMuted     = lipgloss.Color("#6B7280") // gray
	colorBg        = lipgloss.Color("#1E1B2E") // dark bg
	colorSurface   = lipgloss.Color("#2D2B3E") // card bg
	colorText      = lipgloss.Color("#E2E8F0") // light text

	// Styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			MarginBottom(1)

	logoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorSecondary).
			MarginBottom(1)

	menuItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedMenuStyle = lipgloss.NewStyle().
				PaddingLeft(1).
				Foreground(colorPrimary).
				Bold(true)

	descriptionStyle = lipgloss.NewStyle().
				Foreground(colorMuted).
				PaddingLeft(4)

	statusBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(0, 2).
			MarginBottom(1)

	successStyle = lipgloss.NewStyle().
			Foreground(colorSuccess)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorDanger)

	warningStyle = lipgloss.NewStyle().
			Foreground(colorWarning)

	mutedStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			MarginTop(1)

	resultBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorSecondary).
			Padding(1, 2).
			MarginTop(1).
			MaxHeight(20)

	inputLabelStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Bold(true).
			MarginBottom(1)

	spinnerStyle = lipgloss.NewStyle().
			Foreground(colorWarning)

	logBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorMuted).
			Padding(0, 1)
)
