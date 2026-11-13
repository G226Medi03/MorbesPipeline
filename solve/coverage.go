package csolve

import (
	"encoding/json"
	"strings"

	eprocedure "github.com/G226Medi03/MedicalEntity/ename/emedical_procedure"
	centitylabel "github.com/G226Medi03/MorbesPipeline/label"
	entity "github.com/G226Medi03/MorbesPipeline/legacy"
)

// coverageObservedStatuses adds statuses that the status extractor found but
// situation omitted, when their label is explicitly present in the problem
// stem. Answer/explanation are intentionally excluded from solve evidence.
func coverageObservedStatuses(state *pipelineState, observed []entity.Status) ([]entity.Status, []string) {
	if state == nil || len(state.Statuses) == 0 {
		return observed, nil
	}
	evidence := coverageEvidence(state.Case.Problem)
	if strings.TrimSpace(evidence) == "" {
		return observed, nil
	}
	seen := make(map[string]struct{}, len(observed))
	for _, status := range observed {
		seen[coverageStatusKey(status)] = struct{}{}
	}
	addedLabels := make([]string, 0)
	for _, status := range state.Statuses {
		label := coverageStatusLabel(status)
		if label == "" {
			continue
		}
		key := coverageStatusKey(status)
		if _, exists := seen[key]; exists {
			continue
		}
		matched := false
		for _, term := range coverageStatusTerms(status, label) {
			if term != "" && strings.Contains(evidence, term) {
				matched = true
				break
			}
		}
		if !matched {
			continue
		}
		observed = append(observed, status)
		seen[key] = struct{}{}
		addedLabels = append(addedLabels, label)
	}
	return observed, addedLabels
}

func coverageEvidence(problem string) string {
	return problem_stem(problem)
}

func coverageStatusLabel(status entity.Status) string {
	data, err := json.Marshal(status.Value)
	if err != nil {
		return ""
	}
	label := strings.TrimSpace(centitylabel.StatusLabel(status.Variant, string(data)))
	if label == "" || label == status.Variant {
		return ""
	}
	return label
}

func coverageStatusTerms(status entity.Status, label string) []string {
	terms := []string{label}
	encoded, err := json.Marshal(status.Value)
	if err != nil {
		return terms
	}
	if status.Variant == "increased_index" || status.Variant == "decreased_index" {
		var payload struct {
			Index json.RawMessage `json:"index"`
			Name  string          `json:"name"`
		}
		if err := json.Unmarshal(encoded, &payload); err == nil {
			if value := indexTerm(payload.Index); value != "" {
				terms = append(terms, value)
			}
			if value := strings.TrimSpace(payload.Name); value != "" {
				terms = append(terms, value)
			}
		}
		return terms
	}
	if status.Variant != "received_medical_procedure" {
		return terms
	}
	var payload struct {
		MedicalProcedure json.RawMessage `json:"medical_procedure"`
	}
	if err := json.Unmarshal(encoded, &payload); err != nil || len(payload.MedicalProcedure) == 0 {
		return terms
	}
	value, err := entity.UnmarshalProcedure(payload.MedicalProcedure)
	if err != nil {
		return terms
	}
	procedure_label := centitylabel.ProcedureLabel(value)
	if procedure_label != "" {
		terms = append(terms, procedure_label)
	}
	if cleaning, ok := value.(eprocedure.Cleaning); ok && cleaning.Target != nil {
		device := strings.TrimSpace(cleaning.Target.ActorName())
		if device != "" {
			terms = append(terms, device+" 세척", device)
		}
	}
	return terms
}

func indexTerm(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return strings.TrimSpace(text)
	}
	var payload struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ""
	}
	return strings.TrimSpace(payload.Name)
}

func coverageStatusKey(status entity.Status) string {
	data, _ := json.Marshal(status.Value)
	return status.Variant + "|" + string(data)
}
