// Package protocol defines all inter-agent communication types.
// Machine-verifiable structured Handoff — no natural language pass-through.
package protocol

import (
	"encoding/json"
	"time"
)

// ── Enums ────────────────────────────────────────────────────

type ChangeType string

const (
	ChangeLogicFix   ChangeType = "logic_fix"
	ChangeNewFeature ChangeType = "new_feature"
	ChangeRefactor   ChangeType = "refactor"
	ChangeConfig     ChangeType = "config"
	ChangeDependency ChangeType = "dependency"
	ChangeDocs       ChangeType = "docs"
)

type Severity string

const (
	SevCritical Severity = "CRITICAL"
	SevHigh     Severity = "HIGH"
	SevMedium   Severity = "MEDIUM"
	SevLow      Severity = "LOW"
	SevInfo     Severity = "INFO"
)

type Priority string

const (
	PriCritical Priority = "CRITICAL"
	PriHigh     Priority = "HIGH"
	PriMedium   Priority = "MEDIUM"
	PriLow      Priority = "LOW"
)

type VerdictOverall string

const (
	OverPass    VerdictOverall = "PASS"
	OverWarn    VerdictOverall = "WARN"
	OverFail    VerdictOverall = "FAIL"
	OverConflict VerdictOverall = "CONFLICT"
)

type VerifierType string

const (
	VerStatic  VerifierType = "static"
	VerDynamic VerifierType = "dynamic"
	VerSemantic VerifierType = "semantic"
)

type Stage string

const (
	StageSpecify  Stage = "SPECIFY"
	StageExecute  Stage = "EXECUTE"
	StageVerifyL1 Stage = "VERIFY_L1"
	StageVerifyL2 Stage = "VERIFY_L2"
	StageMerge    Stage = "MERGE"
)

var StageOrder = []Stage{StageSpecify, StageExecute, StageVerifyL1, StageVerifyL2, StageMerge}

type ConflictType string

const (
	ConflictFact           ConflictType = "fact"
	ConflictInterpretation ConflictType = "interpretation"
	ConflictSpecAmbiguity  ConflictType = "spec_ambiguity"
)

type ConcurrencyAction string

const (
	ActionPass      ConcurrencyAction = "PASS"
	ActionFix       ConcurrencyAction = "FIX"
	ActionEscalate  ConcurrencyAction = "ESCALATE"
	ActionConflict  ConcurrencyAction = "CONFLICT"
)

type Alignment string

const (
	AlignAligned  Alignment = "ALIGNED"
	AlignEnhanced Alignment = "ENHANCED"
	AlignDeviated Alignment = "DEVIATED"
	AlignConflict Alignment = "CONFLICT"
)

type SuggestedAction string

const (
	SuggRollback  SuggestedAction = "ROLLBACK"
	SuggFix       SuggestedAction = "FIX"
	SuggRetest    SuggestedAction = "RETEST"
	SuggEscalate  SuggestedAction = "ESCALATE"
	SuggIgnore    SuggestedAction = "IGNORE"
)

type AgentType string

const (
	AgentExecute         AgentType = "execute"
	AgentStaticVerifier  AgentType = "static_verifier"
	AgentDynamicVerifier AgentType = "dynamic_verifier"
	AgentSemanticVerifier AgentType = "semantic_verifier"
)

// ── Task / Spec ──────────────────────────────────────────────

type TaskSpec struct {
	Title             string   `json:"title"`
	Description       string   `json:"description"`
	AcceptanceCriteria []string `json:"acceptance_criteria,omitempty"`
	ScopeConstraints  []string `json:"scope_constraints,omitempty"`
	SensitiveModules  []string `json:"sensitive_modules,omitempty"`
	Priority          Priority `json:"priority"`
}

type Task struct {
	ID           string    `json:"id"`
	Spec         TaskSpec  `json:"spec"`
	Branch       string    `json:"branch"`
	BaseCommit   string    `json:"base_commit"`
	CreatedAt    time.Time `json:"created_at"`
	CurrentStage Stage     `json:"current_stage"`
	Status       string    `json:"status"`
}

// ── Handoff ──────────────────────────────────────────────────

type RiskAssessment struct {
	Risk     string   `json:"risk"`
	Severity Severity `json:"severity"`
}

type FileChange struct {
	File               string            `json:"file"`
	DiffRange          map[string]int    `json:"diff_range"`
	ChangeType         ChangeType        `json:"change_type"`
	Summary            string            `json:"summary"`
	RiskSelfAssessment []RiskAssessment  `json:"risk_self_assessment,omitempty"`
}

