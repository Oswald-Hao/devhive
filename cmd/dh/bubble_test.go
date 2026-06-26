package main

import (
	"os"
	"strings"
	"testing"

	"github.com/Oswald-Hao/devhive/internal/api"
)

func newTestClient() *apiClient {
	os.Setenv("ANTHROPIC_BASE_URL", "http://test.example.com")
	os.Setenv("ANTHROPIC_AUTH_TOKEN", "test-token")
	c, err := newAPIClient("")
	if err != nil {
		panic(err)
	}
	return c
}

func TestNewAPIClient(t *testing.T) {
	c := newTestClient()
	if c.model == "" {
		t.Error("model should not be empty")
	}
	if len(c.history) != 0 {
		t.Error("history should be empty")
	}

	c2, _ := newAPIClient("claude-sonnet-4-6")
	if c2.model != "claude-sonnet-4-6" {
		t.Errorf("model should be claude-sonnet-4-6, got %s", c2.model)
	}
}

func TestApiClientHistoryTrim(t *testing.T) {
	client := newTestClient()
	for i := 0; i < 11; i++ {
		client.history = append(client.history,
			api.Message{Role: "user", Content: "msg"},
			api.Message{Role: "assistant", Content: "resp"},
		)
	}
	if len(client.history) != 22 {
		t.Errorf("expected 22 history entries, got %d", len(client.history))
	}
}

func TestFormatAPIError(t *testing.T) {
	tests := []struct {
		input    string
		contains string
	}{
		{"connection timed out", "timed out"},
		{"dial tcp: connection refused", "Cannot reach"},
		{"401 Unauthorized", "Authentication failed"},
		{"429 Too Many Requests", "Rate limit"},
		{"500 Internal Server Error", "API server error"},
		{"something unexpected happened", "API request failed"},
	}
	for _, tt := range tests {
		result := formatAPIError(tt.input)
		if !strings.Contains(result, tt.contains) {
			t.Errorf("formatAPIError(%q): expected to contain %q, got %q", tt.input, tt.contains, result)
		}
		if !strings.Contains(result, "✗") {
			t.Errorf("formatAPIError(%q): should contain ✗ prefix", tt.input)
		}
	}
}

func TestRenderMsg(t *testing.T) {
	u := renderMsg(roleUser, "hello", 80)
	if !strings.Contains(u, "hello") {
		t.Error("user msg should contain content")
	}
	if !strings.Contains(u, "You") {
		t.Error("user msg should have You label")
	}

	a := renderMsg(roleAssistant, "hi", 80)
	if !strings.Contains(a, "hi") {
		t.Error("assistant msg should contain content")
	}
	if !strings.Contains(a, "│") {
		t.Error("assistant msg should have left bar prefix")
	}

	s := renderMsg(roleSystem, "note", 80)
	if !strings.Contains(s, "note") {
		t.Error("system msg should contain content")
	}
}

func TestSaveLoadSession(t *testing.T) {
	messages := []chatMsg{
		{Role: roleUser, Content: "hello"},
		{Role: roleAssistant, Content: "hi"},
	}
	history := []api.Message{
		{Role: "user", Content: "hello"},
		{Role: "assistant", Content: "hi"},
	}

	saveSession(messages, history, "test-model")

	s, err := loadSession()
	if err != nil {
		t.Fatalf("loadSession failed: %v", err)
	}
	if s.Model != "test-model" {
		t.Errorf("model should be test-model, got %s", s.Model)
	}
	if len(s.Messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(s.Messages))
	}
	if len(s.History) != 2 {
		t.Errorf("expected 2 history entries, got %d", len(s.History))
	}

	os.Remove(sessionDir() + "/latest.json")
}

func TestHistoryBytes(t *testing.T) {
	messages := []chatMsg{
		{Role: roleUser, Content: "q"},
		{Role: roleAssistant, Content: "a"},
		{Role: roleSystem, Content: "s"},
	}

	output := string(historyBytes(messages, "m"))
	if !strings.Contains(output, "▸ q") {
		t.Error("should have user prefix")
	}
	if !strings.Contains(output, "│ a") {
		t.Error("should have assistant prefix")
	}
	if !strings.Contains(output, "· s") {
		t.Error("should have system prefix")
	}
	if !strings.Contains(output, "DevHive") {
		t.Error("should have DevHive header")
	}
}

func TestHistoryBytesEmpty(t *testing.T) {
	output := historyBytes(nil, "m")
	if len(output) == 0 {
		t.Error("empty history should still produce header")
	}
}
