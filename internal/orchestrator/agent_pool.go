package orchestrator

import (
	"fmt"
	"sync"

	"github.com/Oswald-Hao/devhive/internal/protocol"
)

// AgentHandle manages a running agent instance.
type AgentHandle struct {
	ID        string
	AgentType protocol.AgentType
	Busy      bool
}

// AgentPool manages agent lifecycle and dispatch.
type AgentPool struct {
	mu      sync.RWMutex
	agents  map[string]*AgentHandle
	counter int
}

// NewAgentPool creates a new agent pool.
func NewAgentPool() *AgentPool {
	return &AgentPool{
		agents: make(map[string]*AgentHandle),
	}
}

// StartAgent creates and registers a new agent.
func (ap *AgentPool) StartAgent(agentType protocol.AgentType) *AgentHandle {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	ap.counter++
	id := fmt.Sprintf("%s-%d", agentType, ap.counter)
	handle := &AgentHandle{
		ID:        id,
		AgentType: agentType,
		Busy:      false,
	}
	ap.agents[id] = handle
	return handle
}

// GetIdle returns an idle agent of the given type, or nil.
func (ap *AgentPool) GetIdle(agentType protocol.AgentType) *AgentHandle {
	ap.mu.RLock()
	defer ap.mu.RUnlock()
	for _, h := range ap.agents {
		if h.AgentType == agentType && !h.Busy {
			return h
		}
	}
	return nil
}

// MarkBusy marks an agent as busy.
func (ap *AgentPool) MarkBusy(agentID string) {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	if a, ok := ap.agents[agentID]; ok {
		a.Busy = true
	}
}

// MarkIdle marks an agent as idle.
func (ap *AgentPool) MarkIdle(agentID string) {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	if a, ok := ap.agents[agentID]; ok {
		a.Busy = false
	}
}

// StopAll removes all agents.
func (ap *AgentPool) StopAll() {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	ap.agents = make(map[string]*AgentHandle)
}

// Count returns the count of agents by type.
func (ap *AgentPool) Count() map[protocol.AgentType]int {
	ap.mu.RLock()
	defer ap.mu.RUnlock()
	counts := make(map[protocol.AgentType]int)
	for _, h := range ap.agents {
		counts[h.AgentType]++
	}
	return counts
}
