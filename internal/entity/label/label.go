// Package centitylabel builds short, human-readable labels for typed entity values.
package centitylabel

import (
	"strings"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
)

// StatusLabelValue returns a human-readable label for a typed status value.
func StatusLabelValue(value repo.IEStatus) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case repo.IncreasedIndex:
		return state_suffix(strings.TrimSpace(typed.Index.Name), "증가")
	case repo.DecreasedIndex:
		return state_suffix(strings.TrimSpace(typed.Index.Name), "감소")
	case repo.HaveDisease:
		return strings.TrimSpace(string(typed.Disease))
	case repo.HaveSymptom:
		return symptom_label_value(typed.Symptom)
	case repo.Damaged:
		return first_nonempty(actor_label(typed.Actor), "손상")
	case repo.Inflammation:
		return first_nonempty(actor_label(typed.Actor), "염증")
	case repo.ReceivedMedicalProcedure:
		label := first_nonempty(procedure_label(typed.MedicalProcedure), "시술")
		label = state_suffix(label, "받음")
		if period := strings.TrimSpace(typed.PostProcedurePeriod); period != "" {
			label += " (" + period + ")"
		}
		return strings.TrimSpace(label)
	default:
		return strings.TrimSpace(value.Kind())
	}
}

// ActionLabelValue returns a human-readable label for a typed action value.
func ActionLabelValue(value repo.IEAction) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case repo.CauseStatus:
		if typed.Status == nil {
			return "원인 상태 정보 없음"
		}
		return "cause_status: " + StatusLabelValue(typed.Status)
	case repo.Block:
		return first_nonempty(actor_label(typed.Actor), "block")
	case repo.Relax:
		return first_nonempty(actor_label(typed.Actor), "relax")
	default:
		return strings.TrimSpace(value.Kind())
	}
}

func symptom_label_value(value repo.IESymptom) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case repo.EOtherSymptom:
		return strings.TrimSpace(typed.Name)
	case repo.EPain:
		return pain_label_value(typed)
	default:
		return strings.TrimSpace(value.Kind())
	}
}

func pain_label_value(value repo.EPain) string {
	site := actor_label(value.Site)
	details := make([]string, 0, len(value.Aspects))
	for _, aspect := range value.Aspects {
		switch typed := aspect.(type) {
		case repo.SymptomDuration:
			if text := strings.TrimSpace(strings.TrimSpace(typed.Value) + strings.TrimSpace(typed.Measure)); text != "" {
				details = append(details, text)
			}
		case repo.Intensity:
			if text := strings.TrimSpace(strings.TrimSpace(typed.Value) + strings.TrimSpace(typed.Measure)); text != "" {
				details = append(details, text)
			}
		case repo.Quality:
			if text := strings.TrimSpace(typed.Name); text != "" {
				details = append(details, text)
			}
		case repo.Radiation:
			if text := strings.TrimSpace(typed.Name); text != "" {
				details = append(details, text+" 방사")
			}
		}
	}
	label := "통증"
	if site != "" {
		label = site + " 통증"
	}
	if len(details) > 0 {
		label += " (" + strings.Join(details, ", ") + ")"
	}
	return label
}

func actor_label(value repo.IEActor) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(value.ActorName())
}

func state_suffix(value, suffix string) string {
	text := strings.TrimSpace(value)
	if text == "" {
		return suffix
	}
	if strings.HasSuffix(text, suffix) {
		return text
	}
	return text + " " + suffix
}

func first_nonempty(values ...string) string {
	for _, value := range values {
		if text := strings.TrimSpace(value); text != "" {
			return text
		}
	}
	return ""
}
