package centitylabel

import (
	"strings"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
)

func ProcedureLabel(value repo.IMedicalProcedure) string {
	return procedure_label(value)
}

func procedure_label(value repo.IMedicalProcedure) string {
	switch typed := value.(type) {
	case repo.Transplant:
		target := strings.TrimSpace(typed.Target.ActorName())
		if target == "" {
			return "이식"
		}
		return target + " 이식"
	case repo.Infuse:
		return first_nonempty(strings.TrimSpace(typed.Material.ActorName()), "주입")
	case repo.Surgery:
		return first_nonempty(strings.TrimSpace(typed.Name), "수술")
	case repo.Insertion:
		device := ""
		if typed.Device.Name != "" {
			device = strings.TrimSpace(typed.Device.Name)
		}
		return first_nonempty(device, "삽입")
	case repo.Cleaning:
		target := ""
		if typed.Target != nil {
			target = strings.TrimSpace(typed.Target.ActorName())
		}
		if target == "" {
			return "세척"
		}
		return "세척 (" + target + ")"
	case repo.DiagnosticTest:
		return first_nonempty(strings.TrimSpace(typed.Name), "검사")
	case repo.Education:
		return first_nonempty(strings.TrimSpace(typed.Name), strings.TrimSpace(typed.Topic), "교육")
	case repo.Instruction:
		return first_nonempty(strings.TrimSpace(typed.Name), strings.TrimSpace(typed.Action), "지시")
	case repo.UnknownMedicalProcedure:
		return first_nonempty(strings.TrimSpace(typed.Name), "알 수 없는 시술")
	default:
		return value.Kind()
	}
}
