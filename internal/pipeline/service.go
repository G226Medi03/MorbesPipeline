package csolve

import (
	"context"
	"errors"
	"fmt"

	morbesllm "github.com/G226Medi03/MorbesLLMService"
	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
	"github.com/G226Medi03/MorbesPipeline/internal/persistence"
	core "github.com/G226Medi03/PipeLineCore"
	sourceentity "github.com/G226Medi03/SourceEntity"
	cllmcall "github.com/G226Medi03/llm-client/call"
)

type Service struct {
	ctx            context.Context
	store          *store.Store
	llm            *morbesllm.Service
	question_store QuestionStore

	replay_live        *cllmcall.Client
	replay_live_stages map[string]bool
	cache_mode         core.CacheMode
}

func (self *Service) SetReplayLive(live *cllmcall.Client, stages map[string]bool) {
	self.replay_live = live
	self.replay_live_stages = stages
}

func New_Service(ctx context.Context, deps Dependencies) (*Service, error) {
	llmService := deps.LLM
	if llmService == nil && (deps.Solver != nil || deps.ExtractLLM != nil || deps.SituationLLM != nil) {
		if deps.Store == nil {
			return nil, errors.New("store is required to build the LLM service")
		}
		extract := deps.ExtractLLM
		if extract == nil {
			extract = deps.Solver
		}
		if extract == nil {
			extract = deps.SituationLLM
		}
		created, err := morbesllm.New(
			deps.Solver,
			extract,
			deps.Store,
			deps.Store,
			llmLogger{},
		)
		if err != nil {
			return nil, err
		}
		llmService = created
	}
	return &Service{
		ctx:            ctx,
		store:          deps.Store,
		llm:            llmService,
		question_store: deps.QuestionStore,
	}, nil
}

func (self *Service) set_stage(id domain.CaseID, stage string) error {
	if err := self.store.SetStage(id, stage); err != nil {
		_ = self.store.SetFailed(id, err)
		return err
	}
	return nil
}

func (self *Service) explain(ctx context.Context, problem, answer string) (*morbesllm.ExplainResult, error) {
	if self.llm == nil {
		return nil, fmt.Errorf("solver LLM 이 없습니다")
	}
	return self.llm.ExplainResult(ctx, problem, answer, sourceentity.SourceId{}, self.cache_mode)
}
