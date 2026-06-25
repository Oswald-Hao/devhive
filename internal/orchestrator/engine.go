package orchestrator

import (
	"fmt"
	"time"

	"github.com/Oswald-Hao/devhive/internal/protocol"
	"github.com/google/uuid"
)

// Engine is the central orchestrator for DevHive.
type Engine struct {
	eventBus    *EventBus
	taskQueue   *TaskQueue
	agentPool   *AgentPool
	convergence *ConvergenceGate
	attempts    map[string]map[string]int
	listeners   []func(event string, task *protocol.Task)
}

// NewEngine creates a new orchestrator engine.
func NewEngine() *Engine {
	return &Engine{
		eventBus:    NewEventBus(),
		taskQueue:   NewTaskQueue(),
		agentPool:   NewAgentPool(),
		convergence: NewConvergenceGate(DefaultConvergenceConfig()),
		attempts:    make(map[string]map[string]int),
	}
}

// OnChange registers a listener for task state changes.
func (e *Engine) OnChange(fn func(event string, task *protocol.Task)) {
	e.listeners = append(e.listeners, fn)
}

func (e *Engine) notify(event string, task *protocol.Task) {
	for _, fn := range e.listeners {
		fn(event, task)
	}
}

// Start begins the orchestrator event loop.
func (e *Engine) Start() {
	e.eventBus.Start()

	taskCreatedCh := e.eventBus.Subscribe("task.created")
	agentIdleCh := e.eventBus.Subscribe("agent.idle")
	handoffCh := e.eventBus.Subscribe("handoff.emitted")

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
}

// Stop shuts down the orchestrator.
func (e *Engine) Stop() {
	e.agentPool.StopAll()
	e.eventBus.Stop()
}

// SubmitTask creates and enqueues a new task.
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
	task := e.taskQueue.Peek(event.TaskID)
	if task == nil {
		return
	}

	// Advance past SPECIFY immediately
	task.CurrentStage = protocol.StageExecute
	task.Status = "running"
	e.notify("stage_changed", task)

	// Simulate agent execution with staged progression
	go e.simulatePipeline(task)
}

func (e *Engine) simulatePipeline(task *protocol.Task) {
	// EXECUTE → 模拟执行
	time.Sleep(300 * time.Millisecond)
	task.CurrentStage = protocol.StageVerifyL1
	e.notify("stage_changed", task)

	// VERIFY_L1 → 模拟静态+动态验证
	time.Sleep(200 * time.Millisecond)
	task.CurrentStage = protocol.StageVerifyL2
	e.notify("stage_changed", task)

	// VERIFY_L2 → 模拟语义验证
	time.Sleep(150 * time.Millisecond)
	task.CurrentStage = protocol.StageMerge
	task.Status = "completed"
	e.notify("stage_changed", task)
}

func (e *Engine) onAgentIdle(event *Event) {}

func (e *Engine) onHandoff(event *Event) {}

func (e *Engine) tryDispatch(task *protocol.Task) {}

// GetTasks returns all tasks.
func (e *Engine) GetTasks() []*protocol.Task {
	return e.taskQueue.All()
}

// GetTask returns a task by ID. Supports partial suffix matching.
func (e *Engine) GetTask(taskID string) *protocol.Task {
	if t := e.taskQueue.Peek(taskID); t != nil {
		return t
	}
	// Search partial match by suffix
	for _, t := range e.taskQueue.All() {
		if len(t.ID) >= len(taskID) && t.ID[len(t.ID)-len(taskID):] == taskID {
			return t
		}
	}
	// Search by containment
	for _, t := range e.taskQueue.All() {
		if len(taskID) >= 6 && len(t.ID) >= len(taskID) {
			for i := 0; i <= len(t.ID)-len(taskID); i++ {
				if t.ID[i:i+len(taskID)] == taskID {
					return t
				}
			}
		}
	}
	return nil
}
