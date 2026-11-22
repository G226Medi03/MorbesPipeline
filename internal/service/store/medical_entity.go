package store

import (
	"errors"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"github.com/G226Medi03/MorbesPipeline/internal/entity/entitykey"
	source "github.com/G226Medi03/SourceEntity"
)

func (s *MedicalEntityService) SaveProcedure(value repo.IMedicalProcedure, sources []source.Source) error {
	unlock := s.store.LockWrite()
	defer unlock()
	entities, err := s.store.MedicalEntities()
	if err != nil {
		return err
	}
	_, err = entities.InsertProcedure(
		value,
		nil,
		sourceIDs(sources),
	)
	return err
}

func (s *MedicalEntityService) SaveMedicalProcedureOption(option repo.MedicalProcedureOption) (repo.MedicalProcedureOption, error) {
	unlock := s.store.LockWrite()
	defer unlock()
	entities, err := s.store.MedicalEntities()
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

func (s *MedicalEntityService) SaveMedicalProcedureOptionByValues(
	procedure repo.IMedicalProcedure,
	category string,
	indications []repo.IEStatus,
	effects []repo.IEStatus,
	sources []source.Source,
) (repo.MedicalProcedureOption, error) {
	unlock := s.store.LockWrite()
	defer unlock()
	entities, err := s.store.MedicalEntities()
	if err != nil {
		return repo.MedicalProcedureOption{}, err
	}
	key := entitykey.OptionKey(procedure, category, indications)
	option := repo.MedicalProcedureOption{
		Key:              key,
		Category:         category,
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

func (s *MedicalEntityService) LinkActionResultStatusValues(action repo.IEAction, status repo.IEStatus, sourceID source.SourceId) error {
	unlock := s.store.LockWrite()
	defer unlock()
	entities, err := s.store.MedicalEntities()
	if err != nil {
		return err
	}
	return entities.LinkActionResultStatus(action, status, sourceID)
}

func (s *MedicalEntityService) LinkStatusActionValues(status repo.IEStatus, action repo.IEAction, sourceID source.SourceId) error {
	unlock := s.store.LockWrite()
	defer unlock()
	entities, err := s.store.MedicalEntities()
	if err != nil {
		return err
	}
	return entities.LinkStatusAction(status, action, sourceID)
}

func (s *MedicalEntityService) DeleteStatusActionValues(status repo.IEStatus, action repo.IEAction) error {
	unlock := s.store.LockWrite()
	defer unlock()
	entities, err := s.store.MedicalEntities()
	if err != nil {
		return err
	}
	return entities.UnlinkStatusAction(status, action)
}

func (s *MedicalEntityService) DeleteActionResultStatusValues(action repo.IEAction, status repo.IEStatus) error {
	unlock := s.store.LockWrite()
	defer unlock()
	entities, err := s.store.MedicalEntities()
	if err != nil {
		return err
	}
	return entities.UnlinkActionResultStatus(action, status)
}

func (s *MedicalEntityService) GetActorByName(name string) (repo.Actor, bool, error) {
	entities, err := s.store.MedicalEntities()
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

func (s *MedicalEntityService) GetDeviceByName(name string) (repo.Actor, bool, error) {
	actor, found, err := s.GetActorByName(name)
	if err != nil || !found {
		return repo.Actor{}, found, err
	}
	if actor.Name().Kind() != string(repo.ActorKindDevice) {
		return repo.Actor{}, false, nil
	}
	return actor, true, nil
}

func (s *MedicalEntityService) GetAction(value repo.IEAction) (repo.Action, bool, error) {
	entities, err := s.store.MedicalEntities()
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

func (s *MedicalEntityService) GetStatus(value repo.IEStatus) (repo.Status, bool, error) {
	entities, err := s.store.MedicalEntities()
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

func (s *MedicalEntityService) GetProcedure(value repo.IMedicalProcedure) (repo.Procedure, bool, error) {
	entities, err := s.store.MedicalEntities()
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
