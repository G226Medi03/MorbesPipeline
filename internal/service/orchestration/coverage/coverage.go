package coverage

import (
	"strings"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	entitylabel "github.com/G226Medi03/MorbesPipeline/internal/entity/label"
	"github.com/G226Medi03/MorbesPipeline/internal/entity/solve/input"
	"github.com/G226Medi03/MorbesPipeline/internal/service/orchestration/state"
)

// ObservedStatuses adds statuses that the status extractor found but
// situation omitted, when their label is explicitly present in the problem
// stem. Answer/explanation are intentionally excluded from solve evidence.
func ObservedStatuses(state *state.State, observed []repo.IEStatus) ([]repo.IEStatus, []string) {
	if state == nil || len(state.Statuses) == 0 {
		return observed, nil
	}
	evidence := coverageEvidence(state.Case.Problem)
	if strings.TrimSpace(evidence) == "" {
		return observed, nil
	}
	seen := make([]repo.IEStatus, 0, len(observed)+len(state.Statuses))
	for _, status := range observed {
		if !containsStatus(seen, status) {
			seen = append(seen, status)
		}
	}
	addedLabels := make([]string, 0)
	for _, status := range state.Statuses {
		label := coverageStatusLabel(status)
		if label == "" {
			continue
		}
		if containsStatus(seen, status) {
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
		seen = append(seen, status)
		addedLabels = append(addedLabels, label)
	}
	return observed, addedLabels
}

func coverageEvidence(problem string) string {
	return input.Stem(problem)
}

func coverageStatusLabel(status repo.IEStatus) string {
	label := strings.TrimSpace(entitylabel.StatusLabelValue(status))
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
		procedureLabel := entitylabel.ProcedureLabel(typed.MedicalProcedure)
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

func containsStatus(values []repo.IEStatus, target repo.IEStatus) bool {
	for _, value := range values {
		if equalStatus(value, target) {
			return true
		}
	}
	return false
}

func equalStatus(a, b repo.IEStatus) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.Equal(b)
}
