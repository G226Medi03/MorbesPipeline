package csolve

import (
	"strings"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	entitykey "github.com/G226Medi03/MorbesPipeline/internal/domain/entitykey"
	centitylabel "github.com/G226Medi03/MorbesPipeline/internal/domain/label"
)

// coverageObservedStatuses adds statuses that the status extractor found but
// situation omitted, when their label is explicitly present in the problem
// stem. Answer/explanation are intentionally excluded from solve evidence.
func coverageObservedStatuses(state *pipelineState, observed []repo.IEStatus) ([]repo.IEStatus, []string) {
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

func coverageStatusLabel(status repo.IEStatus) string {
	label := strings.TrimSpace(centitylabel.StatusLabelValue(status))
	if label == "" || label == status.Kind() {
		return ""
	}
	return label
}

func coverageStatusTerms(status repo.IEStatus, label string) []string {
	terms := []string{label}
	switch typed := status.(type) {
	case repo.IncreasedIndex:
		terms = appendTerm(terms, typed.Index.Name)
	case repo.DecreasedIndex:
		terms = appendTerm(terms, typed.Index.Name)
	case repo.ReceivedMedicalProcedure:
		procedureLabel := centitylabel.ProcedureLabel(typed.MedicalProcedure)
		terms = appendTerm(terms, procedureLabel)
		if cleaning, ok := typed.MedicalProcedure.(repo.Cleaning); ok && cleaning.Target != nil {
			device := strings.TrimSpace(cleaning.Target.ActorName())
			if device != "" {
				terms = append(terms, device+" 세척", device)
			}
		}
	}
	return terms
}

func appendTerm(values []string, add ...string) []string {
	for _, value := range add {
		if value = strings.TrimSpace(value); value != "" {
			values = append(values, value)
		}
	}
	return values
}

func coverageStatusKey(status repo.IEStatus) string {
	return entitykey.Status(status)
}
