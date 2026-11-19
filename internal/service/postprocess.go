package service

import (
	"fmt"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	postprocessrepo "github.com/G226Medi03/MorbesPipeline/internal/store/repository/postprocess"
	source "github.com/G226Medi03/SourceEntity"
)

// RecordPostprocess writes one postprocess audit row and returns the source
// that represents that operation.
func (s *PostprocessService) RecordPostprocess(
	name string,
	operation postprocessrepo.Operation,
	entityType postprocessrepo.EntityType,
	entityID int64,
	targetType postprocessrepo.EntityType,
	targetID int64,
) (source.SourceId, error) {
	postprocessID, err := s.store.PostprocessRepo().Append(s.store.Context(), s.store.MainExecutor(), postprocessrepo.Record{
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
	return s.store.SourceRepo().Ensure(s.store.Context(), s.store.SourceExecutor(), source.SourceKindMorbesPostprocess, postprocessID)
}

// AuditPostprocessStatus records one status observation.
func (s *PostprocessService) AuditPostprocessStatus(name string, operation postprocessrepo.Operation, value repo.IEStatus) error {
	id, err := s.entity.StatusIDForValue(value)
	if err != nil {
		return err
	}
	_, err = s.RecordPostprocess(name, operation, postprocessrepo.EntityStatus, id, "", 0)
	return err
}

// AuditPostprocessStatusCreated records a status creation and attaches the
// postprocess source to it.
func (s *PostprocessService) AuditPostprocessStatusCreated(name string, value repo.IEStatus) error {
	id, err := s.entity.StatusIDForValue(value)
	if err != nil {
		return err
	}
	sourceID, err := s.RecordPostprocess(name, postprocessrepo.OperationCreated, postprocessrepo.EntityStatus, id, "", 0)
	if err != nil {
		return err
	}
	return s.AddPostprocessSource(postprocessrepo.EntityStatus, id, sourceID)
}

// AuditPostprocessAction records one action observation.
func (s *PostprocessService) AuditPostprocessAction(name string, operation postprocessrepo.Operation, value repo.IEAction) error {
	id, err := s.entity.ActionIDForValue(value)
	if err != nil {
		return err
	}
	_, err = s.RecordPostprocess(name, operation, postprocessrepo.EntityAction, id, "", 0)
	return err
}

// AuditPostprocessActionCreated records an action creation and attaches the
// postprocess source to it.
func (s *PostprocessService) AuditPostprocessActionCreated(name string, value repo.IEAction) error {
	id, err := s.entity.ActionIDForValue(value)
	if err != nil {
		return err
	}
	sourceID, err := s.RecordPostprocess(name, postprocessrepo.OperationCreated, postprocessrepo.EntityAction, id, "", 0)
	if err != nil {
		return err
	}
	return s.AddPostprocessSource(postprocessrepo.EntityAction, id, sourceID)
}

// AuditPostprocessStatusAction records one status-action link operation.
func (s *PostprocessService) AuditPostprocessStatusAction(
	name string,
	operation postprocessrepo.Operation,
	status repo.IEStatus,
	action repo.IEAction,
) error {
	statusID, err := s.entity.StatusIDForValue(status)
	if err != nil {
		return err
	}
	actionID, err := s.entity.ActionIDForValue(action)
	if err != nil {
		return err
	}
	_, err = s.RecordPostprocess(
		name,
		operation,
		postprocessrepo.EntityStatus,
		statusID,
		postprocessrepo.EntityAction,
		actionID,
	)
	return err
}

// AuditPostprocessActionResultStatus records one action-result link operation.
func (s *PostprocessService) AuditPostprocessActionResultStatus(
	name string,
	operation postprocessrepo.Operation,
	action repo.IEAction,
	status repo.IEStatus,
) error {
	actionID, err := s.entity.ActionIDForValue(action)
	if err != nil {
		return err
	}
	statusID, err := s.entity.StatusIDForValue(status)
	if err != nil {
		return err
	}
	_, err = s.RecordPostprocess(
		name,
		operation,
		postprocessrepo.EntityAction,
		actionID,
		postprocessrepo.EntityStatus,
		statusID,
	)
	return err
}

// AddPostprocessSource links a postprocess source to an entity row.
func (s *PostprocessService) AddPostprocessSource(
	entityType postprocessrepo.EntityType,
	entityID int64,
	sourceID source.SourceId,
) error {
	table, column, ok := postprocessSourceTable(entityType)
	if !ok {
		return fmt.Errorf("unsupported postprocess entity type: %s", entityType)
	}
	return s.store.SourceRepo().AddSourceIDs(
		s.store.Context(),
		s.store.SourceExecutor(),
		table,
		column,
		entityID,
		[]source.SourceId{sourceID},
	)
}

func postprocessSourceTable(entityType postprocessrepo.EntityType) (string, string, bool) {
	switch entityType {
	case postprocessrepo.EntityStatus:
		return "status_sources", "estatus_id", true
	case postprocessrepo.EntityAction:
		return "action_sources", "eaction_id", true
	case postprocessrepo.EntityActor:
		return "actor_sources", "eactor_id", true
	case postprocessrepo.EntityProcedure:
		return "procedure_sources", "emedical_procedure_id", true
	case postprocessrepo.EntityOption:
		return "medical_procedure_option_sources", "medical_procedure_option_id", true
	default:
		return "", "", false
	}
}
