package csolve

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/G226Medi03/MorbesPipeline/canonical"
	"github.com/G226Medi03/MorbesPipeline/ports"
	cservice "github.com/G226Medi03/MorbesPipeline/service"
	"github.com/G226Medi03/MorbesPipeline/solve/domain"
	cllmcall "github.com/G226Medi03/llm-client/call"
)

type Service struct {
	ctx            context.Context
	store          ports.Store
	canonical      *canonical.Canonical
	solver         *cllmcall.Client
	extract_llm    *cllmcall.Client
	situation_llm  *cllmcall.Client
	question_store QuestionStore

	replay_live        *cllmcall.Client
	replay_live_stages map[string]bool
	cache_mode         cservice.LLMCacheMode
}

func (self *Service) SetReplayLive(live *cllmcall.Client, stages map[string]bool) {
	self.replay_live = live
	self.replay_live_stages = stages
}

func New_Service(ctx context.Context, deps Dependencies) *Service {
	return &Service{
		ctx:            ctx,
		store:          deps.Store,
		canonical:      deps.Canonical,
		solver:         deps.Solver,
		extract_llm:    deps.ExtractLLM,
		situation_llm:  deps.SituationLLM,
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
	if self.solver == nil {
		return "", "", "", fmt.Errorf("solver LLM 이 없습니다")
	}
	if strings.TrimSpace(template) == "" {
		return "", "", "", fmt.Errorf("explain prompt 가 DB에 없습니다")
	}
	prompt := RenderPrompt(template, map[string]string{
		"problem": problem,
		"answer":  answer,
	})

	raw, err := self.solver.Generate(ctx, prompt)
	if err != nil {
		return "", prompt, raw, err
	}

	var parsed struct {
		Answer      string `json:"answer"`
		Explanation string `json:"explanation"`
	}
	if err := decode_json_object(raw, &parsed); err == nil && strings.TrimSpace(parsed.Explanation) != "" {
		return strings.TrimSpace(parsed.Explanation), prompt, raw, nil
	}
	// Solver may return prose despite the JSON instruction.
	if strings.TrimSpace(raw) == "" {
		return "", prompt, raw, fmt.Errorf("solver explanation 이 비어 있습니다")
	}
	return strings.TrimSpace(raw), prompt, raw, nil
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
