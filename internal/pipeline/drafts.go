package csolve

import (
	"encoding/json"

	morbesllm "github.com/G226Medi03/MorbesLLmPipeline"
	source "github.com/G226Medi03/MorbesPipeline/internal/external/sourceentity"
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
