// Package ccause traces possible causes through the status/action graph.
package ccause

import (
	"fmt"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	centitylabel "github.com/G226Medi03/MorbesPipeline/internal/domain/label"
	source "github.com/G226Medi03/MorbesPipeline/internal/external/sourceentity"
	"github.com/G226Medi03/MorbesPipeline/internal/support/ids"
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

func (s *Service) actions_causing(status_id ids.StatusID) ([]actionRow, error) {
	if s.container == nil {
		return nil, fmt.Errorf("morbes graph container 가 없습니다")
	}
	var result []actionRow
	for _, id := range s.container.Get_StatusCauseActions(status_id) {
		value, ok := s.container.FindActionById(id)
		if !ok {
			continue
		}
		result = append(result, actionRow{
			ID:        id,
			Value:     value,
			SourceIDs: append([]source.SourceId(nil), s.container.ActionSourceIDs[id]...),
		})
	}
	return result, nil
}

func (s *Service) statuses_causing(action_id ids.ActionID) ([]statusRow, error) {
	if s.container == nil {
		return nil, fmt.Errorf("morbes graph container 가 없습니다")
	}
	var result []statusRow
	for _, id := range s.container.StatusesCausingAction(action_id) {
		value, ok := s.container.FindStatusById(id)
		if !ok {
			continue
		}
		result = append(result, statusRow{
			ID:    id,
			Value: value,
		})
	}
	return result, nil
}
