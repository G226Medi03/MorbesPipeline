package medicalentity

import (
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
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
