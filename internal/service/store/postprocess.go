package store

import (
	"encoding/json"

	"github.com/G226Medi03/MedicalEntityParser/llmparser"
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
	entityPayload string,
	targetType postprocessrepo.EntityType,
	targetPayload string,
) (source.SourceId, error) {
	postprocessID, err := s.store.PostprocessRepo().Append(s.store.Context(), s.store.MainExecutor(), postprocessrepo.Record{
		PostprocessorName:   name,
		Operation:           operation,
		EntityType:          entityType,
		EntityPayload:       entityPayload,
		TargetEntityType:    targetType,
		TargetEntityPayload: targetPayload,
	})
	if err != nil {
		return source.SourceId{}, err
	}
	return s.store.SourceRepo().Ensure(s.store.Context(), s.store.SourceExecutor(), source.SourceKindMorbesPostprocess, postprocessID)
}

// AuditPostprocessStatus records one status observation.
func (s *PostprocessService) AuditPostprocessStatus(name string, operation postprocessrepo.Operation, value repo.IEStatus) error {
	payload, err := statusPayload(value)
	if err != nil {
		return err
	}
	_, err = s.RecordPostprocess(name, operation, postprocessrepo.EntityStatus, payload, "", "")
	return err
}

// AuditPostprocessStatusCreated records a status creation and attaches the
// postprocess source to it.
func (s *PostprocessService) AuditPostprocessStatusCreated(name string, value repo.IEStatus) error {
	payload, err := statusPayload(value)
	if err != nil {
		return err
	}
	sourceID, err := s.RecordPostprocess(name, postprocessrepo.OperationCreated, postprocessrepo.EntityStatus, payload, "", "")
	if err != nil {
		return err
	}
	return s.AddStatusSource(value, sourceID)
}

// AuditPostprocessAction records one action observation.
func (s *PostprocessService) AuditPostprocessAction(name string, operation postprocessrepo.Operation, value repo.IEAction) error {
	payload, err := actionPayload(value)
	if err != nil {
		return err
	}
	_, err = s.RecordPostprocess(name, operation, postprocessrepo.EntityAction, payload, "", "")
	return err
}

// AuditPostprocessActionCreated records an action creation and attaches the
// postprocess source to it.
func (s *PostprocessService) AuditPostprocessActionCreated(name string, value repo.IEAction) error {
	payload, err := actionPayload(value)
	if err != nil {
		return err
	}
	sourceID, err := s.RecordPostprocess(name, postprocessrepo.OperationCreated, postprocessrepo.EntityAction, payload, "", "")
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
	statusValue, err := statusPayload(status)
	if err != nil {
		return err
	}
	actionValue, err := actionPayload(action)
	if err != nil {
		return err
	}
	_, err = s.RecordPostprocess(
		name,
		operation,
		postprocessrepo.EntityStatus,
		statusValue,
		postprocessrepo.EntityAction,
		actionValue,
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
	actionValue, err := actionPayload(action)
	if err != nil {
		return err
	}
	statusValue, err := statusPayload(status)
	if err != nil {
		return err
	}
	_, err = s.RecordPostprocess(
		name,
		operation,
		postprocessrepo.EntityAction,
		actionValue,
		postprocessrepo.EntityStatus,
		statusValue,
	)
	return err
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

func statusPayload(value repo.IEStatus) (string, error) {
	encoded, err := json.Marshal(llmparser.StatusToDTO(value))
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func actionPayload(value repo.IEAction) (string, error) {
	encoded, err := json.Marshal(llmparser.ActionToDTO(value))
	if err != nil {
		return "", err
	}
	return string(encoded), nil
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
