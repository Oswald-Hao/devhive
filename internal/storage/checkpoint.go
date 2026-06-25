package storage

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const schemaSQL = `
CREATE TABLE IF NOT EXISTS checkpoints (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL,
    stage TEXT NOT NULL,
    agent_id TEXT,
    handoff_json TEXT,
    verdict_json TEXT,
    state_before TEXT,
    state_after TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    duration_ms INTEGER,
    outcome TEXT,
    escalation_id TEXT
);

CREATE INDEX IF NOT EXISTS idx_checkpoint_task ON checkpoints(task_id, created_at);

CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    spec_json TEXT NOT NULL,
    branch TEXT NOT NULL,
    base_commit TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    current_stage TEXT DEFAULT 'SPECIFY',
    status TEXT DEFAULT 'pending'
);

CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status, created_at);

CREATE TABLE IF NOT EXISTS escalation_log (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL,
    report_json TEXT NOT NULL,
    resolved_by TEXT,
    resolved_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

// CheckpointStore provides persistent task-state storage via SQLite.
type CheckpointStore struct {
	db *sql.DB
}

// NewCheckpointStore creates a new checkpoint store.
func NewCheckpointStore(dbPath string) (*CheckpointStore, error) {
	dir := filepath.Dir(dbPath)
	if dir != "." {
		os.MkdirAll(dir, 0755)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(schemaSQL); err != nil {
		db.Close()
		return nil, err
	}

	return &CheckpointStore{db: db}, nil
}

// Close closes the database connection.
func (cs *CheckpointStore) Close() {
	cs.db.Close()
}

// SaveCheckpoint persists a checkpoint.
func (cs *CheckpointStore) SaveCheckpoint(checkpointID, taskID, stage, agentID, handoffJSON, verdictJSON, stateBefore, stateAfter string, durationMs int, outcome, escalationID string) error {
	_, err := cs.db.Exec(
		`INSERT OR REPLACE INTO checkpoints (id, task_id, stage, agent_id, handoff_json, verdict_json, state_before, state_after, duration_ms, outcome, escalation_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		checkpointID, taskID, stage, agentID, handoffJSON, verdictJSON, stateBefore, stateAfter, durationMs, outcome, escalationID,
	)
	return err
}

// GetTaskHistory retrieves the checkpoint history for a task.
func (cs *CheckpointStore) GetTaskHistory(taskID string) ([]map[string]interface{}, error) {
	rows, err := cs.db.Query(
		"SELECT id, task_id, stage, agent_id, outcome, created_at FROM checkpoints WHERE task_id = ? ORDER BY created_at",
		taskID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []map[string]interface{}
	for rows.Next() {
		var id, tID, stage, agentID, outcome, createdAt string
		if err := rows.Scan(&id, &tID, &stage, &agentID, &outcome, &createdAt); err != nil {
			continue
		}
		history = append(history, map[string]interface{}{
			"id":         id,
			"task_id":    tID,
			"stage":      stage,
			"agent_id":   agentID,
			"outcome":    outcome,
			"created_at": createdAt,
		})
	}
	return history, nil
}

// SaveTask persists a task.
func (cs *CheckpointStore) SaveTask(taskID, specJSON, branch, baseCommit string) error {
	_, err := cs.db.Exec(
		`INSERT OR REPLACE INTO tasks (id, spec_json, branch, base_commit) VALUES (?, ?, ?, ?)`,
		taskID, specJSON, branch, baseCommit,
	)
	return err
}

// UpdateTaskStage updates a task's current stage.
func (cs *CheckpointStore) UpdateTaskStage(taskID, stage string) error {
	_, err := cs.db.Exec("UPDATE tasks SET current_stage = ? WHERE id = ?", stage, taskID)
	return err
}

// GetPendingTasks returns all pending tasks.
func (cs *CheckpointStore) GetPendingTasks() ([]map[string]interface{}, error) {
	rows, err := cs.db.Query("SELECT id, spec_json, current_stage, status, created_at FROM tasks WHERE status = 'pending' ORDER BY created_at")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []map[string]interface{}
	for rows.Next() {
		var id, specJSON, stage, status, createdAt string
		if err := rows.Scan(&id, &specJSON, &stage, &status, &createdAt); err != nil {
			continue
		}
		tasks = append(tasks, map[string]interface{}{
			"id":            id,
			"spec_json":     specJSON,
			"current_stage": stage,
			"status":        status,
			"created_at":    createdAt,
		})
	}
	return tasks, nil
}

// SaveEscalation records an escalation.
func (cs *CheckpointStore) SaveEscalation(escalationID, taskID, reportJSON string) error {
	_, err := cs.db.Exec(
		`INSERT OR REPLACE INTO escalation_log (id, task_id, report_json) VALUES (?, ?, ?)`,
		escalationID, taskID, reportJSON,
	)
	return err
}

// ResolveEscalation marks an escalation as resolved.
func (cs *CheckpointStore) ResolveEscalation(escalationID, resolvedBy string) error {
	_, err := cs.db.Exec(
		"UPDATE escalation_log SET resolved_by = ?, resolved_at = ? WHERE id = ?",
		resolvedBy, time.Now().UTC().Format(time.RFC3339), escalationID,
	)
	return err
}

// GetOpenEscalations returns unresolved escalations.
func (cs *CheckpointStore) GetOpenEscalations() ([]map[string]interface{}, error) {
	rows, err := cs.db.Query("SELECT id, task_id, report_json, created_at FROM escalation_log WHERE resolved_by IS NULL ORDER BY created_at")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var escalations []map[string]interface{}
	for rows.Next() {
		var id, taskID, reportJSON, createdAt string
		if err := rows.Scan(&id, &taskID, &reportJSON, &createdAt); err != nil {
			continue
		}
		// Parse the report JSON for display
		var report map[string]interface{}
		json.Unmarshal([]byte(reportJSON), &report)

		escalations = append(escalations, map[string]interface{}{
			"id":         id,
			"task_id":    taskID,
			"report":     report,
			"created_at": createdAt,
		})
	}
	return escalations, nil
}
