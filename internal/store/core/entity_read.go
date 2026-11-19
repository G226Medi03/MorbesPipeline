package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/G226Medi03/MedicalEntityParser/llmdto"
	"github.com/G226Medi03/MedicalEntityParser/llmparser"
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"strings"
)

func (s *StoreContainer) statusValueByID(id int64) (repo.IEStatus, error) {
	var kind string
	var ptr int64
	err := s.QueryRowContext(s.ctx,
		`SELECT kind, ptr FROM estatus WHERE estatus_id = ?`, id).Scan(&kind, &ptr)
	if err != nil {
		return nil, err
	}
	return s.statusValueByKindPtr(kind, ptr)
}

func (s *StoreContainer) statusValueByKindPtr(kind string, ptr int64) (repo.IEStatus, error) {
	switch repo.StatusKind(kind) {
	case repo.StatusKindIncreasedIndex:
		indexID, err := s.subtypeValue("status_increased_index", "index_id", ptr)
		if err != nil {
			return nil, err
		}
		name, err := s.indexNameByID(indexID)
		if err != nil {
			return nil, err
		}
		return repo.IncreasedIndex{Index: repo.EIndex{Name: name}}, nil
	case repo.StatusKindDecreasedIndex:
		indexID, err := s.subtypeValue("status_decreased_index", "index_id", ptr)
		if err != nil {
			return nil, err
		}
		name, err := s.indexNameByID(indexID)
		if err != nil {
			return nil, err
		}
		return repo.DecreasedIndex{Index: repo.EIndex{Name: name}}, nil
	case repo.StatusKindHaveDisease:
		diseaseID, err := s.subtypeValue("status_have_disease", "disease_id", ptr)
		if err != nil {
			return nil, err
		}
		name, err := s.diseaseNameByID(diseaseID)
		if err != nil {
			return nil, err
		}
		return repo.HaveDisease{Disease: repo.EDisease(name)}, nil
	case repo.StatusKindHaveSymptom:
		var symptomID int64
		err := s.QueryRowContext(s.ctx,
			`SELECT symptom_id FROM status_have_symptom WHERE id = ?`, ptr).Scan(&symptomID)
		if err != nil {
			return nil, err
		}
		symptom, err := s.symptomValueByID(symptomID)
		if err != nil {
			return nil, err
		}
		return repo.HaveSymptom{Symptom: symptom}, nil
	case repo.StatusKindDamaged, repo.StatusKindInflammation:
		table := "status_damaged"
		if repo.StatusKind(kind) == repo.StatusKindInflammation {
			table = "status_inflammation"
		}
		actorID, err := s.subtypeValue(table, "eactor_id", ptr)
		if err != nil {
			return nil, err
		}
		actor, err := s.actorValueByID(actorID)
		if err != nil {
			return nil, err
		}
		if repo.StatusKind(kind) == repo.StatusKindDamaged {
			return repo.Damaged{Actor: actor}, nil
		}
		return repo.Inflammation{Actor: actor}, nil
	case repo.StatusKindReceivedMedicalProcedure:
		var procedureID int64
		var period string
		err := s.QueryRowContext(s.ctx,
			`SELECT emedical_procedure_id, post_procedure_period FROM status_received_medical_procedure WHERE id = ?`, ptr).
			Scan(&procedureID, &period)
		if err != nil {
			return nil, err
		}
		procedure, err := s.procedureValueByID(procedureID)
		if err != nil {
			return nil, err
		}
		return repo.ReceivedMedicalProcedure{MedicalProcedure: procedure, PostProcedurePeriod: period}, nil
	default:
		return nil, fmt.Errorf("지원하지 않는 status kind 입니다: %s", kind)
	}
}

func (s *StoreContainer) subtypeValue(table, column string, ptr int64) (int64, error) {
	var value int64
	err := s.QueryRowContext(s.ctx,
		`SELECT `+column+` FROM `+table+` WHERE id = ?`, ptr).Scan(&value)
	return value, err
}

func (s *StoreContainer) diseaseNameByID(id int64) (string, error) {
	var name string
	err := s.QueryRowContext(s.ctx, `SELECT name FROM edisease WHERE disease_id = ?`, id).Scan(&name)
	return name, err
}

