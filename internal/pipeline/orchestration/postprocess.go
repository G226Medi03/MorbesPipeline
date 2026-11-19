package csolve

import (
	"context"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"github.com/G226Medi03/MorbesPipeline/internal/pipeline/postprocess"
	store "github.com/G226Medi03/MorbesPipeline/internal/store/facade"
	source "github.com/G226Medi03/SourceEntity"
)

const RejectionStatusName = postprocess.RejectionStatusName

// Postprocess applies deterministic postprocess rules to the current case.
func Postprocess(ctx context.Context, st *store.StoreContainer, state *pipelineState) error {
	if state == nil {
		return nil
	}
	return postprocess.Run(ctx, st, buildPostprocessState(state))
}

func buildPostprocessState(state *pipelineState) postprocess.State {
	result := postprocess.State{
		Statuses:            currentStatuses(state),
		Actions:             currentActions(state),
		StatusSources:       make(map[string]source.SourceId, len(state.Extraction.Statuses)),
		ActionSources:       make(map[string]source.SourceId, len(state.Extraction.Actions)),
		RelationSource:      state.RelationSource.SourceID,
		ExplanationSourceID: state.ExplanationSourceID,
	}
	for _, draft := range state.Extraction.Statuses {
		if draft.SourceID.ToI64() > 0 {
			result.StatusSources[draft.Key] = draft.SourceID
			if result.FallbackSource.ToI64() <= 0 {
				result.FallbackSource = draft.SourceID
			}
		}
	}
	for _, draft := range state.Extraction.Actions {
		if draft.SourceID.ToI64() > 0 {
			result.ActionSources[draft.Key] = draft.SourceID
			if result.FallbackSource.ToI64() <= 0 {
				result.FallbackSource = draft.SourceID
			}
		}
	}
	return result
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
