package service

import (
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"github.com/G226Medi03/MorbesPipeline/internal/domain/trait"
	source "github.com/G226Medi03/SourceEntity"
)

func (s *EntityService) SaveProcedure(value repo.IMedicalProcedure, traits []trait.Trait, sources []source.Source) (int64, error) {
	unlock := s.store.LockWrite()
	defer unlock()
	entities, err := s.store.Entities()
	if err != nil {
		return 0, err
	}
	return s.store.EntityRepo().SaveProcedure(s.store.Context(), s.store.MainExecutor(), entities, value, traits, sources)
}

func (s *EntityService) SaveMedicalProcedureOption(option repo.MedicalProcedureOption) (repo.MedicalProcedureOption, error) {
	unlock := s.store.LockWrite()
	defer unlock()
	entities, err := s.store.Entities()
	if err != nil {
		return option, err
	}
	return s.store.EntityRepo().SaveMedicalProcedureOption(s.store.Context(), s.store.MainExecutor(), entities, option)
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
	return s.store.EntityRepo().SaveMedicalProcedureOptionByValues(
		s.store.Context(), s.store.MainExecutor(), entities,
		procedure, category, traits, nil, indications, effects, sources,
	)
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
	return s.store.EntityRepo().LinkStatusActionValues(s.store.Context(), s.store.MainExecutor(), status, action, sourceID)
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
	return s.store.EntityRepo().GetActorByName(s.store.Context(), s.store.MainExecutor(), entities, name)
}

func (s *EntityService) GetDeviceByName(name string) (repo.Actor, bool, error) {
	entities, err := s.store.Entities()
	if err != nil {
		return repo.Actor{}, false, err
	}
	return s.store.EntityRepo().GetDeviceByName(s.store.Context(), s.store.MainExecutor(), entities, name)
}

func (s *EntityService) GetAction(value repo.IEAction) (repo.Action, bool, error) {
	entities, err := s.store.Entities()
	if err != nil {
		return repo.Action{}, false, err
	}
	return s.store.EntityRepo().GetAction(s.store.Context(), s.store.MainExecutor(), entities, value)
}

func (s *EntityService) GetStatus(value repo.IEStatus) (repo.Status, bool, error) {
	entities, err := s.store.Entities()
	if err != nil {
		return repo.Status{}, false, err
	}
	return s.store.EntityRepo().GetStatus(s.store.Context(), s.store.MainExecutor(), entities, value)
}

func (s *EntityService) GetProcedure(value repo.IMedicalProcedure) (repo.Procedure, bool, error) {
	entities, err := s.store.Entities()
	if err != nil {
		return repo.Procedure{}, false, err
	}
	return s.store.EntityRepo().GetProcedure(s.store.Context(), s.store.MainExecutor(), entities, value)
}

func (s *EntityService) StatusIDForValue(value repo.IEStatus) (int64, error) {
	return s.store.EntityRepo().StatusIDForValue(s.store.Context(), s.store.MainExecutor(), value)
}

func (s *EntityService) ActionIDForValue(value repo.IEAction) (int64, error) {
	return s.store.EntityRepo().ActionIDForValue(s.store.Context(), s.store.MainExecutor(), value)
}
