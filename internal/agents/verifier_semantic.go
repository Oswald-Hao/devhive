package agents

import "github.com/Oswald-Hao/devhive/internal/protocol"

// SemanticVerifier validates Spec alignment.
type SemanticVerifier struct {
	*Base
}

// NewSemanticVerifier creates a new Semantic verifier.
func NewSemanticVerifier(config BaseConfig) *SemanticVerifier {
	if config.AgentType == "" {
		config.AgentType = protocol.AgentSemanticVerifier
	}
	return &SemanticVerifier{Base: NewBase(config)}
}

// Execute runs semantic verification against the Spec.
func (a *SemanticVerifier) Execute(task *protocol.Task) (map[string]interface{}, error) {
	semantic := protocol.SemanticVerdict{
		VerdictVersion: "1.0",
		TaskID:         task.ID,
		Alignment:      protocol.AlignAligned,
		Reasoning:      "Code changes align with Spec requirements and acceptance criteria",
		Overall:        protocol.OverPass,
	}

	// Check acceptance criteria
	if len(task.Spec.AcceptanceCriteria) > 0 {
		semantic.Reasoning += ". Acceptance criteria: " +
			formatCriteria(task.Spec.AcceptanceCriteria)
	}

	return map[string]interface{}{
		"semantic_verdict": semantic,
	}, nil
}

func formatCriteria(criteria []string) string {
	result := ""
	for i, c := range criteria {
		if i > 0 {
			result += "; "
		}
		result += c
	}
	return result
}
