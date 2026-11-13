package store

import (
	"database/sql"
	"fmt"
	"strings"

	entity "github.com/G226Medi03/MorbesPipeline/legacy"
	"github.com/G226Medi03/SourceEntity"
)

// EnsureProcedureActor creates (or reuses) the actor row for an eprocedure.
// The actor kind is procedure and its ptr points to actor_procedure.id.
func (s *Store) EnsureProcedureActor(procedureID entity.ProcedureID, name string, sourceID source.SourceId) (entity.ActorID, error) {
	unlock := s.lockWrite()
	defer unlock()
	return s.ensureProcedureActor(procedureID, name, sourceID)
}

func (s *Store) ensureProcedureActor(procedureID entity.ProcedureID, name string, sourceID source.SourceId) (entity.ActorID, error) {
	if procedureID <= 0 {
		return 0, fmt.Errorf("procedure_id가 비어 있습니다")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		if label, ok := s.ProcedureLabelByID(procedureID); ok {
			name = strings.TrimSpace(label)
		}
	}
	if name == "" {
		name = fmt.Sprintf("procedure:%d", procedureID)
	}
	var ptr int64
	err := s.QueryRowContext(s.ctx, `SELECT id FROM actor_procedure WHERE emedical_procedure_id = ?`, int64(procedureID)).Scan(&ptr)
	if err == sql.ErrNoRows {
		result, insertErr := s.ExecContext(s.ctx, `
			INSERT INTO actor_procedure (emedical_procedure_id, name) VALUES (?, ?)`,
			int64(procedureID), name)
		if insertErr != nil {
			return 0, fmt.Errorf("actor_procedure 저장 실패: %w", insertErr)
		}
		ptr, err = result.LastInsertId()
		if err != nil {
			return 0, err
		}
	} else if err != nil {
		return 0, fmt.Errorf("actor_procedure 조회 실패: %w", err)
	}
	if ptr > 0 {
		if _, err := s.ExecContext(s.ctx, `UPDATE actor_procedure SET name = ? WHERE id = ?`, name, ptr); err != nil {
			return 0, fmt.Errorf("actor_procedure 이름 갱신 실패: %w", err)
		}
	}

	var actorID entity.ActorID
	err = s.QueryRowContext(s.ctx,
		`SELECT eactor_id FROM eactor WHERE kind = 'procedure' AND ptr = ?`, ptr).Scan(&actorID)
	if err == sql.ErrNoRows {
		result, insertErr := s.ExecContext(s.ctx, `
			INSERT INTO eactor (name, kind, ptr, source_id) VALUES (?, 'procedure', ?, ?)`,
			name, ptr, sourceID.ToI64())
		if insertErr != nil {
			return 0, fmt.Errorf("procedure actor 저장 실패: %w", insertErr)
		}
		rawID, idErr := result.LastInsertId()
		if idErr != nil {
			return 0, idErr
		}
		actorID = entity.ActorID(rawID)
	} else if err != nil {
		return 0, fmt.Errorf("procedure actor 조회 실패: %w", err)
	}
	if _, err := s.ExecContext(s.ctx,
		`UPDATE eactor SET name = ?, source_id = CASE WHEN source_id = 0 THEN ? ELSE source_id END
		 WHERE eactor_id = ?`, name, sourceID.ToI64(), int64(actorID)); err != nil {
		return 0, fmt.Errorf("procedure actor 갱신 실패: %w", err)
	}
	if sourceID.ToI64() > 0 {
		if _, err := s.ExecContext(s.ctx,
			`INSERT OR IGNORE INTO actor_sources (eactor_id, source_id) VALUES (?, ?)`,
			actorID.ToI64(), sourceID.ToI64()); err != nil {
			return 0, err
		}
	}
	if err := s.validateKindPtr("eactor", "emedical_procedure"); err != nil {
		return 0, err
	}
	return actorID, nil
}
