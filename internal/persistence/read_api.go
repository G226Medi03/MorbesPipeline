package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	centitylabel "github.com/G226Medi03/MorbesPipeline/internal/domain/label"
	medicalprocedureoption "github.com/G226Medi03/MorbesPipeline/internal/domain/procedureoption"
	source "github.com/G226Medi03/MorbesPipeline/internal/external/sourceentity"
	"github.com/G226Medi03/MorbesPipeline/internal/support/ids"
)

// NamedRecord is a name-based entity row with its numeric ID and sources.
type NamedRecord struct {
	ID      int64
	Name    string
	Sources []source.SourceRef
}

// ActorRecord is the read model for an actor row and its linked actions.
type ActorRecord struct {
	ID         ids.ActorID
	Value      repo.IEActor
	OtherNames []string
	Sources    []source.SourceRef
	Actions    []ActorActionLink
}

// ActorActionLink connects an actor to one action.
type ActorActionLink struct {
	Action   ActionRecord
	SourceID source.SourceId
}

// ActionRecord is the read model for an action row and its result statuses.
type ActionRecord struct {
	ID             ids.ActionID
	Value          repo.IEAction
	Sources        []source.SourceRef
	ResultStatuses []ActionResultStatusLink
}

// ActionResultStatusLink connects an action to one result status.
type ActionResultStatusLink struct {
	Status   StatusRecord
	SourceID source.SourceId
}

// StatusRecord is the read model for a status row and its linked actions.
type StatusRecord struct {
	ID      ids.StatusID
	Value   repo.IEStatus
	Sources []source.SourceRef
	Actions []StatusActionLink
}

// StatusActionLink connects a status to one action.
type StatusActionLink struct {
	Action   ActionRecord
	SourceID source.SourceId
}

// ProcedureRecord is the read model for a procedure row.
type ProcedureRecord struct {
	ID         ids.ProcedureID
	Value      repo.IMedicalProcedure
	OtherNames []string
	Traits     []repo.Trait
	Sources    []source.SourceRef
}

// EntityNeighbors is the previous/next existing ID in an entity table.
type EntityNeighbors struct {
	PreviousID int64
	NextID     int64
}

// ErrUnknownEntityType is returned when an entity type has no table mapping.
var ErrUnknownEntityType = errors.New("지원하지 않는 entity type")

