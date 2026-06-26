package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Oswald-Hao/devhive/internal/api"
	"github.com/Oswald-Hao/devhive/internal/tui"
	"github.com/charmbracelet/lipgloss"
)

// ---- spinner frames ----
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func runChat(modelName string, resume bool) {
	// --- load state ---
	client, err := newAPIClient(modelName)
	if err != nil {
		fmt.Fprintln(os.Stderr, tui.ErrorPrefix.Render()+" "+err.Error())
		os.Exit(1)
	}
	messages := []chatMsg{}

	if resume {
		s, err := loadSession()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s No saved session found.\n", tui.WarningPrefix.Render())
		} else {
			if modelName == "" {
				client.model = s.Model
			}
			messages = s.Messages
			client.history = s.History
		}
	}

	// --- print header ---
	termWidth := termWidth()
	printHeader(client.model, termWidth)

	// --- print resumed messages ---
	for _, msg := range messages {
		fmt.Print(renderMsg(msg.Role, msg.Content, termWidth))
		fmt.Println()
	}

	// --- Ctrl+C handling (double-press to quit) ---
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT)
	ctrlCPressed := false
	go func() {
		for range sigCh {
			if ctrlCPressed {
				// Second Ctrl+C: quit
				saveSession(messages, client.history, client.model)
				fmt.Printf("\n%s Goodbye.\n", tui.InfoPrefix.Render())
				os.Exit(0)
			}
			ctrlCPressed = true
			fmt.Printf("\n%s Press Ctrl+C again to exit.\n", tui.WarningPrefix.Render())
			// Reset after 2 seconds
			go func() {
				time.Sleep(2 * time.Second)
				ctrlCPressed = false
			}()
		}
	}()

	// --- main input loop ---
	scanner := bufio.NewScanner(os.Stdin)
	prompt := lipgloss.NewStyle().Foreground(tui.Primary).Bold(true).Render("❯") + " "

	for {
		// Print prompt
		fmt.Print(prompt)
		if !scanner.Scan() {
			break // EOF
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		// Slash commands
		if strings.HasPrefix(input, "/") {
			if input == "/quit" || input == "/q" {
				saveSession(messages, client.history, client.model)
				fmt.Printf("\n%s Goodbye.\n", tui.InfoPrefix.Render())
				return
			}
			handleSlashCmd(input, client, &messages, termWidth)
			saveSession(messages, client.history, client.model)
			continue
		}

		// Print user message
		userContent := input
		fmt.Print("\n" + renderMsg(roleUser, userContent, termWidth) + "\n")

		// Call API with spinner
		aiContent, err := callAPIWithSpinner(client, userContent)
		if err != nil {
			errContent := formatAPIError(err.Error())
			messages = append(messages,
				chatMsg{Role: roleUser, Content: userContent},
				chatMsg{Role: roleSystem, Content: errContent},
			)
			fmt.Print(renderMsg(roleSystem, errContent, termWidth) + "\n")
		} else if aiContent != "" {
			messages = append(messages,
				chatMsg{Role: roleUser, Content: userContent},
				chatMsg{Role: roleAssistant, Content: aiContent},
			)
			fmt.Print(renderMsg(roleAssistant, aiContent, termWidth) + "\n")
		}
		fmt.Println()

		saveSession(messages, client.history, client.model)
	}
}

func callAPIWithSpinner(client *apiClient, input string) (string, error) {
	client.history = append(client.history, api.Message{Role: "user", Content: input})
	if len(client.history) > 20 {
		client.history = client.history[len(client.history)-20:]
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start spinner goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	stopSpinner := make(chan struct{})
	go func() {
		defer wg.Done()
		i := 0
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stopSpinner:
				fmt.Fprint(os.Stderr, "\r\033[K") // clear spinner line
				return
			case <-ticker.C:
				fmt.Fprintf(os.Stderr, "\r  %s Thinking...", spinnerFrames[i%len(spinnerFrames)])
				i++
			}
		}
	}()

	// Actually call API
	_ = ctx
	eventCh, errCh := client.client.CreateMessageStream(systemPrompt, client.history, 4096, client.model)

	var full strings.Builder
	for event := range eventCh {
		if event.Type == "content_block_delta" && event.Delta != nil {
			if event.Delta.Type == "text_delta" {
				full.WriteString(event.Delta.Text)
			}
		}
	}

	close(stopSpinner)
	wg.Wait()

	select {
	case err := <-errCh:
		if err != nil {
			return "", err
		}
	default:
	}

	result := full.String()
	if result != "" {
		client.history = append(client.history, api.Message{Role: "assistant", Content: result})
	}
	return result, nil
}

