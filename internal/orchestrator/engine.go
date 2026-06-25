package orchestrator

import (
	"fmt"
	"log"
	"time"

	"github.com/Oswald-Hao/devhive/internal/protocol"
	"github.com/Oswald-Hao/devhive/internal/verification"
	"github.com/google/uuid"
)

// Engine is the central orchestrator for DevHive.
type Engine struct {
	eventBus    *EventBus
	taskQueue   *TaskQueue
	agentPool   *AgentPool
	convergence *ConvergenceGate
	verification *verification.Pipeline
	diagnostic  *verification.DiagnosticAggregator
	attempts    map[string]map[string]int // taskID -> stage -> count
}

// NewEngine creates a new orchestrator engine.
func NewEngine() *Engine {
	return &Engine{
		eventBus:    NewEventBus(),
		taskQueue:   NewTaskQueue(),
		agentPool:   NewAgentPool(),
		convergence: NewConvergenceGate(DefaultConvergenceConfig()),
		verification: verification.NewPipeline(),
		diagnostic:  verification.NewDiagnosticAggregator(),
		attempts:    make(map[string]map[string]int),
	}
}

// Start begins the orchestrator event loop.
func (e *Engine) Start() {
	// Start the event bus
	e.eventBus.Start()

	// Subscribe to events
	taskCreatedCh := e.eventBus.Subscribe("task.created")
	agentIdleCh := e.eventBus.Subscribe("agent.idle")
	handoffCh := e.eventBus.Subscribe("handoff.emitted")

	// Main event loop
	go func() {
		for {
			select {
			case event := <-taskCreatedCh:
				e.onTaskCreated(event)
			case event := <-agentIdleCh:
				e.onAgentIdle(event)
			case event := <-handoffCh:
				e.onHandoff(event)
			}
		}
	}()

	log.Println("[orchestrator] Engine started")
}

// Stop shuts down the orchestrator.
func (e *Engine) Stop() {
	e.agentPool.StopAll()
	e.eventBus.Stop()
	log.Println("[orchestrator] Engine stopped")
}

// SubmitTask creates and enqueues a new task from a spec.
func (e *Engine) SubmitTask(spec *protocol.TaskSpec) string {
	taskID := fmt.Sprintf("task-%s-%s",
		time.Now().UTC().Format("20060102-150405"),
		uuid.New().String()[:6])

	task := &protocol.Task{
		ID:           taskID,
		Spec:         *spec,
		Branch:       "main",
		BaseCommit:   "HEAD",
		CreatedAt:    time.Now().UTC(),
		CurrentStage: protocol.StageSpecify,
		Status:       "pending",
	}

	e.taskQueue.Enqueue(task)

	e.eventBus.Publish(&Event{
		Type:   "task.created",
		TaskID: taskID,
		Payload: map[string]interface{}{
			"task": task,
		},
	})

	return taskID
}

func (e *Engine) onTaskCreated(event *Event) {
	// Start an execute agent if none running
	if e.agentPool.GetIdle(protocol.AgentExecute) == nil {
		e.agentPool.StartAgent(protocol.AgentExecute)
	}

	task := e.taskQueue.Peek(event.TaskID)
	if task == nil {
		return
	}

	// Try to dispatch immediately
	e.tryDispatch(task)
}

func (e *Engine) onAgentIdle(event *Event) {
	agentType, _ := event.Payload["agent_type"].(string)
	task := e.taskQueue.NextPending(agentType)
	if task != nil {
		handle := e.agentPool.GetIdle(protocol.AgentType(agentType))
		if handle != nil {
			e.agentPool.MarkBusy(handle.ID)
			e.incrementAttempt(task.ID, "execute")

			// Publish handoff directly (simulated agent execution)
			go func() {
				time.Sleep(500 * time.Millisecond) // Simulate work

				e.eventBus.Publish(&Event{
					Type:   "handoff.emitted",
					TaskID: task.ID,
					Payload: map[string]interface{}{
						"handoff": map[string]interface{}{
							"handoff_version": "1.0",
							"source":          handle.ID,
							"task_id":         task.ID,
							"intent":          task.Spec.Title,
							"changes":         []interface{}{},
							"verification_focus": []interface{}{},
							"env_changes": map[string]interface{}{
								"new_dependencies":  []string{},
								"config_changes":    []string{},
								"migration_needed":  false,
							},
							"execution_trace": map[string]interface{}{
								"commands_run":       []string{},
								"self_check_passed":  true,
							},
						},
						"agent_id": handle.ID,
					},
				})

				e.agentPool.MarkIdle(handle.ID)
			}()
		}
	}
}

