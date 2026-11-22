package entity

import (
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	entitykey "github.com/G226Medi03/MorbesPipeline/internal/entity/key"
	"github.com/G226Medi03/MorbesPipeline/internal/entity/trait"
	source "github.com/G226Medi03/SourceEntity"
)

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
	optionSources := make([]source.SourceId, 0, len(sources))
	for _, item := range sources {
		if item.SourceID.ToI64() > 0 {
			optionSources = append(optionSources, item.SourceID)
		}
	}
	if err := s.entities.SaveMedicalProcedureOption(option, optionSources); err != nil {
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
