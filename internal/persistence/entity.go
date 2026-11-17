package store

import (
	"fmt"
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"

	"github.com/G226Medi03/MorbesPipeline/internal/domain/procedureoption"
	"github.com/G226Medi03/MorbesPipeline/internal/domain/trait"
	"github.com/G226Medi03/MorbesPipeline/internal/support/ids"
	"github.com/G226Medi03/SourceEntity"
)

func actorValueByKind(kind, name string) (repo.IEActor, error) {
	switch repo.ParseActorKind(kind) {
	case repo.ActorKindDevice:
		return repo.NewDevice(name), nil
	case repo.ActorKindProcedure:
		return repo.NewActorProcedure(name), nil
	default:
		return repo.NewUnknownActor(name), nil
	}
}

func (s *Store) SaveProcedure(value repo.IMedicalProcedure, traits []trait.Trait, sources []source.SourceRef) (ids.ProcedureID, error) {
	if value == nil {
		return 0, fmt.Errorf("procedure value is nil")
	}
	unlock := s.lockWrite()
	defer unlock()
	entityStore, err := s.entities()
	if err != nil {
		return 0, err
	}
	if _, err := entityStore.InsertProcedure(value, nil, medicalEntityTraits(traits), sourceIDs(sources)); err != nil {
		return 0, err
	}
	return s.procedureIDForValue(value)
}

func (s *Store) SaveMedicalProcedureOption(option medicalprocedureoption.MedicalProcedureOption) (medicalprocedureoption.MedicalProcedureOption, error) {
	unlock := s.lockWrite()
	defer unlock()
	if err := option.Validate(); err != nil {
		return option, err
	}
	if option.ProcedureID <= 0 {
		return option, fmt.Errorf("medical procedure option procedure_id is required")
	}
	return s.OptionRepository.Save(option)
}

func (s *Store) SaveMedicalProcedureOptionByValues(
	procedureID ids.ProcedureID,
	category string,
	traits []trait.Trait,
	contextTraits []trait.Trait,
	indications []repo.IEStatus,
	effects []repo.IEStatus,
	sources []source.SourceRef,
) (medicalprocedureoption.MedicalProcedureOption, error) {
	indicationIDs, err := s.statusIDsForValues(indications)
	if err != nil {
		return medicalprocedureoption.MedicalProcedureOption{}, err
	}
	effectIDs, err := s.statusIDsForValues(effects)
	if err != nil {
		return medicalprocedureoption.MedicalProcedureOption{}, err
	}
	return s.SaveMedicalProcedureOption(medicalprocedureoption.NewMedicalProcedureOption(
		0,
		procedureID,
		category,
		traits,
		contextTraits,
		indicationIDs,
		effectIDs,
		sources,
		nil,
		nil,
	))
}

func (s *Store) statusIDsForValues(values []repo.IEStatus) ([]ids.StatusID, error) {
	result := make([]ids.StatusID, 0, len(values))
	for _, value := range values {
		id, err := s.statusIDForValue(value)
		if err != nil {
			return nil, err
		}
		if id > 0 {
			result = append(result, id)
		}
	}
	return result, nil
}

// LinkActionResultStatusValues links typed action/status values using their
// persisted row identifiers.
func medicalEntityTraits(values []trait.Trait) []repo.Trait {
	result := make([]repo.Trait, len(values))
	for index, value := range values {
		result[index] = repo.Trait(value)
	}
	return result
}

func (s *Store) LinkActionResultStatusValues(action repo.IEAction, status repo.IEStatus, sourceID source.SourceId) error {
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
