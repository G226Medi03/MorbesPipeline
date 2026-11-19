package store

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/G226Medi03/MedicalEntity/ename/eactor"
	"github.com/G226Medi03/MedicalEntity/ename/emedical_procedure"
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	source "github.com/G226Medi03/SourceEntity"
)

// UnresolvedTransplantTarget is a transplant procedure whose target actor kind
// is still unknown.
type UnresolvedTransplantTarget struct {
	Procedure    emedical_procedure.IMedicalProcedure
	TargetName   string
	DetectedKind string
}

// ListUnresolvedTransplantTargets returns transplant procedures with unknown
// target actor kind. The procedure is returned as an ename value.
func (s *StoreContainer) ListUnresolvedTransplantTargets() ([]UnresolvedTransplantTarget, error) {
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
		var procedureID int64
		var targetName string
		if err := rows.Scan(&procedureID, &targetName); err != nil {
			return nil, err
		}
		procedure, valueErr := s.procedureValueByID(procedureID)
		if valueErr != nil {
			continue
		}
		item := UnresolvedTransplantTarget{
			Procedure:  procedure,
			TargetName: strings.TrimSpace(targetName),
		}
		if item.TargetName == "" {
			continue
		}
		if record, found, actorErr := s.GetActorByName(item.TargetName); actorErr == nil && found {
			item.DetectedKind = record.Name().Kind()
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

// ResolveTransplantTarget resolves a transplant procedure target by ename value.
func (s *StoreContainer) ResolveTransplantTarget(procedure emedical_procedure.IMedicalProcedure, kind string) error {
	kind = normalizeTransplantTargetKind(kind)
	if kind == "" {
		return fmt.Errorf("지원하지 않는 transplant target kind 입니다: %s", kind)
	}
	procedureID, err := s.procedureIDForValue(procedure)
	if err != nil {
		return fmt.Errorf("transplant procedure를 찾을 수 없습니다: %w", err)
	}
	var rowKind string
	var ptr int64
	err = s.QueryRowContext(s.ctx,
		`SELECT kind, ptr FROM emedical_procedure WHERE emedical_procedure_id = ?`, procedureID).
		Scan(&rowKind, &ptr)
	if err == sql.ErrNoRows || rowKind != "Transplant" {
		return fmt.Errorf("transplant procedure를 찾을 수 없습니다: %d", procedureID)
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
		return fmt.Errorf("target_name이 비어 있습니다: %d", procedureID)
	}
	var sourceID source.SourceId
	err = s.QueryRowContext(s.ctx,
		`SELECT source_id FROM procedure_sources WHERE emedical_procedure_id = ? ORDER BY id LIMIT 1`, procedureID).
		Scan(&sourceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("procedure %d source가 없습니다", procedureID)
		}
		return err
	}
	var actorValue eactor.IEActor
	if kind == "device" {
		actorValue = eactor.NewDevice(targetName)
	} else {
		actorValue = eactor.NewUnknownActor(targetName)
	}
	entityStore, err := s.Entities()
	if err != nil {
		return err
	}
	if _, err := entityStore.InsertActor(actorValue, nil, []source.SourceId{sourceID}); err != nil {
		return err
	}
	actorID, found, err := s.actorIDByName(targetName)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("transplant target actor를 찾을 수 없습니다: %s", targetName)
	}
	unlock := s.lockWrite()
	defer unlock()
	if _, err := s.ExecContext(s.ctx,
		`UPDATE procedure_transplant SET target_kind = ?, target_eactor_id = ?, target_name = '' WHERE id = ?`,
		kind, actorID, ptr); err != nil {
		return err
	}
	return nil
}

func normalizeTransplantTargetKind(kind string) string {
	switch repo.ParseActorKind(kind) {
	case repo.ActorKindDevice:
		return "device"
	case repo.ActorKindUnknownActor:
		return "unknown_actor"
	default:
		return ""
	}
}
