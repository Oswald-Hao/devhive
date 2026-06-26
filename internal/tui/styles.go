// Package tui provides terminal UI styling via lipgloss.
package tui

import (
	"fmt"
	"strings"

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

// Status prefix styles
var (
	SuccessPrefix = lipgloss.NewStyle().Foreground(Success).SetString("✓")
	ErrorPrefix   = lipgloss.NewStyle().Foreground(Error).SetString("✗")
	WarningPrefix = lipgloss.NewStyle().Foreground(Warning).SetString("⚠")
	InfoPrefix    = lipgloss.NewStyle().Foreground(Blue).SetString("ℹ")

	// Status box styles
	ErrorBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Error).
			Padding(1, 2)

	SuccessBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Success).
			Padding(1, 2)

	HelpBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Primary).
			Padding(1, 2)

	// --- TUI layout styles ---

	// UserBubble wraps user messages in a cyan-bordered bubble.
	UserBubble = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Primary).
			Padding(0, 1)

	// SystemBubble wraps system messages in a dim border.
	SystemBubble = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Dim).
			Padding(0, 1)

	// HeaderStyle is the top bar of the TUI.
	HeaderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Primary).
			Padding(0, 1)

	// FooterStyle is the bottom status bar.
	FooterStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Dim).
			Padding(0, 1)

	// InputStyle wraps the textarea input.
	InputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Primary).
			Padding(0, 1)
)

// HumanError formats an error with reason and suggestion in a styled box.
func HumanError(what, reason, suggestion string) string {
	var b strings.Builder
	b.WriteString(ErrorPrefix.Render() + " " + ErrorStyle.Render(what))
	if reason != "" {
		b.WriteString("\n   Reason: " + DimStyle.Render(reason))
	}
	if suggestion != "" {
		b.WriteString("\n   Suggestion: " + DimStyle.Render(suggestion))
	}
	return b.String()
}

// FormatHelpBox renders help text in a bordered panel.
func FormatHelpBox(title, content string) string {
	header := lipgloss.NewStyle().Bold(true).Foreground(Primary).Render(title)
	return header + "\n" + HelpBoxStyle.Render(content)
}

// messageWidth returns the width for message content, capped at 80.
func messageWidth(termWidth int) int {
	w := termWidth - 4
	if w > 80 {
		return 80
	}
	if w < 20 {
		return 20
	}
	return w
}

// RenderUserMsg renders a user message in a cyan bubble with "You" in the top border.
func RenderUserMsg(content string, width int) string {
	w := messageWidth(width)
	b := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Primary).
		Padding(0, 1).
		Width(w)
	// Build top border with " You " label
	label := lipgloss.NewStyle().Foreground(Primary).Bold(true).Render(" You ")
	top := "╭─" + label + strings.Repeat("─", max(0, w+2-lipgloss.Width(label)-2)) + "╮"
	body := b.Render(content)
	// Replace the default top border line with our labeled one
	lines := strings.Split(body, "\n")
	lines[0] = top
	return strings.Join(lines, "\n")
}

// RenderAssistMsg renders an assistant message with a colored left bar.
func RenderAssistMsg(content string, width int) string {
	w := messageWidth(width)
	bar := lipgloss.NewStyle().Foreground(Highlight).SetString("│").String()
	body := lipgloss.NewStyle().Width(w).Padding(0, 1).Render(content)
	var b strings.Builder
	for _, line := range strings.Split(body, "\n") {
		b.WriteString(bar + " " + line + "\n")
	}
	return strings.TrimSuffix(b.String(), "\n")
}

// RenderSystemMsg renders a system message in a dim-bordered panel.
func RenderSystemMsg(content string, width int) string {
	w := messageWidth(width)
	return SystemBubble.Copy().Width(w).Render(content)
}

// RenderHeader builds the top bar.
func RenderHeader(version, model string, width int) string {
	left := BannerStyle.Render("⬡") + " " + BannerStyle.Render("DevHive")
	right := DimStyle.Render("v" + version + " · " + model)
	innerWidth := width - 6
	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	spacer := innerWidth - leftW - rightW
	if spacer < 1 {
		spacer = 1
	}
	content := left + strings.Repeat(" ", spacer) + right
	return HeaderStyle.Copy().Width(width - 2).Render(content)
}

// RenderFooter builds the bottom status bar.
func RenderFooter(width, msgCount int, model string) string {
	left := DimStyle.Render("/help · /clear · /model · /save")
	right := DimStyle.Render(fmt.Sprintf("msgs: %d · %s", msgCount, model))
	innerWidth := width - 6
	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	spacer := innerWidth - leftW - rightW
	if spacer < 1 {
		spacer = 1
	}
	content := left + strings.Repeat(" ", spacer) + right
	return FooterStyle.Copy().Width(width - 2).Render(content)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
