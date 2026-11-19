package store

import (
	"fmt"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"github.com/G226Medi03/MorbesPipeline/internal/support/ids"
	source "github.com/G226Medi03/SourceEntity"
)

// LinkStatusAction connects a status and an action with one source.
func (s *StoreContainer) LinkStatusAction(statusID ids.StatusID, actionID ids.ActionID, sourceID source.SourceId) error {
	if sourceID.ToI64() <= 0 {
		return fmt.Errorf("status-action source가 필요합니다")
	}
	unlock := s.lockWrite()
	defer unlock()
	_, err := s.ExecContext(s.ctx,
		`INSERT OR IGNORE INTO status_actions (estatus_id, eaction_id, source_id) VALUES (?, ?, ?)`,
		statusID.ToI64(), actionID.ToI64(), sourceID.ToI64())
	return err
}

// LinkStatusActionValues resolves typed values and connects them.
func (s *StoreContainer) LinkStatusActionValues(status repo.IEStatus, action repo.IEAction, sourceID source.SourceId) error {
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
func (s *StoreContainer) LinkActionResultStatus(actionID ids.ActionID, statusID ids.StatusID, sourceID source.SourceId) error {
	if sourceID.ToI64() <= 0 {
		return fmt.Errorf("action-result source가 필요합니다")
	}
	unlock := s.lockWrite()
	defer unlock()
	_, err := s.ExecContext(s.ctx,
		`INSERT OR IGNORE INTO action_result_statuses (eaction_id, estatus_id, source_id) VALUES (?, ?, ?)`,
		actionID.ToI64(), statusID.ToI64(), sourceID.ToI64())
	return err
}

// DeleteStatusAction removes one status-action edge.
func (s *StoreContainer) DeleteStatusAction(statusID ids.StatusID, actionID ids.ActionID) error {
	unlock := s.lockWrite()
	defer unlock()
	_, err := s.ExecContext(s.ctx,
		`DELETE FROM status_actions WHERE estatus_id = ? AND eaction_id = ?`,
		statusID.ToI64(), actionID.ToI64())
	return err
}

// DeleteStatusActionValues resolves typed values and removes one edge.
func (s *StoreContainer) DeleteStatusActionValues(status repo.IEStatus, action repo.IEAction) error {
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
func (s *StoreContainer) DeleteActionResultStatus(actionID ids.ActionID, statusID ids.StatusID) error {
	unlock := s.lockWrite()
	defer unlock()
	_, err := s.ExecContext(s.ctx,
		`DELETE FROM action_result_statuses WHERE eaction_id = ? AND estatus_id = ?`,
		actionID.ToI64(), statusID.ToI64())
	return err
}

// DeleteActionResultStatusValues resolves typed values and removes one edge.
func (s *StoreContainer) DeleteActionResultStatusValues(action repo.IEAction, status repo.IEStatus) error {
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
