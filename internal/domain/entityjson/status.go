package entityjson

import (
	"encoding/json"
	"fmt"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
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

func encodeStatusIndex(value repo.EIndex) json.RawMessage {
	encoded, _ := json.Marshal(value)
	return encoded
}
