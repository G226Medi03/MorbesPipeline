package postprocess

import (
	"fmt"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	storerepo "github.com/G226Medi03/MorbesPipeline/internal/store/repository"
	source "github.com/G226Medi03/SourceEntity"
)

// RecordPostprocess writes one postprocess audit row and returns the source
// that represents that operation.
func (s *PostprocessService) RecordPostprocess(
	name string,
	operation storerepo.PostprocessOperation,
	entityType storerepo.PostprocessEntityType,
	entityID int64,
	targetType storerepo.PostprocessEntityType,
	targetID int64,
) (source.SourceId, error) {
	postprocessID, err := s.store.PostprocessRepository().Append(s.store.Context(), s.store.MainExecutor(), storerepo.PostprocessRecord{
		PostprocessorName: name,
		Operation:         operation,
		EntityType:        entityType,
		EntityID:          entityID,
		TargetEntityType:  targetType,
		TargetEntityID:    targetID,
	})
	if err != nil {
		return source.SourceId{}, err
	}
	return s.store.EnsureSource(source.SourceKindMorbesPostprocess, postprocessID)
}

// AuditPostprocessStatus records one status observation.
func (s *PostprocessService) AuditPostprocessStatus(name string, operation storerepo.PostprocessOperation, value repo.IEStatus) error {
	id, err := s.store.StatusIDForValue(value)
	if err != nil {
		return err
	}
	_, err = s.RecordPostprocess(name, operation, storerepo.PostprocessEntityStatus, id, "", 0)
	return err
}

// AuditPostprocessStatusCreated records a status creation and attaches the
// postprocess source to it.
func (s *PostprocessService) AuditPostprocessStatusCreated(name string, value repo.IEStatus) error {
	id, err := s.store.StatusIDForValue(value)
	if err != nil {
		return err
	}
	sourceID, err := s.RecordPostprocess(name, storerepo.PostprocessOperationCreated, storerepo.PostprocessEntityStatus, id, "", 0)
	if err != nil {
		return err
	}
	return s.AddPostprocessSource(storerepo.PostprocessEntityStatus, id, sourceID)
}

// AuditPostprocessAction records one action observation.
func (s *PostprocessService) AuditPostprocessAction(name string, operation storerepo.PostprocessOperation, value repo.IEAction) error {
	id, err := s.store.ActionIDForValue(value)
	if err != nil {
		return err
	}
	_, err = s.RecordPostprocess(name, operation, storerepo.PostprocessEntityAction, id, "", 0)
	return err
}

// AuditPostprocessActionCreated records an action creation and attaches the
// postprocess source to it.
func (s *PostprocessService) AuditPostprocessActionCreated(name string, value repo.IEAction) error {
	id, err := s.store.ActionIDForValue(value)
	if err != nil {
		return err
	}
	sourceID, err := s.RecordPostprocess(name, storerepo.PostprocessOperationCreated, storerepo.PostprocessEntityAction, id, "", 0)
	if err != nil {
		return err
	}
	return s.AddPostprocessSource(storerepo.PostprocessEntityAction, id, sourceID)
}

// AuditPostprocessStatusAction records one status-action link operation.
func (s *PostprocessService) AuditPostprocessStatusAction(
	name string,
	operation storerepo.PostprocessOperation,
	status repo.IEStatus,
	action repo.IEAction,
) error {
	statusID, err := s.store.StatusIDForValue(status)
	if err != nil {
		return err
	}
	actionID, err := s.store.ActionIDForValue(action)
	if err != nil {
		return err
	}
	_, err = s.RecordPostprocess(
		name,
		operation,
		storerepo.PostprocessEntityStatus,
		statusID,
		storerepo.PostprocessEntityAction,
		actionID,
	)
	return err
}

// AuditPostprocessActionResultStatus records one action-result link operation.
func (s *PostprocessService) AuditPostprocessActionResultStatus(
	name string,
	operation storerepo.PostprocessOperation,
	action repo.IEAction,
	status repo.IEStatus,
) error {
	actionID, err := s.store.ActionIDForValue(action)
	if err != nil {
		return err
	}
	statusID, err := s.store.StatusIDForValue(status)
	if err != nil {
		return err
	}
	_, err = s.RecordPostprocess(
		name,
		operation,
		storerepo.PostprocessEntityAction,
		actionID,
		storerepo.PostprocessEntityStatus,
		statusID,
	)
	return err
}

// AddPostprocessSource links a postprocess source to an entity row.
func (s *PostprocessService) AddPostprocessSource(
	entityType storerepo.PostprocessEntityType,
	entityID int64,
	sourceID source.SourceId,
) error {
	table, column, ok := postprocessSourceTable(entityType)
	if !ok {
		return fmt.Errorf("unsupported postprocess entity type: %s", entityType)
	}
	return s.store.SourceRepository().AddSourceIDs(
		s.store.Context(),
		s.store.SourceExecutor(),
		table,
		column,
		entityID,
		[]source.SourceId{sourceID},
	)
}

func postprocessSourceTable(entityType storerepo.PostprocessEntityType) (string, string, bool) {
	switch entityType {
	case storerepo.PostprocessEntityStatus:
		return "status_sources", "estatus_id", true
	case storerepo.PostprocessEntityAction:
		return "action_sources", "eaction_id", true
	case storerepo.PostprocessEntityActor:
		return "actor_sources", "eactor_id", true
	case storerepo.PostprocessEntityProcedure:
		return "procedure_sources", "emedical_procedure_id", true
	case storerepo.PostprocessEntityOption:
		return "medical_procedure_option_sources", "medical_procedure_option_id", true
	default:
		return "", "", false
	}
}
