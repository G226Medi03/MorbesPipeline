package csolve

import (
	"encoding/json"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	morbesllm "github.com/G226Medi03/MorbesLLMService"
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
	return morbesllm.ParseStatusResponse(string(encoded))
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