type VerificationFocus struct {
	What        string   `json:"what"`
	HowToVerify string   `json:"how_to_verify"`
	Priority    Priority `json:"priority"`
}

type EnvChanges struct {
	NewDependencies []string `json:"new_dependencies,omitempty"`
	ConfigChanges   []string `json:"config_changes,omitempty"`
	MigrationNeeded bool     `json:"migration_needed"`
}

type ExecutionTrace struct {
	CommandsRun        []string `json:"commands_run,omitempty"`
	SelfCheckPassed    bool     `json:"self_check_passed"`
	SelfCheckOutputHash string  `json:"self_check_output_hash,omitempty"`
}

type ExecutionHandoff struct {
	HandoffVersion    string              `json:"handoff_version"`
	Source            string              `json:"source"`
	TaskID            string              `json:"task_id"`
	Timestamp         time.Time           `json:"timestamp"`
	Intent            string              `json:"intent"`
	Changes           []FileChange        `json:"changes"`
	VerificationFocus []VerificationFocus `json:"verification_focus"`
	EnvChanges        EnvChanges          `json:"env_changes"`
	ExecutionTrace    ExecutionTrace      `json:"execution_trace"`
}

// ── Verdict ──────────────────────────────────────────────────

type FindingEvidence struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

type Finding struct {
	ID               string          `json:"id"`
	Severity         Severity        `json:"severity"`
	Category         string          `json:"category"`
	Title            string          `json:"title"`
	Detail           string          `json:"detail"`
	Evidence         FindingEvidence `json:"evidence"`
	MatchedSignature string          `json:"matched_signature,omitempty"`
	SuggestedAction  SuggestedAction `json:"suggested_action"`
}

type Verdict struct {
	VerdictVersion string          `json:"verdict_version"`
	VerifierType   VerifierType    `json:"verifier_type"`
	TaskID         string          `json:"task_id"`
	Timestamp      time.Time       `json:"timestamp"`
	Overall        VerdictOverall  `json:"overall"`
	Findings       []Finding       `json:"findings,omitempty"`
}

type SemanticVerdict struct {
	VerdictVersion string         `json:"verdict_version"`
	VerifierType   string         `json:"verifier_type"`
	TaskID         string         `json:"task_id"`
	Timestamp      time.Time      `json:"timestamp"`
	Alignment      Alignment      `json:"alignment"`
	Reasoning      string         `json:"reasoning"`
	Concerns       []string       `json:"concerns,omitempty"`
	Overall        VerdictOverall `json:"overall"`
}

// ── Convergence ──────────────────────────────────────────────

type EscalationReport struct {
	EscalationID        string     `json:"escalation_id"`
	TaskID              string     `json:"task_id"`
	TriggeredBy         string     `json:"triggered_by"`
	CurrentState        map[string]any `json:"current_state"`
	History             []map[string]any `json:"history,omitempty"`
	WhatAgentTried      []string   `json:"what_agent_tried,omitempty"`
	BlockingFinding     map[string]any `json:"blocking_finding,omitempty"`
	SuggestedHumanAction string    `json:"suggested_human_action,omitempty"`
}

type ConvergenceDecision struct {
	Action      ConcurrencyAction  `json:"action"`
	Reason      string             `json:"reason"`
	FixStrategy string             `json:"fix_strategy,omitempty"`
	Escalation  *EscalationReport  `json:"escalation,omitempty"`
}

// ── Events ───────────────────────────────────────────────────

type DevHiveEvent struct {
	EventType string         `json:"event_type"`
	TaskID    string         `json:"task_id"`
	Timestamp time.Time      `json:"timestamp"`
	Payload   map[string]any `json:"payload"`
}

// ── Helpers ──────────────────────────────────────────────────

func NewTaskSpec(title, description string, priority Priority) TaskSpec {
	return TaskSpec{
		Title:       title,
		Description: description,
		Priority:    priority,
	}
}

func NewHandoff() *ExecutionHandoff {
	return &ExecutionHandoff{
		HandoffVersion: "1.0",
		Timestamp:      time.Now().UTC(),
		EnvChanges:     EnvChanges{},
		ExecutionTrace: ExecutionTrace{},
	}
}

func NewVerdict(vtype VerifierType, taskID string, overall VerdictOverall) Verdict {
	return Verdict{
		VerdictVersion: "1.0",
		VerifierType:   vtype,
		TaskID:         taskID,
		Timestamp:      time.Now().UTC(),
		Overall:        overall,
	}
}

// ToJSON marshals any value to indented JSON.
func ToJSON(v any) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
