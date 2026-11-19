package csolve

import (
	"encoding/json"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	morbesllm "github.com/G226Medi03/MorbesLLMService"
	entityjson "github.com/G226Medi03/MorbesPipeline/internal/domain/entityjson"
	source "github.com/G226Medi03/SourceEntity"
)

// CaseExtraction is the single structured output of the solve v2 extraction
// stage. Draft keys are resolved to persisted IDs by the resolve stage.
type CaseExtraction struct {
	PatientDescription      *PatientDescriptionDraft  `json:"patientDescription,omitempty"`
	PatientDescriptions     []PatientDescriptionDraft `json:"patientDescriptions,omitempty"`
	Actors                  []ActorDraft              `json:"actors,omitempty"`
	Statuses                []StatusDraft             `json:"statuses,omitempty"`
	Actions                 []ActionDraft             `json:"actions,omitempty"`
	Relations               []RelationDraft           `json:"relations,omitempty"`
	MedicalProcedureOptions []OptionDraft             `json:"medicalProcedureOptions,omitempty"`
}

type ActorDraft = morbesllm.ActorDraft

type StatusDraft struct {
	Key      string          `json:"key"`
	Kind     string          `json:"kind"`
	Value    json.RawMessage `json:"value"`
	SourceID source.SourceId `json:"sourceId,omitempty"`
}

type ActionDraft struct {
	Key      string          `json:"key"`
	Variant  string          `json:"variant"`
	Value    json.RawMessage `json:"value"`
	SourceID source.SourceId `json:"sourceId,omitempty"`
}

type RelationDraft struct {
	CauseStatusKey  string          `json:"causeStatusKey"`
	ActionKey       string          `json:"actionKey"`
	ResultStatusKey string          `json:"resultStatusKey"`
	SourceID        source.SourceId `json:"sourceId,omitempty"`
}

type OptionDraft struct {
	Key              string          `json:"key"`
	Category         string          `json:"category"`
	Traits           []string        `json:"traits"`
	MedicalProcedure json.RawMessage `json:"medicalProcedure"`
	Indications      []string        `json:"indications"`
	ExpectedEffects  []string        `json:"expectedEffects,omitempty"`
	SourceID         source.SourceId `json:"sourceId,omitempty"`
}

type PatientDescriptionDraft struct {
	Choice      string          `json:"choice,omitempty"`
	Description []StatusDraft   `json:"description"`
	Knowledge   []string        `json:"knowledge,omitempty"`
	SourceID    source.SourceId `json:"sourceId,omitempty"`
}

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
	return entityjson.DecodeStatus(encoded)
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
