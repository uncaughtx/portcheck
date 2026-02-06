package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	primaryColor   = lipgloss.Color("#7C3AED") // Purple
	successColor   = lipgloss.Color("#10B981") // Green
	warningColor   = lipgloss.Color("#F59E0B") // Amber
	dangerColor    = lipgloss.Color("#EF4444") // Red
	mutedColor     = lipgloss.Color("#6B7280") // Gray
	textColor      = lipgloss.Color("#F3F4F6") // Light gray
	backgroundColor = lipgloss.Color("#1F2937") // Dark gray

	// Box style for headers
	BoxStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(primaryColor).
		Padding(0, 2).
		MarginBottom(1)

	// Title style
	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		MarginBottom(1)

	// Label style for key-value pairs
	LabelStyle = lipgloss.NewStyle().
		Foreground(mutedColor).
		Width(12)

	// Value style
	ValueStyle = lipgloss.NewStyle().
		Foreground(textColor)

	// Highlight style for important values
	HighlightStyle = lipgloss.NewStyle().
		Foreground(successColor).
		Bold(true)

	// Warning style
	WarningStyle = lipgloss.NewStyle().
		Foreground(warningColor).
		Bold(true)

	// Danger style
	DangerStyle = lipgloss.NewStyle().
		Foreground(dangerColor).
		Bold(true)

	// Help text style
	HelpStyle = lipgloss.NewStyle().
		Foreground(mutedColor).
		MarginTop(1)

	// Help key style
	HelpKeyStyle = lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true)

	// Success message style
	SuccessStyle = lipgloss.NewStyle().
		Foreground(successColor).
		Bold(true)

	// Error style
	ErrorStyle = lipgloss.NewStyle().
		Foreground(dangerColor).
		Bold(true)

	// Info panel style
	InfoPanelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(1, 2).
		MarginTop(1)

	// Confirm box style
	ConfirmBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(dangerColor).
		Padding(1, 2).
		MarginTop(1)
)

// Truncate truncates a string to max length with ellipsis
func Truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}
