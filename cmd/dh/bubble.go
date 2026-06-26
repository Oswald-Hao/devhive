package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/Oswald-Hao/devhive/internal/api"
	"github.com/Oswald-Hao/devhive/internal/tui"
)

const defaultModel = "deepseek-v4-pro"

const systemPrompt = `You are DevHive, a multi-agent software development system.

DevHive's architecture:
- **Orchestrator Engine**: central coordinator that manages task queues, agent pools, and event bus via goroutines and channels
- **Pipeline**: SPECIFY → EXECUTE → VERIFY_L1 (static) → VERIFY_L2 (dynamic/semantic) → MERGE
- **Execute Agent**: calls the AI model, produces code changes, outputs structured Handoff JSON
- **Verifier Agents**: Static (rule engine), Dynamic (test runner), Semantic (spec alignment check)
- **Convergence Gate**: loop detection and escalation
- **Signature Engine**: Pure Go weighted similarity matching
- **Checkpoint Store**: SQLite-based task state persistence
- **Distribution**: single Go binary, zero runtime dependencies

You are the chat interface for DevHive. Answer questions about DevHive's internals accurately. For general programming questions, be concise and provide working code over explanations.`

const slashHelp = `Available Commands

Chat:
  /help           Show this help
  /clear          Clear conversation history
  /model <name>   Switch AI model
  /save <file>    Save conversation to file
  /quit, /q       Exit DevHive

Pipeline:
  /specify <desc> Create a task specification
  /execute        Run the Execute Agent
  /verify         Trigger full verification
  /merge          Merge approved changes
  /status         Show task queue and agent pool state
  /config         View current configuration
  /checkpoint     List or rollback to checkpoints
  /converge       Force convergence check
  /signature      Show error pattern matches`

type msgRole string

const (
	roleUser      msgRole = "user"
	roleAssistant msgRole = "assistant"
	roleSystem    msgRole = "system"
)

type chatMsg struct {
	Role    msgRole `json:"role"`
	Content string  `json:"content"`
}

type sessionFile struct {
	Version  string        `json:"version"`
	Model    string        `json:"model"`
	Messages []chatMsg     `json:"messages"`
	History  []api.Message `json:"history"`
}

type apiClient struct {
	client  *api.Client
	history []api.Message
	model   string
}

func newAPIClient(modelName string) *apiClient {
	if modelName == "" {
		modelName = defaultModel
	}
	return &apiClient{
		client:  api.NewClient("", "", modelName),
		model:   modelName,
		history: []api.Message{},
	}
}

func sessionDir() string {
	d := os.ExpandEnv("$HOME/.devhive/sessions")
	os.MkdirAll(d, 0700)
	return d
}

func saveSession(messages []chatMsg, history []api.Message, modelName string) {
	s := sessionFile{
		Version:  version,
		Model:    modelName,
		Messages: messages,
		History:  history,
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(filepath.Join(sessionDir(), "latest.json"), data, 0644)
}

func loadSession() (*sessionFile, error) {
	data, err := os.ReadFile(filepath.Join(sessionDir(), "latest.json"))
	if err != nil {
		return nil, err
	}
	var s sessionFile
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func formatAPIError(errText string) string {
	lower := strings.ToLower(errText)
	switch {
	case strings.Contains(lower, "timeout") || strings.Contains(lower, "deadline"):
		return tui.HumanError("API request timed out", "the request took too long to complete", "check your network connection or try again with a shorter prompt")
	case strings.Contains(lower, "connection refused") || strings.Contains(lower, "no such host"):
		return tui.HumanError("Cannot reach API server", errText, "verify your network and ANTHROPIC_BASE_URL setting")
	case strings.Contains(lower, "401") || strings.Contains(lower, "unauthorized") || strings.Contains(lower, "403"):
		return tui.HumanError("Authentication failed", errText, "check your ANTHROPIC_AUTH_TOKEN or LEJU_TOKEN")
	case strings.Contains(lower, "429") || strings.Contains(lower, "rate limit"):
		return tui.HumanError("Rate limit exceeded", errText, "wait a moment and try again")
	case strings.Contains(lower, "500") || strings.Contains(lower, "502") || strings.Contains(lower, "503"):
		return tui.HumanError("API server error", errText, "the server is temporarily unavailable; try again in a few moments")
	default:
		return tui.HumanError("API request failed", errText, "run 'dh --help' for usage information")
	}
}

func renderMsg(role msgRole, content string, width int) string {
	switch role {
	case roleUser:
		return tui.RenderUserMsg(content, width)
	case roleAssistant:
		return tui.RenderAssistMsg(content, width)
	case roleSystem:
		return tui.RenderSystemMsg(content, width)
	}
	return content
}
