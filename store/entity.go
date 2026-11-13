package store

import (
	"encoding/json"
	"fmt"
	medicalcodec "github.com/G226Medi03/MedicalEntity/codec"

	"github.com/G226Medi03/MedicalEntity/ename/edisease"
	"github.com/G226Medi03/MedicalEntity/ename/eindex"
	"github.com/G226Medi03/MedicalEntity/ename/esymptom"
	medicalentity "github.com/G226Medi03/MedicalEntity/entity"
	entity "github.com/G226Medi03/MorbesPipeline/legacy"
	"github.com/G226Medi03/MorbesPipeline/store/medicalprocedureoption"
	"github.com/G226Medi03/SourceEntity"
)

func (s *Store) GetDiseaseByID(id int64) (medicalentity.Disease, bool, error) {
	value, err := s.DiseaseRepository.Get(typedDiseaseID(id))
	if err != nil {
		if err == medicalentity.ErrNotFound {
			return medicalentity.Disease{}, false, nil
		}
		return medicalentity.Disease{}, false, err
	}
	refs, err := s.sourceRefsFor("disease_sources", "disease_id", id)
	if err != nil {
		return medicalentity.Disease{}, false, err
	}
	value.Sources = refs
	return value, true, nil
}

func (s *Store) GetIndexByID(id int64) (medicalentity.Index, bool, error) {
	value, err := s.IndexRepository.Get(typedIndexID(id))
	if err != nil {
		if err == medicalentity.ErrNotFound {
			return medicalentity.Index{}, false, nil
		}
		return medicalentity.Index{}, false, err
	}
	refs, err := s.sourceRefsFor("index_sources", "index_id", id)
	if err != nil {
		return medicalentity.Index{}, false, err
	}
	value.Sources = refs
	return value, true, nil
}

func (s *Store) GetSymptomByID(id int64) (medicalentity.Symptom, bool, error) {
	value, err := s.SymptomRepository.Get(typedSymptomID(id))
	if err != nil {
		if err == medicalentity.ErrNotFound {
			return medicalentity.Symptom{}, false, nil
		}
		return medicalentity.Symptom{}, false, err
	}
	refs, err := s.sourceRefsFor("symptom_sources", "symptom_id", id)
	if err != nil {
		return medicalentity.Symptom{}, false, err
	}
	value.Sources = refs
	return value, true, nil
}

func (s *Store) GetProcedureByID(id entity.ProcedureID) (entity.Procedure, bool, error) {
	value, err := s.ProcedureRepository.Get(entityProcedureID(id))
	if err != nil {
		if err == medicalentity.ErrNotFound {
			return entity.NewProcedure("", nil, nil), false, nil
		}
		return entity.NewProcedure("", nil, nil), false, err
	}
	local, err := localProcedureFromExternal(value)
	if err != nil {
		return entity.NewProcedure("", nil, nil), false, err
	}
	refs, err := s.sourceRefsFor("procedure_sources", "emedical_procedure_id", id.ToI64())
	if err != nil {
		return entity.NewProcedure("", nil, nil), false, err
	}
	local.Sources = refs
	return local, true, nil
}

func (s *Store) GetMedicalProcedureOptionByID(id entity.MedicalProcedureOptionID) (medicalprocedureoption.MedicalProcedureOption, bool, error) {
	value, err := s.OptionRepository.Get(medicalprocedureoption.MedicalProcedureOptionID(id))
	if err != nil {
		if err == medicalentity.ErrNotFound {
			return medicalprocedureoption.MedicalProcedureOption{}, false, nil
		}
		return medicalprocedureoption.MedicalProcedureOption{}, false, err
	}
	if procedureValue, ok, procedureErr := s.GetProcedureByID(value.ProcedureID); procedureErr == nil && ok {
		if encoded, encodeErr := json.Marshal(procedureValue); encodeErr == nil {
			if parsed, parseErr := entity.UnmarshalProcedure(encoded); parseErr == nil {
				value.Procedure = parsed
			}
		}
	}
	for _, statusID := range value.IndicationStatusIDs {
		status, statusErr := s.GetStatusByID(statusID)
		if statusErr != nil {
			continue
		}
		external, encodeErr := externalStatusFromLocal(status)
		if encodeErr != nil {
			continue
		}
		raw, marshalErr := medicalcodec.MarshalStatus(external)
		if marshalErr != nil {
			continue
		}
		parsed, parseErr := entity.UnmarshalStatus(raw)
		if parseErr == nil {
			value.Indication = append(value.Indication, parsed)
		}
	}
	return value, true, nil
}

func (s *Store) SaveIndex(name string, sources []source.SourceRef) (medicalentity.Index, error) {
	unlock := s.lockWrite()
	defer unlock()
	saved, err := s.IndexRepository.Insert(eindex.EIndex{Name: name}, nil, sourceIDs(sources))
	if err != nil {
		return medicalentity.Index{}, err
	}
	return saved, nil
}

func (s *Store) SaveDisease(name string, sources []source.SourceRef) (medicalentity.Disease, error) {
	unlock := s.lockWrite()
	defer unlock()
	saved, err := s.DiseaseRepository.Insert(edisease.EDisease(name), nil, sourceIDs(sources))
	if err != nil {
		return medicalentity.Disease{}, err
	}
	return saved, nil
}

func (s *Store) SaveSymptom(name string, sources []source.SourceRef) (medicalentity.Symptom, error) {
	unlock := s.lockWrite()
	defer unlock()
	saved, err := s.SymptomRepository.Insert(esymptom.EOtherSymptom{Name: name}, nil, sourceIDs(sources))
	if err != nil {
		return medicalentity.Symptom{}, err
	}
	return saved, nil
}

func (s *Store) SaveProcedureWithTraits(kind, data string, traits []entity.Trait, sources []source.SourceRef) (entity.ProcedureID, error) {
	unlock := s.lockWrite()
	defer unlock()
	if err := entity.ValidateSources(sources); err != nil {
		return 0, err
	}
	value := entity.Procedure{Variant: kind, Value: decode_data_value(data), Sources: sources}
	external, err := externalProcedureFromLocal(value)
	if err != nil {
		return 0, err
	}
	saved, err := s.ProcedureRepository.Insert(external, nil, traits, sourceIDs(sources))
	if err != nil {
		return 0, err
	}
	if err := s.validateKindPtr("emedical_procedure"); err != nil {
		return 0, err
	}
	id := localProcedureID(saved.Id())
	name := ""
	if label, ok := s.ProcedureLabelByID(id); ok {
		name = label
	}
	if _, err := s.ensureProcedureActor(id, name, firstSourceID(sources)); err != nil {
		return 0, err
	}
	return id, nil
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
	saved, err := s.OptionRepository.Save(option)
	if err != nil {
		return option, err
	}
	return saved, nil
}
