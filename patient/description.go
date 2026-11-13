// Package patient contains structured descriptions of what a patient says or
// understands.
package patient

import (
	"encoding/json"
	"fmt"

	"github.com/G226Medi03/MedicalEntityParser/parser"
	entity "github.com/G226Medi03/MorbesPipeline/legacy"
)

// PatientDescription is the minimal patient-reported context for one answer
// choice or for a narrative patient statement.
type PatientDescription struct {
	Choice      string          `json:"choice,omitempty"`
	Description []entity.Status `json:"description"`
	Knowledge   []string        `json:"knowledge,omitempty"`
}

// ParseStatus converts an LLM status object into an entity.Status. It performs
// the same normalization as the situation and status extractors, but does not
// persist the status.
func ParseStatus(raw json.RawMessage) (entity.Status, error) {
	payload := map[string]any{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return entity.NewStatus("", nil, nil), fmt.Errorf("patient description status JSON 파싱 실패: %w", err)
	}
	kind, _ := payload["kind"].(string)
	if kind == "" {
		return entity.NewStatus("", nil, nil), fmt.Errorf("patient description status.kind 가 비어 있습니다")
	}
	delete(payload, "kind")

	normalized, err := parser.NormalizeStatus(kind, payload)
	if err != nil {
		return entity.NewStatus("", nil, nil), fmt.Errorf("patient description status 정규화 실패: %w", err)
	}
	return entity.NewStatus(normalized.Kind, normalized.Value, nil), nil
}
