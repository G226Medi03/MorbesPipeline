package store

import (
	"database/sql"
	"fmt"

	medicalentity "github.com/G226Medi03/MedicalEntity/entity"
	entity "github.com/G226Medi03/MorbesPipeline/legacy"
	"github.com/G226Medi03/SourceEntity"
)

func (s *Store) SaveStatus(value entity.Status) (entity.Status, error) {
	unlock := s.lockWrite()
	defer unlock()
	if err := value.Validate(); err != nil {
		return value, err
	}
	external, err := s.externalStatus(value)
	if err != nil {
		return value, fmt.Errorf("status data encode failed: %w", err)
	}
	saved, err := s.StatusRepository.Insert(external, sourceIDs(value.Sources))
	if err != nil {
		return value, err
	}
	if err := s.validateKindPtr("estatus"); err != nil {
		return value, err
	}
	return localStatusFromExternal(saved)
}

func (s *Store) GetStatusByID(id entity.StatusID) (entity.Status, error) {
	status, err := s.getStatusShallow(id)
	if err != nil {
		return status, err
	}
	links, err := s.ActionRepository.StatusActions(entityStatusID(id))
	if err != nil {
		return entity.NewStatus("", nil, nil), err
	}
	for _, link := range links {
		action, err := s.getActionShallow(localActionID(link.ActionID))
		if err != nil {
			continue
		}
		status.Actions = append(status.Actions, struct {
			Action   entity.Action   `json:"action"`
			SourceID source.SourceId `json:"sourceId"`
		}{Action: action, SourceID: link.SourceID})
	}
	return status, nil
}

func (s *Store) getStatusShallow(id entity.StatusID) (entity.Status, error) {
	value, err := s.StatusRepository.Get(entityStatusID(id))
	if err != nil {
		if err == medicalentity.ErrNotFound {
			return entity.NewStatus("", nil, nil), sql.ErrNoRows
		}
		return entity.NewStatus("", nil, nil), err
	}
	status, err := localStatusFromExternal(value)
	if err != nil {
		return entity.NewStatus("", nil, nil), err
	}
	refs, err := s.sourceRefsFor("status_sources", "estatus_id", id.ToI64())
	if err != nil {
		return entity.NewStatus("", nil, nil), err
	}
	status.Sources = refs
	return status, nil
}

func (s *Store) LinkStatusAction(statusID entity.StatusID, actionID entity.ActionID, sourceID source.SourceId) error {
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

func (s *Store) DeleteStatusAction(statusID entity.StatusID, actionID entity.ActionID) error {
	unlock := s.lockWrite()
	defer unlock()
	_, err := s.ExecContext(s.ctx,
		`DELETE FROM status_actions WHERE estatus_id = ? AND eaction_id = ?`,
		statusID.ToI64(), actionID.ToI64())
	return err
}
