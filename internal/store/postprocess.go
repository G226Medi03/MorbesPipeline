package store

import (
	"context"
	"fmt"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"github.com/G226Medi03/MorbesPipeline/internal/store/postprocess"
	source "github.com/G226Medi03/SourceEntity"
	storesource "github.com/G226Medi03/SourceEntityRepository"
)

// RecordPostprocess writes one postprocess audit row and returns the source
// that represents that operation.
func (s *StoreContainer) RecordPostprocess(
	name string,
	operation postprocess.Operation,
	entityType postprocess.EntityType,
	entityID int64,
	targetType postprocess.EntityType,
	targetID int64,
) (source.SourceId, error) {
	postprocessID, err := s.Postprocess.Append(postprocess.Record{
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
func (s *StoreContainer) AuditPostprocessStatus(name string, operation postprocess.Operation, value repo.IEStatus) error {
	id, err := s.statusIDForValue(value)
	if err != nil {
		return err
	}
	_, err = s.RecordPostprocess(name, operation, postprocess.EntityStatus, id.ToI64(), "", 0)
	return err
}

// AuditPostprocessStatusCreated records a status creation and attaches the
// postprocess source to it.
func (s *StoreContainer) AuditPostprocessStatusCreated(name string, value repo.IEStatus) error {
	id, err := s.statusIDForValue(value)
	if err != nil {
		return err
	}
	sourceID, err := s.RecordPostprocess(name, postprocess.OperationCreated, postprocess.EntityStatus, id.ToI64(), "", 0)
	if err != nil {
		return err
	}
	return s.AddPostprocessSource(postprocess.EntityStatus, id.ToI64(), sourceID)
}

// AuditPostprocessAction records one action observation.
func (s *StoreContainer) AuditPostprocessAction(name string, operation postprocess.Operation, value repo.IEAction) error {
	id, err := s.actionIDForValue(value)
	if err != nil {
		return err
	}
	_, err = s.RecordPostprocess(name, operation, postprocess.EntityAction, id.ToI64(), "", 0)
	return err
}

// AuditPostprocessActionCreated records an action creation and attaches the
// postprocess source to it.
func (s *StoreContainer) AuditPostprocessActionCreated(name string, value repo.IEAction) error {
	id, err := s.actionIDForValue(value)
	if err != nil {
		return err
	}
	sourceID, err := s.RecordPostprocess(name, postprocess.OperationCreated, postprocess.EntityAction, id.ToI64(), "", 0)
	if err != nil {
		return err
	}
	return s.AddPostprocessSource(postprocess.EntityAction, id.ToI64(), sourceID)
}

// AuditPostprocessStatusAction records one status-action link operation.
func (s *StoreContainer) AuditPostprocessStatusAction(
	name string,
	operation postprocess.Operation,
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
		postprocess.EntityStatus,
		statusID.ToI64(),
		postprocess.EntityAction,
		actionID.ToI64(),
	)
	return err
}

// AuditPostprocessActionResultStatus records one action-result link operation.
func (s *StoreContainer) AuditPostprocessActionResultStatus(
	name string,
	operation postprocess.Operation,
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
		postprocess.EntityAction,
		actionID.ToI64(),
		postprocess.EntityStatus,
		statusID.ToI64(),
	)
	return err
}

// AddPostprocessSource links a postprocess source to an entity row.
func (s *StoreContainer) AddPostprocessSource(
	entityType postprocess.EntityType,
	entityID int64,
	sourceID source.SourceId,
) error {
	table, column, ok := postprocessSourceTable(entityType)
	if !ok {
		return fmt.Errorf("unsupported postprocess entity type: %s", entityType)
	}
	return storesource.AddSourceIDs(
		s.ctx,
		s.sourceExecutor(),
		table,
		column,
		entityID,
		[]source.SourceId{sourceID},
	)
}

func postprocessSourceTable(entityType postprocess.EntityType) (string, string, bool) {
	switch entityType {
	case postprocess.EntityStatus:
		return "status_sources", "estatus_id", true
	case postprocess.EntityAction:
		return "action_sources", "eaction_id", true
	case postprocess.EntityActor:
		return "actor_sources", "eactor_id", true
	case postprocess.EntityProcedure:
		return "procedure_sources", "emedical_procedure_id", true
	case postprocess.EntityOption:
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
