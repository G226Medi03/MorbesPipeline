package postprocess

import (
	"context"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"github.com/G226Medi03/MorbesPipeline/internal/service/orchestration/state"
	enginepostprocess "github.com/G226Medi03/MorbesPipeline/internal/service/postprocess"
	store "github.com/G226Medi03/MorbesPipeline/internal/service/store"
	source "github.com/G226Medi03/SourceEntity"
)

const RejectionStatusName = enginepostprocess.RejectionStatusName

// Postprocess applies deterministic postprocess rules to the current case.
func Run(ctx context.Context, st *store.StoreContainer, state *state.State) error {
	if state == nil {
		return nil
	}
	return enginepostprocess.Run(ctx, st, buildPostprocessState(state))
}

func buildPostprocessState(state *state.State) enginepostprocess.State {
	statuses := currentStatuses(state)
	actions := currentActions(state)
	result := enginepostprocess.State{
		Statuses:            statuses,
		Actions:             actions,
		StatusSources:       make([]source.SourceId, len(statuses)),
		ActionSources:       make([]source.SourceId, len(actions)),
		RelationSource:      state.RelationSource.SourceID,
		ExplanationSourceID: state.ExplanationSourceID,
	}
	for index, draft := range state.Extraction.Statuses() {
		if index >= len(result.StatusSources) || draft.SourceID.ToI64() <= 0 {
			continue
		}
		result.StatusSources[index] = draft.SourceID
		if result.FallbackSource.ToI64() <= 0 {
			result.FallbackSource = draft.SourceID
		}
	}
	for index, draft := range state.Extraction.Actions() {
		if index >= len(result.ActionSources) || draft.SourceID.ToI64() <= 0 {
			continue
		}
		result.ActionSources[index] = draft.SourceID
		if result.FallbackSource.ToI64() <= 0 {
			result.FallbackSource = draft.SourceID
		}
	}
	return result
}

func currentStatuses(state *state.State) []repo.IEStatus {
	if len(state.Statuses) > 0 {
		return state.Statuses
	}
	return state.Resolved.Statuses
}

func currentActions(state *state.State) []repo.IEAction {
	if len(state.Actions) > 0 {
		return state.Actions
	}
	return state.Resolved.Actions
}
