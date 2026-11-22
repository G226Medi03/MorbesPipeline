package store

import (
	"fmt"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	entitykey "github.com/G226Medi03/MorbesPipeline/internal/domain/entitykey"
	source "github.com/G226Medi03/SourceEntity"
)

func (s *StoreContainer) SaveProcedure(value repo.IMedicalProcedure, sources []source.Source) (int64, error) {
	if value == nil {
		return 0, fmt.Errorf("procedure value is nil")
	}
	unlock := s.lockWrite()
	defer unlock()
	entityStore, err := s.entities()
	if err != nil {
		return 0, err
	}
	if _, err := entityStore.InsertProcedure(value, nil, sourceIDs(sources)); err != nil {
		return 0, err
	}
	return s.procedureIDForValue(value)
}

func (s *StoreContainer) SaveMedicalProcedureOption(option repo.MedicalProcedureOption) (repo.MedicalProcedureOption, error) {
	unlock := s.lockWrite()
	defer unlock()
	entityStore, err := s.entities()
	if err != nil {
		return option, err
	}
	sources := make([]source.SourceId, 0, 1)
	if option.SourceID.ToI64() > 0 {
		sources = append(sources, option.SourceID)
	}
	if err := entityStore.SaveMedicalProcedureOption(option, sources); err != nil {
		return option, err
	}
	return option, nil
}

func (s *StoreContainer) SaveMedicalProcedureOptionByValues(
	procedure repo.IMedicalProcedure,
	category string,
	indications []repo.IEStatus,
	effects []repo.IEStatus,
	sources []source.Source,
) (repo.MedicalProcedureOption, error) {
	key := entitykey.Procedure(procedure) + "|" + category
	for _, indication := range indications {
		key += "|" + entitykey.Status(indication)
	}
	option := repo.MedicalProcedureOption{
		Key:              key,
		Category:         category,
		MedicalProcedure: procedure,
		Indications:      indications,
		ExpectedEffects:  effects,
	}
	if len(sources) > 0 {
		option.SourceID = sources[0].SourceID
	}
	return s.SaveMedicalProcedureOption(option)
}

// LinkActionResultStatusValues links typed action/status values using their
// persisted row identifiers.
func (s *StoreContainer) LinkActionResultStatusValues(action repo.IEAction, status repo.IEStatus, sourceID source.SourceId) error {
	actionID, err := s.actionIDForValue(action)
	if err != nil {
		return err
	}
	statusID, err := s.statusIDForValue(status)
	if err != nil {
		return err
	}
	return s.LinkActionResultStatus(actionID, statusID, sourceID)
}
