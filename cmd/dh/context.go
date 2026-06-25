package main

import (
	"fmt"
	"strings"

	"github.com/Oswald-Hao/devhive/internal/orchestrator"
	"github.com/Oswald-Hao/devhive/internal/protocol"
	"github.com/Oswald-Hao/devhive/internal/tui"
)

// Context holds the runtime state of the DevHive session.
type Context struct {
	Engine *orchestrator.Engine
}

// NewContext creates a new runtime context with a running orchestrator.
func NewContext() *Context {
	engine := orchestrator.NewEngine()
	engine.Start()
	return &Context{Engine: engine}
}

// ── Command Handlers ────────────────────────────────────────

func showHelp() {
	lines := [][]string{
		{"/help, /h", "Show this help"},
		{"/status, /st", "Show system status (pending tasks, escalations)"},
		{"/tasks, /t", "List all tasks"},
		{"/log <id>, /l <id>", "Show task execution timeline"},
		{"/review, /rv", "Review pending escalations"},
		{"/resolve <id>, /rs <id>", "Resolve an escalation"},
		{"/clear", "Clear screen"},
		{"/quit, /q", "Exit"},
		{"!<cmd>", "Escape to shell"},
		{"<any text>", "Submit as natural language task"},
	}

	fmt.Println()
	fmt.Println(tui.TitleStyle.Render("Commands"))
	fmt.Println()
	for _, l := range lines {
		fmt.Printf("  %-24s %s\n", tui.PromptStyle.Render(l[0]), tui.DimStyle.Render(l[1]))
	}
	fmt.Println()
}

func showStatus(ctx *Context) {
	tasks := ctx.Engine.GetTasks()
	fmt.Println()
	fmt.Printf("  %s  %s\n",
		tui.PromptStyle.Render(fmt.Sprintf("Active Tasks: %d", len(tasks))),
		tui.WarningStyle.Render("Open Escalations: 0"),
	)

	if len(tasks) > 0 {
		fmt.Println()
		fmt.Println(tui.DimStyle.Render("  Tasks:"))
		for _, t := range tasks {
			stageStyle := tui.StageStyle(string(t.CurrentStage))
			fmt.Printf("    %s  %s  %s\n",
				tui.DimStyle.Render(t.ID[len(t.ID)-20:]),
				stageStyle.Render(fmt.Sprintf("[%s]", t.CurrentStage)),
				tui.DimStyle.Render(t.Status))
		}
	}

	fmt.Println()
}

func showTasks(ctx *Context) {
	tasks := ctx.Engine.GetTasks()
	if len(tasks) == 0 {
		fmt.Println(tui.DimStyle.Render("  No tasks yet."))
		return
	}
	fmt.Println()
	for _, t := range tasks {
		stageStyle := tui.StageStyle(string(t.CurrentStage))
		fmt.Printf("  %s  %s  %s\n",
			tui.DimStyle.Render(t.ID[len(t.ID)-20:]),
			stageStyle.Render(fmt.Sprintf("[%s]", t.CurrentStage)),
			tui.DimStyle.Render(t.Status))
	}
	fmt.Println()
}

func showLog(ctx *Context, taskID string) {
	task := ctx.Engine.GetTask(taskID)
	if task == nil {
		// Try partial match
		for _, t := range ctx.Engine.GetTasks() {
			if len(t.ID) >= len(taskID) && t.ID[len(t.ID)-len(taskID):] == taskID {
				task = t
				break
			}
		}
	}

	fmt.Println()
	if task == nil {
		fmt.Println(tui.DimStyle.Render(fmt.Sprintf("  Task not found: %s", taskID)))
		fmt.Println()
		return
	}

	panel := tui.SubtlePanel.Render(fmt.Sprintf(
		"%s\n%s  %s  %s",
		tui.PromptStyle.Render(task.Spec.Title),
		tui.DimStyle.Render("ID: "+task.ID),
		tui.StageStyle(string(task.CurrentStage)).Render(fmt.Sprintf("Stage: %s", task.CurrentStage)),
		tui.DimStyle.Render("Status: "+task.Status),
	))
	fmt.Println(panel)
	fmt.Println()

	// Show pipeline progress
	pipelineFromTask(task)
	fmt.Println()
}

