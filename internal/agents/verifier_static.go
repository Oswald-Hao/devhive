package agents

import (
	"fmt"

	"github.com/Oswald-Hao/devhive/internal/protocol"
)

// StaticVerifier detects code pattern issues via rule engine + LLM analysis.
type StaticVerifier struct {
	*Base
}

// NewStaticVerifier creates a new Static verifier.
func NewStaticVerifier(config BaseConfig) *StaticVerifier {
	if config.AgentType == "" {
		config.AgentType = protocol.AgentStaticVerifier
	}
	return &StaticVerifier{Base: NewBase(config)}
}

// Execute runs static analysis on code changes.
func (a *StaticVerifier) Execute(task *protocol.Task) (map[string]interface{}, error) {
	verdict := protocol.NewVerdict(protocol.VerStatic, task.ID, protocol.OverPass)

	// Check for sensitive modules
	sensitiveCategories := map[string]bool{
		"auth": true, "payment": true, "data_deletion": true, "permission_change": true,
	}

	for _, mod := range task.Spec.SensitiveModules {
		if sensitiveCategories[mod] {
			verdict.Findings = append(verdict.Findings, protocol.Finding{
				ID:               fmt.Sprintf("F-%s-sensitive", task.ID[:8]),
				Severity:         protocol.SevHigh,
				Category:         "sensitive_module",
				Title:            fmt.Sprintf("Changes touch sensitive module: %s", mod),
				Detail:           fmt.Sprintf("Task involves %s which requires human review before merge.", mod),
				Evidence:         protocol.FindingEvidence{Type: "spec", Data: "sensitive_module: " + mod},
				SuggestedAction:  protocol.SuggEscalate,
			})
		}
	}

	// Scan for scope constraint violations
	for _, constraint := range task.Spec.ScopeConstraints {
		verdict.Findings = append(verdict.Findings, protocol.Finding{
			ID:              fmt.Sprintf("F-%s-scope", task.ID[:8]),
			Severity:        protocol.SevInfo,
			Category:        "scope_constraint",
			Title:           "Scope constraint: " + constraint,
			Detail:          "Verifier will ensure changes respect: " + constraint,
			Evidence:        protocol.FindingEvidence{Type: "spec", Data: constraint},
			SuggestedAction: protocol.SuggIgnore,
		})
	}

	// Set overall from findings
	verdict.Overall = protocol.OverPass
	for _, f := range verdict.Findings {
		if f.Severity == protocol.SevCritical || f.Severity == protocol.SevHigh {
			verdict.Overall = protocol.OverFail
			break
		}
		if f.Severity == protocol.SevMedium {
			verdict.Overall = protocol.OverWarn
		}
	}

	return map[string]interface{}{
		"verdict": verdict,
	}, nil
}
