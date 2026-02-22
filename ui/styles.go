// Package ui exposes shared Lip Gloss styles used across all commands.
// All styles are exported package-level variables so commands can render
// styled output without duplicating style definitions.
package ui

import "github.com/charmbracelet/lipgloss"

var (
	// colorPrimary is the brand purple, adaptive for light and dark terminals.
	colorPrimary = lipgloss.AdaptiveColor{Light: "#5C06D4", Dark: "#A673FF"}
	colorMuted   = lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#9CA3AF"}
	colorSuccess = lipgloss.AdaptiveColor{Light: "#16A34A", Dark: "#4ADE80"}
	colorWarning = lipgloss.AdaptiveColor{Light: "#D97706", Dark: "#FCD34D"}
	colorDanger  = lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#F87171"}

	// Text styles

	// Title renders bold text in the primary brand color.
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(colorPrimary)

	// Subtitle renders italic text in the muted color.
	Subtitle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Italic(true)

	// Bold renders text in bold weight with no color change.
	Bold = lipgloss.NewStyle().Bold(true)

	// Muted renders de-emphasized text in a grey tone.
	Muted = lipgloss.NewStyle().Foreground(colorMuted)

	// Status styles

	// Success renders bold green text, used for positive confirmations.
	Success = lipgloss.NewStyle().Foreground(colorSuccess).Bold(true)

	// Warning renders bold yellow text, used for non-critical alerts.
	Warning = lipgloss.NewStyle().Foreground(colorWarning).Bold(true)

	// Danger renders bold red text, used for errors or destructive actions.
	Danger = lipgloss.NewStyle().Foreground(colorDanger).Bold(true)

	// Block styles

	// Banner renders a filled pill-shaped label with white text on the primary color.
	Banner = lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(colorPrimary).
		Bold(true)

	// Box wraps content in a rounded border using the primary color.
	Box = lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPrimary)
)
