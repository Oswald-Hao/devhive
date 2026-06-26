package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Oswald-Hao/devhive/internal/api"
	"github.com/Oswald-Hao/devhive/internal/tui"
)

func runNoTUI(modelName string, jsonOut bool, quiet bool, isTerminal bool) {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, tui.HumanError("Failed to read stdin", err.Error(), ""))
		os.Exit(1)
	}

	prompt := strings.TrimSpace(string(input))
	if prompt == "" {
		fmt.Fprintln(os.Stderr, tui.ErrorPrefix.Render()+" No input provided on stdin")
		os.Exit(1)
	}

	if !quiet && isTerminal {
		fmt.Fprintf(os.Stderr, "%s Sending prompt (%d chars)...\n", tui.InfoPrefix.Render(), len(prompt))
	}

	client, err := api.NewClient("", "", modelName)
	if err != nil {
		fmt.Fprintln(os.Stderr, tui.ErrorPrefix.Render()+" "+err.Error())
		os.Exit(1)
	}
	system := systemPrompt
	messages := []api.Message{{Role: "user", Content: prompt}}

	eventCh, errCh := client.CreateMessageStream(system, messages, 4096, modelName)

	var full strings.Builder
	for event := range eventCh {
		if event.Type == "content_block_delta" && event.Delta != nil {
			if event.Delta.Type == "text_delta" {
				full.WriteString(event.Delta.Text)
			}
		}
	}

	select {
	case e := <-errCh:
		if e != nil {
			fmt.Fprintln(os.Stderr, formatAPIError(e.Error()))
			os.Exit(1)
		}
	default:
	}

	result := full.String()
	if result == "" {
		fmt.Fprintln(os.Stderr, tui.ErrorPrefix.Render()+" No response received")
		os.Exit(1)
	}

	if jsonOut {
		output := map[string]string{
			"prompt":   prompt,
			"response": result,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(output)
	} else {
		// Print result; wrap long lines
		for _, line := range strings.Split(result, "\n") {
			fmt.Println(line)
		}
	}
}
