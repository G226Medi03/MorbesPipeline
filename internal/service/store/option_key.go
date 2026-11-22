// Package entitykey builds stable identity keys for typed medical entity values.
package service

import (
	"fmt"
	"sort"
	"strings"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
)

// Status returns a stable identity key for a status value.
func statusOptionKey(value repo.IEStatus) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case repo.IncreasedIndex:
		return "increased_index|" + strings.TrimSpace(typed.Index.Name)
	case repo.DecreasedIndex:
		return "decreased_index|" + strings.TrimSpace(typed.Index.Name)
	case repo.HaveDisease:
		return "have_disease|" + strings.TrimSpace(string(typed.Disease))
	case repo.HaveSymptom:
		return "have_symptom|" + symptomOptionKey(typed.Symptom)
	case repo.Damaged:
		return "damaged|" + actorOptionKey(typed.Actor)
	case repo.Inflammation:
		return "inflammation|" + actorOptionKey(typed.Actor)
	case repo.ReceivedMedicalProcedure:
		return "received_medical_procedure|" + procedureOptionKey(typed.MedicalProcedure) + "|" + strings.TrimSpace(typed.PostProcedurePeriod)
	default:
		return strings.TrimSpace(value.Kind())
	}
}

// Action returns a stable identity key for an action value.
func actionOptionKey(value repo.IEAction) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case repo.CauseStatus:
		return "cause_status|" + statusOptionKey(typed.Status)
	case repo.Block:
		return "block|" + actorOptionKey(typed.Actor)
	case repo.Relax:
		return "relax|" + actorOptionKey(typed.Actor)
	default:
		return strings.TrimSpace(value.Kind())
	}
}

// Symptom returns a stable identity key for a symptom value.
func symptomOptionKey(value repo.IESymptom) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case repo.EOtherSymptom:
		return "other_symptom|" + strings.TrimSpace(typed.Name)
	case repo.EPain:
		items := make([]string, 0, len(typed.Aspects)+1)
		items = append(items, "site|"+actorOptionKey(typed.Site))
		for _, aspect := range typed.Aspects {
			items = append(items, aspectOptionKey(aspect))
		}
		sort.Strings(items)
		return "pain|" + strings.Join(items, ";")
	default:
		return strings.TrimSpace(value.Kind())
	}
}

// Procedure returns a stable identity key for a procedure value.
func procedureOptionKey(value repo.IMedicalProcedure) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case repo.Surgery:
		return "surgery|" + strings.TrimSpace(typed.Name)
	case repo.DiagnosticTest:
		return "diagnostic_test|" + strings.TrimSpace(typed.Name)
	case repo.UnknownMedicalProcedure:
		return "unknown|" + strings.TrimSpace(typed.Name)
	case repo.Education:
		return "education|" + strings.TrimSpace(typed.Name) + "|" + strings.TrimSpace(typed.Topic)
	case repo.Instruction:
		return "instruction|" + strings.TrimSpace(typed.Name) + "|" + strings.TrimSpace(typed.Action) + "|" + strings.TrimSpace(typed.Duration) + "|" + strings.TrimSpace(typed.Position)
	case repo.Insertion:
		return "insertion|" + actorOptionKey(typed.Device) + "|" + actorOptionKey(typed.Site)
	case repo.Infuse:
		return "infuse|" + actorOptionKey(typed.Material) + "|" + string(typed.Route)
	case repo.Transplant:
		return "transplant|" + actorOptionKey(typed.Target)
	case repo.Cleaning:
		return "cleaning|" + actorOptionKey(typed.Target)
	default:
		return strings.TrimSpace(value.Kind())
	}
}

// Actor returns a stable identity key for an actor value.
func actorOptionKey(value repo.IEActor) string {
	if value == nil {
		return "nil"
	}
	return strings.TrimSpace(value.Kind()) + "|" + strings.TrimSpace(value.ActorName())
}

func aspectOptionKey(value repo.IPainAspect) string {
	if value == nil {
		return "nil"
	}
	switch typed := value.(type) {
	case repo.SymptomDuration:
		return fmt.Sprintf("duration|%s|%s", strings.TrimSpace(typed.Value), strings.TrimSpace(typed.Measure))
	case repo.Intensity:
		return fmt.Sprintf("intensity|%s|%s", strings.TrimSpace(typed.Value), strings.TrimSpace(typed.Measure))
	case repo.Quality:
		return "quality|" + strings.TrimSpace(typed.Name)
	case repo.Radiation:
		return "radiation|" + strings.TrimSpace(typed.Name)
	default:
		return strings.TrimSpace(value.Kind())
	}
}
