package checkpoint

import (
	"bytes"
	"encoding/json"
	"sort"

	"github.com/G226Medi03/MedicalEntityParser/llmdto"
	"github.com/G226Medi03/MedicalEntityParser/llmparser"
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"github.com/G226Medi03/MorbesLLMService"
	"github.com/G226Medi03/MorbesPipeline/internal/entity/patient"
)

type ResolvedCaseExtraction struct {
	Actors                  []llmdto.ActorItem            `json:"actors,omitempty"`
	Statuses                []repo.IEStatus               `json:"statuses,omitempty"`
	Actions                 []repo.IEAction               `json:"actions,omitempty"`
	Relations               []morbesllm.RelationDraft     `json:"relations,omitempty"`
	MedicalProcedureOptions []repo.MedicalProcedureOption `json:"medicalProcedureOptions,omitempty"`
	PatientDescriptions     []patient.PatientDescription  `json:"patientDescriptions,omitempty"`
}

func (value ResolvedCaseExtraction) MarshalJSON() ([]byte, error) {
	statuses := make([]json.RawMessage, 0, len(value.Statuses))
	for _, status := range value.Statuses {
		encoded, err := json.Marshal(llmparser.StatusToDTO(status))
		if err != nil {
			return nil, err
		}
		statuses = append(statuses, encoded)
	}
	actions := make([]json.RawMessage, 0, len(value.Actions))
	for _, action := range value.Actions {
		encoded, err := json.Marshal(llmparser.ActionToDTO(action))
		if err != nil {
			return nil, err
		}
		actions = append(actions, encoded)
	}
	return json.Marshal(struct {
		Actors                  []llmdto.ActorItem            `json:"actors,omitempty"`
		Statuses                []json.RawMessage             `json:"statuses,omitempty"`
		Actions                 []json.RawMessage             `json:"actions,omitempty"`
		Relations               []morbesllm.RelationDraft     `json:"relations,omitempty"`
		MedicalProcedureOptions []repo.MedicalProcedureOption `json:"medicalProcedureOptions,omitempty"`
		PatientDescriptions     []patient.PatientDescription  `json:"patientDescriptions,omitempty"`
	}{value.Actors, statuses, actions, value.Relations, value.MedicalProcedureOptions, value.PatientDescriptions})
}

func (value *ResolvedCaseExtraction) UnmarshalJSON(data []byte) error {
	aux := struct {
		Actors                  []llmdto.ActorItem            `json:"actors,omitempty"`
		Statuses                json.RawMessage               `json:"statuses,omitempty"`
		Actions                 json.RawMessage               `json:"actions,omitempty"`
		Relations               json.RawMessage               `json:"relations,omitempty"`
		MedicalProcedureOptions []repo.MedicalProcedureOption `json:"medicalProcedureOptions,omitempty"`
		PatientDescriptions     []patient.PatientDescription  `json:"patientDescriptions,omitempty"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	statuses, err := decodeStatuses(aux.Statuses)
	if err != nil {
		return err
	}
	actions, err := decodeActions(aux.Actions)
	if err != nil {
		return err
	}
	relations, err := decodeRelations(aux.Relations)
	if err != nil {
		return err
	}
	value.Actors = aux.Actors
	value.Statuses = statuses
	value.Actions = actions
	value.Relations = relations
	value.MedicalProcedureOptions = aux.MedicalProcedureOptions
	value.PatientDescriptions = aux.PatientDescriptions
	return nil
}

func decodeStatuses(raw json.RawMessage) ([]repo.IEStatus, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || string(trimmed) == "null" {
		return nil, nil
	}
	if trimmed[0] == '[' {
		var items []json.RawMessage
		if err := json.Unmarshal(trimmed, &items); err != nil {
			return nil, err
		}
		result := make([]repo.IEStatus, 0, len(items))
		for _, item := range items {
			var dto llmdto.Status
			if err := json.Unmarshal(item, &dto); err != nil {
				return nil, err
			}
			status, err := llmparser.StatusFromDTO(dto)
			if err != nil {
				return nil, err
			}
			result = append(result, status)
		}
		return result, nil
	}
	var items map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &items); err != nil {
		return nil, err
	}
	keys := sortedKeys(items)
	result := make([]repo.IEStatus, 0, len(keys))
	for _, key := range keys {
		var dto llmdto.Status
		if err := json.Unmarshal(items[key], &dto); err != nil {
			return nil, err
		}
		status, err := llmparser.StatusFromDTO(dto)
		if err != nil {
			return nil, err
		}
		result = append(result, status)
	}
	return result, nil
}

func decodeActions(raw json.RawMessage) ([]repo.IEAction, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || string(trimmed) == "null" {
		return nil, nil
	}
	if trimmed[0] == '[' {
		var items []json.RawMessage
		if err := json.Unmarshal(trimmed, &items); err != nil {
			return nil, err
		}
		result := make([]repo.IEAction, 0, len(items))
		for _, item := range items {
			var dto llmdto.Action
			if err := json.Unmarshal(item, &dto); err != nil {
				return nil, err
			}
			action, err := llmparser.ActionFromDTO(dto)
			if err != nil {
				return nil, err
			}
			result = append(result, action)
		}
		return result, nil
	}
	var items map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &items); err != nil {
		return nil, err
	}
	keys := sortedKeys(items)
	result := make([]repo.IEAction, 0, len(keys))
	for _, key := range keys {
		var dto llmdto.Action
		if err := json.Unmarshal(items[key], &dto); err != nil {
			return nil, err
		}
		action, err := llmparser.ActionFromDTO(dto)
		if err != nil {
			return nil, err
		}
		result = append(result, action)
	}
	return result, nil
}

func decodeRelations(raw json.RawMessage) ([]morbesllm.RelationDraft, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || string(trimmed) == "null" {
		return nil, nil
	}
	var result []morbesllm.RelationDraft
	if err := json.Unmarshal(trimmed, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func sortedKeys(values map[string]json.RawMessage) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
