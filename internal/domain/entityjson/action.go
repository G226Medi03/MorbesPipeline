package entityjson

import (
	"encoding/json"
	"fmt"
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"strings"
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

// decodeAction decodes a kind-tagged action value.
func decodeAction(data []byte) (repo.IEAction, error) {
	var raw actionDocument
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("action JSON 파싱 실패: %w", err)
	}
	switch repo.ActionKind(canonicalActionKind(raw.Kind)) {
	case repo.ActionKindBlock:
		actor, err := decodeActionActor(raw.Actor)
		if err != nil {
			return nil, err
		}
		return repo.Block{Actor: actor}, nil
	case repo.ActionKindRelax:
		actor, err := decodeActionActor(raw.Actor)
		if err != nil {
			return nil, err
		}
		return repo.Relax{Actor: actor}, nil
	case repo.ActionKindCauseStatus:
		var status repo.IEStatus
		if len(raw.Status) > 0 {
			decoded, err := decodeStatus(raw.Status)
			if err != nil {
				return nil, err
			}
			status = decoded
		}
		return repo.CauseStatus{Status: status}, nil
	default:
		return nil, fmt.Errorf("지원하지 않는 action kind 입니다: %s", raw.Kind)
	}
}

func decodeActionActor(data json.RawMessage) (repo.IEActor, error) {
	if len(data) == 0 {
		return actorFromKind("", ""), nil
	}
	return decodeActor(data)
}

// canonicalActionKind normalizes action kind aliases.
func canonicalActionKind(kind string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "block":
		return "block"
	case "relax":
		return "relax"
	case "cause_status", "cause", "status":
		return "cause_status"
	default:
		return strings.ToLower(strings.TrimSpace(kind))
	}
}
