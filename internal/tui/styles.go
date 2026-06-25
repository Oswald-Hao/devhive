// Package tui provides terminal UI styling via lipgloss.
package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Base colors
	Primary   = lipgloss.Color("39")  // Cyan
	Success   = lipgloss.Color("42")  // Green
	Warning   = lipgloss.Color("226") // Yellow
	Error     = lipgloss.Color("196") // Red
	Dim       = lipgloss.Color("245") // Gray
	Highlight = lipgloss.Color("213") // Pink
	Blue      = lipgloss.Color("33")  // Blue

	// Styles
	BannerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary)

	PromptStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary)

	DimStyle = lipgloss.NewStyle().
			Foreground(Dim)

	SuccessStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Success)

	ErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Error)

	WarningStyle = lipgloss.NewStyle().
			Foreground(Warning)

	HighlightStyle = lipgloss.NewStyle().
			Foreground(Highlight)

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary).
			Padding(0, 1)

	PanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Primary).
			Padding(1, 2)

	SubtlePanel = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Dim).
			Padding(0, 1)

	ErrorPanel = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Error).
			Padding(1, 2)

	SuccessPanel = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Success).
			Padding(1, 2)

	// Pipeline stage colors
	StageColors = map[string]lipgloss.Style{
		"SPECIFY":   lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		"EXECUTE":   lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true),
		"VERIFY_L1": lipgloss.NewStyle().Foreground(lipgloss.Color("213")),
		"VERIFY_L2": lipgloss.NewStyle().Foreground(lipgloss.Color("33")),
		"MERGE":     lipgloss.NewStyle().Foreground(lipgloss.Color("42")),
	}

	// Activity source colors
	SourceColors = map[string]lipgloss.Style{
		"execute-1":  lipgloss.NewStyle().Foreground(lipgloss.Color("226")),
		"static-v":   lipgloss.NewStyle().Foreground(lipgloss.Color("213")),
		"dynamic-v":  lipgloss.NewStyle().Foreground(lipgloss.Color("33")),
		"semantic-v": lipgloss.NewStyle().Foreground(lipgloss.Color("39")),
	}
)

// Panel renders content inside a styled border.
func Panel(title, content string) string {
	if title != "" {
		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary).
			Padding(0, 1)
		return titleStyle.Render(title) + "\n" + PanelStyle.Render(content)
	}
	return PanelStyle.Render(content)
}

// StageStyle returns the lipgloss style for a pipeline stage.
func StageStyle(stage string) lipgloss.Style {
	if s, ok := StageColors[stage]; ok {
		return s
	}
	return DimStyle
}

// SourceStyle returns the lipgloss style for an activity source.
func SourceStyle(source string) lipgloss.Style {
	if s, ok := SourceColors[source]; ok {
		return s
	}
	return DimStyle
}
