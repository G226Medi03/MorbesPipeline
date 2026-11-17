package csolve

import (
	"encoding/json"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	entityjson "github.com/G226Medi03/MorbesPipeline/internal/domain/entityjson"
	medicalentityparser "github.com/G226Medi03/MorbesPipeline/internal/external/medicalentityparser"
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
	value, err := medicalentityparser.DecodeSingleStatus(string(encoded))
	if err != nil {
		return nil, err
	}
	return value, nil
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
