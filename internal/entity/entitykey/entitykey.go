// Package entitykey builds stable storage keys for typed medical entity values.
//
// These keys are for persistence, audit records, and option identity only.
// Entity equality must use the ename Equal methods, not these strings.
package entitykey

import (
	"fmt"
	"sort"
	"strings"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
)

// Status returns a stable storage key for a status value.
func Status(value repo.IEStatus) string {
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
		return "have_symptom|" + Symptom(typed.Symptom)
	case repo.Damaged:
		return "damaged|" + Actor(typed.Actor)
	case repo.Inflammation:
		return "inflammation|" + Actor(typed.Actor)
	case repo.ReceivedMedicalProcedure:
		return "received_medical_procedure|" + Procedure(typed.MedicalProcedure) + "|" + strings.TrimSpace(typed.PostProcedurePeriod)
	default:
		return strings.TrimSpace(value.Kind())
	}
}

// Action returns a stable storage key for an action value.
func Action(value repo.IEAction) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case repo.CauseStatus:
		return "cause_status|" + Status(typed.Status)
	case repo.Block:
		return "block|" + Actor(typed.Actor)
	case repo.Relax:
		return "relax|" + Actor(typed.Actor)
	default:
		return strings.TrimSpace(value.Kind())
	}
}

// Symptom returns a stable storage key for a symptom value.
func Symptom(value repo.IESymptom) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case repo.EOtherSymptom:
		return "other_symptom|" + strings.TrimSpace(typed.Name)
	case repo.EPain:
		items := make([]string, 0, len(typed.Aspects)+1)
		items = append(items, "site|"+Actor(typed.Site))
		for _, aspect := range typed.Aspects {
			items = append(items, aspectKey(aspect))
		}
		sort.Strings(items)
		return "pain|" + strings.Join(items, ";")
	default:
		return strings.TrimSpace(value.Kind())
	}
}

// Procedure returns a stable storage key for a procedure value.
func Procedure(value repo.IMedicalProcedure) string {
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
		return "insertion|" + Actor(typed.Device) + "|" + Actor(typed.Site)
	case repo.Infuse:
		return "infuse|" + Actor(typed.Material) + "|" + string(typed.Route)
	case repo.Transplant:
		return "transplant|" + Actor(typed.Target)
	case repo.Cleaning:
		return "cleaning|" + Actor(typed.Target)
	default:
		return strings.TrimSpace(value.Kind())
	}
}

// Actor returns a stable storage key for an actor value.
func Actor(value repo.IEActor) string {
	if value == nil {
		return "nil"
	}
	return strings.TrimSpace(value.Kind()) + "|" + strings.TrimSpace(value.ActorName())
}

// OptionKey returns the stable key used for a medical procedure option.
func OptionKey(procedure repo.IMedicalProcedure, category string, indications []repo.IEStatus) string {
	key := Procedure(procedure) + "|" + category
	for _, indication := range indications {
		key += "|" + Status(indication)
	}
	return key
}

func aspectKey(value repo.IPainAspect) string {
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
