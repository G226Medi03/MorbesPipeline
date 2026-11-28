package label

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
		material := ""
		if typed.Material != nil {
			material = strings.TrimSpace(typed.Material.ActorName())
		}
		if isBloodProduct(material) {
			return "수혈"
		}
		return first_nonempty(material, "주입")
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
	case repo.Palpation:
		site := ""
		if typed.Site != nil {
			site = strings.TrimSpace(typed.Site.ActorName())
		}
		if site == "" {
			return "촉진"
		}
		return "촉진 (" + site + ")"
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
func isBloodProduct(name string) bool {
	lower := strings.ToLower(strings.TrimSpace(name))
	if lower == "" {
		return false
	}
	for _, token := range []string{"혈액", "혈장", "혈소판", "적혈구", "농축적혈구", "blood", "plasma", "platelet", "erythrocyte"} {
		if strings.Contains(lower, token) {
			return true
		}
	}
	return false
}
