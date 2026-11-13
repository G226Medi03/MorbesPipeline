package csituation

import (
	"encoding/json"
	"fmt"

	"github.com/G226Medi03/MedicalEntityParser/parser"
	entity "github.com/G226Medi03/MorbesPipeline/legacy"
)

// ToStatuses converts findings.status into observable status DTOs.
func (s Situation) ToStatuses() ([]entity.Status, error) {
	result := make([]entity.Status, 0, len(s.Findings.Status))
	for index, raw := range s.Findings.Status {
		status, err := status_from_raw(raw)
		if err != nil {
			return nil, fmt.Errorf("findings.status[%d]: %w", index, err)
		}
		result = append(result, status)
	}
	return result, nil
}

func status_from_raw(raw json.RawMessage) (entity.Status, error) {
	payload := map[string]any{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return entity.NewStatus("", nil, nil), fmt.Errorf("status JSON 파싱 실패: %w", err)
	}
	kind, _ := payload["kind"].(string)
	if kind == "" {
		return entity.NewStatus("", nil, nil), fmt.Errorf("status.kind 가 비어 있습니다")
	}
	delete(payload, "kind")

	normalized, err := parser.NormalizeStatus(kind, payload)
	if err != nil {
		return entity.NewStatus("", nil, nil), err
	}
	return entity.NewStatus(normalized.Kind, normalized.Value, nil), nil
}
