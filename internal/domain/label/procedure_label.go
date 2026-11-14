package centitylabel

import (
	"encoding/json"
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"strings"

	"github.com/G226Medi03/MorbesPipeline/internal/domain/trait"
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

func is_procedure_kind(kind string) bool {
	switch trait.ProcedureType(kind) {
	case trait.KindSurgery, trait.KindTransplant,
		trait.KindInfuse, trait.KindCleaning, trait.KindDiagnosticTest,
		trait.KindEducation, trait.KindInstruction, trait.KindInsertion, trait.KindUnknown:
		return true
	}
	return false
}

// unmarshalProcedure accepts both the current codec shape and the legacy
// nested shape used by older databases.
func unmarshalProcedure(raw json.RawMessage) (repo.IMedicalProcedure, error) {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	var kind string
	_ = json.Unmarshal(payload["kind"], &kind)
	kind = strings.ToLower(strings.TrimSpace(kind))
	nestedKeys := []string{kind}
	switch kind {
	case "diagnostictest", "diagnostic_test", "diagnostic", "test":
		nestedKeys = []string{"diagnostic_test", "diagnostic"}
	case "unknownmedicalprocedure":
		nestedKeys = []string{"unknown"}
	}
	for _, key := range nestedKeys {
		nestedRaw, ok := payload[key]
		if !ok {
			continue
		}
		nested := map[string]json.RawMessage{}
		if err := json.Unmarshal(nestedRaw, &nested); err != nil {
			return nil, err
		}
		for nestedKey, value := range nested {
			payload[nestedKey] = value
		}
		encoded, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		return repo.UnmarshalProcedure(encoded)
	}
	return repo.UnmarshalProcedure(raw)
}
