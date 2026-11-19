package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/G226Medi03/MedicalEntity/ename/edisease"
	"github.com/G226Medi03/MedicalEntity/ename/eindex"
	"github.com/G226Medi03/MedicalEntity/ename/estatus"
	"github.com/G226Medi03/MedicalEntity/ename/esymptom"
)

// StatusNameRef is the result of converting one status entity to another type.
type StatusNameRef struct {
	Type  string
	Name  string
	Value any
}

type statusTableInfo struct {
	table       string
	idColumn    string
	pkColumn    string
	sourceTable string
}

var statusTables = map[string]statusTableInfo{
	"symptom": {"esymptom", "symptom_id", "symptom_id", "symptom_sources"},
	"disease": {"edisease", "disease_id", "disease_id", "disease_sources"},
	"index":   {"eindex", "index_id", "index_id", "index_sources"},
}

// ConvertStatus converts a symptom/disease/index ename value to another
// status entity type. Numeric row IDs stay internal.
func (s *StoreContainer) ConvertStatus(sourceType string, sourceValue any, targetType, direction string) (StatusNameRef, error) {
	sourceInfo, ok := statusTables[sourceType]
	if !ok {
		return StatusNameRef{}, fmt.Errorf("지원하지 않는 status type: %s", sourceType)
	}
	targetInfo, ok := statusTables[targetType]
	if !ok || sourceType == targetType {
		return StatusNameRef{}, fmt.Errorf("지원하지 않는 status 변환: %s -> %s", sourceType, targetType)
	}
	if targetType == "index" && direction != "increased_index" && direction != "decreased_index" {
		return StatusNameRef{}, fmt.Errorf("index 변환 방향이 필요합니다: %s", direction)
	}
	name, err := statusValueName(sourceType, sourceValue)
	if err != nil {
		return StatusNameRef{}, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return StatusNameRef{}, fmt.Errorf("변환 source 이름이 비어 있습니다")
	}

	var ref StatusNameRef
	err = s.WithTx(s.ctx, func(tx *StoreContainer) error {
		var sourceID int64
		var sourceIDValue int64
		if err := tx.QueryRowContext(s.ctx,
			`SELECT `+sourceInfo.pkColumn+`, COALESCE(source_id, 0) FROM `+sourceInfo.table+` WHERE name = ? LIMIT 1`,
			name).Scan(&sourceID, &sourceIDValue); err != nil {
			return fmt.Errorf("변환 source entity 조회 실패: %w", err)
		}

		var targetID int64
		err := tx.QueryRowContext(s.ctx,
			`SELECT `+targetInfo.pkColumn+` FROM `+targetInfo.table+` WHERE name = ? LIMIT 1`, name).Scan(&targetID)
		if err == sql.ErrNoRows {
			result, insertErr := tx.ExecContext(s.ctx,
				`INSERT INTO `+targetInfo.table+` (name, source_id) VALUES (?, ?)`, name, sourceIDValue)
			if insertErr != nil {
				return fmt.Errorf("변환 target entity 생성 실패: %w", insertErr)
			}
			targetID, err = result.LastInsertId()
			if err != nil {
				return err
			}
		} else if err != nil {
			return fmt.Errorf("변환 target entity 조회 실패: %w", err)
		}

		if err := tx.moveStatusRows(name, sourceType, sourceID, targetType, direction); err != nil {
			return err
		}
		if err := tx.moveStatusEntitySources(sourceInfo, targetInfo, sourceID, targetID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(s.ctx,
			`DELETE FROM `+sourceInfo.table+` WHERE `+sourceInfo.pkColumn+` = ?`, sourceID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(s.ctx, `
			UPDATE status_other_name
			SET status_type = ?, status_id = ?
			WHERE name = ?`, targetType, targetID, name); err != nil {
			return err
		}
		ref = StatusNameRef{
			Type:  targetType,
			Name:  name,
			Value: statusValueForName(targetType, name, direction),
		}
		return nil
	})
	return ref, err
}

func statusValueName(sourceType string, value any) (string, error) {
	switch sourceType {
	case "disease":
		switch typed := value.(type) {
		case edisease.EDisease:
			return string(typed), nil
		case estatus.HaveDisease:
			return string(typed.Disease), nil
		}
	case "index":
		switch typed := value.(type) {
		case eindex.EIndex:
			return typed.Name, nil
		case estatus.IncreasedIndex:
			return typed.Index.Name, nil
		case estatus.DecreasedIndex:
			return typed.Index.Name, nil
		}
	case "symptom":
		switch typed := value.(type) {
		case esymptom.EOtherSymptom:
			return typed.Name, nil
		case estatus.HaveSymptom:
			if other, ok := typed.Symptom.(esymptom.EOtherSymptom); ok {
				return other.Name, nil
			}
			return "", fmt.Errorf("pain symptom 변환은 지원하지 않습니다")
		}
	}
	return "", fmt.Errorf("지원하지 않는 status source value: %T", value)
}

func statusValueForName(targetType, name, direction string) any {
	switch targetType {
	case "disease":
		return estatus.HaveDisease{Disease: edisease.EDisease(name)}
	case "index":
		index := eindex.EIndex{Name: name}
		if direction == "decreased_index" {
			return estatus.DecreasedIndex{Index: index}
		}
		return estatus.IncreasedIndex{Index: index}
	case "symptom":
		return estatus.HaveSymptom{Symptom: esymptom.EOtherSymptom{Name: name}}
	default:
		return nil
	}
}

func (s *StoreContainer) moveStatusRows(name, sourceType string, sourceID int64, targetType, direction string) error {
	query := ""
	switch sourceType {
	case "symptom":
		query = `SELECT s.estatus_id FROM estatus s JOIN status_have_symptom n ON n.id = s.ptr WHERE n.symptom_id = ?`
	case "disease":
		query = `SELECT s.estatus_id FROM estatus s JOIN status_have_disease n ON n.id = s.ptr WHERE n.disease_id = ?`
	case "index":
		query = `SELECT s.estatus_id FROM estatus s JOIN (SELECT id, index_id FROM status_increased_index UNION ALL SELECT id, index_id FROM status_decreased_index) n ON n.id = s.ptr WHERE n.index_id = ?`
	default:
		return fmt.Errorf("지원하지 않는 status type: %s", sourceType)
	}
	rows, err := s.QueryContext(s.ctx, query, sourceID)
	if err != nil {
		return err
	}
	statusIDs := []int64{}
	for rows.Next() {
		var statusID int64
		if err := rows.Scan(&statusID); err != nil {
			_ = rows.Close()
			return err
		}
		statusIDs = append(statusIDs, statusID)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return err
	}
	_ = rows.Close()

	entities, err := s.Entities()
	if err != nil {
		return err
	}
	for _, oldStatusID := range statusIDs {
		var value estatus.IEStatus
		switch targetType {
		case "symptom":
			value = estatus.HaveSymptom{Symptom: esymptom.EOtherSymptom{Name: name}}
		case "disease":
			value = estatus.HaveDisease{Disease: edisease.EDisease(name)}
		case "index":
			index := eindex.EIndex{Name: name}
			if direction == "decreased_index" {
				value = estatus.DecreasedIndex{Index: index}
			} else {
				value = estatus.IncreasedIndex{Index: index}
			}
		}
		if _, err := entities.InsertStatus(value, nil); err != nil {
			return err
		}
		newStatusID, err := s.statusIDForValue(value)
		if err != nil {
			return err
		}
		if newStatusID == oldStatusID {
			continue
		}
		if err := s.mergeStatusReferences(oldStatusID, newStatusID); err != nil {
			return err
		}
		if _, err := s.ExecContext(s.ctx, `DELETE FROM estatus WHERE estatus_id = ?`, oldStatusID); err != nil {
			return err
		}
	}
	return nil
}

func (s *StoreContainer) mergeStatusReferences(sourceID, targetID int64) error {
	for _, item := range []struct{ table, column string }{
		{"status_sources", "estatus_id"},
		{"status_actions", "estatus_id"},
		{"action_result_statuses", "estatus_id"},
	} {
		if _, err := s.ExecContext(s.ctx,
			`UPDATE OR IGNORE `+item.table+` SET `+item.column+` = ? WHERE `+item.column+` = ?`,
			targetID, sourceID); err != nil {
			return err
		}
		if _, err := s.ExecContext(s.ctx,
			`DELETE FROM `+item.table+` WHERE `+item.column+` = ?`, sourceID); err != nil {
			return err
		}
	}
	if _, err := s.ExecContext(s.ctx, `
		UPDATE OR IGNORE entity_sources SET entity_id = ?
		WHERE entity_kind = 'status' AND entity_id = ?`, targetID, sourceID); err != nil {
		return err
	}
	if _, err := s.ExecContext(s.ctx, `
		DELETE FROM entity_sources WHERE entity_kind = 'status' AND entity_id = ?`, sourceID); err != nil {
		return err
	}
	if _, err := s.ExecContext(s.ctx, `
		UPDATE OR IGNORE status_other_name SET status_id = ? WHERE status_id = ?`, targetID, sourceID); err != nil {
		return err
	}
	if _, err := s.ExecContext(s.ctx, `
		DELETE FROM status_other_name WHERE status_id = ?`, sourceID); err != nil {
		return err
	}
	rows, err := s.QueryContext(s.ctx, `
		SELECT medical_procedure_option_id, indication_status_ids_json FROM medical_procedure_options`)
	if err != nil {
		return err
	}
	defer rows.Close() //nolint:errcheck // deferred close
	type update struct {
		id     int64
		values []int64
	}
	updates := make([]update, 0)
	for rows.Next() {
		var id int64
		var raw string
		if err := rows.Scan(&id, &raw); err != nil {
			return err
		}
		var values []int64
		if err := json.Unmarshal([]byte(raw), &values); err != nil {
			return err
		}
		changed := false
		for index, value := range values {
			if value == sourceID {
				values[index] = targetID
				changed = true
			}
		}
		if changed {
			updates = append(updates, update{id: id, values: values})
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, item := range updates {
		encoded, err := json.Marshal(item.values)
		if err != nil {
			return err
		}
		if _, err := s.ExecContext(s.ctx,
			`UPDATE medical_procedure_options SET indication_status_ids_json = ? WHERE medical_procedure_option_id = ?`,
			string(encoded), item.id); err != nil {
			return err
		}
	}
	return nil
}

func (s *StoreContainer) moveStatusEntitySources(sourceInfo, targetInfo statusTableInfo, sourceID, targetID int64) error {
	if _, err := s.ExecContext(s.ctx, `
		INSERT OR IGNORE INTO `+targetInfo.sourceTable+`
			(`+targetInfo.idColumn+`, source_id)
		SELECT ?, source_id FROM `+sourceInfo.sourceTable+` WHERE `+sourceInfo.idColumn+` = ?`,
		targetID, sourceID); err != nil {
		return err
	}
	if _, err := s.ExecContext(s.ctx,
		`DELETE FROM `+sourceInfo.sourceTable+` WHERE `+sourceInfo.idColumn+` = ?`, sourceID); err != nil {
		return err
	}
	if _, err := s.ExecContext(s.ctx, `
		INSERT OR IGNORE INTO entity_sources (entity_kind, entity_id, source_id, created_at)
		SELECT ?, ?, source_id, created_at FROM entity_sources
		WHERE entity_kind = ? AND entity_id = ?`,
		targetInfo.table[1:], targetID, sourceInfo.table[1:], sourceID); err != nil {
		return err
	}
	if _, err := s.ExecContext(s.ctx,
		`DELETE FROM entity_sources WHERE entity_kind = ? AND entity_id = ?`,
		sourceInfo.table[1:], sourceID); err != nil {
		return err
	}
	return nil
}