func handleSlashCmd(input string, client *apiClient, messages *[]chatMsg, width int) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return
	}
	cmd := parts[0]
	args := parts[1:]

	var content string

	switch cmd {
	case "/help":
		content = tui.FormatHelpBox("Commands", slashHelp)
	case "/clear":
		client.history = nil
		*messages = nil
		content = tui.SuccessPrefix.Render() + " Conversation cleared."
	case "/model":
		if len(args) > 0 {
			old := client.model
			client.model = args[0]
			content = tui.SuccessPrefix.Render() + fmt.Sprintf(" Switched from %s to %s", old, args[0])
		} else {
			content = tui.InfoPrefix.Render() + " Current model: " + client.model + "\nUsage: /model <name>"
		}
	case "/save":
		if len(args) > 0 {
			buf := historyBytes(*messages, client.model)
			if err := os.WriteFile(args[0], buf, 0644); err != nil {
				content = tui.ErrorPrefix.Render() + " Failed to save: " + err.Error()
			} else {
				content = tui.SuccessPrefix.Render() + " Saved to " + args[0]
			}
		} else {
			content = tui.InfoPrefix.Render() + " Usage: /save <file>"
		}
	case "/specify", "/execute", "/verify", "/merge",
		"/status", "/config", "/checkpoint", "/converge", "/signature":
		content = tui.WarningPrefix.Render() + fmt.Sprintf(" %s: orchestrator not connected in CLI mode.", cmd)
	default:
		content = tui.WarningPrefix.Render() + " Unknown command: " + cmd + "\nType /help for available commands."
	}

	*messages = append(*messages, chatMsg{Role: roleSystem, Content: content})
	fmt.Print("\n" + renderMsg(roleSystem, content, width) + "\n\n")
}

func printHeader(modelName string, width int) {
	headerStyle := lipgloss.NewStyle().Foreground(tui.Primary).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(tui.Dim)

	fmt.Println(headerStyle.Render("╭───── DevHive v" + version + " ─────") + dimStyle.Render(strings.Repeat("─", max(0, width-30))) + "╮")
	fmt.Println(headerStyle.Render("│ ⬡") + "  " + dimStyle.Render(modelName) + strings.Repeat(" ", max(0, width-8-len(modelName))) + "│")
	fmt.Println(dimStyle.Render("╰" + strings.Repeat("─", width-2) + "╯"))
	fmt.Println()
}

func historyBytes(messages []chatMsg, modelName string) []byte {
	var buf strings.Builder
	buf.WriteString("DevHive v" + version + "\n")
	buf.WriteString("Model: " + modelName + "\n")
	buf.WriteString(strings.Repeat("─", 60) + "\n")
	for _, msg := range messages {
		switch msg.Role {
		case roleUser:
			buf.WriteString("▸ " + msg.Content + "\n")
		case roleAssistant:
			for _, line := range strings.Split(msg.Content, "\n") {
				buf.WriteString("│ " + line + "\n")
			}
		case roleSystem:
			buf.WriteString("· " + msg.Content + "\n")
		}
	}
	buf.WriteString(strings.Repeat("─", 60) + "\n")
	return []byte(buf.String())
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func termWidth() int {
	// Try to detect terminal width, default to 80
	if w, _, err := getTermSize(); err == nil && w > 0 {
		return w
	}
	return 80
}

func getTermSize() (int, int, error) {
	// Use simple ioctl or fallback
	fd := int(os.Stdout.Fd())
	return getTermSizeFd(fd)
}