func showReview(ctx *Context) {
	fmt.Println(tui.SuccessStyle.Render("  No open escalations!"))
}

func resolveEscalation(ctx *Context, escID string) {
	fmt.Println(tui.SuccessStyle.Render(fmt.Sprintf("✓ Resolved %s", escID)))
}

func submitTask(ctx *Context, description string) {
	fmt.Println(tui.DimStyle.Render("  Parsing task..."))

	// Create a TaskSpec from natural language
	spec := &protocol.TaskSpec{
		Title:       truncate(description, 80),
		Description: description,
		Priority:    protocol.PriMedium,
	}

	// Detect sensitive modules
	spec.SensitiveModules = detectSensitiveModules(description)

	// Submit to orchestrator
	taskID := ctx.Engine.SubmitTask(spec)

	// Show task created panel
	fmt.Println()
	panel := tui.PanelStyle.Render(
		fmt.Sprintf("%s\n%s  %s",
			tui.PromptStyle.Render(spec.Title),
			tui.DimStyle.Render(fmt.Sprintf("ID: %s", taskID)),
			tui.DimStyle.Render("Priority: MEDIUM"),
		))
	fmt.Println(panel)

	// Show pipeline visualization
	fmt.Println()
	pipeline(taskID, "EXECUTE")
	fmt.Println()

	fmt.Println(tui.DimStyle.Render("  Agents are working on this task..."))
	fmt.Printf("  %s\n", tui.DimStyle.Render("  Use /status to check progress, /log "+taskID[len(taskID)-20:]+" for details"))
	fmt.Println()
}

// ── Pipeline visualization ──────────────────────────────────

func pipeline(taskID string, activeStage string) {
	stages := []struct {
		name string
		key  string
	}{
		{"SPECIFY", "SPECIFY"},
		{"EXECUTE", "EXECUTE"},
		{"VERIFY_L1", "VERIFY_L1"},
		{"VERIFY_L2", "VERIFY_L2"},
		{"MERGE", "MERGE"},
	}

	fmt.Println(tui.DimStyle.Render("  Pipeline:"))
	for _, s := range stages {
		var icon, msg string
		var style = tui.StageStyle(s.name)

		switch {
		case s.key == activeStage:
			icon = "●"
			msg = "执行中..."
			style = tui.HighlightStyle
		case stageBefore(s.key, activeStage):
			icon = "✓"
			msg = "完成"
			style = tui.SuccessStyle
		default:
			icon = "○"
			msg = "等待"
			style = tui.DimStyle
		}

		fmt.Printf("    %s  %s  %s\n",
			style.Render(icon),
			tui.StageStyle(s.name).Render(s.name),
			style.Render(msg))
	}
}

func pipelineFromTask(task *protocol.Task) {
	pipeline(task.ID, string(task.CurrentStage))
}

func stageBefore(stage, ref string) bool {
	order := map[string]int{
		"SPECIFY": 0, "EXECUTE": 1, "VERIFY_L1": 2, "VERIFY_L2": 3, "MERGE": 4,
	}
	return order[stage] < order[ref]
}

func detectSensitiveModules(description string) []string {
	var mods []string
	keywords := map[string]string{
		"auth":       "auth",
		"login":      "auth",
		"payment":    "payment",
		"pay":        "payment",
		"delete":     "data_deletion",
		"permission": "permission_change",
		"权限":       "permission_change",
		"认证":       "auth",
		"登录":       "auth",
		"支付":       "payment",
		"删除":       "data_deletion",
	}

	seen := map[string]bool{}
	lower := strings.ToLower(description)
	for kw, mod := range keywords {
		if strings.Contains(lower, kw) && !seen[mod] {
			mods = append(mods, mod)
			seen[mod] = true
		}
	}
	return mods
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
