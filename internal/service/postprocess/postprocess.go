package postprocess

import (
	"context"
	"errors"
	"fmt"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	store "github.com/G226Medi03/MorbesPipeline/internal/service/store"
	source "github.com/G226Medi03/SourceEntity"
)

const RejectionStatusName = "급성면역반응"

// Postprocess applies the deterministic postprocess rules to the statuses and
// actions owned by the current case.
func Run(ctx context.Context, st *store.StoreContainer, state State) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	entityStore, err := st.MedicalEntities()
	if err != nil {
		return err
	}
	rules := []func(*store.StoreContainer, *repo.Store, State, []repo.IEStatus, []repo.IEAction) error{
		transplantRejectionRule,
		increasedIndexActionRule,
		causeStatusResultRule,
	}
	for _, rule := range rules {
		if err := rule(st, entityStore, state, state.Statuses, state.Actions); err != nil {
			return err
		}
	}
	return nil
}

func transplantRejectionRule(
	st *store.StoreContainer,
	entityStore *repo.Store,
	state State,
	statuses []repo.IEStatus,
	_ []repo.IEAction,
) error {
	transplantIndexes := make([]int, 0)
	for index, status := range statuses {
		received, ok := status.(repo.ReceivedMedicalProcedure)
		if !ok {
			continue
		}
		if _, ok := received.MedicalProcedure.(repo.Transplant); !ok {
			continue
		}
		transplantIndexes = append(transplantIndexes, index)
	}
	if len(transplantIndexes) == 0 {
		return nil
	}

	sourceID := source.SourceId{}
	for _, index := range transplantIndexes {
		if value := sourceForStatus(state, index); value.ToI64() > 0 {
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
	for _, index := range transplantIndexes {
		if err := st.AuditPostprocessStatus(store.PostprocessorTransplantRejection, store.OperationObserved, statuses[index]); err != nil {
			return err
		}
	}
	if _, err := ensureStatus(st, entityStore, store.PostprocessorTransplantRejection, effect, sourceID); err != nil {
		return err
	}
	action := repo.CauseStatus{Status: effect}
	if _, err := ensureAction(st, entityStore, store.PostprocessorTransplantRejection, action, sourceID); err != nil {
		return err
	}
	for _, index := range transplantIndexes {
		status := statuses[index]
		if err := st.LinkStatusActionValues(status, action, sourceID); err != nil {
			return err
		}
		if err := st.AuditPostprocessStatusAction(
			store.PostprocessorTransplantRejection,
			store.OperationLinkAdded,
			status,
			action,
		); err != nil {
			return err
		}
		if err := st.LinkActionResultStatusValues(action, effect, sourceID); err != nil {
			return err
		}
		if err := st.AuditPostprocessActionResultStatus(
			store.PostprocessorTransplantRejection,
			store.OperationLinkAdded,
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
	st *store.StoreContainer,
	entityStore *repo.Store,
	state State,
	statuses []repo.IEStatus,
	_ []repo.IEAction,
) error {
	for index, status := range statuses {
		if status.Kind() != "increased_index" {
			continue
		}
		if err := st.AuditPostprocessStatus(store.PostprocessorIncreasedIndexAction, store.OperationObserved, status); err != nil {
			return err
		}
		action := repo.CauseStatus{Status: status}
		sourceID := sourceForStatus(state, index)
		if _, err := ensureAction(st, entityStore, store.PostprocessorIncreasedIndexAction, action, sourceID); err != nil {
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
			if equalAction(other, action) {
				continue
			}
			if err := st.DeleteStatusActionValues(status, other); err != nil {
				return err
			}
			if err := st.AuditPostprocessStatusAction(
				store.PostprocessorIncreasedIndexAction,
				store.OperationLinkRemoved,
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
			if equalAction(other, action) {
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
					store.PostprocessorIncreasedIndexAction,
					store.OperationLinkAdded,
					pred,
					action,
				); err != nil {
					return err
				}
				if !equalStatus(pred, status) {
					if err := st.DeleteStatusActionValues(pred, other); err != nil {
						return err
					}
					if err := st.AuditPostprocessStatusAction(
						store.PostprocessorIncreasedIndexAction,
						store.OperationLinkRemoved,
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
				store.PostprocessorIncreasedIndexAction,
				store.OperationLinkRemoved,
				other,
				status,
			); err != nil {
				return err
			}
		}
		if sourceID.ToI64() <= 0 {
			return fmt.Errorf("increased_index status source is missing: %d", index)
		}
		if err := st.LinkActionResultStatusValues(action, status, sourceID); err != nil {
			return err
		}
		if err := st.AuditPostprocessActionResultStatus(
			store.PostprocessorIncreasedIndexAction,
			store.OperationLinkAdded,
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
	st *store.StoreContainer,
	entityStore *repo.Store,
	state State,
	statuses []repo.IEStatus,
	actions []repo.IEAction,
) error {
	for actionIndex, action := range actions {
		cause, ok := action.(repo.CauseStatus)
		if !ok || cause.Status == nil {
			continue
		}
		statusIndex, status, ok := findStatus(statuses, cause.Status)
		if !ok {
			continue
		}
		results, err := entityStore.ActionResultStatuses(action)
		if err != nil {
			return err
		}
		for _, link := range results {
			if equalStatus(link.Status, status) {
				continue
			}
			if err := st.DeleteActionResultStatusValues(action, link.Status); err != nil {
				return err
			}
			if err := st.AuditPostprocessActionResultStatus(
				store.PostprocessorCauseStatusResult,
				store.OperationLinkRemoved,
				action,
				link.Status,
			); err != nil {
				return err
			}
		}
		sourceID := sourceForAction(state, actionIndex)
		if sourceID.ToI64() <= 0 {
			sourceID = sourceForStatus(state, statusIndex)
		}
		if sourceID.ToI64() <= 0 {
			return fmt.Errorf("cause_status action source is missing: %d", actionIndex)
		}
		if err := st.LinkActionResultStatusValues(action, status, sourceID); err != nil {
			return err
		}
		if err := st.AuditPostprocessActionResultStatus(
			store.PostprocessorCauseStatusResult,
			store.OperationLinkAdded,
			action,
			status,
		); err != nil {
			return err
		}
	}
	return nil
}

// State is the typed postprocess input owned by the current case.
type State struct {
	Statuses            []repo.IEStatus
	Actions             []repo.IEAction
	StatusSources       []source.SourceId
	ActionSources       []source.SourceId
	FallbackSource      source.SourceId
	RelationSource      source.SourceId
	ExplanationSourceID source.SourceId
}

func sourceForStatus(state State, index int) source.SourceId {
	if index >= 0 && index < len(state.StatusSources) {
		if value := state.StatusSources[index]; value.ToI64() > 0 {
			return value
		}
	}
	return sourceFallback(state)
}

func sourceForAction(state State, index int) source.SourceId {
	if index >= 0 && index < len(state.ActionSources) {
		if value := state.ActionSources[index]; value.ToI64() > 0 {
			return value
		}
	}
	return sourceFallback(state)
}

func sourceFallback(state State) source.SourceId {
	if state.FallbackSource.ToI64() > 0 {
		return state.FallbackSource
	}
	if state.RelationSource.ToI64() > 0 {
		return state.RelationSource
	}
	return state.ExplanationSourceID
}

func findStatus(statuses []repo.IEStatus, target repo.IEStatus) (int, repo.IEStatus, bool) {
	for index, status := range statuses {
		if status.Kind() != target.Kind() {
			continue
		}
		if equalStatus(status, target) {
			return index, status, true
		}
	}
	return -1, nil, false
}

func ensureStatus(
	st *store.StoreContainer,
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
	st *store.StoreContainer,
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
