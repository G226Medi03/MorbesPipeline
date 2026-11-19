package entity

import (
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"github.com/G226Medi03/MorbesPipeline/internal/domain/trait"
)

func ActorValueByKind(kind, name string) (repo.IEActor, error) {
	switch repo.ParseActorKind(kind) {
	case repo.ActorKindDevice:
		return repo.NewDevice(name), nil
	case repo.ActorKindProcedure:
		return repo.NewActorProcedure(name), nil
	default:
		return repo.NewUnknownActor(name), nil
	}
}

func MedicalEntityTraits(values []trait.Trait) []repo.Trait {
	result := make([]repo.Trait, len(values))
	for index, value := range values {
		result[index] = repo.Trait(value)
	}
	return result
}