func (s *StoreContainer) indexNameByID(id int64) (string, error) {
	var name string
	err := s.QueryRowContext(s.ctx, `SELECT name FROM eindex WHERE index_id = ?`, id).Scan(&name)
	return name, err
}

func (s *StoreContainer) symptomValueByID(id int64) (repo.IESymptom, error) {
	var kind string
	var ptr int64
	if err := s.QueryRowContext(s.ctx,
		`SELECT kind, ptr FROM esymptom WHERE symptom_id = ?`, id).Scan(&kind, &ptr); err != nil {
		return nil, err
	}
	switch repo.SymptomKind(kind) {
	case repo.SymptomKindOtherSymptom:
		var name string
		if err := s.QueryRowContext(s.ctx,
			`SELECT name FROM symptom_other WHERE id = ?`, ptr).Scan(&name); err != nil {
			return nil, err
		}
		return repo.EOtherSymptom{Name: name}, nil
	case repo.SymptomKindPain:
		var siteKind, siteName, aspectsJSON string
		if err := s.QueryRowContext(s.ctx,
			`SELECT site_kind, site_name, aspects_json FROM symptom_pain WHERE id = ?`, ptr).
			Scan(&siteKind, &siteName, &aspectsJSON); err != nil {
			return nil, err
		}
		var rawAspects []json.RawMessage
		if err := json.Unmarshal([]byte(aspectsJSON), &rawAspects); err != nil {
			return nil, err
		}
		aspects := make([]repo.IPainAspect, 0, len(rawAspects))
		for _, raw := range rawAspects {
			var aspectDTO llmdto.Aspect
			if err := json.Unmarshal(raw, &aspectDTO); err != nil {
				return nil, err
			}
			aspect, err := llmparser.AspectFromDTO(aspectDTO)
			if err != nil {
				return nil, err
			}
			aspects = append(aspects, aspect)
		}
		return repo.EPain{Site: actorFromKind(siteKind, siteName), Aspects: aspects}, nil
	default:
		return nil, fmt.Errorf("지원하지 않는 symptom kind 입니다: %s", kind)
	}
}

func actorFromKind(kind, name string) repo.IEActor {
	switch repo.ParseActorKind(kind) {
	case repo.ActorKindDevice:
		return repo.NewDevice(name)
	case repo.ActorKindProcedure:
		return repo.NewActorProcedure(name)
	default:
		return repo.NewUnknownActor(name)
	}
}

func (s *StoreContainer) actorValueByID(id int64) (repo.IEActor, error) {
	var kind, name string
	err := s.QueryRowContext(s.ctx,
		`SELECT kind, name FROM eactor WHERE eactor_id = ?`, id).Scan(&kind, &name)
	if err == sql.ErrNoRows {
		return repo.NewUnknownActor(""), nil
	}
	if err != nil {
		return nil, err
	}
	switch repo.ParseActorKind(kind) {
	case repo.ActorKindDevice:
		return repo.NewDevice(name), nil
	case repo.ActorKindProcedure:
		return repo.NewActorProcedure(name), nil
	default:
		return repo.NewUnknownActor(name), nil
	}
}

func (s *StoreContainer) procedureValueByID(id int64) (repo.IMedicalProcedure, error) {
	var kind string
	var ptr int64
	err := s.QueryRowContext(s.ctx,
		`SELECT kind, ptr FROM emedical_procedure WHERE emedical_procedure_id = ?`, id).Scan(&kind, &ptr)
	if err != nil {
		return nil, err
	}
	return s.procedureValueByKindPtr(kind, ptr)
}

