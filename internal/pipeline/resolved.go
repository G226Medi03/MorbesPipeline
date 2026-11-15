package csolve

import (
	"encoding/json"
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	entityjson "github.com/G226Medi03/MorbesPipeline/internal/domain/entityjson"

	"github.com/G226Medi03/MorbesPipeline/internal/domain/patient"
	medicalprocedureoption "github.com/G226Medi03/MorbesPipeline/internal/domain/procedureoption"
)

type ResolvedCaseExtraction struct {
	Actors                  []ActorDraft                                    `json:"actors,omitempty"`
	Statuses                map[string]repo.IEStatus                        `json:"statuses,omitempty"`
	Actions                 map[string]repo.IEAction                        `json:"actions,omitempty"`
	Relations               []RelationDraft                                 `json:"relations,omitempty"`
	MedicalProcedureOptions []medicalprocedureoption.MedicalProcedureOption `json:"medicalProcedureOptions,omitempty"`
	PatientDescriptions     []patient.PatientDescription                    `json:"patientDescriptions,omitempty"`
}

func (value ResolvedCaseExtraction) MarshalJSON() ([]byte, error) {
	statuses := make(map[string]json.RawMessage, len(value.Statuses))
	for key, status := range value.Statuses {
		encoded, err := entityjson.EncodeStatus(status)
		if err != nil {
			return nil, err
		}
		statuses[key] = encoded
	}
	actions := make(map[string]json.RawMessage, len(value.Actions))
	for key, action := range value.Actions {
		encoded, err := entityjson.EncodeAction(action)
		if err != nil {
			return nil, err
		}
		actions[key] = encoded
	}
	return json.Marshal(struct {
		Actors                  []ActorDraft                                    `json:"actors,omitempty"`
		Statuses                map[string]json.RawMessage                      `json:"statuses,omitempty"`
		Actions                 map[string]json.RawMessage                      `json:"actions,omitempty"`
		Relations               []RelationDraft                                 `json:"relations,omitempty"`
		MedicalProcedureOptions []medicalprocedureoption.MedicalProcedureOption `json:"medicalProcedureOptions,omitempty"`
		PatientDescriptions     []patient.PatientDescription                    `json:"patientDescriptions,omitempty"`
	}{value.Actors, statuses, actions, value.Relations, value.MedicalProcedureOptions, value.PatientDescriptions})
}

func (value *ResolvedCaseExtraction) UnmarshalJSON(data []byte) error {
	aux := struct {
		Actors                  []ActorDraft                                    `json:"actors,omitempty"`
		Statuses                map[string]json.RawMessage                      `json:"statuses,omitempty"`
		Actions                 map[string]json.RawMessage                      `json:"actions,omitempty"`
		Relations               []RelationDraft                                 `json:"relations,omitempty"`
		MedicalProcedureOptions []medicalprocedureoption.MedicalProcedureOption `json:"medicalProcedureOptions,omitempty"`
		PatientDescriptions     []patient.PatientDescription                    `json:"patientDescriptions,omitempty"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	value.Actors = aux.Actors
	value.Relations = aux.Relations
	value.MedicalProcedureOptions = aux.MedicalProcedureOptions
	value.PatientDescriptions = aux.PatientDescriptions
	value.Statuses = make(map[string]repo.IEStatus, len(aux.Statuses))
	for key, raw := range aux.Statuses {
		status, err := entityjson.DecodeStatus(raw)
		if err != nil {
			return err
		}
		value.Statuses[key] = status
	}
	value.Actions = make(map[string]repo.IEAction, len(aux.Actions))
	for key, raw := range aux.Actions {
		action, err := entityjson.DecodeAction(raw)
		if err != nil {
			return err
		}
		value.Actions[key] = action
	}
	return nil
}
