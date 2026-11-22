package service

import (
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"github.com/G226Medi03/MorbesPipeline/internal/entity/entitykey"
	postprocessrepo "github.com/G226Medi03/MorbesPipeline/internal/repository/postprocess"
	source "github.com/G226Medi03/SourceEntity"
)

// RecordPostprocess writes one postprocess audit row and returns the source
// that represents that operation.
func (s *PostprocessService) RecordPostprocess(
	name string,
	operation postprocessrepo.Operation,
	entityType postprocessrepo.EntityType,
	entityKey string,
	targetType postprocessrepo.EntityType,
	targetKey string,
) (source.SourceId, error) {
	postprocessID, err := s.store.PostprocessRepo().Append(s.store.Context(), s.store.MainExecutor(), postprocessrepo.Record{
		PostprocessorName: name,
		Operation:         operation,
		EntityType:        entityType,
		EntityKey:         entityKey,
		TargetEntityType:  targetType,
		TargetEntityKey:   targetKey,
	})
	if err != nil {
		return source.SourceId{}, err
	}
	return s.store.SourceRepo().Ensure(s.store.Context(), s.store.SourceExecutor(), source.SourceKindMorbesPostprocess, postprocessID)
}

// AuditPostprocessStatus records one status observation.
func (s *PostprocessService) AuditPostprocessStatus(name string, operation postprocessrepo.Operation, value repo.IEStatus) error {
	_, err := s.RecordPostprocess(name, operation, postprocessrepo.EntityStatus, entitykey.Status(value), "", "")
	return err
}

// AuditPostprocessStatusCreated records a status creation and attaches the
// postprocess source to it.
func (s *PostprocessService) AuditPostprocessStatusCreated(name string, value repo.IEStatus) error {
	sourceID, err := s.RecordPostprocess(name, postprocessrepo.OperationCreated, postprocessrepo.EntityStatus, entitykey.Status(value), "", "")
	if err != nil {
		return err
	}
	return s.AddStatusSource(value, sourceID)
}

// AuditPostprocessAction records one action observation.
func (s *PostprocessService) AuditPostprocessAction(name string, operation postprocessrepo.Operation, value repo.IEAction) error {
	_, err := s.RecordPostprocess(name, operation, postprocessrepo.EntityAction, entitykey.Action(value), "", "")
	return err
}

// AuditPostprocessActionCreated records an action creation and attaches the
// postprocess source to it.
func (s *PostprocessService) AuditPostprocessActionCreated(name string, value repo.IEAction) error {
	sourceID, err := s.RecordPostprocess(name, postprocessrepo.OperationCreated, postprocessrepo.EntityAction, entitykey.Action(value), "", "")
	if err != nil {
		return err
	}
	return s.AddActionSource(value, sourceID)
}

// AuditPostprocessStatusAction records one status-action link operation.
func (s *PostprocessService) AuditPostprocessStatusAction(
	name string,
	operation postprocessrepo.Operation,
	status repo.IEStatus,
	action repo.IEAction,
) error {
	_, err := s.RecordPostprocess(
		name,
		operation,
		postprocessrepo.EntityStatus,
		entitykey.Status(status),
		postprocessrepo.EntityAction,
		entitykey.Action(action),
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
	_, err := s.RecordPostprocess(
		name,
		operation,
		postprocessrepo.EntityAction,
		entitykey.Action(action),
		postprocessrepo.EntityStatus,
		entitykey.Status(status),
	)
	return err
}

// AddStatusSource attaches a postprocess source to a status.
func (s *PostprocessService) AddStatusSource(value repo.IEStatus, sourceID source.SourceId) error {
	entities, err := s.store.Entities()
	if err != nil {
		return err
	}
	status, err := entities.GetStatus(value)
	if err != nil {
		return err
	}
	status.Sources = appendSource(status.Sources, sourceID)
	return entities.SaveStatus(status)
}

// AddActionSource attaches a postprocess source to an action.
func (s *PostprocessService) AddActionSource(value repo.IEAction, sourceID source.SourceId) error {
	entities, err := s.store.Entities()
	if err != nil {
		return err
	}
	action, err := entities.GetAction(value)
	if err != nil {
		return err
	}
	action.Sources = appendSource(action.Sources, sourceID)
	return entities.SaveAction(action)
}

func appendSource(sources []source.Source, sourceID source.SourceId) []source.Source {
	if sourceID.ToI64() <= 0 {
		return sources
	}
	for _, item := range sources {
		if item.SourceID == sourceID {
			return sources
		}
	}
	return append(sources, source.Source{SourceID: sourceID})
}
