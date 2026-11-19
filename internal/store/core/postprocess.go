package store

import (
	"context"
	"fmt"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	storerepo "github.com/G226Medi03/MorbesPipeline/internal/store/repository"
	source "github.com/G226Medi03/SourceEntity"
)

// RecordPostprocess writes one postprocess audit row and returns the source
// that represents that operation.
func (s *StoreContainer) RecordPostprocess(
	name string,
	operation storerepo.PostprocessOperation,
	entityType storerepo.PostprocessEntityType,
	entityID int64,
	targetType storerepo.PostprocessEntityType,
	targetID int64,
) (source.SourceId, error) {
	postprocessID, err := s.Postprocess.Append(s.ctx, s, storerepo.PostprocessRecord{
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
	return s.EnsureSource(source.SourceKindMorbesPostprocess, postprocessID)
}

// AuditPostprocessStatus records one status observation.
func (s *StoreContainer) AuditPostprocessStatus(name string, operation storerepo.PostprocessOperation, value repo.IEStatus) error {
	id, err := s.statusIDForValue(value)
	if err != nil {
		return err
	}
	_, err = s.RecordPostprocess(name, operation, storerepo.PostprocessEntityStatus, id, "", 0)
	return err
}

// AuditPostprocessStatusCreated records a status creation and attaches the
// postprocess source to it.
func (s *StoreContainer) AuditPostprocessStatusCreated(name string, value repo.IEStatus) error {
	id, err := s.statusIDForValue(value)
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
func (s *StoreContainer) AuditPostprocessAction(name string, operation storerepo.PostprocessOperation, value repo.IEAction) error {
	id, err := s.actionIDForValue(value)
	if err != nil {
		return err
	}
	_, err = s.RecordPostprocess(name, operation, storerepo.PostprocessEntityAction, id, "", 0)
	return err
}

// AuditPostprocessActionCreated records an action creation and attaches the
// postprocess source to it.
func (s *StoreContainer) AuditPostprocessActionCreated(name string, value repo.IEAction) error {
	id, err := s.actionIDForValue(value)
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
func (s *StoreContainer) AuditPostprocessStatusAction(
	name string,
	operation storerepo.PostprocessOperation,
	status repo.IEStatus,
	action repo.IEAction,
) error {
	statusID, err := s.statusIDForValue(status)
	if err != nil {
		return err
	}
	actionID, err := s.actionIDForValue(action)
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
func (s *StoreContainer) AuditPostprocessActionResultStatus(
	name string,
	operation storerepo.PostprocessOperation,
	action repo.IEAction,
	status repo.IEStatus,
) error {
	actionID, err := s.actionIDForValue(action)
	if err != nil {
		return err
	}
	statusID, err := s.statusIDForValue(status)
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
func (s *StoreContainer) AddPostprocessSource(
	entityType storerepo.PostprocessEntityType,
	entityID int64,
	sourceID source.SourceId,
) error {
	table, column, ok := postprocessSourceTable(entityType)
	if !ok {
		return fmt.Errorf("unsupported postprocess entity type: %s", entityType)
	}
	return s.Source.AddSourceIDs(
		s.ctx,
		s.sourceExecutor(),
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

// WithPostprocessTx runs postprocess writes in one main.db transaction.
func (s *StoreContainer) WithPostprocessTx(ctx context.Context, fn func(*StoreContainer) error) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("store db is required")
	}
	if s.sourceDB != s.db || s.solveDB != s.db {
		return fmt.Errorf("postprocess requires a single database")
	}
	return s.WithTx(ctx, fn)
}
