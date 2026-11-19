package checkpoint

import (
	"encoding/json"

	"github.com/G226Medi03/MedicalEntityParser/llmdto"
	"github.com/G226Medi03/MedicalEntityParser/llmparser"
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"github.com/G226Medi03/MorbesLLMService"
	"github.com/G226Medi03/MorbesPipeline/internal/domain/patient"
)

type ResolvedCaseExtraction struct {
	Actors                  []llmdto.ActorItem            `json:"actors,omitempty"`
	Statuses                map[string]repo.IEStatus      `json:"statuses,omitempty"`
	Actions                 map[string]repo.IEAction      `json:"actions,omitempty"`
	Relations               []morbesllm.RelationDraft     `json:"relations,omitempty"`
	MedicalProcedureOptions []repo.MedicalProcedureOption `json:"medicalProcedureOptions,omitempty"`
	PatientDescriptions     []patient.PatientDescription  `json:"patientDescriptions,omitempty"`
}

func (value ResolvedCaseExtraction) MarshalJSON() ([]byte, error) {
	statuses := make(map[string]json.RawMessage, len(value.Statuses))
	for key, status := range value.Statuses {
		encoded, err := json.Marshal(llmparser.StatusToDTO(status))
		if err != nil {
			return nil, err
		}
		statuses[key] = encoded
	}
	actions := make(map[string]json.RawMessage, len(value.Actions))
	for key, action := range value.Actions {
		encoded, err := json.Marshal(llmparser.ActionToDTO(action))
		if err != nil {
			return nil, err
		}
		actions[key] = encoded
	}
	return json.Marshal(struct {
		Actors                  []llmdto.ActorItem            `json:"actors,omitempty"`
		Statuses                map[string]json.RawMessage    `json:"statuses,omitempty"`
		Actions                 map[string]json.RawMessage    `json:"actions,omitempty"`
		Relations               []morbesllm.RelationDraft     `json:"relations,omitempty"`
		MedicalProcedureOptions []repo.MedicalProcedureOption `json:"medicalProcedureOptions,omitempty"`
		PatientDescriptions     []patient.PatientDescription  `json:"patientDescriptions,omitempty"`
	}{value.Actors, statuses, actions, value.Relations, value.MedicalProcedureOptions, value.PatientDescriptions})
}

func (value *ResolvedCaseExtraction) UnmarshalJSON(data []byte) error {
	aux := struct {
		Actors                  []llmdto.ActorItem            `json:"actors,omitempty"`
		Statuses                map[string]json.RawMessage    `json:"statuses,omitempty"`
		Actions                 map[string]json.RawMessage    `json:"actions,omitempty"`
		Relations               []morbesllm.RelationDraft     `json:"relations,omitempty"`
		MedicalProcedureOptions []repo.MedicalProcedureOption `json:"medicalProcedureOptions,omitempty"`
		PatientDescriptions     []patient.PatientDescription  `json:"patientDescriptions,omitempty"`
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
		var dto llmdto.Status
		if err := json.Unmarshal(raw, &dto); err != nil {
			return err
		}
		status, err := llmparser.StatusFromDTO(dto)
		if err != nil {
			return err
		}
		value.Statuses[key] = status
	}
	value.Actions = make(map[string]repo.IEAction, len(aux.Actions))
	for key, raw := range aux.Actions {
		var dto llmdto.Action
		if err := json.Unmarshal(raw, &dto); err != nil {
			return err
		}
		action, err := llmparser.ActionFromDTO(dto)
		if err != nil {
			return err
		}
		value.Actions[key] = action
	}
	return nil
}
