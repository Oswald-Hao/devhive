package agents

import (
	"github.com/Oswald-Hao/devhive/internal/api"
	"github.com/Oswald-Hao/devhive/internal/protocol"
)

// Agent defines the interface for all agent types.
type Agent interface {
	ID() string
	Type() protocol.AgentType
	Execute(task *protocol.Task) (map[string]interface{}, error)
}

// BaseConfig holds common agent configuration.
type BaseConfig struct {
	ID          string
	AgentType   protocol.AgentType
	APIClient   *api.Client
	Model       string
	MaxContext  int
	TimeoutSecs int
}

// Base provides common agent functionality.
type Base struct {
	config BaseConfig
}

// NewBase creates a new base agent.
func NewBase(config BaseConfig) *Base {
	return &Base{config: config}
}

// ID returns the agent's ID.
func (b *Base) ID() string {
	return b.config.ID
}

// Type returns the agent's type.
func (b *Base) Type() protocol.AgentType {
	return b.config.AgentType
}

// Client returns the API client.
func (b *Base) Client() *api.Client {
	return b.config.APIClient
}

// Model returns the configured model.
func (b *Base) Model() string {
	if b.config.Model != "" {
		return b.config.Model
	}
	return "deepseek/deepseek-v4-pro"
}
