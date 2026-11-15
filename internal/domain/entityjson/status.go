package entityjson

import (
	"encoding/json"
	"fmt"
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"strings"
)

type statusDocument struct {
	Kind                string          `json:"kind"`
	Actor               json.RawMessage `json:"actor,omitempty"`
	Index               json.RawMessage `json:"index,omitempty"`
	Disease             string          `json:"disease,omitempty"`
	Symptom             json.RawMessage `json:"symptom,omitempty"`
	MedicalProcedure    json.RawMessage `json:"medical_procedure,omitempty"`
	Procedure           json.RawMessage `json:"procedure,omitempty"`
	PostProcedurePeriod string          `json:"post_procedure_period,omitempty"`
	// Name is the shorthand form emitted by the extractors. It is accepted as
	// a fallback for actor, index, disease, symptom, and procedure values.
	Name string `json:"name,omitempty"`
}

// encodeStatus encodes a status interface value.
func encodeStatus(value repo.IEStatus) ([]byte, error) {
	if value == nil {
		return nil, fmt.Errorf("status is nil")
	}
	switch typed := value.(type) {
	case repo.Damaged:
		actor, err := encodeActor(typed.Actor)
		if err != nil {
			return nil, err
		}
		return json.Marshal(statusDocument{Kind: typed.Kind(), Actor: actor})
	case repo.Inflammation:
		actor, err := encodeActor(typed.Actor)
		if err != nil {
			return nil, err
		}
		return json.Marshal(statusDocument{Kind: typed.Kind(), Actor: actor})
	case repo.IncreasedIndex:
		return json.Marshal(statusDocument{Kind: typed.Kind(), Index: encodeStatusIndex(typed.Index)})
	case repo.DecreasedIndex:
		return json.Marshal(statusDocument{Kind: typed.Kind(), Index: encodeStatusIndex(typed.Index)})
	case repo.HaveDisease:
		return json.Marshal(statusDocument{Kind: typed.Kind(), Disease: string(typed.Disease)})
	case repo.HaveSymptom:
		symptom, err := encodeSymptom(typed.Symptom)
		if err != nil {
			return nil, err
		}
		return json.Marshal(statusDocument{Kind: typed.Kind(), Symptom: symptom})
	case repo.ReceivedMedicalProcedure:
		procedure, err := encodeProcedure(typed.MedicalProcedure)
		if err != nil {
			return nil, err
		}
		return json.Marshal(statusDocument{
			Kind:                typed.Kind(),
			MedicalProcedure:    procedure,
			PostProcedurePeriod: typed.PostProcedurePeriod,
		})
	default:
		return nil, fmt.Errorf("지원하지 않는 status 타입입니다: %T", value)
	}
}

// decodeStatus decodes a kind-tagged status value.
func decodeStatus(data []byte) (repo.IEStatus, error) {
	var raw statusDocument
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("status JSON 파싱 실패: %w", err)
	}
	switch repo.StatusKind(canonicalStatusKind(raw.Kind)) {
	case repo.StatusKindDamaged:
		actor, err := decodeStatusActorFallback(raw.Actor, raw.Name)
		if err != nil {
			return nil, err
		}
		return repo.Damaged{Actor: actor}, nil
	case repo.StatusKindInflammation:
		actor, err := decodeStatusActorFallback(raw.Actor, raw.Name)
		if err != nil {
			return nil, err
		}
		return repo.Inflammation{Actor: actor}, nil
	case repo.StatusKindIncreasedIndex:
		return repo.IncreasedIndex{Index: decodeStatusIndex(raw.Index, raw.Name)}, nil
	case repo.StatusKindDecreasedIndex:
		return repo.DecreasedIndex{Index: decodeStatusIndex(raw.Index, raw.Name)}, nil
	case repo.StatusKindHaveDisease:
		return repo.HaveDisease{Disease: repo.EDisease(firstNonEmptyStatus(raw.Disease, raw.Name))}, nil
	case repo.StatusKindHaveSymptom:
		symptom, err := decodeStatusSymptom(raw.Symptom, raw.Name)
		if err != nil {
			return nil, err
		}
		return repo.HaveSymptom{Symptom: symptom}, nil
	case repo.StatusKindReceivedMedicalProcedure:
		procedureRaw := raw.MedicalProcedure
		if isEmptyStatusRaw(procedureRaw) {
			procedureRaw = raw.Procedure
		}
		var procedure repo.IMedicalProcedure
		var err error
		if isEmptyStatusRaw(procedureRaw) && strings.TrimSpace(raw.Name) != "" {
			procedure, err = decodeOrParseProcedure([]byte(raw.Name))
		} else {
			procedure, err = decodeProcedure(procedureRaw)
		}
		if err != nil {
			return nil, err
		}
		return repo.ReceivedMedicalProcedure{
			MedicalProcedure:    procedure,
			PostProcedurePeriod: raw.PostProcedurePeriod,
		}, nil
	default:
		return nil, fmt.Errorf("지원하지 않는 status kind 입니다: %s", raw.Kind)
	}
}

func decodeStatusActorFallback(data json.RawMessage, fallback string) (repo.IEActor, error) {
	if isEmptyStatusRaw(data) {
		return actorFromKind("", fallback), nil
	}
	return decodeActor(data)
}

func decodeStatusIndex(data json.RawMessage, fallback string) repo.EIndex {
	var value repo.EIndex
	if !isEmptyStatusRaw(data) {
		if err := json.Unmarshal(data, &value); err == nil && strings.TrimSpace(value.Name) != "" {
			return value
		}
	}
	if strings.TrimSpace(fallback) != "" {
		return repo.EIndex{Name: strings.TrimSpace(fallback)}
	}
	return value
}

func decodeStatusSymptom(data json.RawMessage, fallback string) (repo.IESymptom, error) {
	if isEmptyStatusRaw(data) {
		return repo.EOtherSymptom{Name: strings.TrimSpace(fallback)}, nil
	}
	return decodeSymptom(data)
}

func isEmptyStatusRaw(data json.RawMessage) bool {
	if len(data) == 0 || string(data) == "null" {
		return true
	}
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		return strings.TrimSpace(text) == ""
	}
	return false
}

func firstNonEmptyStatus(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func encodeStatusIndex(value repo.EIndex) json.RawMessage {
	encoded, _ := json.Marshal(value)
	return encoded
}

// canonicalStatusKind normalizes status kind aliases.
func canonicalStatusKind(kind string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "damaged":
		return "damaged"
	case "inflammation":
		return "inflammation"
	case "increased_index":
		return "increased_index"
	case "decreased_index":
		return "decreased_index"
	case "have_disease", "disease":
		return "have_disease"
	case "have_symptom", "symptom":
		return "have_symptom"
	case "received_medical_procedure", "procedure", "medical_procedure":
		return "received_medical_procedure"
	default:
		return strings.ToLower(strings.TrimSpace(kind))
	}
}

// hasStatusKind reports whether the kind maps to a supported status type.
func hasStatusKind(kind string) bool {
	switch repo.StatusKind(canonicalStatusKind(kind)) {
	case repo.StatusKindDamaged, repo.StatusKindInflammation, repo.StatusKindIncreasedIndex, repo.StatusKindDecreasedIndex,
		repo.StatusKindHaveDisease, repo.StatusKindHaveSymptom, repo.StatusKindReceivedMedicalProcedure:
		return true
	default:
		return false
	}
}
