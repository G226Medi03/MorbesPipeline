// Package patient contains structured descriptions of what a patient says or
// understands.
package patient

import (
	"encoding/json"
	"fmt"
	"github.com/G226Medi03/MedicalEntityParser/llmparser"
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	entityjson "github.com/G226Medi03/MorbesPipeline/internal/domain/entityjson"
)

// PatientDescription is the minimal patient-reported context for one answer
// choice or for a narrative patient statement.
type PatientDescription struct {
	Choice      string          `json:"choice,omitempty"`
	Description []repo.IEStatus `json:"description"`
	Knowledge   []string        `json:"knowledge,omitempty"`
}

func (value PatientDescription) MarshalJSON() ([]byte, error) {
	descriptions := make([]json.RawMessage, 0, len(value.Description))
	for _, status := range value.Description {
		encoded, err := entityjson.EncodeStatus(status)
		if err != nil {
			return nil, err
		}
		descriptions = append(descriptions, encoded)
	}
	return json.Marshal(struct {
		Choice      string            `json:"choice,omitempty"`
		Description []json.RawMessage `json:"description"`
		Knowledge   []string          `json:"knowledge,omitempty"`
	}{value.Choice, descriptions, value.Knowledge})
}

func (value *PatientDescription) UnmarshalJSON(data []byte) error {
	aux := struct {
		Choice      string            `json:"choice,omitempty"`
		Description []json.RawMessage `json:"description"`
		Knowledge   []string          `json:"knowledge,omitempty"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	value.Choice = aux.Choice
	value.Knowledge = aux.Knowledge
	value.Description = make([]repo.IEStatus, 0, len(aux.Description))
	for _, raw := range aux.Description {
		status, err := entityjson.DecodeStatus(raw)
		if err != nil {
			return err
		}
		value.Description = append(value.Description, status)
	}
	return nil
}

// ParseStatus converts an LLM status object into a typed status.
func ParseStatus(raw json.RawMessage) (repo.IEStatus, error) {
	var payload struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("patient description status JSON 파싱 실패: %w", err)
	}
	if payload.Kind == "" {
		return nil, fmt.Errorf("patient description status.kind 가 비어 있습니다")
	}
	statuses, err := llmparser.DecodeStatuses(string(raw))
	if err != nil {
		return nil, fmt.Errorf("patient description status 정규화 실패: %w", err)
	}
	if len(statuses) != 1 {
		return nil, fmt.Errorf("patient description status response must contain exactly one status: got %d", len(statuses))
	}
	for _, value := range statuses {
		return value, nil
	}
	return nil, fmt.Errorf("patient description status response is empty")
}
