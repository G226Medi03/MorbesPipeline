package entity

import (
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	source "github.com/G226Medi03/SourceEntity"
)

// LinkActionResultStatusValues links typed action/status values using their
// persisted row identifiers.
func (s *entityTable) LinkActionResultStatusValues(action repo.IEAction, status repo.IEStatus, sourceID source.SourceId) error {
	actionID, err := s.actionIDForValue(action)
	if err != nil {
		return err
	}
	statusID, err := s.statusIDForValue(status)
	if err != nil {
		return err
	}
	return s.LinkActionResultStatus(actionID, statusID, sourceID)
}
