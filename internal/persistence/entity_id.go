package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"github.com/G226Medi03/MorbesPipeline/internal/support/ids"
)

func (s *Store) statusIDForValue(value repo.IEStatus) (ids.StatusID, error) {
	kind, ptr, err := s.statusPtr(value)
	if err != nil {
		return 0, err
	}
	return s.statusIDByKindPtr(kind, ptr)
}

func (s *Store) statusIDByKindPtr(kind string, ptr int64) (ids.StatusID, error) {
	var id ids.StatusID
	err := s.QueryRowContext(s.ctx,
		`SELECT estatus_id FROM estatus WHERE kind = ? AND ptr = ?`, kind, ptr).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("status row not found: kind=%s ptr=%d", kind, ptr)
	}
	return id, err
}

func (s *Store) statusPtr(value repo.IEStatus) (string, int64, error) {
	switch typed := value.(type) {
	case repo.HaveDisease:
		diseaseID, err := s.diseaseIDByName(string(typed.Disease))
		if err != nil {
			return "", 0, err
		}
		return s.subtypePtr("status_have_disease", "disease_id", diseaseID, string(repo.StatusKindHaveDisease))
	case repo.IncreasedIndex:
		indexID, err := s.indexIDByName(typed.Index.Name)
		if err != nil {
			return "", 0, err
		}
		return s.subtypePtr("status_increased_index", "index_id", indexID, string(repo.StatusKindIncreasedIndex))
	case repo.DecreasedIndex:
		indexID, err := s.indexIDByName(typed.Index.Name)
		if err != nil {
			return "", 0, err
		}
		return s.subtypePtr("status_decreased_index", "index_id", indexID, string(repo.StatusKindDecreasedIndex))
	case repo.HaveSymptom:
		symptomID, err := s.symptomIDForValue(typed.Symptom)
		if err != nil {
			return "", 0, err
		}
		return s.subtypePtr("status_have_symptom", "symptom_id", symptomID, string(repo.StatusKindHaveSymptom))
	case repo.Damaged:
		actorID, err := s.actorIDForValue(typed.Actor)
		if err != nil {
			return "", 0, err
		}
		return s.subtypePtr("status_damaged", "eactor_id", actorID.ToI64(), string(repo.StatusKindDamaged))
	case repo.Inflammation:
		actorID, err := s.actorIDForValue(typed.Actor)
		if err != nil {
			return "", 0, err
		}
		return s.subtypePtr("status_inflammation", "eactor_id", actorID.ToI64(), string(repo.StatusKindInflammation))
	case repo.ReceivedMedicalProcedure:
		procedureID, err := s.procedureIDForValue(typed.MedicalProcedure)
		if err != nil {
			return "", 0, err
		}
		return s.subtypePtr("status_received_medical_procedure", "emedical_procedure_id", procedureID.ToI64(), string(repo.StatusKindReceivedMedicalProcedure))
	default:
		return "", 0, fmt.Errorf("지원하지 않는 status 타입입니다: %T", value)
	}
}

func (s *Store) subtypePtr(table, column string, value int64, kind string) (string, int64, error) {
	var ptr int64
	err := s.QueryRowContext(s.ctx,
		`SELECT id FROM `+table+` WHERE `+column+` = ?`, value).Scan(&ptr)
	if err == sql.ErrNoRows {
		return "", 0, fmt.Errorf("%s row not found: %s=%d", table, column, value)
	}
	return kind, ptr, err
}

func (s *Store) actionIDForValue(value repo.IEAction) (ids.ActionID, error) {
	kind, ptr, err := s.actionPtr(value)
	if err != nil {
		return 0, err
	}
	var id ids.ActionID
	err = s.QueryRowContext(s.ctx,
		`SELECT eaction_id FROM eaction WHERE kind = ? AND ptr = ?`, kind, ptr).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("action row not found: kind=%s ptr=%d", kind, ptr)
	}
	return id, err
}

func (s *Store) actionPtr(value repo.IEAction) (string, int64, error) {
	switch typed := value.(type) {
	case repo.CauseStatus:
		if typed.Status == nil {
			return "", 0, fmt.Errorf("cause_status status가 비어 있습니다")
		}
		statusID, err := s.statusIDForValue(typed.Status)
		if err != nil {
			return "", 0, err
		}
		return s.subtypePtr("action_cause_status", "estatus_id", statusID.ToI64(), string(repo.ActionKindCauseStatus))
	case repo.Block:
		actorID, err := s.actorIDForValue(typed.Actor)
		if err != nil {
			return "", 0, err
		}
		return s.subtypePtr("action_block", "eactor_id", actorID.ToI64(), string(repo.ActionKindBlock))
	case repo.Relax:
		actorID, err := s.actorIDForValue(typed.Actor)
		if err != nil {
			return "", 0, err
		}
		return s.subtypePtr("action_relax", "eactor_id", actorID.ToI64(), string(repo.ActionKindRelax))
	default:
		return "", 0, fmt.Errorf("지원하지 않는 action 타입입니다: %T", value)
	}
}

