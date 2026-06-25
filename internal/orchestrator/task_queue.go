package orchestrator

import (
	"sort"
	"sync"

	"github.com/Oswald-Hao/devhive/internal/protocol"
)

// TaskQueue is a priority-based task queue.
type TaskQueue struct {
	mu    sync.Mutex
	tasks map[string]*protocol.Task
	order []string // task IDs in insertion order
}

// NewTaskQueue creates a new task queue.
func NewTaskQueue() *TaskQueue {
	return &TaskQueue{
		tasks: make(map[string]*protocol.Task),
		order: make([]string, 0),
	}
}

// Enqueue adds a task to the queue.
func (tq *TaskQueue) Enqueue(task *protocol.Task) {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	tq.tasks[task.ID] = task
	tq.order = append(tq.order, task.ID)
}

// Peek returns a task by ID without removing it.
func (tq *TaskQueue) Peek(taskID string) *protocol.Task {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	return tq.tasks[taskID]
}

// Dequeue removes and returns the highest-priority task.
func (tq *TaskQueue) Dequeue() *protocol.Task {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	if len(tq.order) == 0 {
		return nil
	}
	// Sort by priority then insertion order
	sort.SliceStable(tq.order, func(i, j int) bool {
		ti, tj := tq.tasks[tq.order[i]], tq.tasks[tq.order[j]]
		return priorityWeight(ti.Spec.Priority) > priorityWeight(tj.Spec.Priority)
	})
	id := tq.order[0]
	tq.order = tq.order[1:]
	task := tq.tasks[id]
	delete(tq.tasks, id)
	return task
}

// NextPending returns the highest-priority pending task for a given agent type.
func (tq *TaskQueue) NextPending(agentType string) *protocol.Task {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	if len(tq.order) == 0 {
		return nil
	}
	// Find highest-priority pending task
	best := ""
	bestWeight := -1
	for _, id := range tq.order {
		task := tq.tasks[id]
		if task.Status == "pending" {
			w := priorityWeight(task.Spec.Priority)
			if w > bestWeight {
				bestWeight = w
				best = id
			}
		}
	}
	if best == "" {
		return nil
	}
	return tq.tasks[best]
}

// AllPending returns all pending tasks.
func (tq *TaskQueue) AllPending() []*protocol.Task {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	var tasks []*protocol.Task
	for _, id := range tq.order {
		task := tq.tasks[id]
		if task.Status == "pending" {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

// All returns all tasks.
func (tq *TaskQueue) All() []*protocol.Task {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	var tasks []*protocol.Task
	for _, t := range tq.tasks {
		tasks = append(tasks, t)
	}
	return tasks
}

// Len returns the number of pending tasks.
func (tq *TaskQueue) Len() int {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	return len(tq.tasks)
}

func priorityWeight(p protocol.Priority) int {
	switch p {
	case protocol.PriCritical:
		return 4
	case protocol.PriHigh:
		return 3
	case protocol.PriMedium:
		return 2
	case protocol.PriLow:
		return 1
	default:
		return 0
	}
}
