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

	// UserLabel is the "You" tag on user bubbles.
	UserLabel = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true).
			Padding(0, 1)

	// AssistBar is the left colored bar for assistant messages.
	AssistBar = lipgloss.NewStyle().
			Foreground(Highlight).
			Bold(true).
			SetString("│")

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

// RenderUserMsg renders a user message in a bubble.
func RenderUserMsg(content string, width int) string {
	label := UserLabel.Render("You")
	body := UserBubble.Copy().Width(width - 4).Render(content)
	return label + "\n" + body
}

// RenderAssistMsg renders an assistant message with a left bar prefix.
func RenderAssistMsg(content string, width int) string {
	label := lipgloss.NewStyle().Foreground(Highlight).Bold(true).Render("DevHive")
	barWidth := lipgloss.Width(AssistBar.String())
	bodyWidth := width - barWidth - 4
	if bodyWidth < 10 {
		bodyWidth = 10
	}
	var b strings.Builder
	b.WriteString(label + "\n")
	for _, line := range strings.Split(content, "\n") {
		b.WriteString(AssistBar.String() + " " + DimStyle.Copy().Width(bodyWidth).Render(line) + "\n")
	}
	return b.String()
}

// RenderSystemMsg renders a system message in a dim-bordered panel.
func RenderSystemMsg(content string, width int) string {
	return SystemBubble.Copy().Width(width - 4).Render(content)
}

// RenderHeader builds the top bar.
func RenderHeader(version, model string, width int) string {
	left := BannerStyle.Render("⬡") + " " + BannerStyle.Render("DevHive")
	right := DimStyle.Render("v" + version + " · " + model)
	// right-align the version/model info
	innerWidth := width - 6 // account for border+padding
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
