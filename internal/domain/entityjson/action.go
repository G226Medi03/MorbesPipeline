package entityjson

import (
	"encoding/json"
	"fmt"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
)

type actionDocument struct {
	Kind   string          `json:"kind"`
	Actor  json.RawMessage `json:"actor,omitempty"`
	Status json.RawMessage `json:"status,omitempty"`
}

// encodeAction encodes an action interface value.
func encodeAction(value repo.IEAction) ([]byte, error) {
	if value == nil {
		return nil, fmt.Errorf("action is nil")
	}
	switch typed := value.(type) {
	case repo.Block:
		actor, err := encodeActor(typed.Actor)
		if err != nil {
			return nil, err
		}
		return json.Marshal(actionDocument{Kind: typed.Kind(), Actor: actor})
	case repo.Relax:
		actor, err := encodeActor(typed.Actor)
		if err != nil {
			return nil, err
		}
		return json.Marshal(actionDocument{Kind: typed.Kind(), Actor: actor})
	case repo.CauseStatus:
		status, err := encodeStatus(typed.Status)
		if err != nil {
			return nil, err
		}
		return json.Marshal(actionDocument{Kind: typed.Kind(), Status: status})
	default:
		return nil, fmt.Errorf("지원하지 않는 action 타입입니다: %T", value)
	}
}