func (s *StoreContainer) procedureValueByKindPtr(kind string, ptr int64) (repo.IMedicalProcedure, error) {
	switch kind {
	case "Surgery", "DiagnosticTest", "Unknown", "Procedure":
		var name string
		err := s.QueryRowContext(s.ctx, `SELECT name FROM procedure_named WHERE id = ?`, ptr).Scan(&name)
		if err != nil {
			return nil, err
		}
		switch kind {
		case "Surgery":
			return repo.Surgery{Name: name}, nil
		case "DiagnosticTest":
			return repo.DiagnosticTest{Name: name}, nil
		default:
			return repo.UnknownMedicalProcedure{Name: name}, nil
		}
	case "Education":
		var name, topic string
		err := s.QueryRowContext(s.ctx,
			`SELECT name, topic FROM procedure_education WHERE id = ?`, ptr).Scan(&name, &topic)
		if err != nil {
			return nil, err
		}
		return repo.Education{Name: name, Topic: topic}, nil
	case "Instruction":
		var name, action, duration, position string
		err := s.QueryRowContext(s.ctx,
			`SELECT name, action, duration, position FROM procedure_instruction WHERE id = ?`, ptr).
			Scan(&name, &action, &duration, &position)
		if err != nil {
			return nil, err
		}
		return repo.Instruction{Name: name, Action: action, Duration: duration, Position: position}, nil
	case "Insertion":
		var name, tube, site string
		err := s.QueryRowContext(s.ctx,
			`SELECT name, tube, site FROM procedure_insertion WHERE id = ?`, ptr).Scan(&name, &tube, &site)
		if err != nil {
			return nil, err
		}
		var siteActor repo.IEActor = repo.NewUnknownActor("")
		if strings.TrimSpace(site) != "" {
			siteActor = repo.NewUnknownActor(site)
		}
		return repo.Insertion{Device: repo.NewDevice(name), Site: siteActor}, nil
	case "Infuse":
		var materialID int64
		var route string
		err := s.QueryRowContext(s.ctx,
			`SELECT material_eactor_id, route FROM procedure_infuse WHERE id = ?`, ptr).Scan(&materialID, &route)
		if err != nil {
			return nil, err
		}
		material, err := s.actorValueByID(materialID)
		if err != nil {
			return nil, err
		}
		return repo.Infuse{Material: material, Route: repo.InfuseRoute(route)}, nil
	case "Transplant", "Cleaning":
		var targetKind, targetName string
		var targetID sql.NullInt64
		table := "procedure_transplant"
		if kind == "Cleaning" {
			table = "procedure_cleaning"
		}
		err := s.QueryRowContext(s.ctx,
			`SELECT target_kind, target_eactor_id, target_name FROM `+table+` WHERE id = ?`, ptr).
			Scan(&targetKind, &targetID, &targetName)
		if err != nil {
			return nil, err
		}
		target, err := s.procedureTargetValue(targetKind, targetID, targetName)
		if err != nil {
			return nil, err
		}
		if kind == "Transplant" {
			return repo.Transplant{Target: target}, nil
		}
		return repo.Cleaning{Target: target}, nil
	default:
		return nil, fmt.Errorf("지원하지 않는 procedure kind 입니다: %s", kind)
	}
}

func (s *StoreContainer) procedureTargetValue(kind string, id sql.NullInt64, name string) (repo.IEActor, error) {
	if id.Valid && id.Int64 > 0 {
		return s.actorValueByID(id.Int64)
	}
	if strings.TrimSpace(name) == "" {
		return repo.NewUnknownActor(""), nil
	}
	if repo.ParseActorKind(kind) == repo.ActorKindDevice {
		return repo.NewDevice(name), nil
	}
	return repo.NewUnknownActor(name), nil
}

func (s *StoreContainer) actionValueByKindPtr(kind string, ptr int64) (repo.IEAction, error) {
	switch repo.ActionKind(kind) {
	case repo.ActionKindCauseStatus:
		var statusID int64
		err := s.QueryRowContext(s.ctx,
			`SELECT estatus_id FROM action_cause_status WHERE id = ?`, ptr).Scan(&statusID)
		if err != nil {
			return nil, err
		}
		status, err := s.statusValueByID(statusID)
		if err != nil {
			return nil, err
		}
		return repo.CauseStatus{Status: status}, nil
	case repo.ActionKindBlock, repo.ActionKindRelax:
		table := "action_block"
		if repo.ActionKind(kind) == repo.ActionKindRelax {
			table = "action_relax"
		}
		var actorID int64
		err := s.QueryRowContext(s.ctx,
			`SELECT eactor_id FROM `+table+` WHERE id = ?`, ptr).Scan(&actorID)
		if err != nil {
			return nil, err
		}
		actor, err := s.actorValueByID(actorID)
		if err != nil {
			return nil, err
		}
		if repo.ActionKind(kind) == repo.ActionKindBlock {
			return repo.Block{Actor: actor}, nil
		}
		return repo.Relax{Actor: actor}, nil
	default:
		return nil, fmt.Errorf("지원하지 않는 action kind 입니다: %s", kind)
	}
}
