package entity

import (
	"fmt"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	source "github.com/G226Medi03/SourceEntity"
)

// LinkStatusAction connects a status and an action with one source.
func (s *EntityService) LinkStatusAction(statusID int64, actionID int64, sourceID source.SourceId) error {
	if sourceID.ToI64() <= 0 {
		return fmt.Errorf("status-action source가 필요합니다")
	}
	unlock := s.store.LockWrite()
	defer unlock()
	_, err := s.store.MainExecutor().ExecContext(s.store.Context(),
		`INSERT OR IGNORE INTO status_actions (estatus_id, eaction_id, source_id) VALUES (?, ?, ?)`,
		statusID, actionID, sourceID.ToI64())
	return err
}

// LinkStatusActionValues resolves typed values and connects them.
func (s *EntityService) LinkStatusActionValues(status repo.IEStatus, action repo.IEAction, sourceID source.SourceId) error {
	statusID, err := s.statusIDForValue(status)
	if err != nil {
		return err
	}
	actionID, err := s.actionIDForValue(action)
	if err != nil {
		return err
	}
	return s.LinkStatusAction(statusID, actionID, sourceID)
}

// LinkActionResultStatus connects an action and its result status with one
// source.
func (s *EntityService) LinkActionResultStatus(actionID int64, statusID int64, sourceID source.SourceId) error {
	if sourceID.ToI64() <= 0 {
		return fmt.Errorf("action-result source가 필요합니다")
	}
	unlock := s.store.LockWrite()
	defer unlock()
	_, err := s.store.MainExecutor().ExecContext(s.store.Context(),
		`INSERT OR IGNORE INTO action_result_statuses (eaction_id, estatus_id, source_id) VALUES (?, ?, ?)`,
		actionID, statusID, sourceID.ToI64())
	return err
}

// DeleteStatusAction removes one status-action edge.
func (s *EntityService) DeleteStatusAction(statusID int64, actionID int64) error {
	unlock := s.store.LockWrite()
	defer unlock()
	_, err := s.store.MainExecutor().ExecContext(s.store.Context(),
		`DELETE FROM status_actions WHERE estatus_id = ? AND eaction_id = ?`,
		statusID, actionID)
	return err
}

// DeleteStatusActionValues resolves typed values and removes one edge.
func (s *EntityService) DeleteStatusActionValues(status repo.IEStatus, action repo.IEAction) error {
	statusID, err := s.statusIDForValue(status)
	if err != nil {
		return err
	}
	actionID, err := s.actionIDForValue(action)
	if err != nil {
		return err
	}
	return s.DeleteStatusAction(statusID, actionID)
}

// DeleteActionResultStatus removes one action-result edge.
func (s *EntityService) DeleteActionResultStatus(actionID int64, statusID int64) error {
	unlock := s.store.LockWrite()
	defer unlock()
	_, err := s.store.MainExecutor().ExecContext(s.store.Context(),
		`DELETE FROM action_result_statuses WHERE eaction_id = ? AND estatus_id = ?`,
		actionID, statusID)
	return err
}

// DeleteActionResultStatusValues resolves typed values and removes one edge.
func (s *EntityService) DeleteActionResultStatusValues(action repo.IEAction, status repo.IEStatus) error {
	actionID, err := s.actionIDForValue(action)
	if err != nil {
		return err
	}
	statusID, err := s.statusIDForValue(status)
	if err != nil {
		return err
	}
	return s.DeleteActionResultStatus(actionID, statusID)
}
