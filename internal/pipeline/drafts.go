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
	Key      string
	Value    repo.IEStatus
	SourceID source.SourceId
}

type ActionDraft struct {
	Key      string
	Value    repo.IEAction
	SourceID source.SourceId
}

type RelationDraft struct {
	CauseStatusKey  string          `json:"causeStatusKey"`
	ActionKey       string          `json:"actionKey"`
	ResultStatusKey string          `json:"resultStatusKey"`
	SourceID        source.SourceId `json:"sourceId,omitempty"`
}

type OptionDraft struct {
	Key              string
	Category         string
	Traits           []string
	MedicalProcedure repo.IMedicalProcedure
	Indications      []string
	ExpectedEffects  []string
	SourceID         source.SourceId
}

type PatientDescriptionDraft struct {
	Choice      string          `json:"choice,omitempty"`
	Description []StatusDraft   `json:"description"`
	Knowledge   []string        `json:"knowledge,omitempty"`
	SourceID    source.SourceId `json:"sourceId,omitempty"`
}

type statusDraftJSON struct {
	Key      string          `json:"key"`
	Kind     string          `json:"kind"`
	Value    json.RawMessage `json:"value"`
	SourceID source.SourceId `json:"sourceId,omitempty"`
}

func (draft StatusDraft) MarshalJSON() ([]byte, error) {
	value, err := entityjson.EncodeStatus(draft.Value)
	if err != nil {
		return nil, err
	}
	return json.Marshal(statusDraftJSON{
		Key:      draft.Key,
		Kind:     draft.Value.Kind(),
		Value:    value,
		SourceID: draft.SourceID,
	})
}

func (draft *StatusDraft) UnmarshalJSON(data []byte) error {
	var encoded statusDraftJSON
	if err := json.Unmarshal(data, &encoded); err != nil {
		return err
	}
	value, err := entityjson.DecodeStatus(encoded.Value)
	if err != nil {
		return err
	}
	draft.Key = encoded.Key
	draft.Value = value
	draft.SourceID = encoded.SourceID
	return nil
}

type actionDraftJSON struct {
	Key      string          `json:"key"`
	Variant  string          `json:"variant"`
	Value    json.RawMessage `json:"value"`
	SourceID source.SourceId `json:"sourceId,omitempty"`
}

func (draft ActionDraft) MarshalJSON() ([]byte, error) {
	value, err := entityjson.EncodeAction(draft.Value)
	if err != nil {
		return nil, err
	}
	return json.Marshal(actionDraftJSON{
		Key:      draft.Key,
		Variant:  draft.Value.Kind(),
		Value:    value,
		SourceID: draft.SourceID,
	})
}

func (draft *ActionDraft) UnmarshalJSON(data []byte) error {
	var encoded actionDraftJSON
	if err := json.Unmarshal(data, &encoded); err != nil {
		return err
	}
	value, err := entityjson.DecodeAction(encoded.Value)
	if err != nil {
		return err
	}
	draft.Key = encoded.Key
	draft.Value = value
	draft.SourceID = encoded.SourceID
	return nil
}

type optionDraftJSON struct {
	Key              string          `json:"key"`
	Category         string          `json:"category"`
	Traits           []string        `json:"traits"`
	MedicalProcedure json.RawMessage `json:"medicalProcedure"`
	Indications      []string        `json:"indications"`
	ExpectedEffects  []string        `json:"expectedEffects,omitempty"`
	SourceID         source.SourceId `json:"sourceId,omitempty"`
}

func (draft OptionDraft) MarshalJSON() ([]byte, error) {
	value, err := entityjson.EncodeProcedure(draft.MedicalProcedure)
	if err != nil {
		return nil, err
	}
	return json.Marshal(optionDraftJSON{
		Key:              draft.Key,
		Category:         draft.Category,
		Traits:           draft.Traits,
		MedicalProcedure: value,
		Indications:      draft.Indications,
		ExpectedEffects:  draft.ExpectedEffects,
		SourceID:         draft.SourceID,
	})
}

func (draft *OptionDraft) UnmarshalJSON(data []byte) error {
	var encoded optionDraftJSON
	if err := json.Unmarshal(data, &encoded); err != nil {
		return err
	}
	value, err := entityjson.DecodeProcedure(encoded.MedicalProcedure)
	if err != nil {
		return err
	}
	draft.Key = encoded.Key
	draft.Category = encoded.Category
	draft.Traits = encoded.Traits
	draft.MedicalProcedure = value
	draft.Indications = encoded.Indications
	draft.ExpectedEffects = encoded.ExpectedEffects
	draft.SourceID = encoded.SourceID
	return nil
}
