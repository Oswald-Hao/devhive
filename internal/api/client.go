// Package api provides an Anthropic-compatible HTTP client.
package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Config holds API configuration loaded from config file.
type Config struct {
	BaseURL   string
	AuthToken string
	Model     string
}

// Client is an async-friendly HTTP client for the Anthropic Messages API.
type Client struct {
	BaseURL      string
	AuthToken    string
	DefaultModel string
	HTTP         *http.Client
}

// NewClient creates a new API client. Configuration priority:
//  1. Function arguments (non-empty values)
//  2. Environment variables (ANTHROPIC_BASE_URL, ANTHROPIC_AUTH_TOKEN, LEJU_TOKEN, DEVHIVE_MODEL)
//  3. ~/.devhive/config.yaml
func NewClient(baseURL, authToken, model string) (*Client, error) {
	cfg, cfgErr := LoadConfig(configPath())

	if baseURL == "" {
		baseURL = os.Getenv("ANTHROPIC_BASE_URL")
	}
	if baseURL == "" && cfgErr == nil {
		baseURL = cfg.BaseURL
	}
	if authToken == "" {
		authToken = os.Getenv("ANTHROPIC_AUTH_TOKEN")
	}
	if authToken == "" {
		authToken = os.Getenv("LEJU_TOKEN")
	}
	if authToken == "" {
		if data, err := os.ReadFile(os.ExpandEnv("$HOME/.devhive/token")); err == nil {
			authToken = strings.TrimSpace(string(data))
		}
	}
	if authToken == "" && cfgErr == nil {
		authToken = cfg.AuthToken
	}
	if model == "" {
		model = os.Getenv("DEVHIVE_MODEL")
	}
	if model == "" && cfgErr == nil {
		model = cfg.Model
	}

	if baseURL == "" || authToken == "" {
		return nil, errors.New("API configuration missing. Run 'dh --init' to create a config file, then edit ~/.devhive/config.yaml with your API credentials.")
	}

	return &Client{
		BaseURL:      baseURL,
		AuthToken:    authToken,
		DefaultModel: model,
		HTTP: &http.Client{
			Timeout: 600 * time.Second,
		},
	}, nil
}

func configPath() string {
	return os.ExpandEnv("$HOME/.devhive/config.yaml")
}

// LoadConfig reads API configuration from a YAML file.
// Only parses api.base_url, api.auth_token, api.default_model.
func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	inAPI := false
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if trimmed == "api:" {
			inAPI = true
			continue
		}
		if !inAPI {
			continue
		}
		// Exit api section when a non-indented key at top level appears
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			inAPI = false
			continue
		}
		key, val := parseYAMLKeyValue(trimmed)
		switch key {
		case "base_url":
			cfg.BaseURL = val
		case "auth_token":
			cfg.AuthToken = val
		case "default_model":
			cfg.Model = val
		}
	}
	return cfg, nil
}

// parseYAMLKeyValue parses a simple "key: value" or "key: "value"" line.
// Also resolves ${ENV_VAR} references.
func parseYAMLKeyValue(line string) (key, val string) {
	idx := strings.Index(line, ":")
	if idx < 0 {
		return "", ""
	}
	key = strings.TrimSpace(line[:idx])
	val = strings.TrimSpace(line[idx+1:])
	// Strip surrounding quotes
	if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
		val = val[1 : len(val)-1]
	}
	// Resolve ${ENV_VAR}
	if strings.HasPrefix(val, "${") && strings.HasSuffix(val, "}") {
		envName := val[2 : len(val)-1]
		val = os.Getenv(envName)
	}
	return key, val
}

// ConfigTemplate returns the content for a new config file.
func ConfigTemplate() string {
	return strings.TrimLeft(`
# DevHive Configuration
# Fill in your API credentials below.

api:
  # Your API endpoint (required)
  base_url: "https://api.example.com"
  # Your API token (required)
  auth_token: "your-token-here"
  # Default model (optional, leave empty for API default)
  default_model: ""
`, "\n")
}

