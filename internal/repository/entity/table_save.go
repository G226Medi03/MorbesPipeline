package entity

import (
	"fmt"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	entitykey "github.com/G226Medi03/MorbesPipeline/internal/entity/key"
	"github.com/G226Medi03/MorbesPipeline/internal/entity/trait"
	source "github.com/G226Medi03/SourceEntity"
)

func (s *entityTable) SaveProcedure(value repo.IMedicalProcedure, traits []trait.Trait, sources []source.Source) (int64, error) {
	if value == nil {
		return 0, fmt.Errorf("procedure value is nil")
	}
	if _, err := s.entities.InsertProcedure(value, nil, MedicalEntityTraits(traits), sourceIDs(sources)); err != nil {
		return 0, err
	}
	return s.procedureIDForValue(value)
}

func (s *entityTable) SaveMedicalProcedureOption(option repo.MedicalProcedureOption) (repo.MedicalProcedureOption, error) {
	sources := make([]source.SourceId, 0, 1)
	if option.SourceID.ToI64() > 0 {
		sources = append(sources, option.SourceID)
	}
	if err := s.entities.SaveMedicalProcedureOption(option, sources); err != nil {
		return option, err
	}
	return option, nil
}

func (s *entityTable) SaveMedicalProcedureOptionByValues(
	procedure repo.IMedicalProcedure,
	category string,
	traits []trait.Trait,
	_ []trait.Trait,
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
		Traits:           traitStrings(traits),
		MedicalProcedure: procedure,
		Indications:      indications,
		ExpectedEffects:  effects,
	}
	if len(sources) > 0 {
		option.SourceID = sources[0].SourceID
	}
	return s.SaveMedicalProcedureOption(option)
}

func traitStrings(values []trait.Trait) []string {
	result := make([]string, len(values))
	for index, value := range values {
		result[index] = string(value)
	}
	return result
}

// LinkActionResultStatusValues links typed action/status values using their
// persisted row identifiers.
func (s *entityTable) LinkActionResultStatusValues(action repo.IEAction, status repo.IEStatus, sourceID source.SourceId) error {
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
