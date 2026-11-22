package service

import (
	"errors"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"github.com/G226Medi03/MorbesPipeline/internal/entity/trait"
	entityrepo "github.com/G226Medi03/MorbesPipeline/internal/repository/entity"
	source "github.com/G226Medi03/SourceEntity"
)

func (s *EntityService) SaveProcedure(value repo.IMedicalProcedure, traits []trait.Trait, sources []source.Source) error {
	unlock := s.store.LockWrite()
	defer unlock()
	entities, err := s.store.Entities()
	if err != nil {
		return err
	}
	_, err = entities.InsertProcedure(
		value,
		nil,
		entityrepo.MedicalEntityTraits(traits),
		sourceIDs(sources),
	)
	return err
}

func (s *EntityService) SaveMedicalProcedureOption(option repo.MedicalProcedureOption) (repo.MedicalProcedureOption, error) {
	unlock := s.store.LockWrite()
	defer unlock()
	entities, err := s.store.Entities()
	if err != nil {
		return option, err
	}
	sources := make([]source.SourceId, 0, 1)
	if option.SourceID.ToI64() > 0 {
		sources = append(sources, option.SourceID)
	}
	if err := entities.SaveMedicalProcedureOption(option, sources); err != nil {
		return option, err
	}
	return option, nil
}

func (s *EntityService) SaveMedicalProcedureOptionByValues(
	procedure repo.IMedicalProcedure,
	category string,
	traits []trait.Trait,
	_ []trait.Trait,
	indications []repo.IEStatus,
	effects []repo.IEStatus,
	sources []source.Source,
) (repo.MedicalProcedureOption, error) {
	unlock := s.store.LockWrite()
	defer unlock()
	entities, err := s.store.Entities()
	if err != nil {
		return repo.MedicalProcedureOption{}, err
	}
	key := procedureOptionKey(procedure) + "|" + category
	for _, indication := range indications {
		key += "|" + statusOptionKey(indication)
	}
	option := repo.MedicalProcedureOption{
		Key:              key,
		Category:         category,
		Traits:           traitStrings(traits),
		MedicalProcedure: procedure,
		Indications:      indications,
		ExpectedEffects:  effects,
	}
	optionSources := make([]source.SourceId, 0, len(sources))
	for _, item := range sources {
		if item.SourceID.ToI64() > 0 {
			optionSources = append(optionSources, item.SourceID)
		}
	}
	if err := entities.SaveMedicalProcedureOption(option, optionSources); err != nil {
		return option, err
	}
	if len(sources) > 0 {
		option.SourceID = sources[0].SourceID
	}
	return option, nil
}

func traitStrings(values []trait.Trait) []string {
	result := make([]string, len(values))
	for index, value := range values {
		result[index] = string(value)
	}
	return result
}

func (s *EntityService) LinkActionResultStatusValues(action repo.IEAction, status repo.IEStatus, sourceID source.SourceId) error {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.EntityRepo().LinkActionResultStatusValues(s.store.Context(), s.store.MainExecutor(), action, status, sourceID)
}

func (s *EntityService) LinkStatusAction(statusID int64, actionID int64, sourceID source.SourceId) error {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.EntityRepo().LinkStatusAction(s.store.Context(), s.store.MainExecutor(), statusID, actionID, sourceID)
}

func (s *EntityService) LinkStatusActionValues(status repo.IEStatus, action repo.IEAction, sourceID source.SourceId) error {
	unlock := s.store.LockWrite()
	defer unlock()
	entities, err := s.store.Entities()
	if err != nil {
		return err
	}
	return entities.LinkStatusAction(status, action, sourceID)
}

func (s *EntityService) LinkActionResultStatus(actionID int64, statusID int64, sourceID source.SourceId) error {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.EntityRepo().LinkActionResultStatus(s.store.Context(), s.store.MainExecutor(), actionID, statusID, sourceID)
}

func (s *EntityService) DeleteStatusAction(statusID int64, actionID int64) error {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.EntityRepo().DeleteStatusAction(s.store.Context(), s.store.MainExecutor(), statusID, actionID)
}

func (s *EntityService) DeleteStatusActionValues(status repo.IEStatus, action repo.IEAction) error {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.EntityRepo().DeleteStatusActionValues(s.store.Context(), s.store.MainExecutor(), status, action)
}

func (s *EntityService) DeleteActionResultStatus(actionID int64, statusID int64) error {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.EntityRepo().DeleteActionResultStatus(s.store.Context(), s.store.MainExecutor(), actionID, statusID)
}

func (s *EntityService) DeleteActionResultStatusValues(action repo.IEAction, status repo.IEStatus) error {
	unlock := s.store.LockWrite()
	defer unlock()
	return s.store.EntityRepo().DeleteActionResultStatusValues(s.store.Context(), s.store.MainExecutor(), action, status)
}

func (s *EntityService) GetActorByName(name string) (repo.Actor, bool, error) {
	entities, err := s.store.Entities()
	if err != nil {
		return repo.Actor{}, false, err
	}
	actor, err := entities.GetActorByName(name)
	if errors.Is(err, repo.ErrNotFound) {
		return repo.Actor{}, false, nil
	}
	if err != nil {
		return repo.Actor{}, false, err
	}
	return actor, true, nil
}

func (s *EntityService) GetDeviceByName(name string) (repo.Actor, bool, error) {
	actor, found, err := s.GetActorByName(name)
	if err != nil || !found {
		return repo.Actor{}, found, err
	}
	if actor.Name().Kind() != string(repo.ActorKindDevice) {
		return repo.Actor{}, false, nil
	}
	return actor, true, nil
}

func (s *EntityService) GetAction(value repo.IEAction) (repo.Action, bool, error) {
	entities, err := s.store.Entities()
	if err != nil {
		return repo.Action{}, false, err
	}
	action, err := entities.GetAction(value)
	if errors.Is(err, repo.ErrNotFound) {
		return repo.Action{}, false, nil
	}
	if err != nil {
		return repo.Action{}, false, err
	}
	return action, true, nil
}

func (s *EntityService) GetStatus(value repo.IEStatus) (repo.Status, bool, error) {
	entities, err := s.store.Entities()
	if err != nil {
		return repo.Status{}, false, err
	}
	status, err := entities.GetStatus(value)
	if errors.Is(err, repo.ErrNotFound) {
		return repo.Status{}, false, nil
	}
	if err != nil {
		return repo.Status{}, false, err
	}
	return status, true, nil
}

func (s *EntityService) GetProcedure(value repo.IMedicalProcedure) (repo.Procedure, bool, error) {
	entities, err := s.store.Entities()
	if err != nil {
		return repo.Procedure{}, false, err
	}
	procedure, err := entities.GetProcedure(value)
	if errors.Is(err, repo.ErrNotFound) {
		return repo.Procedure{}, false, nil
	}
	if err != nil {
		return repo.Procedure{}, false, err
	}
	return procedure, true, nil
}

func (s *EntityService) StatusIDForValue(value repo.IEStatus) (int64, error) {
	return s.store.EntityRepo().StatusIDForValue(s.store.Context(), s.store.MainExecutor(), value)
}

func (s *EntityService) ActionIDForValue(value repo.IEAction) (int64, error) {
	return s.store.EntityRepo().ActionIDForValue(s.store.Context(), s.store.MainExecutor(), value)
}
