package entityjson

import (
	"encoding/json"
	"fmt"
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"strings"
)

type actorDocument struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

// encodeActor encodes an actor interface value.
func encodeActor(value repo.IEActor) ([]byte, error) {
	if value == nil {
		return nil, fmt.Errorf("actor is nil")
	}
	return json.Marshal(actorDocument{Kind: canonicalActorKind(value.Kind()), Name: actorName(value)})
}

// decodeActor decodes a kind-tagged actor value.
func decodeActor(data []byte) (repo.IEActor, error) {
	var raw actorDocument
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("actor JSON 파싱 실패: %w", err)
	}
	return actorFromKind(raw.Kind, raw.Name), nil
}

// actorName returns the display name of an actor.
func actorName(value repo.IEActor) string {
	switch typed := value.(type) {
	case repo.Device:
		return typed.Name
	case repo.ActorProcedure:
		return typed.Name
	case repo.UnknownActor:
		return typed.Name
	default:
		return ""
	}
}

// actorFromKind rebuilds an actor from its kind and name.
func actorFromKind(kind, name string) repo.IEActor {
	name = strings.TrimSpace(name)
	switch canonicalActorKind(kind) {
	case string(repo.ActorKindDevice):
		return repo.NewDevice(name)
	case string(repo.ActorKindProcedure):
		return repo.NewActorProcedure(name)
	default:
		return repo.NewUnknownActor(name)
	}
}

// canonicalActorKind normalizes actor kind aliases.
func canonicalActorKind(kind string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "eprocedure":
		return string(repo.ActorKindProcedure)
	default:
		return repo.ParseActorKind(kind).String()
	}
}