func (s *Store) actorIDForValue(value repo.IEActor) (ids.ActorID, error) {
	if value == nil {
		return 0, fmt.Errorf("actor value is nil")
	}
	name := strings.TrimSpace(value.ActorName())
	if name == "" {
		return 0, fmt.Errorf("actor name이 비어 있습니다")
	}
	kind := string(repo.ParseActorKind(value.Kind()))
	var id ids.ActorID
	err := s.QueryRowContext(s.ctx,
		`SELECT eactor_id FROM eactor WHERE kind = ? AND name = ?`, kind, name).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}
	found, foundOK, foundErr := s.actorIDByName(name)
	if foundErr != nil {
		return 0, foundErr
	}
	if foundOK {
		return found, nil
	}
	return 0, fmt.Errorf("actor row not found: kind=%s name=%s", kind, name)
}

func (s *Store) actorIDByName(name string) (ids.ActorID, bool, error) {
	var id ids.ActorID
	err := s.QueryRowContext(s.ctx,
		`SELECT eactor_id FROM actor_other_name WHERE name = ?`, strings.TrimSpace(name)).Scan(&id)
	if err == nil {
		return id, true, nil
	}
	if err != sql.ErrNoRows {
		return 0, false, err
	}
	err = s.QueryRowContext(s.ctx,
		`SELECT eactor_id FROM eactor WHERE name = ? AND kind <> 'procedure'`, strings.TrimSpace(name)).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	return id, err == nil, err
}

func (s *Store) procedureIDForValue(value repo.IMedicalProcedure) (ids.ProcedureID, error) {
	if value == nil {
		return 0, fmt.Errorf("procedure value is nil")
	}
	kind, err := procedureDBKind(value)
	if err != nil {
		return 0, err
	}
	table, where, args, err := procedureSubtypeSpec(value)
	if err != nil {
		return 0, err
	}
	var ptr int64
	err = s.QueryRowContext(s.ctx, `SELECT id FROM `+table+` WHERE `+where, args...).Scan(&ptr)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("procedure subtype not found: kind=%s", kind)
	}
	if err != nil {
		return 0, err
	}
	var id ids.ProcedureID
	err = s.QueryRowContext(s.ctx,
		`SELECT emedical_procedure_id FROM emedical_procedure WHERE kind = ? AND ptr = ?`, kind, ptr).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("procedure row not found: kind=%s ptr=%d", kind, ptr)
	}
	return id, err
}

func procedureDBKind(value repo.IMedicalProcedure) (string, error) {
	switch value.(type) {
	case repo.Surgery:
		return "Surgery", nil
	case repo.DiagnosticTest:
		return "DiagnosticTest", nil
	case repo.UnknownMedicalProcedure:
		return "Unknown", nil
	case repo.Education:
		return "Education", nil
	case repo.Instruction:
		return "Instruction", nil
	case repo.Insertion:
		return "Insertion", nil
	case repo.Infuse:
		return "Infuse", nil
	case repo.Transplant:
		return "Transplant", nil
	case repo.Cleaning:
		return "Cleaning", nil
	default:
		return "", fmt.Errorf("지원하지 않는 procedure 타입입니다: %T", value)
	}
}

func procedureSubtypeSpec(value repo.IMedicalProcedure) (string, string, []any, error) {
	switch typed := value.(type) {
	case repo.Surgery:
		return "procedure_named", "name = ?", []any{strings.TrimSpace(typed.Name)}, nil
	case repo.DiagnosticTest:
		return "procedure_named", "name = ?", []any{strings.TrimSpace(typed.Name)}, nil
	case repo.UnknownMedicalProcedure:
		return "procedure_named", "name = ?", []any{strings.TrimSpace(typed.Name)}, nil
	case repo.Education:
		return "procedure_education", "name = ? AND topic = ?",
			[]any{strings.TrimSpace(typed.Name), strings.TrimSpace(typed.Topic)}, nil
	case repo.Instruction:
		return "procedure_instruction", "name = ? AND action = ? AND duration = ? AND position = ?",
			[]any{strings.TrimSpace(typed.Name), strings.TrimSpace(typed.Action), strings.TrimSpace(typed.Duration), strings.TrimSpace(typed.Position)}, nil
	case repo.Insertion:
		site := ""
		if typed.Site != nil {
			site = strings.TrimSpace(typed.Site.ActorName())
		}
		return "procedure_insertion", "name = ? AND tube = ? AND site = ?",
			[]any{strings.TrimSpace(typed.Device.ActorName()), "", site}, nil
	case repo.Infuse:
		if typed.Material == nil {
			return "", "", nil, fmt.Errorf("infuse material이 비어 있습니다")
		}
		return "procedure_infuse", "material_eactor_id = (SELECT eactor_id FROM eactor WHERE kind = ? AND name = ?) AND route = ?",
			[]any{string(repo.ParseActorKind(typed.Material.Kind())), strings.TrimSpace(typed.Material.ActorName()), string(typed.Route)}, nil
	case repo.Transplant:
		return procedureTargetSpec("procedure_transplant", typed.Target)
	case repo.Cleaning:
		return procedureTargetSpec("procedure_cleaning", typed.Target)
	default:
		return "", "", nil, fmt.Errorf("지원하지 않는 procedure 타입입니다: %T", value)
	}
}

