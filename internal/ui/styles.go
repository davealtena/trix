package ui

import "github.com/charmbracelet/lipgloss"

// ANSI 256 color codes: https://www.ditig.com/256-colors-cheat-sheet
// Using subtle colors that work on both light and dark terminals.

var (
	// Severity colors - muted/subtle versions
	ColorCritical = lipgloss.Color("167") // Muted red
	ColorHigh     = lipgloss.Color("173") // Muted orange
	ColorMedium   = lipgloss.Color("179") // Muted yellow
	ColorLow      = lipgloss.Color("246") // Gray
	ColorUnknown  = lipgloss.Color("240") // Dark gray

	// Accent colors
	ColorMuted  = lipgloss.Color("241") // Subtle gray
	ColorBorder = lipgloss.Color("238") // Darker border

	// Severity styles - no bold, subtle colors
	Critical = lipgloss.NewStyle().Foreground(ColorCritical)
	High     = lipgloss.NewStyle().Foreground(ColorHigh)
	Medium   = lipgloss.NewStyle().Foreground(ColorMedium)
	Low      = lipgloss.NewStyle().Foreground(ColorLow)
	Unknown  = lipgloss.NewStyle().Foreground(ColorUnknown)

	// Text styles
	Title = lipgloss.NewStyle().Bold(true)
	Muted = lipgloss.NewStyle().Foreground(ColorMuted)
	Info  = lipgloss.NewStyle() // No color, just plain text
)

// Severity returns the appropriate style for a severity string.
// Usage: ui.Severity("CRITICAL").Render("CRITICAL")
func Severity(sev string) lipgloss.Style {
	switch sev {
	case "CRITICAL":
		return Critical
	case "HIGH":
		return High
	case "MEDIUM":
		return Medium
	case "LOW":
		return Low
	default:
		return Unknown
	}
}
