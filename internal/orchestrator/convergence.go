package orchestrator

import (
	"fmt"

	"github.com/Oswald-Hao/devhive/internal/protocol"
)

// ConvergenceConfig holds convergence thresholds.
type ConvergenceConfig struct {
	L1MaxRetries            int
	L2MaxRetries            int
	LoopDetectionWindow     int
	LoopSimilarityThreshold float64
}

// DefaultConvergenceConfig returns sensible defaults.
func DefaultConvergenceConfig() ConvergenceConfig {
	return ConvergenceConfig{
		L1MaxRetries:            3,
		L2MaxRetries:            2,
		LoopDetectionWindow:     5,
		LoopSimilarityThreshold: 0.8,
	}
}

// ConvergenceGate evaluates convergence conditions for L1 and L2.
type ConvergenceGate struct {
	config   ConvergenceConfig
	history  map[string][]convergenceEntry
}

type convergenceEntry struct {
	stage   string
	outcome string
}

// NewConvergenceGate creates a new convergence gate.
func NewConvergenceGate(config ConvergenceConfig) *ConvergenceGate {
	return &ConvergenceGate{
		config:  config,
		history: make(map[string][]convergenceEntry),
	}
}

// EvaluateL1 evaluates static + dynamic verdicts for L1 convergence.
func (cg *ConvergenceGate) EvaluateL1(task *protocol.Task, static, dynamic protocol.Verdict, attempt int) *protocol.ConvergenceDecision {
	cg.record(task.ID, "VERIFY_L1", string(static.Overall)+"/"+string(dynamic.Overall))

	// Both pass → advance
	if static.Overall == protocol.OverPass && dynamic.Overall == protocol.OverPass {
		return &protocol.ConvergenceDecision{
			Action: protocol.ActionPass,
			Reason: "L1: Static and Dynamic verification both passed",
		}
	}

	// Retry limit reached → escalate
	if attempt >= cg.config.L1MaxRetries {
		return &protocol.ConvergenceDecision{
			Action: protocol.ActionEscalate,
			Reason: fmt.Sprintf("L1: Retry limit (%d) reached", cg.config.L1MaxRetries),
			Escalation: &protocol.EscalationReport{
				EscalationID:        fmt.Sprintf("esc-%s", task.ID),
				TaskID:              task.ID,
				TriggeredBy:         "l1_max_retries",
				SuggestedHumanAction: "Review verification findings and determine if code changes need manual intervention",
			},
		}
	}

	// Check if any findings match known signatures
	allFindings := append(static.Findings, dynamic.Findings...)
	matchedFixed := false
	for _, f := range allFindings {
		if f.MatchedSignature != "" && (f.Severity == protocol.SevCritical || f.Severity == protocol.SevHigh) {
			matchedFixed = true
			break
		}
	}

	if matchedFixed {
		return &protocol.ConvergenceDecision{
			Action:      protocol.ActionFix,
			Reason:      "L1: Critical findings matched known signatures — auto-fix available",
			FixStrategy: "SIGNATURE_MATCH",
		}
	}

	// Has critical/high findings without known fix → escalate
	for _, f := range allFindings {
		if f.Severity == protocol.SevCritical || f.Severity == protocol.SevHigh {
			return &protocol.ConvergenceDecision{
				Action: protocol.ActionEscalate,
				Reason: fmt.Sprintf("L1: %s finding: %s", f.Severity, f.Title),
				Escalation: &protocol.EscalationReport{
					EscalationID:        fmt.Sprintf("esc-%s", task.ID),
					TaskID:              task.ID,
					TriggeredBy:         "unknown_critical_finding",
					SuggestedHumanAction: f.Detail,
				},
			}
		}
	}

	// Only WARN/LOW → pass with caution
	return &protocol.ConvergenceDecision{
		Action: protocol.ActionPass,
		Reason: "L1: No critical findings, passing with warnings",
	}
}

// EvaluateL2 evaluates L1 verdicts + semantic + mutation for L2 convergence.
func (cg *ConvergenceGate) EvaluateL2(task *protocol.Task, static, dynamic protocol.Verdict, semantic protocol.SemanticVerdict, mutation *protocol.Verdict, attempt int) *protocol.ConvergenceDecision {
	// Check L1 first
	l1Decision := cg.EvaluateL1(task, static, dynamic, attempt)
	if l1Decision.Action != protocol.ActionPass {
		return l1Decision
	}

	// Check semantic alignment
	switch semantic.Alignment {
	case protocol.AlignDeviated:
		return &protocol.ConvergenceDecision{
			Action: protocol.ActionEscalate,
			Reason: fmt.Sprintf("L2: Semantic alignment DEVIATED — %s", semantic.Reasoning),
			Escalation: &protocol.EscalationReport{
				EscalationID: fmt.Sprintf("esc-%s", task.ID),
				TaskID:       task.ID,
				TriggeredBy:  "semantic_deviation",
				SuggestedHumanAction: semantic.Reasoning,
			},
		}
	case protocol.AlignConflict:
		return &protocol.ConvergenceDecision{
			Action: protocol.ActionConflict,
			Reason: fmt.Sprintf("L2: Semantic CONFLICT — %s", semantic.Reasoning),
		}
	case protocol.AlignEnhanced:
		return &protocol.ConvergenceDecision{
			Action: protocol.ActionPass,
			Reason: "L2: Changes exceed Spec but are reasonable enhancements. Human review recommended.",
		}
	}

	// Check mutation testing
	if mutation != nil && mutation.Overall == protocol.OverFail {
		if attempt >= cg.config.L2MaxRetries {
			return &protocol.ConvergenceDecision{
				Action: protocol.ActionEscalate,
				Reason: "L2: Mutation testing failures persist after retries",
				Escalation: &protocol.EscalationReport{
					EscalationID: fmt.Sprintf("esc-%s", task.ID),
					TaskID:       task.ID,
					TriggeredBy:  "mutation_testing_failure",
				},
			}
		}
		return &protocol.ConvergenceDecision{
			Action:      protocol.ActionFix,
			Reason:      "L2: Mutation testing found coverage gaps",
			FixStrategy: "ADD_TESTS",
		}
	}

	return &protocol.ConvergenceDecision{
		Action: protocol.ActionPass,
		Reason: "L2: All verification layers passed",
	}
}

// DetectLoop checks if the system is stuck in a loop.
func (cg *ConvergenceGate) DetectLoop(taskID string) bool {
	cg.history[taskID] = nil // not yet implemented; placeholder
	return false
}

func (cg *ConvergenceGate) record(taskID, stage, outcome string) {
	cg.history[taskID] = append(cg.history[taskID], convergenceEntry{stage, outcome})
	if len(cg.history[taskID]) > cg.config.LoopDetectionWindow {
		cg.history[taskID] = cg.history[taskID][1:]
	}
}
