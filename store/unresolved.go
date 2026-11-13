package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/G226Medi03/MedicalEntity/ename/eactor"
	entity "github.com/G226Medi03/MorbesPipeline/legacy"
	"github.com/G226Medi03/SourceEntity"
)

type UnresolvedTransplantTarget struct {
	ProcedureID  entity.ProcedureID `json:"procedureId"`
	TargetName   string             `json:"targetName"`
	DetectedKind string             `json:"detectedKind,omitempty"`
}

func (s *Store) ListUnresolvedTransplantTargets() ([]UnresolvedTransplantTarget, error) {
	rows, err := s.QueryContext(s.ctx, `
		SELECT p.emedical_procedure_id, t.target_name
		FROM emedical_procedure p
		JOIN procedure_transplant t ON t.id = p.ptr
		WHERE p.kind = 'Transplant'
		  AND t.target_kind = 'unknown'
		ORDER BY p.emedical_procedure_id`)
	if err != nil {
		return nil, fmt.Errorf("unresolved transplant 조회 실패: %w", err)
	}
	defer rows.Close() //nolint:errcheck // deferred close
	result := make([]UnresolvedTransplantTarget, 0)
	for rows.Next() {
		var item UnresolvedTransplantTarget
		var targetName string
		if err := rows.Scan(&item.ProcedureID, &targetName); err != nil {
			return nil, err
		}
		item.TargetName = strings.TrimSpace(targetName)
		if item.TargetName == "" {
			continue
		}
		if record, found, actorErr := s.GetActorRecordByName(item.TargetName); actorErr == nil && found {
			if kind, kindErr := actorKind(record.Value); kindErr == nil {
				item.DetectedKind = kind
			}
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (s *Store) ResolveTransplantTarget(id entity.ProcedureID, kind string) error {
	kind = normalizeTransplantTargetKind(kind)
	if kind == "" {
		return fmt.Errorf("지원하지 않는 transplant target kind 입니다: %s", kind)
	}
	var rowKind string
	var ptr int64
	err := s.QueryRowContext(s.ctx,
		`SELECT kind, ptr FROM emedical_procedure WHERE emedical_procedure_id = ?`, id.ToI64()).
		Scan(&rowKind, &ptr)
	if err == sql.ErrNoRows || rowKind != "Transplant" {
		return fmt.Errorf("transplant procedure를 찾을 수 없습니다: %d", id)
	}
	if err != nil {
		return err
	}
	var targetName string
	if err := s.QueryRowContext(s.ctx,
		`SELECT target_name FROM procedure_transplant WHERE id = ?`, ptr).Scan(&targetName); err != nil {
		return err
	}
	targetName = strings.TrimSpace(targetName)
	if targetName == "" {
		return fmt.Errorf("target_name이 비어 있습니다: %d", id)
	}
	var sourceID source.SourceId
	err = s.QueryRowContext(s.ctx,
		`SELECT source_id FROM procedure_sources WHERE emedical_procedure_id = ? ORDER BY id LIMIT 1`, id.ToI64()).
		Scan(&sourceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("procedure %d source가 없습니다", id)
		}
		return err
	}
	sources := []source.SourceRef{{SourceID: sourceID}}
	kind = normalizeTransplantTargetKind(kind)
	var actorID entity.ActorID
	switch kind {
	case "device":
		saved, err := s.SaveDevice(eactor.Device{Name: targetName}, sources)
		if err != nil {
			return err
		}
		actorID = saved.ID
	case "unknown_actor":
		saved, err := s.SaveUnknownActor(eactor.UnknownActor{Name: targetName}, sources)
		if err != nil {
			return err
		}
		actorID = saved.ID
	default:
		return fmt.Errorf("지원하지 않는 transplant target kind 입니다: %s", kind)
	}
	unlock := s.lockWrite()
	defer unlock()
	if _, err = s.ExecContext(s.ctx,
		`UPDATE procedure_transplant SET target_kind = ?, target_eactor_id = ?, target_name = '' WHERE id = ?`,
		kind, actorID.ToI64(), ptr); err != nil {
		return err
	}
	return s.validateKindPtr("eactor", "emedical_procedure")
}

func normalizeTransplantTargetKind(kind string) string {
	switch string(eactor.ParseKind(kind)) {
	case "device":
		return "device"
	case "unknown_actor":
		return "unknown_actor"
	default:
		return ""
	}
}

var _ = json.Marshal
