package csituation

import (
	"encoding/json"
	"fmt"
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"

	"github.com/G226Medi03/MedicalEntityParser/parser"
)

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

func status_from_raw(raw json.RawMessage) (repo.IEStatus, error) {
	payload := map[string]any{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("status JSON 파싱 실패: %w", err)
	}
	if kind, _ := payload["kind"].(string); kind == "" {
		return nil, fmt.Errorf("status.kind 가 비어 있습니다")
	}
	value, err := parser.ParseStatus(payload)
	if err != nil {
		return nil, err
	}
	return value, nil
}
