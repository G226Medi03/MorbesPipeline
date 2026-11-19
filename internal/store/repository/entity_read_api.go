package repository

import (
	"errors"
	"strings"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
)

// GetActorByName returns the actor entity for a name.
func (s *entityTable) GetActorByName(name string) (repo.Actor, bool, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return repo.Actor{}, false, nil
	}
	actor, err := s.entities.GetActorByName(name)
	if errors.Is(err, repo.ErrNotFound) {
		return repo.Actor{}, false, nil
	}
	if err != nil {
		return repo.Actor{}, false, err
	}
	return actor, true, nil
}

// GetDeviceByName returns the device actor entity for a name.
func (s *entityTable) GetDeviceByName(name string) (repo.Actor, bool, error) {
	actor, found, err := s.GetActorByName(name)
	if err != nil || !found {
		return repo.Actor{}, found, err
	}
	if actor.Name().Kind() != string(repo.ActorKindDevice) {
		return repo.Actor{}, false, nil
	}
	return actor, true, nil
}

// GetAction returns the action entity for a typed action value.
func (s *entityTable) GetAction(value repo.IEAction) (repo.Action, bool, error) {
	action, err := s.entities.GetAction(value)
	if errors.Is(err, repo.ErrNotFound) {
		return repo.Action{}, false, nil
	}
	if err != nil {
		return repo.Action{}, false, err
	}
	return action, true, nil
}

// GetStatus returns the status entity for a typed status value.
func (s *entityTable) GetStatus(value repo.IEStatus) (repo.Status, bool, error) {
	status, err := s.entities.GetStatus(value)
	if errors.Is(err, repo.ErrNotFound) {
		return repo.Status{}, false, nil
	}
	if err != nil {
		return repo.Status{}, false, err
	}
	return status, true, nil
}

// GetProcedure returns the procedure entity for a typed procedure value.
func (s *entityTable) GetProcedure(value repo.IMedicalProcedure) (repo.Procedure, bool, error) {
	procedure, err := s.entities.GetProcedure(value)
	if errors.Is(err, repo.ErrNotFound) {
		return repo.Procedure{}, false, nil
	}
	if err != nil {
		return repo.Procedure{}, false, err
	}
	return procedure, true, nil
}
