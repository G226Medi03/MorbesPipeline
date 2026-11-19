package store

import (
	"database/sql"
	"strings"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"github.com/G226Medi03/MorbesPipeline/internal/domain/entitykey"
	"github.com/G226Medi03/MorbesPipeline/internal/store/read"
	source "github.com/G226Medi03/SourceEntity"
)

// ActorRecord is the name-based read model for an actor row.
type (
	ActorRecord            = read.ActorRecord
	ActorActionLink        = read.ActorActionLink
	ActionRecord           = read.ActionRecord
	ActionResultStatusLink = read.ActionResultStatusLink
	StatusRecord           = read.StatusRecord
	StatusActionLink       = read.StatusActionLink
	ProcedureRecord        = read.ProcedureRecord
)

func (s *StoreContainer) sourceRefsFor(table, column string, id int64) ([]source.Source, error) {
	rows, err := s.QueryContext(s.ctx,
		`SELECT source_id FROM `+table+` WHERE `+column+` = ? ORDER BY id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck // deferred close
	var sourceIDs []source.SourceId
	for rows.Next() {
		var sourceID source.SourceId
		if err := rows.Scan(&sourceID); err != nil {
			return nil, err
		}
		sourceIDs = append(sourceIDs, sourceID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return s.refs(s.ctx, sourceIDs), nil
}

// GetActorRecordByName returns the actor read model for a name.
func (s *StoreContainer) GetActorRecordByName(name string) (ActorRecord, bool, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return ActorRecord{}, false, nil
	}
	var actorID int64
	err := s.QueryRowContext(s.ctx, `SELECT eactor_id FROM actor_other_name WHERE name = ? LIMIT 1`, name).Scan(&actorID)
	if err == sql.ErrNoRows {
		err = s.QueryRowContext(s.ctx,
			`SELECT eactor_id FROM eactor WHERE name = ? AND kind <> 'procedure' LIMIT 1`, name).Scan(&actorID)
	}
	if err == sql.ErrNoRows {
		return ActorRecord{}, false, nil
	}
	if err != nil {
		return ActorRecord{}, false, err
	}
	return s.loadActorRecord(actorID)
}

// GetDeviceByName returns the device actor read model for a name.
func (s *StoreContainer) GetDeviceByName(name string) (ActorRecord, bool, error) {
	record, found, err := s.GetActorRecordByName(name)
	if err != nil || !found {
		return ActorRecord{}, found, err
	}
	if record.Value.Kind() != string(repo.ActorKindDevice) {
		return ActorRecord{}, false, nil
	}
	return record, true, nil
}

func (s *StoreContainer) loadActorRecord(actorID int64) (ActorRecord, bool, error) {
	value, err := s.actorValueByID(actorID)
	if err != nil {
		return ActorRecord{}, false, err
	}
	refs, err := s.sourceRefsFor("actor_sources", "eactor_id", actorID)
	if err != nil {
		return ActorRecord{}, false, err
	}
	record := ActorRecord{
		Key:     entitykey.Actor(value),
		Value:   value,
		Sources: refs,
	}
	rows, err := s.QueryContext(s.ctx,
		`SELECT name FROM actor_other_name WHERE eactor_id = ? ORDER BY id`, actorID)
	if err != nil {
		return ActorRecord{}, false, err
	}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			_ = rows.Close()
			return ActorRecord{}, false, err
		}
		record.OtherNames = append(record.OtherNames, name)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return ActorRecord{}, false, err
	}
	_ = rows.Close()

	rows, err = s.QueryContext(s.ctx,
		`SELECT eaction_id, source_id FROM actor_actions WHERE eactor_id = ? ORDER BY id`, actorID)
	if err != nil {
		return ActorRecord{}, false, err
	}
	defer rows.Close() //nolint:errcheck // deferred close
	for rows.Next() {
		var actionID int64
		var sourceID source.SourceId
		if err := rows.Scan(&actionID, &sourceID); err != nil {
			return ActorRecord{}, false, err
		}
		action, err := s.getActionShallow(actionID)
		if err != nil {
			continue
		}
		record.Actions = append(record.Actions, ActorActionLink{Action: action, SourceID: sourceID})
	}
	return record, true, rows.Err()
}

// GetActionRecord returns the action read model for a typed action value.
func (s *StoreContainer) GetActionRecord(value repo.IEAction) (ActionRecord, bool, error) {
	actionID, err := s.actionIDForValue(value)
	if err != nil {
		if err == sql.ErrNoRows {
			return ActionRecord{}, false, nil
		}
		return ActionRecord{}, false, err
	}
	action, err := s.getActionShallow(actionID)
	if err != nil {
		return ActionRecord{}, false, err
	}
	rows, err := s.QueryContext(s.ctx,
		`SELECT estatus_id, source_id FROM action_result_statuses WHERE eaction_id = ? ORDER BY id`, actionID)
	if err != nil {
		return ActionRecord{}, false, err
	}
	defer rows.Close() //nolint:errcheck // deferred close
	for rows.Next() {
		var statusID int64
		var sourceID source.SourceId
		if err := rows.Scan(&statusID, &sourceID); err != nil {
			return ActionRecord{}, false, err
		}
		status, err := s.getStatusShallow(statusID)
		if err != nil {
			continue
		}
		action.ResultStatuses = append(action.ResultStatuses, ActionResultStatusLink{
			Status: status, SourceID: sourceID,
		})
	}
	return action, true, rows.Err()
}

func (s *StoreContainer) actionValueByID(actionID int64) (repo.IEAction, error) {
	var kind string
	var ptr int64
	err := s.QueryRowContext(s.ctx,
		`SELECT kind, ptr FROM eaction WHERE eaction_id = ?`, actionID).Scan(&kind, &ptr)
	if err != nil {
		return nil, err
	}
	return s.actionValueByKindPtr(kind, ptr)
}

func (s *StoreContainer) getActionShallow(actionID int64) (ActionRecord, error) {
	value, err := s.actionValueByID(actionID)
	if err != nil {
		return ActionRecord{}, err
	}
	refs, err := s.sourceRefsFor("action_sources", "eaction_id", actionID)
	if err != nil {
		return ActionRecord{}, err
	}
	return ActionRecord{Key: entitykey.Action(value), Value: value, Sources: refs}, nil
}

// GetStatusRecord returns the status read model for a typed status value.
func (s *StoreContainer) GetStatusRecord(value repo.IEStatus) (StatusRecord, bool, error) {
	statusID, err := s.statusIDForValue(value)
	if err != nil {
		if err == sql.ErrNoRows {
			return StatusRecord{}, false, nil
		}
		return StatusRecord{}, false, err
	}
	status, err := s.getStatusShallow(statusID)
	if err != nil {
		return StatusRecord{}, false, err
	}
	rows, err := s.QueryContext(s.ctx,
		`SELECT eaction_id, source_id FROM status_actions WHERE estatus_id = ? ORDER BY id`, statusID)
	if err != nil {
		return StatusRecord{}, false, err
	}
	defer rows.Close() //nolint:errcheck // deferred close
	for rows.Next() {
		var actionID int64
		var sourceID source.SourceId
		if err := rows.Scan(&actionID, &sourceID); err != nil {
			return StatusRecord{}, false, err
		}
		action, err := s.getActionShallow(actionID)
		if err != nil {
			continue
		}
		status.Actions = append(status.Actions, StatusActionLink{Action: action, SourceID: sourceID})
	}
	return status, true, rows.Err()
}

func (s *StoreContainer) getStatusShallow(statusID int64) (StatusRecord, error) {
	value, err := s.statusValueByID(int64(statusID))
	if err != nil {
		return StatusRecord{}, err
	}
	refs, err := s.sourceRefsFor("status_sources", "estatus_id", statusID)
	if err != nil {
		return StatusRecord{}, err
	}
	return StatusRecord{Key: entitykey.Status(value), Value: value, Sources: refs}, nil
}

// GetProcedureRecord returns the procedure read model for a typed value.
func (s *StoreContainer) GetProcedureRecord(value repo.IMedicalProcedure) (ProcedureRecord, bool, error) {
	procedureID, err := s.procedureIDForValue(value)
	if err != nil {
		if err == sql.ErrNoRows {
			return ProcedureRecord{}, false, nil
		}
		return ProcedureRecord{}, false, err
	}
	stored, err := s.procedureValueByID(procedureID)
	if err != nil {
		return ProcedureRecord{}, false, err
	}
	refs, err := s.sourceRefsFor("procedure_sources", "emedical_procedure_id", procedureID)
	if err != nil {
		return ProcedureRecord{}, false, err
	}
	return ProcedureRecord{
		Key:     entitykey.Procedure(stored),
		Value:   stored,
		Sources: refs,
	}, true, nil
}
