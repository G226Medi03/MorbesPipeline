package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	medicalcodec "github.com/G226Medi03/MedicalEntity/codec"
	"strings"

	medicalentity "github.com/G226Medi03/MedicalEntity/entity"
	entity "github.com/G226Medi03/MorbesPipeline/legacy"
)

type StatusNameRef struct {
	Type string
	ID   entity.StatusEntityID
	Name string
}

func normalize_status_name(name string) string {
	return strings.TrimSpace(name)
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

// ConvertStatusEntity moves an entity to another status entity type, merging
// the status rows and all known status references.
func (s *Store) ConvertStatusEntity(sourceType string, sourceID entity.StatusEntityID, targetType, direction string) (StatusNameRef, error) {
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

	var ref StatusNameRef
	err := s.WithTx(s.ctx, func(tx *Store) error {
		var name string
		var sourceIDValue int64
		if err := tx.QueryRowContext(s.ctx,
			`SELECT name, COALESCE(source_id, 0) FROM `+sourceInfo.table+` WHERE `+sourceInfo.pkColumn+` = ?`,
			sourceID.ToI64()).Scan(&name, &sourceIDValue); err != nil {
			return fmt.Errorf("변환 source entity 조회 실패: %w", err)
		}

		var targetID entity.StatusEntityID
		err := tx.QueryRowContext(s.ctx,
			`SELECT `+targetInfo.pkColumn+` FROM `+targetInfo.table+` WHERE name = ?`, name).Scan(&targetID)
		if err == sql.ErrNoRows {
			result, err := tx.ExecContext(s.ctx,
				`INSERT INTO `+targetInfo.table+` (name, source_id) VALUES (?, ?)`,
				name, sourceIDValue)
			if err != nil {
				return fmt.Errorf("변환 target entity 생성 실패: %w", err)
			}
			rawTargetID, err := result.LastInsertId()
			if err != nil {
				return err
			}
			targetID = entity.StatusEntityID(rawTargetID)
		} else if err != nil {
			return fmt.Errorf("변환 target entity 조회 실패: %w", err)
		}

		if err := tx.move_status_rows(name, sourceType, sourceID, targetType, targetID, direction); err != nil {
			return err
		}
		if err := tx.move_status_entity_sources(sourceInfo, targetInfo, sourceID, targetID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(s.ctx,
			`DELETE FROM `+sourceInfo.table+` WHERE `+sourceInfo.pkColumn+` = ?`, sourceID.ToI64()); err != nil {
			return err
		}
		if _, err := tx.ExecContext(s.ctx, `
			UPDATE status_other_name
			SET status_type = ?, status_id = ?
			WHERE name = ?`, targetType, targetID.ToI64(), normalize_status_name(name)); err != nil {
			return err
		}
		ref = StatusNameRef{Type: targetType, ID: targetID, Name: name}
		return nil
	})
	return ref, err
}

func (s *Store) move_status_rows(name, sourceType string, sourceID entity.StatusEntityID, targetType string, targetID entity.StatusEntityID, direction string) error {
	sourceInfo := statusTables[sourceType]
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
	rows, err := s.QueryContext(s.ctx, query, sourceID.ToI64())
	if err != nil {
		return err
	}
	statusIDs := []entity.StatusID{}
	for rows.Next() {
		var statusID entity.StatusID
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
	for _, statusID := range statusIDs {
		current, err := s.StatusRepository.Get(entityStatusID(statusID))
		if err != nil {
			if err == medicalentity.ErrNotFound {
				continue
			}
			return err
		}
		rawCurrent, err := medicalcodec.MarshalStatus(current.Name())
		if err != nil {
			return err
		}
		data := map[string]any{}
		if err := json.Unmarshal(rawCurrent, &data); err != nil {
			return err
		}
		delete(data, sourceInfo.idColumn)
		if current.Name() == nil {
			return fmt.Errorf("status name이 비어 있습니다: %d", statusID.ToI64())
		}
		targetKind := current.Name().Kind()
		switch targetType {
		case "symptom":
			targetKind = "have_symptom"
			data["symptom_id"] = targetID.ToI64()
		case "disease":
			targetKind = "have_disease"
			data["disease_id"] = targetID.ToI64()
		case "index":
			targetKind = direction
			data["index_id"] = targetID.ToI64()
		}
		encoded, err := json.Marshal(data)
		if err != nil {
			return err
		}
		localStatus, err := entity.UnmarshalStatus(with_kind(targetKind, string(encoded)))
		if err != nil {
			return err
		}
		saved, err := s.StatusRepository.Insert(entity.ToExternalStatus(localStatus), nil)
		if err != nil {
			return err
		}
		newID := localStatusID(saved.Id())
		if newID == statusID {
			continue
		}
		if err := s.merge_status_references(statusID, newID); err != nil {
			return err
		}
		if _, err := s.ExecContext(s.ctx, `DELETE FROM estatus WHERE estatus_id = ?`, statusID.ToI64()); err != nil {
			return err
		}
	}
	return s.validateKindPtr("estatus")
}

func (s *Store) merge_status_references(sourceID, targetID entity.StatusID) error {
	for _, item := range []struct {
		table  string
		column string
	}{
		{"status_sources", "estatus_id"},
		{"status_actions", "estatus_id"},
		{"action_result_statuses", "estatus_id"},
	} {
		if _, err := s.ExecContext(s.ctx,
			`UPDATE OR IGNORE `+item.table+` SET `+item.column+` = ? WHERE `+item.column+` = ?`,
			targetID.ToI64(), sourceID.ToI64()); err != nil {
			return err
		}
		if _, err := s.ExecContext(s.ctx,
			`DELETE FROM `+item.table+` WHERE `+item.column+` = ?`, sourceID.ToI64()); err != nil {
			return err
		}
	}
	if _, err := s.ExecContext(s.ctx, `
		UPDATE OR IGNORE entity_sources SET entity_id = ?
		WHERE entity_kind = 'status' AND entity_id = ?`, targetID.ToI64(), sourceID.ToI64()); err != nil {
		return err
	}
	if _, err := s.ExecContext(s.ctx, `
		DELETE FROM entity_sources WHERE entity_kind = 'status' AND entity_id = ?`, sourceID.ToI64()); err != nil {
		return err
	}
	if _, err := s.ExecContext(s.ctx, `
		UPDATE OR IGNORE status_other_name SET status_id = ? WHERE status_id = ?`, targetID.ToI64(), sourceID.ToI64()); err != nil {
		return err
	}
	if _, err := s.ExecContext(s.ctx, `
		DELETE FROM status_other_name WHERE status_id = ?`, sourceID.ToI64()); err != nil {
		return err
	}
	rows, err := s.QueryContext(s.ctx, `
		SELECT medical_procedure_option_id, indication_status_ids_json FROM medical_procedure_options`)
	if err != nil {
		return err
	}
	defer rows.Close() //nolint:errcheck // deferred close
	type update struct {
		id     entity.MedicalProcedureOptionID
		values []entity.StatusID
	}
	updates := make([]update, 0)
	for rows.Next() {
		var id entity.MedicalProcedureOptionID
		var raw string
		if err := rows.Scan(&id, &raw); err != nil {
			return err
		}
		var values []entity.StatusID
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
			string(encoded), item.id.ToI64()); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) move_status_entity_sources(sourceInfo, targetInfo statusTableInfo, sourceID, targetID entity.StatusEntityID) error {
	if _, err := s.ExecContext(s.ctx, `
		INSERT OR IGNORE INTO `+targetInfo.sourceTable+`
			(`+targetInfo.idColumn+`, source_id)
		SELECT ?, source_id FROM `+sourceInfo.sourceTable+` WHERE `+sourceInfo.idColumn+` = ?`,
		targetID.ToI64(), sourceID.ToI64()); err != nil {
		return err
	}
	if _, err := s.ExecContext(s.ctx,
		`DELETE FROM `+sourceInfo.sourceTable+` WHERE `+sourceInfo.idColumn+` = ?`, sourceID.ToI64()); err != nil {
		return err
	}
	if _, err := s.ExecContext(s.ctx, `
		INSERT OR IGNORE INTO entity_sources (entity_kind, entity_id, source_id, created_at)
		SELECT ?, ?, source_id, created_at FROM entity_sources
		WHERE entity_kind = ? AND entity_id = ?`,
		targetInfo.table[1:], targetID.ToI64(), sourceInfo.table[1:], sourceID.ToI64()); err != nil {
		return err
	}
	if _, err := s.ExecContext(s.ctx,
		`DELETE FROM entity_sources WHERE entity_kind = ? AND entity_id = ?`,
		sourceInfo.table[1:], sourceID.ToI64()); err != nil {
		return err
	}
	return nil
}
