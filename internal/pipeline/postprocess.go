package csolve

import (
	"context"
	"errors"
	"fmt"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	entitykey "github.com/G226Medi03/MorbesPipeline/internal/domain/entitykey"
	"github.com/G226Medi03/MorbesPipeline/internal/store"
	"github.com/G226Medi03/MorbesPipeline/internal/store/postprocess"
	source "github.com/G226Medi03/SourceEntity"
)

const RejectionStatusName = "급성면역반응"

// Postprocess applies the deterministic postprocess rules to the statuses and
// actions owned by the current case.
func Postprocess(ctx context.Context, st *store.Store, state *pipelineState) error {
	if state == nil {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	entityStore, err := st.Entities()
	if err != nil {
		return err
	}
	statuses := currentStatuses(state)
	actions := currentActions(state)
	rules := []func(*store.Store, *repo.Store, *pipelineState, map[string]repo.IEStatus, map[string]repo.IEAction) error{
		transplantRejectionRule,
		increasedIndexActionRule,
		causeStatusResultRule,
	}
	for _, rule := range rules {
		if err := rule(st, entityStore, state, statuses, actions); err != nil {
			return err
		}
	}
	return nil
}

func transplantRejectionRule(
	st *store.Store,
	entityStore *repo.Store,
	state *pipelineState,
	statuses map[string]repo.IEStatus,
	_ map[string]repo.IEAction,
) error {
	transplantKeys := make([]string, 0)
	for key, status := range statuses {
		received, ok := status.(repo.ReceivedMedicalProcedure)
		if !ok {
			continue
		}
		if _, ok := received.MedicalProcedure.(repo.Transplant); !ok {
			continue
		}
		transplantKeys = append(transplantKeys, key)
	}
	if len(transplantKeys) == 0 {
		return nil
	}

	sourceID := source.SourceId{}
	for _, key := range transplantKeys {
		if value := sourceForStatus(state, key); value.ToI64() > 0 {
			sourceID = value
			break
		}
	}
	if sourceID.ToI64() <= 0 {
		return fmt.Errorf("transplant status source is missing")
	}

	effect := repo.HaveSymptom{
		Symptom: repo.EOtherSymptom{Name: RejectionStatusName},
	}
	for _, key := range transplantKeys {
		if err := st.AuditPostprocessStatus(postprocess.PostprocessorTransplantRejection, postprocess.OperationObserved, statuses[key]); err != nil {
			return err
		}
	}
	if _, err := ensureStatus(st, entityStore, postprocess.PostprocessorTransplantRejection, effect, sourceID); err != nil {
		return err
	}
	action := repo.CauseStatus{Status: effect}
	if _, err := ensureAction(st, entityStore, postprocess.PostprocessorTransplantRejection, action, sourceID); err != nil {
		return err
	}
	for _, key := range transplantKeys {
		status := statuses[key]
		if err := st.LinkStatusActionValues(status, action, sourceID); err != nil {
			return err
		}
		if err := st.AuditPostprocessStatusAction(
			postprocess.PostprocessorTransplantRejection,
			postprocess.OperationLinkAdded,
			status,
			action,
		); err != nil {
			return err
		}
		if err := st.LinkActionResultStatusValues(action, effect, sourceID); err != nil {
			return err
		}
		if err := st.AuditPostprocessActionResultStatus(
			postprocess.PostprocessorTransplantRejection,
			postprocess.OperationLinkAdded,
			action,
			effect,
		); err != nil {
			return err
		}
	}
	return nil
}

// increasedIndexActionRule ensures every increased_index status is produced by
// exactly one canonical increase_index action. Any other action -> status edge
// is removed, and predecessor statuses are rewired through the canonical action.
func increasedIndexActionRule(
	st *store.Store,
	entityStore *repo.Store,
	state *pipelineState,
	statuses map[string]repo.IEStatus,
	_ map[string]repo.IEAction,
) error {
	for key, status := range statuses {
		if status.Kind() != "increased_index" {
			continue
		}
		if err := st.AuditPostprocessStatus(postprocess.PostprocessorIncreasedIndexAction, postprocess.OperationObserved, status); err != nil {
			return err
		}
		action := repo.CauseStatus{Status: status}
		actionKey := entitykey.Action(action)
		sourceID := sourceForStatus(state, key)
		if _, err := ensureAction(st, entityStore, postprocess.PostprocessorIncreasedIndexAction, action, sourceID); err != nil {
			return err
		}

		// An increased_index status may only use the canonical increase_index
		// action as its outgoing action.
		outgoing, err := entityStore.StatusActions(status)
		if err != nil {
			return err
		}
		for _, link := range outgoing {
			other := link.Action.Name()
			if entitykey.Action(other) == actionKey {
				continue
			}
			if err := st.DeleteStatusActionValues(status, other); err != nil {
				return err
			}
			if err := st.AuditPostprocessStatusAction(
				postprocess.PostprocessorIncreasedIndexAction,
				postprocess.OperationLinkRemoved,
				status,
				other,
			); err != nil {
				return err
			}
		}

		// Remove every other action that still points at this index status and
		// preserve the predecessor links by moving them to the canonical action.
		incoming, err := entityStore.ActionsResultingInStatus(status)
		if err != nil {
			return err
		}
		for _, link := range incoming {
			other := link.Action
			if entitykey.Action(other) == actionKey {
				continue
			}
			predecessors, err := entityStore.StatusesCausingAction(other)
			if err != nil {
				return err
			}
			for _, predecessor := range predecessors {
				pred := predecessor.Status
				linkSource := predecessor.SourceID
				if linkSource.ToI64() <= 0 {
					linkSource = sourceID
				}
				if err := st.LinkStatusActionValues(pred, action, linkSource); err != nil {
					return err
				}
				if err := st.AuditPostprocessStatusAction(
					postprocess.PostprocessorIncreasedIndexAction,
					postprocess.OperationLinkAdded,
					pred,
					action,
				); err != nil {
					return err
				}
				if entitykey.Status(pred) != entitykey.Status(status) {
					if err := st.DeleteStatusActionValues(pred, other); err != nil {
						return err
					}
					if err := st.AuditPostprocessStatusAction(
						postprocess.PostprocessorIncreasedIndexAction,
						postprocess.OperationLinkRemoved,
						pred,
						other,
					); err != nil {
						return err
					}
				}
			}
			if err := st.DeleteActionResultStatusValues(other, status); err != nil {
				return err
			}
			if err := st.AuditPostprocessActionResultStatus(
				postprocess.PostprocessorIncreasedIndexAction,
				postprocess.OperationLinkRemoved,
				other,
				status,
			); err != nil {
				return err
			}
		}
		if sourceID.ToI64() <= 0 {
			return fmt.Errorf("increased_index status source is missing: %s", key)
		}
		if err := st.LinkActionResultStatusValues(action, status, sourceID); err != nil {
			return err
		}
		if err := st.AuditPostprocessActionResultStatus(
			postprocess.PostprocessorIncreasedIndexAction,
			postprocess.OperationLinkAdded,
			action,
			status,
		); err != nil {
			return err
		}
	}
	return nil
}

// causeStatusResultRule ensures every cause_status action exposes its embedded
// status as a result status.
func causeStatusResultRule(
	st *store.Store,
	entityStore *repo.Store,
	state *pipelineState,
	statuses map[string]repo.IEStatus,
	actions map[string]repo.IEAction,
) error {
	for key, action := range actions {
		cause, ok := action.(repo.CauseStatus)
		if !ok || cause.Status == nil {
			continue
		}
		statusKey, status, ok := findStatus(statuses, cause.Status)
		if !ok {
			continue
		}
		results, err := entityStore.ActionResultStatuses(action)
		if err != nil {
			return err
		}
		for _, link := range results {
			if entitykey.Status(link.Status) == entitykey.Status(status) {
				continue
			}
			if err := st.DeleteActionResultStatusValues(action, link.Status); err != nil {
				return err
			}
			if err := st.AuditPostprocessActionResultStatus(
				postprocess.PostprocessorCauseStatusResult,
				postprocess.OperationLinkRemoved,
				action,
				link.Status,
			); err != nil {
				return err
			}
		}
		sourceID := sourceForAction(state, key)
		if sourceID.ToI64() <= 0 {
			sourceID = sourceForStatus(state, statusKey)
		}
		if sourceID.ToI64() <= 0 {
			return fmt.Errorf("cause_status action source is missing: %s", key)
		}
		if err := st.LinkActionResultStatusValues(action, status, sourceID); err != nil {
			return err
		}
		if err := st.AuditPostprocessActionResultStatus(
			postprocess.PostprocessorCauseStatusResult,
			postprocess.OperationLinkAdded,
			action,
			status,
		); err != nil {
			return err
		}
	}
	return nil
}

func currentStatuses(state *pipelineState) map[string]repo.IEStatus {
	if len(state.Statuses) > 0 {
		return state.Statuses
	}
	return state.Resolved.Statuses
}

func currentActions(state *pipelineState) map[string]repo.IEAction {
	if len(state.Actions) > 0 {
		return state.Actions
	}
	return state.Resolved.Actions
}

func sourceForStatus(state *pipelineState, key string) source.SourceId {
	for _, draft := range state.Extraction.Statuses {
		if draft.Key == key && draft.SourceID.ToI64() > 0 {
			return draft.SourceID
		}
	}
	for _, draft := range state.Extraction.Statuses {
		if draft.SourceID.ToI64() > 0 {
			return draft.SourceID
		}
	}
	if state.RelationSource.SourceID.ToI64() > 0 {
		return state.RelationSource.SourceID
	}
	return state.ExplanationSourceID
}

func sourceForAction(state *pipelineState, key string) source.SourceId {
	for _, draft := range state.Extraction.Actions {
		if draft.Key == key && draft.SourceID.ToI64() > 0 {
			return draft.SourceID
		}
	}
	for _, draft := range state.Extraction.Actions {
		if draft.SourceID.ToI64() > 0 {
			return draft.SourceID
		}
	}
	if state.RelationSource.SourceID.ToI64() > 0 {
		return state.RelationSource.SourceID
	}
	return state.ExplanationSourceID
}

func findStatus(statuses map[string]repo.IEStatus, target repo.IEStatus) (string, repo.IEStatus, bool) {
	expected := entitykey.Status(target)
	for key, status := range statuses {
		if status.Kind() != target.Kind() {
			continue
		}
		if entitykey.Status(status) == expected {
			return key, status, true
		}
	}
	return "", nil, false
}

func ensureStatus(
	st *store.Store,
	entityStore *repo.Store,
	postprocessorName string,
	value repo.IEStatus,
	sourceID source.SourceId,
) (bool, error) {
	if sourceID.ToI64() <= 0 {
		return false, fmt.Errorf("status source is required: kind=%s", value.Kind())
	}
	if _, err := entityStore.GetStatus(value); err == nil {
		return false, nil
	} else if !errors.Is(err, repo.ErrNotFound) {
		return false, err
	}
	if _, err := entityStore.InsertStatus(value, []source.SourceId{sourceID}); err != nil {
		return false, err
	}
	if err := st.AuditPostprocessStatusCreated(postprocessorName, value); err != nil {
		return false, err
	}
	return true, nil
}

func ensureAction(
	st *store.Store,
	entityStore *repo.Store,
	postprocessorName string,
	value repo.IEAction,
	sourceID source.SourceId,
) (bool, error) {
	if sourceID.ToI64() <= 0 {
		return false, fmt.Errorf("action source is required: kind=%s", value.Kind())
	}
	if _, err := entityStore.GetAction(value); err == nil {
		return false, nil
	} else if !errors.Is(err, repo.ErrNotFound) {
		return false, err
	}
	if _, err := entityStore.InsertAction(value, nil, []source.SourceId{sourceID}); err != nil {
		return false, err
	}
	if err := st.AuditPostprocessActionCreated(postprocessorName, value); err != nil {
		return false, err
	}
	return true, nil
}