func (s *Store) sourceRefsFor(table, column string, id int64) ([]source.SourceRef, error) {
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

func (s *Store) GetActorRecordByID(id ids.ActorID) (ActorRecord, bool, error) {
	var exists int
	err := s.QueryRowContext(s.ctx, `SELECT 1 FROM eactor WHERE eactor_id = ?`, id.ToI64()).Scan(&exists)
	if err == sql.ErrNoRows {
		return ActorRecord{}, false, nil
	}
	if err != nil {
		return ActorRecord{}, false, err
	}
	value, err := s.actorValueByID(id.ToI64())
	if err != nil {
		return ActorRecord{}, false, err
	}
	refs, err := s.sourceRefsFor("actor_sources", "eactor_id", id.ToI64())
	if err != nil {
		return ActorRecord{}, false, err
	}
	record := ActorRecord{ID: id, Value: value, Sources: refs}
	rows, err := s.QueryContext(s.ctx,
		`SELECT name FROM actor_other_name WHERE eactor_id = ? ORDER BY id`, id.ToI64())
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
		`SELECT eaction_id, source_id FROM actor_actions WHERE eactor_id = ? ORDER BY id`, id.ToI64())
	if err != nil {
		return ActorRecord{}, false, err
	}
	defer rows.Close() //nolint:errcheck // deferred close
	for rows.Next() {
		var actionID ids.ActionID
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

func (s *Store) GetActorRecordByName(name string) (ActorRecord, bool, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return ActorRecord{}, false, nil
	}
	var id ids.ActorID
	err := s.QueryRowContext(s.ctx, `SELECT eactor_id FROM actor_other_name WHERE name = ? LIMIT 1`, name).Scan(&id)
	if err == sql.ErrNoRows {
		err = s.QueryRowContext(s.ctx,
			`SELECT eactor_id FROM eactor WHERE name = ? AND kind <> 'procedure' LIMIT 1`, name).Scan(&id)
	}
	if err == sql.ErrNoRows {
		return ActorRecord{}, false, nil
	}
	if err != nil {
		return ActorRecord{}, false, err
	}
	return s.GetActorRecordByID(id)
}

func (s *Store) GetDeviceByID(id ids.DeviceID) (ActorRecord, bool, error) {
	record, found, err := s.GetActorRecordByID(ids.ActorID(id))
	if err != nil || !found {
		return ActorRecord{}, found, err
	}
	if record.Value.Kind() != string(repo.ActorKindDevice) {
		return ActorRecord{}, false, nil
	}
	return record, true, nil
}

func (s *Store) actionValueByID(id ids.ActionID) (repo.IEAction, error) {
	var kind string
	var ptr int64
	err := s.QueryRowContext(s.ctx,
		`SELECT kind, ptr FROM eaction WHERE eaction_id = ?`, id.ToI64()).Scan(&kind, &ptr)
	if err != nil {
		return nil, err
	}
	return s.actionValueByKindPtr(kind, ptr)
}

func (s *Store) getActionShallow(id ids.ActionID) (ActionRecord, error) {
	value, err := s.actionValueByID(id)
	if err != nil {
		return ActionRecord{}, err
	}
	refs, err := s.sourceRefsFor("action_sources", "eaction_id", id.ToI64())
	if err != nil {
		return ActionRecord{}, err
	}
	return ActionRecord{ID: id, Value: value, Sources: refs}, nil
}

func (s *Store) GetActionByID(id ids.ActionID) (ActionRecord, error) {
	action, err := s.getActionShallow(id)
	if err != nil {
		return ActionRecord{}, err
	}
	rows, err := s.QueryContext(s.ctx,
		`SELECT estatus_id, source_id FROM action_result_statuses WHERE eaction_id = ? ORDER BY id`, id.ToI64())
	if err != nil {
		return ActionRecord{}, err
	}
	defer rows.Close() //nolint:errcheck // deferred close
	for rows.Next() {
		var statusID ids.StatusID
		var sourceID source.SourceId
		if err := rows.Scan(&statusID, &sourceID); err != nil {
			return ActionRecord{}, err
		}
		status, err := s.getStatusShallow(statusID)
		if err != nil {
			continue
		}
		action.ResultStatuses = append(action.ResultStatuses, ActionResultStatusLink{
			Status: status, SourceID: sourceID,
		})
	}
	return action, rows.Err()
}

func (s *Store) getStatusShallow(id ids.StatusID) (StatusRecord, error) {
	value, err := s.statusValueByID(id)
	if err != nil {
		return StatusRecord{}, err
	}
	refs, err := s.sourceRefsFor("status_sources", "estatus_id", id.ToI64())
	if err != nil {
		return StatusRecord{}, err
	}
	return StatusRecord{ID: id, Value: value, Sources: refs}, nil
}

func (s *Store) GetStatusByID(id ids.StatusID) (StatusRecord, error) {
	status, err := s.getStatusShallow(id)
	if err != nil {
		return StatusRecord{}, err
	}
	rows, err := s.QueryContext(s.ctx,
		`SELECT eaction_id, source_id FROM status_actions WHERE estatus_id = ? ORDER BY id`, id.ToI64())
	if err != nil {
		return StatusRecord{}, err
	}
	defer rows.Close() //nolint:errcheck // deferred close
	for rows.Next() {
		var actionID ids.ActionID
		var sourceID source.SourceId
		if err := rows.Scan(&actionID, &sourceID); err != nil {
			return StatusRecord{}, err
		}
		action, err := s.getActionShallow(actionID)
		if err != nil {
			continue
		}
		status.Actions = append(status.Actions, StatusActionLink{Action: action, SourceID: sourceID})
	}
	return status, rows.Err()
}

func (s *Store) GetProcedureByID(id ids.ProcedureID) (ProcedureRecord, bool, error) {
	value, err := s.procedureValueByID(id.ToI64())
	if err == sql.ErrNoRows {
		return ProcedureRecord{}, false, nil
	}
	if err != nil {
		return ProcedureRecord{}, false, err
	}
	refs, err := s.sourceRefsFor("procedure_sources", "emedical_procedure_id", id.ToI64())
	if err != nil {
		return ProcedureRecord{}, false, err
	}
	return ProcedureRecord{ID: id, Value: value, Sources: refs}, true, nil
}

func (s *Store) ProcedureLabelByID(id ids.ProcedureID) (string, bool) {
	value, found, err := s.GetProcedureByID(id)
	if err != nil || !found || value.Value == nil {
		return "", false
	}
	return centitylabel.ProcedureLabel(value.Value), true
}

func (s *Store) GetDiseaseByID(id int64) (NamedRecord, bool, error) {
	return s.namedRecord("edisease", "disease_id", "disease_sources", "disease_id", id)
}

func (s *Store) GetIndexByID(id int64) (NamedRecord, bool, error) {
	return s.namedRecord("eindex", "index_id", "index_sources", "index_id", id)
}

func (s *Store) GetSymptomByID(id int64) (NamedRecord, bool, error) {
	return s.namedRecord("esymptom", "symptom_id", "symptom_sources", "symptom_id", id)
}

func (s *Store) namedRecord(table, idColumn, sourceTable, sourceColumn string, id int64) (NamedRecord, bool, error) {
	var name string
	err := s.QueryRowContext(s.ctx,
		`SELECT name FROM `+table+` WHERE `+idColumn+` = ?`, id).Scan(&name)
	if err == sql.ErrNoRows {
		return NamedRecord{}, false, nil
	}
	if err != nil {
		return NamedRecord{}, false, err
	}
	refs, err := s.sourceRefsFor(sourceTable, sourceColumn, id)
	if err != nil {
		return NamedRecord{}, false, err
	}
	return NamedRecord{ID: id, Name: name, Sources: refs}, true, nil
}

func (s *Store) GetMedicalProcedureOptionByID(id ids.MedicalProcedureOptionID) (medicalprocedureoption.MedicalProcedureOption, bool, error) {
	value, err := s.OptionRepository.Get(medicalprocedureoption.MedicalProcedureOptionID(id))
	if err == repo.ErrNotFound || err == sql.ErrNoRows {
		return medicalprocedureoption.MedicalProcedureOption{}, false, nil
	}
	if err != nil {
		return medicalprocedureoption.MedicalProcedureOption{}, false, err
	}
	return value, true, nil
}

var entityTables = map[string]struct {
	table    string
	idColumn string
}{
	"actor":                    {"actors", "actor_id"},
	"device":                   {"actors", "actor_id"},
	"action":                   {"e_actions", "action_id"},
	"status":                   {"e_statuses", "status_id"},
	"disease":                  {"edisease", "disease_id"},
	"index":                    {"eindex", "index_id"},
	"symptom":                  {"esymptom", "symptom_id"},
	"procedure":                {"eprocedure", "procedure_id"},
	"medical-procedure-option": {"medical_procedure_options", "medical_procedure_option_id"},
	"medical_procedure_option": {"medical_procedure_options", "medical_procedure_option_id"},
}

func (s *Store) GetEntityNeighbors(entityType string, id int64) (EntityNeighbors, bool, error) {
	entityType = strings.ToLower(strings.TrimSpace(entityType))
	info, ok := entityTables[entityType]
	if !ok {
		return EntityNeighbors{}, false, fmt.Errorf("%w: %s", ErrUnknownEntityType, entityType)
	}
	if id <= 0 {
		return EntityNeighbors{}, false, fmt.Errorf("잘못된 entity id: %d", id)
	}
	var existingID int64
	if err := s.QueryRowContext(s.ctx,
		fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", info.idColumn, info.table, info.idColumn),
		id).Scan(&existingID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return EntityNeighbors{}, false, nil
		}
		return EntityNeighbors{}, false, fmt.Errorf("%s neighbor 대상 조회 실패: %w", entityType, err)
	}
	var previousID, nextID sql.NullInt64
	query := fmt.Sprintf(`
		SELECT
			(SELECT %[2]s FROM %[1]s WHERE %[2]s < ? ORDER BY %[2]s DESC LIMIT 1),
			(SELECT %[2]s FROM %[1]s WHERE %[2]s > ? ORDER BY %[2]s ASC LIMIT 1)`,
		info.table, info.idColumn)
	if err := s.QueryRowContext(s.ctx, query, id, id).Scan(&previousID, &nextID); err != nil {
		return EntityNeighbors{}, false, fmt.Errorf("%s neighbor 조회 실패: %w", entityType, err)
	}
	return EntityNeighbors{PreviousID: previousID.Int64, NextID: nextID.Int64}, true, nil
}
