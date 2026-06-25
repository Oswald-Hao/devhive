package verification

import (
	"github.com/Oswald-Hao/devhive/internal/protocol"
)

// DiagnosticAggregator merges multiple Verifier outputs using deterministic rules.
// This is NOT an LLM — it's a pure rule engine for predictable, auditable decisions.
type DiagnosticAggregator struct{}

// NewDiagnosticAggregator creates a new diagnostic aggregator.
func NewDiagnosticAggregator() *DiagnosticAggregator {
	return &DiagnosticAggregator{}
}

// AggregatorResult is the output of verdict aggregation.
type AggregatorResult struct {
	Action      protocol.ConcurrencyAction
	Reason      string
	FixStrategy string
	NeedsHuman  bool
}

// AggregateL1 merges Static + Dynamic verdicts.
func (da *DiagnosticAggregator) AggregateL1(static, dynamic protocol.Verdict, taskID string) AggregatorResult {
	// Both PASS → advance
	if static.Overall == protocol.OverPass && dynamic.Overall == protocol.OverPass {
		return AggregatorResult{
			Action: protocol.ActionPass,
			Reason: "L1: Static and Dynamic verification both passed",
		}
	}

	// Collect all findings
	allFindings := append(static.Findings, dynamic.Findings...)

	// Check for known-signature matches
	var matchedFindings []protocol.Finding
	var criticalFindings []protocol.Finding
	for _, f := range allFindings {
		if f.MatchedSignature != "" {
			matchedFindings = append(matchedFindings, f)
		}
		if f.Severity == protocol.SevCritical || f.Severity == protocol.SevHigh {
			criticalFindings = append(criticalFindings, f)
		}
	}

	if len(matchedFindings) > 0 {
		return AggregatorResult{
			Action:      protocol.ActionFix,
			Reason:      "Known failure signatures matched — auto-fix available",
			FixStrategy: matchedFindings[0].MatchedSignature,
		}
	}

	if len(criticalFindings) > 0 {
		return AggregatorResult{
			Action:     protocol.ActionEscalate,
			Reason:     "Critical findings with no known fix",
			NeedsHuman: true,
		}
	}

	// WARN/LOW only → pass with caution
	return AggregatorResult{
		Action: protocol.ActionPass,
		Reason: "L1: No critical findings, passing with warnings",
	}
}

// AggregateL2 merges L1 + Semantic + Mutation verdicts.
func (da *DiagnosticAggregator) AggregateL2(static, dynamic protocol.Verdict, semantic protocol.SemanticVerdict, mutation *protocol.Verdict, taskID string) AggregatorResult {
	// Check L1 first
	l1 := da.AggregateL1(static, dynamic, taskID)
	if l1.Action != protocol.ActionPass {
		return l1
	}

	// Check semantic alignment
	switch semantic.Alignment {
	case protocol.AlignDeviated:
		return AggregatorResult{
			Action:     protocol.ActionEscalate,
			Reason:     "L2: Semantic alignment DEVIATED — " + semantic.Reasoning,
			NeedsHuman: true,
		}
	case protocol.AlignConflict:
		return AggregatorResult{
			Action:     protocol.ActionConflict,
			Reason:     "L2: Semantic CONFLICT — " + semantic.Reasoning,
			NeedsHuman: true,
		}
	case protocol.AlignEnhanced:
		return AggregatorResult{
			Action: protocol.ActionPass,
			Reason: "L2: Changes exceed Spec but are reasonable enhancements",
		}
	}

	// Check mutation
	if mutation != nil && mutation.Overall == protocol.OverFail {
		return AggregatorResult{
			Action:      protocol.ActionFix,
			Reason:      "L2: Mutation testing found coverage gaps",
			FixStrategy: "ADD_TESTS",
		}
	}

	return AggregatorResult{
		Action: protocol.ActionPass,
		Reason: "L2: All verification layers passed",
	}
}
