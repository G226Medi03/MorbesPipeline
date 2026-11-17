// Package medicalentityparser adapts MedicalEntityParser output to the typed
// entity values used by the application.
package medicalentityparser

import (
	"fmt"

	"github.com/G226Medi03/MedicalEntityParser/llmparser"
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
)

// DecodeSingleStatus decodes a raw LLM response that must contain exactly one
// status and converts it to a typed status.
func DecodeSingleStatus(raw string) (repo.IEStatus, error) {
	statuses, err := llmparser.DecodeStatuses(raw)
	if err != nil {
		return nil, err
	}
	if len(statuses) != 1 {
		return nil, fmt.Errorf("status response must contain exactly one status: got %d", len(statuses))
	}
	for _, status := range statuses {
		return status, nil
	}
	return nil, fmt.Errorf("status response is empty")
}