func procedureTargetSpec(table string, target repo.IEActor) (string, string, []any, error) {
	if target == nil || strings.TrimSpace(target.ActorName()) == "" {
		return table, "target_kind = ? AND target_eactor_id IS NULL AND target_name = ?", []any{"unknown", ""}, nil
	}
	return table,
		"target_kind = ? AND target_eactor_id = (SELECT eactor_id FROM eactor WHERE kind = ? AND name = ?) AND target_name = ?",
		[]any{string(repo.ParseActorKind(target.Kind())), string(repo.ParseActorKind(target.Kind())), strings.TrimSpace(target.ActorName()), strings.TrimSpace(target.ActorName())}, nil
}

func (s *Store) diseaseIDByName(name string) (int64, error) {
	var id int64
	err := s.QueryRowContext(s.ctx, `SELECT disease_id FROM edisease WHERE name = ?`, strings.TrimSpace(name)).Scan(&id)
	return id, err
}

func (s *Store) indexIDByName(name string) (int64, error) {
	var id int64
	err := s.QueryRowContext(s.ctx, `SELECT index_id FROM eindex WHERE name = ?`, strings.TrimSpace(name)).Scan(&id)
	return id, err
}

func (s *Store) symptomIDForValue(value repo.IESymptom) (int64, error) {
	switch typed := value.(type) {
	case repo.EOtherSymptom:
		return s.symptomIDByName(typed.Name)
	case repo.EPain:
		siteKind, siteName := painSite(typed.Site)
		aspects, err := marshalPainAspects(typed.Aspects)
		if err != nil {
			return 0, err
		}
		var id int64
		err = s.QueryRowContext(s.ctx,
			`SELECT s.symptom_id FROM esymptom s JOIN symptom_pain p ON p.id = s.ptr
			 WHERE s.kind = ? AND p.site_kind = ? AND p.site_name = ? AND p.aspects_json = ?`,
			string(repo.SymptomKindPain), siteKind, siteName, aspects).Scan(&id)
		return id, err
	default:
		return 0, fmt.Errorf("지원하지 않는 symptom 타입입니다: %T", value)
	}
}

func (s *Store) symptomIDByName(name string) (int64, error) {
	var id int64
	err := s.QueryRowContext(s.ctx,
		`SELECT s.symptom_id FROM esymptom s JOIN symptom_other o ON o.id = s.ptr
		 WHERE s.kind = ? AND o.name = ?`,
		string(repo.SymptomKindOtherSymptom), strings.TrimSpace(name)).Scan(&id)
	return id, err
}

func painSite(value repo.IEActor) (string, string) {
	if value == nil {
		return "", ""
	}
	return string(repo.ParseActorKind(value.Kind())), value.ActorName()
}

type painAspectDocument struct {
	Kind    string `json:"kind"`
	Value   string `json:"value,omitempty"`
	Measure string `json:"measure,omitempty"`
	Name    string `json:"name,omitempty"`
}

func marshalPainAspects(aspects []repo.IPainAspect) (string, error) {
	documents := make([]painAspectDocument, 0, len(aspects))
	for _, aspect := range aspects {
		switch typed := aspect.(type) {
		case repo.SymptomDuration:
			documents = append(documents, painAspectDocument{Kind: typed.Kind(), Value: typed.Value, Measure: typed.Measure})
		case repo.Intensity:
			documents = append(documents, painAspectDocument{Kind: typed.Kind(), Value: typed.Value, Measure: typed.Measure})
		case repo.Quality:
			documents = append(documents, painAspectDocument{Kind: typed.Kind(), Name: typed.Name})
		case repo.Radiation:
			documents = append(documents, painAspectDocument{Kind: typed.Kind(), Name: typed.Name})
		default:
			return "", fmt.Errorf("지원하지 않는 pain aspect 타입입니다: %T", aspect)
		}
	}
	sort.SliceStable(documents, func(i, j int) bool {
		return painAspectDocumentKey(documents[i]) < painAspectDocumentKey(documents[j])
	})
	encoded, err := json.Marshal(documents)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func painAspectDocumentKey(document painAspectDocument) string {
	return document.Kind + "\x1f" + document.Value + "\x1f" + document.Measure + "\x1f" + document.Name
}
