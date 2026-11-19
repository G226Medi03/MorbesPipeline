// Package csituation parses and converts structured medical situations.
package csituation

import (
	"fmt"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
)

type Situation struct {
	Findings Findings `json:"findings"`
}

// ToStatuses converts findings.status into typed observable statuses.
func (s Situation) ToStatuses() ([]repo.IEStatus, error) {
	result := make([]repo.IEStatus, 0, len(s.Findings.Status))
	for index, raw := range s.Findings.Status {
		status, err := status_from_raw(raw)
		if err != nil {
			return nil, fmt.Errorf("findings.status[%d]: %w", index, err)
		}
		result = append(result, status)
	}
	return result, nil
}
