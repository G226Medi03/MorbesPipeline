// Package ccause traces possible causes through the status/action graph.
package ccause

import (
	"encoding/json"
	"fmt"
	medicalcodec "github.com/G226Medi03/MedicalEntity/codec"
	"sort"
	"strings"

	eprocedure "github.com/G226Medi03/MedicalEntity/ename/emedical_procedure"
	centitylabel "github.com/G226Medi03/MorbesPipeline/label"
	entity "github.com/G226Medi03/MorbesPipeline/legacy"
	source "github.com/G226Medi03/SourceEntity"
)

func ObservedLabels(statuses []entity.Status) []string {
	result := make([]string, 0, len(statuses))
	for _, status := range statuses {
		data, err := json.Marshal(status.Value)
		if err != nil {
			continue
		}
		result = append(result, centitylabel.StatusLabel(status.Variant, string(data)))
	}
	return result
}

func append_unique_ids[T comparable](values []T, add ...T) []T {
	seen := make(map[T]struct{}, len(values))
	for _, v := range values {
		seen[v] = struct{}{}
	}
	for _, v := range add {
		if _, ok := seen[v]; !ok {
			values = append(values, v)
			seen[v] = struct{}{}
		}
	}
	return values
}
func action_key(kind, data string) string {
	var obj map[string]any
	if err := json.Unmarshal([]byte(data), &obj); err != nil {
		return kind + ":" + data
	}
	switch kind {
	case "cause_status":
		if status_id, ok := obj["status_id"]; ok {
			return kind + ":" + strings.TrimSpace(value_text(status_id))
		}
	case "block", "relax":
		if actor, ok := obj["actor"]; ok {
			return kind + ":" + strings.TrimSpace(value_text(actor))
		}
	}
	return kind + ":" + data
}

func status_key(kind, data string) string {
	var obj map[string]any
	if err := json.Unmarshal([]byte(data), &obj); err != nil {
		return kind + ":" + data
	}
	switch kind {
	case "received_medical_procedure":
		if raw, ok := obj["medical_procedure"]; ok {
			encoded, err := json.Marshal(raw)
			if err == nil {
				value, err := entity.UnmarshalProcedure(encoded)
				if err == nil {
					return kind + ":" + procedure_identity(value)
				}
			}
		}
		if name := strings.TrimSpace(value_text(obj["name"])); name != "" {
			return kind + ":" + name
		}
		return kind + ":procedure"
	case "have_symptom":
		symptom, _ := obj["symptom"].(map[string]any)
		if len(symptom) == 0 {
			if name := strings.TrimSpace(value_text(obj["name"])); name != "" {
				return kind + ":" + name
			}
		}
		symptom_kind, _ := symptom["kind"].(string)
		if symptom_kind == "symptom_code" {
			return kind + ":code:" + strings.TrimSpace(value_text(symptom["code"]))
		}
		if symptom_kind == "pain" {
			aspects, _ := symptom["aspects"].([]any)
			parts := make([]string, 0, len(aspects))
			for _, item := range aspects {
				aspect, _ := item.(map[string]any)
				aspect_kind, _ := aspect["kind"].(string)
				aspect_value := aspect["name"]
				if aspect_kind == "site" {
					aspect_value = aspect["site"]
					if strings.TrimSpace(value_text(aspect_value)) == "" {
						aspect_value = aspect["name"]
					}
				}
				parts = append(parts, aspect_kind+"="+strings.TrimSpace(value_text(aspect_value))+strings.TrimSpace(value_text(aspect["value"])))
			}
			sort.Strings(parts)
			return kind + ":pain:" + strings.Join(parts, "|")
		}
		return kind + ":" + strings.TrimSpace(value_text(symptom))
	case "increased_index", "decreased_index":
		return kind + ":" + strings.TrimSpace(value_text(first_nonempty_any(obj["index"], obj["name"])))
	case "have_disease":
		return kind + ":" + strings.TrimSpace(value_text(first_nonempty_any(obj["disease"], obj["name"])))
	case "damaged", "inflammation":
		return kind + ":" + strings.TrimSpace(value_text(first_nonempty_any(obj["actor"], obj["name"])))
	default:
		return centitylabel.NodeLabel(kind, data)
	}
}

func procedure_identity(value eprocedure.IMedicalProcedure) string {
	switch typed := value.(type) {
	case eprocedure.Transplant:
		return "Transplant:" + strings.TrimSpace(typed.Target.ActorName())
	case eprocedure.Infuse:
		return "Infuse:" + strings.TrimSpace(typed.Material.ActorName()) + ":" + string(typed.Route)
	case eprocedure.Surgery:
		return "Surgery:" + strings.TrimSpace(typed.Name)
	case eprocedure.Cleaning:
		name := ""
		if typed.Target != nil {
			name = strings.TrimSpace(typed.Target.ActorName())
		}
		return "Cleaning:" + name
	case eprocedure.DiagnosticTest:
		return "DiagnosticTest:" + strings.TrimSpace(typed.Name)
	case eprocedure.Education:
		return "Education:" + strings.TrimSpace(typed.Name)
	case eprocedure.Instruction:
		return "Instruction:" + strings.TrimSpace(typed.Name)
	case eprocedure.UnknownMedicalProcedure:
		return "Unknown:" + strings.TrimSpace(typed.Name)
	default:
		return value.Kind()
	}
}

func first_nonempty_any(values ...any) any {
	for _, value := range values {
		if strings.TrimSpace(value_text(value)) != "" {
			return value
		}
	}
	return nil
}

func value_text(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case map[string]any:
		for _, key := range []string{"name", "code", "value"} {
			if text, ok := typed[key].(string); ok {
				return text
			}
		}
		encoded, _ := json.Marshal(typed)
		return string(encoded)
	default:
		encoded, _ := json.Marshal(typed)
		return string(encoded)
	}
}

func (s *Service) actions_causing(status_id entity.StatusID) ([]actionRow, error) {
	if s.container == nil {
		return nil, fmt.Errorf("morbes graph container 가 없습니다")
	}
	var result []actionRow
	for _, id := range s.container.Get_StatusCauseActions(status_id) {
		value, ok := s.container.FindActionById(id)
		if !ok {
			continue
		}
		data, err := entity.MarshalAction(value)
		if err != nil {
			continue
		}
		result = append(result, actionRow{
			ID:        id,
			Kind:      value.Kind(),
			Data:      string(data),
			SourceIDs: append([]source.SourceId(nil), s.container.ActionSourceIDs[id]...),
		})
	}
	return result, nil
}

func (s *Service) statuses_causing(action_id entity.ActionID) ([]statusRow, error) {
	if s.container == nil {
		return nil, fmt.Errorf("morbes graph container 가 없습니다")
	}
	var result []statusRow
	for _, id := range s.container.StatusesCausingAction(action_id) {
		value, ok := s.container.FindStatusById(id)
		if !ok {
			continue
		}
		data, err := medicalcodec.MarshalStatus(value)
		if err != nil {
			continue
		}
		result = append(result, statusRow{
			ID:   id,
			Kind: value.Kind(),
			Data: string(data),
		})
	}
	return result, nil
}
