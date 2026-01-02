package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Box creates a rounded box around content.
//
// How lipgloss boxes work:
// 1. Define a Style with Border(), Padding(), Width()
// 2. .Render(string) wraps your content in that box
//
// Border characters from lipgloss.RoundedBorder():
//
//	╭──╮
//	│  │
//	╰──╯
func Box(title, content string, width int) string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(1, 2). // 1 vertical, 2 horizontal
		Width(width)

	titleStyle := lipgloss.NewStyle().Bold(true).MarginBottom(1)

	inner := titleStyle.Render(title) + "\n" + content

	return boxStyle.Render(inner)
}

// Section creates a section header with a subtle line underneath.
// Useful for "By Severity", "By Type" etc.
func Section(title string) string {
	return Title.Render(title) + "\n" + Muted.Render(strings.Repeat("─", len(title)+4))
}

// SeverityLine formats a severity row with colored label and count.
// Example output: "  CRITICAL    12"
func SeverityLine(severity string, count int) string {
	// Color the severity label
	label := Severity(severity).Render(fmt.Sprintf("%-10s", severity))

	// Format count with color, right-aligned
	countStr := Info.Render(fmt.Sprintf("%5d", count))

	return "  " + label + countStr
}

// TypeLine formats a type row with label and count.
// Example output: "  vulnerability   763"
func TypeLine(typeName string, count int) string {
	label := fmt.Sprintf("%-15s", typeName)
	countStr := Info.Render(fmt.Sprintf("%5d", count))
	return "  " + label + countStr
}

// ResourceLine formats a resource row for top affected resources.
func ResourceLine(resource string, count int, maxLen int) string {
	// Truncate resource name if too long
	if len(resource) > maxLen {
		resource = resource[:maxLen-3] + "..."
	}
	label := fmt.Sprintf("%-*s", maxLen, resource)
	countStr := fmt.Sprintf("%d", count)
	return "  " + label + "  " + countStr
}

// Table creates a simple table with headers and rows.
// Each row is a slice of strings, columns are auto-sized.
//
// Example:
//
//	Table(
//	  []string{"Severity", "Type", "Title"},
//	  [][]string{{"CRITICAL", "vuln", "CVE-123"}},
//	)
type Table struct {
	Headers []string
	Rows    [][]string
	Widths  []int // Column widths (auto-calculated if nil)
}

// NewTable creates a table with headers.
func NewTable(headers ...string) *Table {
	return &Table{
		Headers: headers,
		Widths:  make([]int, len(headers)),
	}
}

// AddRow adds a row to the table.
func (t *Table) AddRow(cols ...string) {
	t.Rows = append(t.Rows, cols)
}

// Render outputs the table as a string with box borders.
func (t *Table) Render() string {
	// Calculate column widths (max of header/content)
	for i, h := range t.Headers {
		if len(h) > t.Widths[i] {
			t.Widths[i] = len(h)
		}
	}
	for _, row := range t.Rows {
		for i, col := range row {
			if i < len(t.Widths) && len(col) > t.Widths[i] {
				t.Widths[i] = len(col)
			}
		}
	}

	// Build table string
	var b strings.Builder

	// Header row
	b.WriteString("  ")
	for i, h := range t.Headers {
		b.WriteString(Title.Render(fmt.Sprintf("%-*s", t.Widths[i], h)))
		if i < len(t.Headers)-1 {
			b.WriteString("  ")
		}
	}
	b.WriteString("\n")

	// Separator
	b.WriteString("  ")
	for i, w := range t.Widths {
		b.WriteString(Muted.Render(strings.Repeat("─", w)))
		if i < len(t.Widths)-1 {
			b.WriteString("  ")
		}
	}
	b.WriteString("\n")

	// Data rows
	for _, row := range t.Rows {
		b.WriteString("  ")
		for i, col := range row {
			if i >= len(t.Widths) {
				break
			}
			// Apply severity coloring to first column if it looks like a severity
			if i == 0 && (col == "CRITICAL" || col == "HIGH" || col == "MEDIUM" || col == "LOW") {
				b.WriteString(Severity(col).Render(fmt.Sprintf("%-*s", t.Widths[i], col)))
			} else {
				b.WriteString(fmt.Sprintf("%-*s", t.Widths[i], col))
			}
			if i < len(row)-1 {
				b.WriteString("  ")
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}
