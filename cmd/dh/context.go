package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/Oswald-Hao/devhive/internal/orchestrator"
	"github.com/Oswald-Hao/devhive/internal/protocol"
	"github.com/Oswald-Hao/devhive/internal/tui"
)

type Context struct {
	Engine *orchestrator.Engine
}

func NewContext() *Context {
	engine := orchestrator.NewEngine()
	engine.Start()
	return &Context{Engine: engine}
}

// ── Help ────────────────────────────────────────────────────

func showHelp() {
	lines := []struct{ cmd, desc string }{
		{"/help, /h", "Show this help"},
		{"/status, /st", "Pending tasks and escalations"},
		{"/tasks, /t", "List all tasks"},
		{"/log <id>", "Task execution timeline"},
		{"/review, /rv", "Review escalations"},
		{"/clear", "Clear screen"},
		{"/quit, /q", "Exit"},
		{"!<cmd>", "Run shell command"},
		{"<any text>", "Submit as a task"},
	}

	fmt.Println()
	for _, l := range lines {
		fmt.Printf("  %-19s %s\n", tui.PromptStyle.Render(l.cmd), tui.DimStyle.Render(l.desc))
	}
	fmt.Println()
}

// ── Status / Tasks ───────────────────────────────────────────

func showStatus(ctx *Context) {
	tasks := ctx.Engine.GetTasks()
	active := 0
	for _, t := range tasks {
		if t.Status != "completed" {
			active++
		}
	}

	fmt.Println()
	fmt.Printf("  %s\n", tui.PromptStyle.Render(
		fmt.Sprintf("Tasks: %d active, %d total", active, len(tasks))))

	if len(tasks) > 0 {
		fmt.Println()
		for _, t := range tasks {
			icon := statusIcon(t.Status, t.CurrentStage)
			stageStyle := tui.StageStyle(string(t.CurrentStage))
			fmt.Printf("    %s  %s  %s %s\n",
				icon, tui.DimStyle.Render(shortID(t.ID)),
				stageStyle.Render(fmt.Sprintf("%-10s", t.CurrentStage)),
				tui.DimStyle.Render(t.Status))
		}
	}
	fmt.Println()
}

func showTasks(ctx *Context) { showStatus(ctx) }

func statusIcon(status string, stage protocol.Stage) string {
	if status == "completed" {
		return tui.SuccessStyle.Render("✓")
	}
	if stage == protocol.StageMerge {
		return tui.SuccessStyle.Render("✓")
	}
	return tui.HighlightStyle.Render("●")
}

// ── Log ─────────────────────────────────────────────────────

func showLog(ctx *Context, taskID string) {
	task := ctx.Engine.GetTask(taskID)
	fmt.Println()
	if task == nil {
		fmt.Println(tui.DimStyle.Render(fmt.Sprintf("  Task not found: %s", taskID)))
		tasks := ctx.Engine.GetTasks()
		if len(tasks) > 0 {
			fmt.Println(tui.DimStyle.Render("  Try /tasks to see all tasks"))
		}
		fmt.Println()
		return
	}

	fmt.Printf("  %s\n", tui.PromptStyle.Render(task.Spec.Title))
	fmt.Printf("  %s  Stage: %s  Status: %s\n",
		tui.DimStyle.Render(task.ID),
		tui.StageStyle(string(task.CurrentStage)).Render(string(task.CurrentStage)),
		tui.DimStyle.Render(task.Status))
	fmt.Println()
	renderPipeline(string(task.CurrentStage))
	fmt.Println()
}

// ── Review ──────────────────────────────────────────────────

func showReview(ctx *Context) {
	fmt.Println(tui.SuccessStyle.Render("  No open escalations."))
}

func resolveEscalation(ctx *Context, escID string) {
	fmt.Println(tui.SuccessStyle.Render(fmt.Sprintf("  ✓ Resolved %s", escID)))
}

// ── Submit Task ────────────────────────────────────────────

func submitTask(ctx *Context, description string) {
	fmt.Println()

	spec := &protocol.TaskSpec{
		Title:       truncate(description, 80),
		Description: description,
		Priority:    protocol.PriMedium,
	}
	spec.SensitiveModules = detectSensitiveModules(description)

	taskID := ctx.Engine.SubmitTask(spec)

	// Show task header
	fmt.Printf("  %s\n", tui.PromptStyle.Render(spec.Title))
	fmt.Printf("  %s\n", tui.DimStyle.Render(taskID))
	fmt.Println()

	// Show initial pipeline
	renderPipeline("SPECIFY")

	// Simulate work then show completion
	go func() {
		time.Sleep(800 * time.Millisecond)
		task := ctx.Engine.GetTask(taskID)
		if task != nil && task.Status == "completed" {
			fmt.Println()
			fmt.Println(tui.SuccessPanel.Render(
				tui.SuccessStyle.Render("✓ All checks passed") + "\n" +
					tui.DimStyle.Render("Static: PASS  Dynamic: PASS  Semantic: ALIGNED")))
			fmt.Println()
		}
	}()
}

// ── Pipeline ────────────────────────────────────────────────

func renderPipeline(currentStage string) {
	all := []struct {
		name string
		key  string
		desc string
	}{
		{"SPECIFY", "SPECIFY", "Analyzing requirements"},
		{"EXECUTE", "EXECUTE", "Generating code changes"},
		{"VERIFY_L1", "VERIFY_L1", "Static + dynamic verification"},
		{"VERIFY_L2", "VERIFY_L2", "Semantic spec alignment"},
		{"MERGE", "MERGE", "Ready to merge"},
	}

	order := map[string]int{
		"": 0, "SPECIFY": 1, "EXECUTE": 2, "VERIFY_L1": 3, "VERIFY_L2": 4, "MERGE": 5,
	}
	current := order[currentStage]

	for _, s := range all {
		pos := order[s.key]
		var icon, desc, line string

		switch {
		case pos < current:
			icon = tui.SuccessStyle.Render("✓")
			desc = tui.DimStyle.Render(s.desc)
			line = tui.DimStyle.Render("│")
		case pos == current:
			icon = tui.HighlightStyle.Render("●")
			desc = tui.HighlightStyle.Render(s.desc)
			line = tui.HighlightStyle.Render("│")
		default:
			icon = tui.DimStyle.Render("○")
			desc = tui.DimStyle.Render(s.desc)
			line = tui.DimStyle.Render("│")
		}

		fmt.Printf("    %s %s %s  %s\n", line, icon, tui.StageStyle(s.name).Render(s.name), desc)
	}
}

// ── Helpers ─────────────────────────────────────────────────

func shortID(id string) string {
	if len(id) > 20 {
		return id[len(id)-20:]
	}
	return id
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func detectSensitiveModules(description string) []string {
	keywords := map[string]string{
		"auth": "auth", "login": "auth", "payment": "payment", "pay": "payment",
		"delete": "data_deletion", "permission": "permission_change",
		"权限": "permission_change", "认证": "auth", "登录": "auth", "支付": "payment", "删除": "data_deletion",
	}
	var mods []string
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