// WriteDefaultConfig writes a config template to ~/.devhive/config.yaml.
// Does not overwrite an existing file.
func WriteDefaultConfig() error {
	path := configPath()
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("config already exists at %s", path)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(ConfigTemplate()), 0644)
}

// MessageRequest is the request body for creating a message.
type MessageRequest struct {
	Model       string          `json:"model"`
	MaxTokens   int             `json:"max_tokens"`
	Temperature float64         `json:"temperature,omitempty"`
	System      string          `json:"system"`
	Messages    []Message       `json:"messages"`
	Tools       []map[string]any `json:"tools,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

// Message represents a conversation turn.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// MessageResponse is the API response for a message.
type MessageResponse struct {
	ID      string          `json:"id"`
	Model   string          `json:"model"`
	Content []ContentBlock  `json:"content"`
	Usage   UsageInfo       `json:"usage"`
}

// ContentBlock represents a block in the response content.
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	Name string `json:"name,omitempty"`
	ID   string `json:"id,omitempty"`
	Input map[string]any `json:"input,omitempty"`
}

// UsageInfo contains token usage statistics.
type UsageInfo struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// StreamEvent represents a server-sent event in the streaming response.
type StreamEvent struct {
	Type  string       `json:"type"`
	Delta *ContentDelta `json:"delta,omitempty"`
}

// ContentDelta is a delta update in streaming.
type ContentDelta struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// CreateMessage sends a synchronous message to the API.
func (c *Client) CreateMessage(system string, messages []Message, maxTokens int, temperature float64, model string) (*MessageResponse, error) {
	if model == "" {
		model = c.DefaultModel
	}
	if maxTokens == 0 {
		maxTokens = 4096
	}

	reqBody := MessageRequest{
		Model:       model,
		MaxTokens:   maxTokens,
		Temperature: temperature,
		System:      system,
		Messages:    messages,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/v1/messages", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("x-api-key", c.AuthToken)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var msgResp MessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&msgResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &msgResp, nil
}

// CreateMessageStream streams a message response token-by-token.
func (c *Client) CreateMessageStream(system string, messages []Message, maxTokens int, model string) (<-chan StreamEvent, <-chan error) {
	eventCh := make(chan StreamEvent, 100)
	errCh := make(chan error, 1)

	go func() {
		defer close(eventCh)
		defer close(errCh)

		if model == "" {
			model = c.DefaultModel
		}
		if maxTokens == 0 {
			maxTokens = 4096
		}

		reqBody := MessageRequest{
			Model:     model,
			MaxTokens: maxTokens,
			System:    system,
			Messages:  messages,
			Stream:    true,
		}

		bodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			errCh <- err
			return
		}

		req, err := http.NewRequest("POST", c.BaseURL+"/v1/messages", bytes.NewReader(bodyBytes))
		if err != nil {
			errCh <- err
			return
		}

		req.Header.Set("x-api-key", c.AuthToken)
		req.Header.Set("anthropic-version", "2023-06-01")
		req.Header.Set("content-type", "application/json")

		resp, err := c.HTTP.Do(req)
		if err != nil {
			errCh <- err
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			errCh <- fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")
				if data == "" || data == "[DONE]" {
					continue
				}
				var event StreamEvent
				if err := json.Unmarshal([]byte(data), &event); err == nil {
					eventCh <- event
				}
			}
		}
	}()

	return eventCh, errCh
}

// ExtractText extracts the first text block from a message response.
func ExtractText(resp *MessageResponse) string {
	for _, block := range resp.Content {
		if block.Type == "text" {
			return block.Text
		}
	}
	return ""
}

// ExtractAllText concatenates all text blocks from a response.
func ExtractAllText(resp *MessageResponse) string {
	var sb strings.Builder
	for _, block := range resp.Content {
		if block.Type == "text" {
			sb.WriteString(block.Text)
		}
	}
	return sb.String()
}