func (e *Engine) onHandoff(event *Event) {
	taskID := event.TaskID
	task := e.taskQueue.Peek(taskID)
	if task == nil {
		return
	}

	task.CurrentStage = protocol.StageVerifyL1

	// Run L1 verification
	static := protocol.NewVerdict(protocol.VerStatic, taskID, protocol.OverPass)
	dynamic := protocol.NewVerdict(protocol.VerDynamic, taskID, protocol.OverPass)

	attempt := e.getAttempt(taskID, "l1")
	decision := e.convergence.EvaluateL1(task, static, dynamic, attempt)

	e.handleDecision(taskID, decision)
}

func (e *Engine) handleDecision(taskID string, decision *protocol.ConvergenceDecision) {
	task := e.taskQueue.Peek(taskID)
	if task == nil {
		return
	}

	switch decision.Action {
	case protocol.ActionPass:
		task.CurrentStage = protocol.StageVerifyL2
		// Run L2 verification
		semantic := protocol.SemanticVerdict{
			VerdictVersion: "1.0",
			TaskID:         taskID,
			Alignment:      protocol.AlignAligned,
			Reasoning:      "Changes align with Spec requirements",
			Overall:        protocol.OverPass,
		}
		l2Decision := e.convergence.EvaluateL2(task,
			protocol.NewVerdict(protocol.VerStatic, taskID, protocol.OverPass),
			protocol.NewVerdict(protocol.VerDynamic, taskID, protocol.OverPass),
			semantic, nil, e.getAttempt(taskID, "l2"))

		if l2Decision.Action == protocol.ActionPass {
			task.CurrentStage = protocol.StageMerge
			task.Status = "completed"
		} else {
			e.handleDecision(taskID, l2Decision)
		}

	case protocol.ActionFix:
		e.incrementAttempt(taskID, "retry")
		task.CurrentStage = protocol.StageExecute
		// Re-enqueue for retry

	case protocol.ActionEscalate:
		if decision.Escalation != nil {
			e.eventBus.Publish(&Event{
				Type:   "escalation.needed",
				TaskID: taskID,
				Payload: map[string]interface{}{
					"report": decision.Escalation,
				},
			})
		}
	}
}

func (e *Engine) tryDispatch(task *protocol.Task) {
	handle := e.agentPool.GetIdle(protocol.AgentExecute)
	if handle == nil {
		return
	}

	e.agentPool.MarkBusy(handle.ID)
	e.incrementAttempt(task.ID, "execute")
	task.CurrentStage = protocol.StageExecute
	task.Status = "running"

	// Publish idle event to trigger dispatch
	e.eventBus.Publish(&Event{
		Type:   "agent.idle",
		TaskID: task.ID,
		Payload: map[string]interface{}{
			"agent_type": string(protocol.AgentExecute),
		},
	})
}

// GetTasks returns all tasks.
func (e *Engine) GetTasks() []*protocol.Task {
	return e.taskQueue.All()
}

// GetTask returns a task by ID.
func (e *Engine) GetTask(taskID string) *protocol.Task {
	return e.taskQueue.Peek(taskID)
}

func (e *Engine) getAttempt(taskID, stage string) int {
	if e.attempts[taskID] == nil {
		e.attempts[taskID] = make(map[string]int)
	}
	return e.attempts[taskID][stage]
}

func (e *Engine) incrementAttempt(taskID, stage string) {
	if e.attempts[taskID] == nil {
		e.attempts[taskID] = make(map[string]int)
	}
	e.attempts[taskID][stage]++
}
