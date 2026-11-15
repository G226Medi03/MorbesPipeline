package entityjson

import (
	"encoding/json"
	"fmt"
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"strings"
)

type procedureDocument struct {
	Kind     string          `json:"kind"`
	Name     string          `json:"name,omitempty"`
	Topic    string          `json:"topic,omitempty"`
	Action   string          `json:"action,omitempty"`
	Duration string          `json:"duration,omitempty"`
	Position string          `json:"position,omitempty"`
	Route    string          `json:"route,omitempty"`
	Target   json.RawMessage `json:"target,omitempty"`
	Material json.RawMessage `json:"material,omitempty"`
	Device   json.RawMessage `json:"device,omitempty"`
	Site     json.RawMessage `json:"site,omitempty"`
}

// encodeProcedure encodes a medical procedure interface value.
func encodeProcedure(value repo.IMedicalProcedure) ([]byte, error) {
	if value == nil {
		return nil, fmt.Errorf("medical procedure is nil")
	}
	switch typed := value.(type) {
	case repo.Cleaning:
		target, err := encodeActor(typed.Target)
		if err != nil {
			return nil, err
		}
		return json.Marshal(procedureDocument{Kind: typed.Kind(), Target: target})
	case repo.Transplant:
		target, err := encodeActor(typed.Target)
		if err != nil {
			return nil, err
		}
		return json.Marshal(procedureDocument{Kind: typed.Kind(), Target: target})
	case repo.Infuse:
		material, err := encodeActor(typed.Material)
		if err != nil {
			return nil, err
		}
		return json.Marshal(procedureDocument{Kind: typed.Kind(), Material: material, Route: string(typed.Route)})
	case repo.Insertion:
		device, err := encodeActor(typed.Device)
		if err != nil {
			return nil, err
		}
		site, err := encodeActor(typed.Site)
		if err != nil {
			return nil, err
		}
		return json.Marshal(procedureDocument{Kind: typed.Kind(), Device: device, Site: site})
	case repo.DiagnosticTest:
		return json.Marshal(procedureDocument{Kind: typed.Kind(), Name: typed.Name})
	case repo.Education:
		return json.Marshal(procedureDocument{Kind: typed.Kind(), Name: typed.Name, Topic: typed.Topic})
	case repo.Instruction:
		return json.Marshal(procedureDocument{Kind: typed.Kind(), Name: typed.Name, Action: typed.Action, Duration: typed.Duration, Position: typed.Position})
	case repo.Surgery:
		return json.Marshal(procedureDocument{Kind: typed.Kind(), Name: typed.Name})
	case repo.UnknownMedicalProcedure:
		return json.Marshal(procedureDocument{Kind: typed.Kind(), Name: typed.Name})
	default:
		return nil, fmt.Errorf("지원하지 않는 medical procedure 타입입니다: %T", value)
	}
}

// decodeProcedure decodes a kind-tagged medical procedure value.
func decodeProcedure(data []byte) (repo.IMedicalProcedure, error) {
	var raw procedureDocument
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("medical procedure JSON 파싱 실패: %w", err)
	}
	switch repo.ProcedureKind(canonicalProcedureKind(raw.Kind)) {
	case repo.ProcedureKindCleaning:
		target, err := decodeProcedureActor(raw.Target)
		if err != nil {
			return nil, err
		}
		return repo.Cleaning{Target: target}, nil
	case repo.ProcedureKindTransplant:
		target, err := decodeProcedureActor(raw.Target)
		if err != nil {
			return nil, err
		}
		return repo.Transplant{Target: target}, nil
	case repo.ProcedureKindInfuse:
		material, err := decodeProcedureActor(raw.Material)
		if err != nil {
			return nil, err
		}
		return repo.Infuse{Material: material, Route: repo.InfuseRoute(raw.Route)}, nil
	case repo.ProcedureKindInsertion:
		device, err := decodeProcedureActor(raw.Device)
		if err != nil {
			return nil, err
		}
		site, err := decodeProcedureActor(raw.Site)
		if err != nil {
			return nil, err
		}
		return repo.Insertion{Device: repo.NewDevice(actorName(device)), Site: site}, nil
	case repo.ProcedureKindDiagnosticTest:
		return repo.DiagnosticTest{Name: raw.Name}, nil
	case repo.ProcedureKindEducation:
		return repo.Education{Name: raw.Name, Topic: raw.Topic}, nil
	case repo.ProcedureKindInstruction:
		return repo.Instruction{Name: raw.Name, Action: raw.Action, Duration: raw.Duration, Position: raw.Position}, nil
	case repo.ProcedureKindSurgery:
		return repo.Surgery{Name: raw.Name}, nil
	default:
		return repo.UnknownMedicalProcedure{Name: raw.Name}, nil
	}
}

// decodeOrParseProcedure accepts either a JSON payload or a free-text name.
// Free-text names are preserved as UnknownMedicalProcedure.
func decodeOrParseProcedure(data []byte) (repo.IMedicalProcedure, error) {
	if len(data) == 0 {
		return repo.UnknownMedicalProcedure{}, nil
	}
	if value, err := decodeProcedure(data); err == nil {
		return value, nil
	}
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		return repo.UnknownMedicalProcedure{Name: text}, nil
	}
	return repo.UnknownMedicalProcedure{Name: strings.TrimSpace(string(data))}, nil
}

func decodeProcedureActor(data json.RawMessage) (repo.IEActor, error) {
	if len(data) == 0 {
		return repo.NewUnknownActor(""), nil
	}
	return decodeActor(data)
}

// canonicalProcedureKind normalizes procedure kind aliases.
func canonicalProcedureKind(kind string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "cleaning":
		return "cleaning"
	case "transplant":
		return "transplant"
	case "infuse", "infusion":
		return "infuse"
	case "insertion":
		return "insertion"
	case "diagnostic", "diagnostic_test", "test":
		return "diagnostic_test"
	case "education":
		return "education"
	case "instruction":
		return "instruction"
	case "surgery", "procedure", "operation":
		return "surgery"
	default:
		return "unknown"
	}
}
