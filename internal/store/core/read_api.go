package store

import (
	"errors"
	"strings"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
)

// GetActorRecordByName returns the actor entity for a name.
func (s *StoreContainer) GetActorRecordByName(name string) (repo.Actor, bool, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return repo.Actor{}, false, nil
	}
	entityStore, err := s.Entities()
	if err != nil {
		return repo.Actor{}, false, err
	}
	actor, err := entityStore.GetActorByName(name)
	if errors.Is(err, repo.ErrNotFound) {
		return repo.Actor{}, false, nil
	}
	if err != nil {
		return repo.Actor{}, false, err
	}
	return actor, true, nil
}

// GetDeviceByName returns the device actor entity for a name.
func (s *StoreContainer) GetDeviceByName(name string) (repo.Actor, bool, error) {
	actor, found, err := s.GetActorRecordByName(name)
	if err != nil || !found {
		return repo.Actor{}, found, err
	}
	if actor.Name().Kind() != string(repo.ActorKindDevice) {
		return repo.Actor{}, false, nil
	}
	return actor, true, nil
}

// GetActionRecord returns the action entity for a typed action value.
func (s *StoreContainer) GetActionRecord(value repo.IEAction) (repo.Action, bool, error) {
	entityStore, err := s.Entities()
	if err != nil {
		return repo.Action{}, false, err
	}
	action, err := entityStore.GetAction(value)
	if errors.Is(err, repo.ErrNotFound) {
		return repo.Action{}, false, nil
	}
	if err != nil {
		return repo.Action{}, false, err
	}
	return action, true, nil
}

// GetStatusRecord returns the status entity for a typed status value.
func (s *StoreContainer) GetStatusRecord(value repo.IEStatus) (repo.Status, bool, error) {
	entityStore, err := s.Entities()
	if err != nil {
		return repo.Status{}, false, err
	}
	status, err := entityStore.GetStatus(value)
	if errors.Is(err, repo.ErrNotFound) {
		return repo.Status{}, false, nil
	}
	if err != nil {
		return repo.Status{}, false, err
	}
	return status, true, nil
}

// GetProcedureRecord returns the procedure entity for a typed procedure value.
func (s *StoreContainer) GetProcedureRecord(value repo.IMedicalProcedure) (repo.Procedure, bool, error) {
	entityStore, err := s.Entities()
	if err != nil {
		return repo.Procedure{}, false, err
	}
	procedure, err := entityStore.GetProcedure(value)
	if errors.Is(err, repo.ErrNotFound) {
		return repo.Procedure{}, false, nil
	}
	if err != nil {
		return repo.Procedure{}, false, err
	}
	return procedure, true, nil
}
