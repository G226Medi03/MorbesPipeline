package entity

import (
	"fmt"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	source "github.com/G226Medi03/SourceEntity"
)

// LinkActionResultStatusValues connects an action and its result status using
// typed ename values.
func (s *entityTable) LinkActionResultStatusValues(action repo.IEAction, status repo.IEStatus, sourceID source.SourceId) error {
	actionID, err := s.actionRowIDForValue(action)
	if err != nil {
		return err
	}
	statusID, err := s.statusRowIDForValue(status)
	if err != nil {
		return err
	}
	return s.linkActionResultStatusRowIDs(actionID, statusID, sourceID)
}

// linkActionResultStatusRowIDs connects an action and its result status with
// one source. Row identifiers stay private to this repository.
func (s *entityTable) linkActionResultStatusRowIDs(actionID int64, statusID int64, sourceID source.SourceId) error {
	if sourceID.ToI64() <= 0 {
		return fmt.Errorf("action-result source가 필요합니다")
	}
	_, err := s.db.ExecContext(s.ctx,
		`INSERT OR IGNORE INTO action_result_statuses (eaction_id, estatus_id, source_id) VALUES (?, ?, ?)`,
		actionID, statusID, sourceID.ToI64())
	return err
}

// DeleteStatusActionValues removes one status-action edge using typed ename
// values.
func (s *entityTable) DeleteStatusActionValues(status repo.IEStatus, action repo.IEAction) error {
	statusID, err := s.statusRowIDForValue(status)
	if err != nil {
		return err
	}
	actionID, err := s.actionRowIDForValue(action)
	if err != nil {
		return err
	}
	return s.deleteStatusActionRowIDs(statusID, actionID)
}

// deleteStatusActionRowIDs removes one status-action edge. Row identifiers
// stay private to this repository.
func (s *entityTable) deleteStatusActionRowIDs(statusID int64, actionID int64) error {
	_, err := s.db.ExecContext(s.ctx,
		`DELETE FROM status_actions WHERE estatus_id = ? AND eaction_id = ?`,
		statusID, actionID)
	return err
}

// DeleteActionResultStatusValues removes one action-result edge using typed
// ename values.
func (s *entityTable) DeleteActionResultStatusValues(action repo.IEAction, status repo.IEStatus) error {
	actionID, err := s.actionRowIDForValue(action)
	if err != nil {
		return err
	}
	statusID, err := s.statusRowIDForValue(status)
	if err != nil {
		return err
	}
	return s.deleteActionResultStatusRowIDs(actionID, statusID)
}

// deleteActionResultStatusRowIDs removes one action-result edge. Row
// identifiers stay private to this repository.
func (s *entityTable) deleteActionResultStatusRowIDs(actionID int64, statusID int64) error {
	_, err := s.db.ExecContext(s.ctx,
		`DELETE FROM action_result_statuses WHERE eaction_id = ? AND estatus_id = ?`,
		actionID, statusID)
	return err
}
