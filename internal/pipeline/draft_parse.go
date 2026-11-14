package csolve

import (
	"encoding/json"
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"

	"github.com/G226Medi03/MedicalEntityParser/parser"
)

func parseStatusDraft(draft StatusDraft) (repo.IEStatus, error) {
	payload := map[string]any{}
	if len(draft.Value) > 0 {
		if err := json.Unmarshal(draft.Value, &payload); err != nil {
			return nil, err
		}
	}
	payload["kind"] = draft.Kind
	return parser.ParseStatus(payload)
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
	return repo.UnmarshalAction(encoded)
}
