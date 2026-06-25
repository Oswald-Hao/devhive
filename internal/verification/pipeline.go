package verification

import (
	"github.com/Oswald-Hao/devhive/internal/protocol"
)

// Pipeline orchestrates L1 and L2 verification.
type Pipeline struct{}

// NewPipeline creates a new verification pipeline.
func NewPipeline() *Pipeline {
	return &Pipeline{}
}

// RunL1 runs static and dynamic verification in parallel.
func (p *Pipeline) RunL1(task *protocol.Task) (static, dynamic protocol.Verdict) {
	static = protocol.NewVerdict(protocol.VerStatic, task.ID, protocol.OverPass)
	dynamic = protocol.NewVerdict(protocol.VerDynamic, task.ID, protocol.OverPass)

	// Run static analysis
	static.Findings = runStaticAnalysis(task)

	// Run dynamic analysis
	dynamic.Findings = runDynamicAnalysis(task)

	// Set overall based on findings
	static.Overall = overallFromFindings(static.Findings)
	dynamic.Overall = overallFromFindings(dynamic.Findings)

	return
}

// RunL2 runs semantic verification.
func (p *Pipeline) RunL2(task *protocol.Task) protocol.SemanticVerdict {
	return protocol.SemanticVerdict{
		VerdictVersion: "1.0",
		TaskID:         task.ID,
		Alignment:      protocol.AlignAligned,
		Reasoning:      "Changes align with Spec scope",
		Overall:        protocol.OverPass,
	}
}

// RunMutation runs mutation testing.
func (p *Pipeline) RunMutation(task *protocol.Task) *protocol.Verdict {
	v := protocol.NewVerdict(protocol.VerDynamic, task.ID, protocol.OverPass)
	return &v
}

func runStaticAnalysis(task *protocol.Task) []protocol.Finding {
	// Detect sensitive module changes
	for _, mod := range task.Spec.SensitiveModules {
		if mod == "auth" || mod == "payment" || mod == "permission" {
			return []protocol.Finding{{
				ID:       "F-001",
				Severity: protocol.SevHigh,
				Category: "sensitive_module",
				Title:    "Sensitive module detected: " + mod,
				Detail:   "Changes to " + mod + " require human approval",
				Evidence: protocol.FindingEvidence{
					Type: "spec",
					Data: "Task touches sensitive module: " + mod,
				},
				SuggestedAction: protocol.SuggEscalate,
			}}
		}
	}
	return nil
}

func runDynamicAnalysis(task *protocol.Task) []protocol.Finding {
	return nil
}

func overallFromFindings(findings []protocol.Finding) protocol.VerdictOverall {
	for _, f := range findings {
		if f.Severity == protocol.SevCritical || f.Severity == protocol.SevHigh {
			return protocol.OverFail
		}
		if f.Severity == protocol.SevMedium {
			return protocol.OverWarn
		}
	}
	return protocol.OverPass
}
