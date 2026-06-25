package agents

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/Oswald-Hao/devhive/internal/api"
	"github.com/Oswald-Hao/devhive/internal/protocol"
)

const executeSystemPrompt = `You are an Execute Agent in the DevHive system. Your job is to:

1. Read the task Spec carefully — understand the target state
2. Explore the codebase to understand the current state and relevant code
3. Produce code changes that move the system toward the target state
4. Run self-checks (compile, lint, basic tests) to verify your changes work
5. Output a structured Execution Handoff in JSON format

CRITICAL RULES:
- Never merge code. Your output goes to a Verifier, not directly to the main branch.
- If you are uncertain about any change, note it in risk_self_assessment.
- Always specify verification_focus — what the verifier should check most carefully.
- Report env_changes — new dependencies, config changes, migrations needed.

OUTPUT FORMAT:
You MUST end your response with a JSON block labeled HANDOFF:
` + "```json" + `
{
  "intent": "one-line summary of what you changed and why",
  "changes": [...],
  "verification_focus": [...],
  "env_changes": {...},
  "execution_trace": {...}
}
` + "```"

// ExecuteAgent produces code changes based on Task Spec.
type ExecuteAgent struct {
	*Base
	maxRetries int
}

// NewExecuteAgent creates a new Execute agent.
func NewExecuteAgent(config BaseConfig) *ExecuteAgent {
	if config.AgentType == "" {
		config.AgentType = protocol.AgentExecute
	}
	return &ExecuteAgent{
		Base:       NewBase(config),
		maxRetries: 2,
	}
}

// Execute runs the agent and produces a Handoff.
func (a *ExecuteAgent) Execute(task *protocol.Task) (map[string]interface{}, error) {
	var lastErr error
	for attempt := 0; attempt <= a.maxRetries; attempt++ {
		handoff, err := a.runExecute(task)
		if err != nil {
			lastErr = err
			continue
		}
		if !a.selfCheck(handoff) {
			lastErr = fmt.Errorf("self-check failed")
			continue
		}
		return map[string]interface{}{
			"handoff": handoff,
		}, nil
	}
	return nil, fmt.Errorf("execute agent %s failed after %d attempts: %w", a.ID(), a.maxRetries+1, lastErr)
}

func (a *ExecuteAgent) runExecute(task *protocol.Task) (*protocol.ExecutionHandoff, error) {
	specJSON := protocol.ToJSON(task.Spec)

	userMessage := fmt.Sprintf(`## Task Spec
%s

## Current State
Branch: %s
Base commit: %s

## Instructions
Read the task spec, explore the codebase, make the necessary changes.
Output the HANDOFF JSON when done.`, specJSON, task.Branch, task.BaseCommit)

	resp, err := a.Client().CreateMessage(
		executeSystemPrompt,
		[]api.Message{{Role: "user", Content: userMessage}},
		8192,
		0.1,
		a.Model(),
	)
	if err != nil {
		return nil, fmt.Errorf("model call failed: %w", err)
	}

	text := api.ExtractText(resp)
	return parseHandoff(text, a.ID(), task.ID)
}

func parseHandoff(text, agentID, taskID string) (*protocol.ExecutionHandoff, error) {
	handoff := protocol.NewHandoff()
	handoff.Source = agentID
	handoff.TaskID = taskID
	handoff.Timestamp = time.Now().UTC()

	// Try to extract JSON from ```json ... ``` block
	re := regexp.MustCompile("(?s)```json\\s*\\n(.*?)\\n```")
	match := re.FindStringSubmatch(text)
	if len(match) > 1 {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(match[1]), &data); err != nil {
			return handoff, nil // Return empty handoff on parse error
		}
		if intent, ok := data["intent"].(string); ok {
			handoff.Intent = intent
		}
		return handoff, nil
	}

	// Fallback: try to parse whole text as JSON
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(text), &data); err == nil {
		if intent, ok := data["intent"].(string); ok {
			handoff.Intent = intent
		}
	}

	return handoff, nil
}

func (a *ExecuteAgent) selfCheck(handoff *protocol.ExecutionHandoff) bool {
	return handoff.ExecutionTrace.SelfCheckPassed
}
