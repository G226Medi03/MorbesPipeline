package orchestration

import (
	"context"
	"fmt"

	"github.com/G226Medi03/MorbesPipeline/internal/entity/solve"
	core "github.com/G226Medi03/PipeLineCore"
)

type ReplayMode = solve.ReplayMode

const (
	ReplayModeLLM  = solve.ReplayModeLLM
	ReplayModeCode = solve.ReplayModeCode
)

// PrepareReplay resets the source case so replay overwrites the same case.
func (self *Service) PrepareReplay(_ context.Context, sourceID solve.CaseID, mode ReplayMode, fromNode string) (solve.CaseID, error) {
	if mode != ReplayModeLLM && mode != ReplayModeCode {
		return "", fmt.Errorf("알 수 없는 replay mode 입니다: %s", mode)
	}
	source, found, err := self.store.Get(sourceID)
	if err != nil {
		return "", err
	}
	if !found {
		return "", fmt.Errorf("replay 원본 case 를 찾지 못했습니다: %s", sourceID)
	}
	if source.State == "running" || source.State == "pending" {
		return "", fmt.Errorf("이미 실행 중인 case 는 replay 할 수 없습니다: %s", sourceID)
	}
	keepCalls := mode == ReplayModeCode
	if err := self.store.ResetCaseForReplay(sourceID, string(mode), fromNode, keepCalls); err != nil {
		return "", err
	}
	return sourceID, nil
}

// RunReplay executes the reset case. useCache is ignored for code replay,
// which always bypasses the shared LLM response cache.
func (self *Service) RunReplay(ctx context.Context, sourceID solve.CaseID, mode ReplayMode, fromNode string, useCache bool) error {
	run := *self
	run.ctx = ctx
	if mode == ReplayModeCode {
		calls, err := self.store.ListLLMCalls(sourceID)
		if err != nil {
			_ = self.store.SetFailed(sourceID, err)
			return err
		}
		replayCalls := make([]core.ReplayCall, 0, len(calls))
		for _, call := range calls {
			replayCalls = append(replayCalls, core.ReplayCall{
				Stage:  call.Stage,
				Prompt: call.Prompt,
				Output: call.RawOutput,
			})
		}
		run.llm = self.llm.WithReplayLive(replayCalls, self.replay_live, self.replay_live_stages)
		run.cache_mode = core.CacheBypass
		if item, found, err := self.store.Get(sourceID); err == nil && found && item.SourceID.ToI64() <= 0 && item.SituationSourceID.ToI64() > 0 {
			if err := self.store.SetSource(sourceID, item.SituationSourceID); err != nil {
				return err
			}
		}
		return run.run_v2_from(ctx, sourceID, fromNode)
	}
	run.cache_mode = core.CacheUse
	if !useCache {
		run.cache_mode = core.CacheRefresh
	}
	return run.run_v2_from(ctx, sourceID, fromNode)
}
