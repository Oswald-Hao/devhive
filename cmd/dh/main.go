// Package dh provides the DevHive CLI — a conversational multi-agent development system.
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Oswald-Hao/devhive/internal/tui"
)

const version = "0.2.0"

var banner = tui.BannerStyle.Render("  DevHive") +
	tui.DimStyle.Render(" — Multi-Agent Software Development") + "\n" +
	tui.DimStyle.Render(fmt.Sprintf("  Version %s  │  /help for commands  │  Ctrl+C to quit", version))

func main() {
	// Handle one-shot commands (non-REPL mode)
	if len(os.Args) > 1 {
		runOneShot()
		return
	}

	// Interactive REPL mode
	runREPL()
}

func runOneShot() {
	// TODO: Port one-shot commands (dh do, dh status, dh log, dh review, dh resolve)
	fmt.Println("One-shot commands coming soon. Use interactive mode: dh")
	os.Exit(0)
}

func runREPL() {
	fmt.Println(banner)
	fmt.Println()

	ctx := NewContext()
	reader := bufio.NewReader(os.Stdin)

	// Handle Ctrl+C gracefully
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println()
		os.Exit(0)
	}()

	for {
		fmt.Print(tui.PromptStyle.Render("dh> "))

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println()
			break
		}

		line := strings.TrimSpace(input)
		if line == "" {
			continue
		}

		dispatch(ctx, line)
	}
}

func dispatch(ctx *Context, line string) {
	switch {
	// Shell escape
	case strings.HasPrefix(line, "!"):
		runShell(line[1:])

	// Commands
	case strings.HasPrefix(line, "/"):
		parts := strings.Fields(line)
		cmd := strings.ToLower(strings.TrimPrefix(parts[0], "/"))
		args := parts[1:]

		switch cmd {
		case "help", "h":
			showHelp()
		case "quit", "q", "exit":
			fmt.Println(tui.DimStyle.Render("Goodbye."))
			os.Exit(0)
		case "status", "st":
			showStatus(ctx)
		case "tasks", "t":
			showTasks(ctx)
		case "log", "l":
			if len(args) > 0 {
				showLog(ctx, args[0])
			} else {
				fmt.Println(tui.ErrorStyle.Render("Usage: /log <task-id>"))
			}
		case "review", "rv":
			showReview(ctx)
		case "resolve", "rs":
			if len(args) > 0 {
				resolveEscalation(ctx, args[0])
			} else {
				fmt.Println(tui.ErrorStyle.Render("Usage: /resolve <escalation-id>"))
			}
		case "clear":
			fmt.Print("\033[2J\033[H")
		default:
			fmt.Println(tui.ErrorStyle.Render(fmt.Sprintf("Unknown command: /%s  Type /help for commands.", cmd)))
		}

	// Bare text → submit as task
	default:
		submitTask(ctx, line)
	}
}

func runShell(cmd string) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return
	}
	c := exec.Command("bash", "-c", cmd)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	if err := c.Run(); err != nil {
		fmt.Println(tui.ErrorStyle.Render(fmt.Sprintf("Error: %v", err)))
	}
}
