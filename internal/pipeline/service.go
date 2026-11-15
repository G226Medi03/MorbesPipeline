package csolve

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	morbesllm "github.com/G226Medi03/MorbesLLmPipeline"
	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
	sourceentity "github.com/G226Medi03/MorbesPipeline/internal/external/sourceentity"
	"github.com/G226Medi03/MorbesPipeline/internal/persistence"
)

type Service struct {
	ctx            context.Context
	store          *store.Store
	llm            *morbesllm.Service
	question_store QuestionStore

	replay_live        morbesllm.Generator
	replay_live_stages map[string]bool
	cache_mode         morbesllm.CacheMode
}

func (self *Service) SetReplayLive(live morbesllm.Generator, stages map[string]bool) {
	self.replay_live = live
	self.replay_live_stages = stages
}

func New_Service(ctx context.Context, deps Dependencies) *Service {
	llmService := deps.LLM
	if llmService == nil && (deps.Solver != nil || deps.ExtractLLM != nil || deps.SituationLLM != nil) {
		var cache morbesllm.Cache
		if deps.Store != nil {
			cache = deps.Store
		}
		llmService = morbesllm.NewWithClients(morbesllm.Clients{
			Solver:    deps.Solver,
			Extract:   deps.ExtractLLM,
			Situation: deps.SituationLLM,
		}, cache, llmLogger{})
	}
	return &Service{
		ctx:            ctx,
		store:          deps.Store,
		llm:            llmService,
		question_store: deps.QuestionStore,
	}
}

func (self *Service) set_stage(id domain.CaseID, stage string) error {
	if err := self.store.SetStage(id, stage); err != nil {
		_ = self.store.SetFailed(id, err)
		return err
	}
	return nil
}

func (self *Service) explain(ctx context.Context, problem, answer, template string) (string, string, string, error) {
	if self.llm == nil {
		return "", "", "", fmt.Errorf("solver LLM 이 없습니다")
	}
	if strings.TrimSpace(template) == "" {
		return "", "", "", fmt.Errorf("explain prompt 가 DB에 없습니다")
	}
	prompt := RenderPrompt(template, map[string]string{
		"problem": problem,
		"answer":  answer,
	})

	result, err := self.llm.ExplainResult(ctx, prompt, sourceentity.SourceId{}, self.cache_mode)
	if err != nil {
		if result != nil {
			return "", prompt, result.Raw, err
		}
		return "", prompt, "", err
	}
	return result.Explanation, prompt, result.Raw, nil
}

func decode_json_object(raw string, target any) error {
	trimmed := strings.TrimSpace(raw)
	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start < 0 || end <= start {
		return fmt.Errorf("JSON 객체를 찾지 못했습니다")
	}
	return json.Unmarshal([]byte(trimmed[start:end+1]), target)
}
