package csituation

import (
	"encoding/json"
	"fmt"

	"github.com/G226Medi03/MedicalEntityParser/llmparser"
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
)

func status_from_raw(raw json.RawMessage) (repo.IEStatus, error) {
	var payload struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("status JSON 파싱 실패: %w", err)
	}
	if payload.Kind == "" {
		return nil, fmt.Errorf("status.kind 가 비어 있습니다")
	}
	statuses, err := llmparser.DecodeStatuses(string(raw))
	if err != nil {
		return nil, err
	}
	if len(statuses) != 1 {
		return nil, fmt.Errorf("status response must contain exactly one status: got %d", len(statuses))
	}
	for _, value := range statuses {
		return value, nil
	}
	return nil, fmt.Errorf("status response is empty")
}
