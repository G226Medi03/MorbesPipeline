package csolve

import (
	"encoding/json"
	"fmt"

	"github.com/G226Medi03/MedicalEntityParser/llmparser"
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	entityjson "github.com/G226Medi03/MorbesPipeline/internal/domain/entityjson"
)

func parseStatusDraft(draft StatusDraft) (repo.IEStatus, error) {
	payload := map[string]any{}
	if len(draft.Value) > 0 {
		if err := json.Unmarshal(draft.Value, &payload); err != nil {
			return nil, err
		}
	}
	payload["kind"] = draft.Kind
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	statuses, err := llmparser.DecodeStatuses(string(encoded))
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

func parseActionDraft(draft ActionDraft) (repo.IEAction, error) {
	payload := map[string]any{}
	if len(draft.Value) > 0 {
		if err := json.Unmarshal(draft.Value, &payload); err != nil {
			return nil, err
		}
	}
	payload["kind"] = draft.Variant
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return entityjson.DecodeAction(encoded)
}
