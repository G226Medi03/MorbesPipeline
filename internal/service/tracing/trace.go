package ccause

import (
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	centitylabel "github.com/G226Medi03/MorbesPipeline/internal/entity/label"
)

func ObservedLabels(statuses []repo.IEStatus) []string {
	result := make([]string, 0, len(statuses))
	for _, status := range statuses {
		result = append(result, centitylabel.StatusLabelValue(status))
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
