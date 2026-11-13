package store

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/G226Medi03/MedicalEntity/ename/eaction"
	medicalentity "github.com/G226Medi03/MedicalEntity/entity"
	entity "github.com/G226Medi03/MorbesPipeline/legacy"
	"github.com/G226Medi03/SourceEntity"
)

func (s *Store) SaveAction(value entity.Action) (entity.Action, error) {
	unlock := s.lockWrite()
	defer unlock()
	if err := value.Validate(); err != nil {
		return value, err
	}
	if value.Variant == "cause_status" {
		if statusID, ok := causeStatusID(value.Value); ok && statusID > 0 {
			return s.saveCauseStatusAction(value, statusID)
		}
	}
	external, err := s.externalAction(value)
	if err != nil {
		return value, fmt.Errorf("action data encode failed: %w", err)
	}
	saved, err := s.ActionRepository.Insert(external, nil, sourceIDs(value.Sources))
	if err != nil {
		return value, err
	}
	if err := s.validateKindPtr("eaction"); err != nil {
		return value, err
	}
	return localActionFromExternal(saved)
}

func (s *Store) saveCauseStatusAction(value entity.Action, statusID entity.StatusID) (entity.Action, error) {
	if _, err := s.ExecContext(s.ctx,
		`INSERT OR IGNORE INTO action_cause_status (estatus_id) VALUES (?)`, statusID.ToI64()); err != nil {
		return value, err
	}
	var ptr int64
	if err := s.QueryRowContext(s.ctx,
		`SELECT id FROM action_cause_status WHERE estatus_id = ?`, statusID.ToI64()).Scan(&ptr); err != nil {
		return value, err
	}
	if ptr <= 0 {
		return value, fmt.Errorf("cause_status ptr가 유효하지 않습니다: estatus_id=%d ptr=%d", statusID.ToI64(), ptr)
	}
	if _, err := s.ExecContext(s.ctx,
		`INSERT OR IGNORE INTO eaction (kind, ptr) VALUES ('cause_status', ?)`, ptr); err != nil {
		return value, err
	}
	var id entity.ActionID
	if err := s.QueryRowContext(s.ctx,
		`SELECT eaction_id FROM eaction WHERE kind = 'cause_status' AND ptr = ?`, ptr).Scan(&id); err != nil {
		return value, err
	}
	if err := s.addActionSources(id, value.Sources); err != nil {
		return value, err
	}
	if err := s.validateKindPtr("eaction"); err != nil {
		return value, err
	}
	return entity.RestoreAction(id, value.Variant, value.Value, value.Sources), nil
}

func causeStatusID(value any) (entity.StatusID, bool) {
	raw, err := json.Marshal(value)
	if err != nil {
		return 0, false
	}
	var payload struct {
		StatusID int64 `json:"status_id"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil || payload.StatusID <= 0 {
		return 0, false
	}
	return entity.StatusID(payload.StatusID), true
}

func (s *Store) UpdateAction(value entity.Action) error {
	unlock := s.lockWrite()
	defer unlock()
	if value.Id() <= 0 {
		return fmt.Errorf("action id required")
	}
	external, err := s.externalAction(value)
	if err != nil {
		return fmt.Errorf("action data encode failed: %w", err)
	}
	if err := s.ActionRepository.Update(entityActionID(value.Id()), external); err != nil {
		return err
	}
	if _, err := s.ExecContext(s.ctx,
		`DELETE FROM action_result_statuses WHERE eaction_id = ?`, value.Id().ToI64()); err != nil {
		return err
	}
	for _, relation := range value.ResultStatuses {
		sourceID := relation.SourceID
		if sourceID.ToI64() == 0 {
			sourceID = firstSourceID(value.Sources)
		}
		if err := s.LinkActionResultStatus(value.Id(), relation.Status.Id(), sourceID); err != nil {
			return err
		}
	}
	if err := s.addActionSources(value.Id(), value.Sources); err != nil {
		return err
	}
	return s.validateKindPtr("eaction")
}

func (s *Store) GetActionByID(id entity.ActionID) (entity.Action, error) {
	action, err := s.getActionShallow(id)
	if err != nil {
		return action, err
	}
	links, err := s.ActionRepository.ActionResultStatuses(entityActionID(id))
	if err != nil {
		return entity.NewAction("", nil, nil), err
	}
	for _, link := range links {
		status, err := s.getStatusShallow(localStatusID(link.StatusID))
		if err != nil {
			continue
		}
		action.ResultStatuses = append(action.ResultStatuses, struct {
			Status   entity.Status   `json:"status"`
			SourceID source.SourceId `json:"sourceId"`
		}{Status: status, SourceID: link.SourceID})
	}
	return action, nil
}

func (s *Store) getActionShallow(id entity.ActionID) (entity.Action, error) {
	value, err := s.ActionRepository.Get(entityActionID(id))
	if err != nil {
		if err == medicalentity.ErrNotFound {
			return entity.NewAction("", nil, nil), sql.ErrNoRows
		}
		return entity.NewAction("", nil, nil), err
	}
	action, err := localActionFromExternal(value)
	if err != nil {
		return entity.NewAction("", nil, nil), err
	}
	refs, err := s.sourceRefsFor("action_sources", "eaction_id", id.ToI64())
	if err != nil {
		return entity.NewAction("", nil, nil), err
	}
	action.Sources = refs
	return action, nil
}

func (s *Store) LinkActionResultStatus(actionID entity.ActionID, statusID entity.StatusID, sourceID source.SourceId) error {
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

func (s *Store) DeleteActionResultStatus(actionID entity.ActionID, statusID entity.StatusID) error {
	unlock := s.lockWrite()
	defer unlock()
	_, err := s.ExecContext(s.ctx,
		`DELETE FROM action_result_statuses WHERE eaction_id = ? AND estatus_id = ?`,
		actionID.ToI64(), statusID.ToI64())
	return err
}

func (s *Store) addActionSources(id entity.ActionID, sources []source.SourceRef) error {
	for _, ref := range sources {
		if ref.SourceID.ToI64() <= 0 {
			continue
		}
		if _, err := s.ExecContext(s.ctx,
			`INSERT OR IGNORE INTO action_sources (eaction_id, source_id) VALUES (?, ?)`,
			id.ToI64(), ref.SourceID.ToI64()); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) externalAction(value entity.Action) (medicalentity.IEAction, error) {
	fmt.Printf("DEBUG externalAction variant=%q type=%T value=%v\n", value.Variant, value.Value, value.Value)
	if value.Variant == "cause_status" {
		raw, err := json.Marshal(value.Value)
		if err != nil {
			return nil, err
		}
		var payload struct {
			StatusID int64 `json:"status_id"`
		}
		if err := json.Unmarshal(raw, &payload); err == nil && payload.StatusID > 0 {
			status, err := s.StatusRepository.Get(entityStatusID(entity.StatusID(payload.StatusID)))
			if err != nil {
				return nil, err
			}
			fmt.Printf("DEBUG status name=%T %#v\n", status.Name(), status.Name())
			return eaction.CauseStatus{Status: status.Name()}, nil
		}
	}
	return externalActionFromLocal(value)
}
