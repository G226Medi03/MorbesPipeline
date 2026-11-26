package store

import (
	"fmt"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	postprocessrepo "github.com/G226Medi03/MorbesPipeline/internal/repository/postprocess"
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
	entityID, err := s.statusID(value)
	if err != nil {
		return err
	}
	_, err = s.RecordPostprocess(name, operation, postprocessrepo.EntityStatus, entityID, "", 0)
	return err
}

// AuditPostprocessStatusCreated records a status creation and attaches the
// postprocess source to it.
func (s *PostprocessService) AuditPostprocessStatusCreated(name string, value repo.IEStatus) error {
	entityID, err := s.statusID(value)
	if err != nil {
		return err
	}
	sourceID, err := s.RecordPostprocess(name, postprocessrepo.OperationCreated, postprocessrepo.EntityStatus, entityID, "", 0)
	if err != nil {
		return err
	}
	return s.AddStatusSource(value, sourceID)
}

// AuditPostprocessAction records one action observation.
func (s *PostprocessService) AuditPostprocessAction(name string, operation postprocessrepo.Operation, value repo.IEAction) error {
	entityID, err := s.actionID(value)
	if err != nil {
		return err
	}
	_, err = s.RecordPostprocess(name, operation, postprocessrepo.EntityAction, entityID, "", 0)
	return err
}

// AuditPostprocessActionCreated records an action creation and attaches the
// postprocess source to it.
func (s *PostprocessService) AuditPostprocessActionCreated(name string, value repo.IEAction) error {
	entityID, err := s.actionID(value)
	if err != nil {
		return err
	}
	sourceID, err := s.RecordPostprocess(name, postprocessrepo.OperationCreated, postprocessrepo.EntityAction, entityID, "", 0)
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
	statusID, err := s.statusID(status)
	if err != nil {
		return err
	}
	actionID, err := s.actionID(action)
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
	actionID, err := s.actionID(action)
	if err != nil {
		return err
	}
	statusID, err := s.statusID(status)
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

func (s *PostprocessService) statusID(value repo.IEStatus) (int64, error) {
	entities, err := s.store.MedicalEntities()
	if err != nil {
		return 0, err
	}
	id, found, err := entities.GetStatusID(value)
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, fmt.Errorf("postprocess status id를 찾지 못했습니다: %s", value.Kind())
	}
	return id, nil
}

func (s *PostprocessService) actionID(value repo.IEAction) (int64, error) {
	entities, err := s.store.MedicalEntities()
	if err != nil {
		return 0, err
	}
	id, found, err := entities.GetActionID(value)
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, fmt.Errorf("postprocess action id를 찾지 못했습니다: %s", value.Kind())
	}
	return id, nil
}

// AddStatusSource attaches a postprocess source to a status.
func (s *PostprocessService) AddStatusSource(value repo.IEStatus, sourceID source.SourceId) error {
	entities, err := s.store.MedicalEntities()
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
	entities, err := s.store.MedicalEntities()
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
