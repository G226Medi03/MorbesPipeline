package csolve

import (
	"context"
	"errors"

	morbesllm "github.com/G226Medi03/MorbesLLMService"
	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
	"github.com/G226Medi03/MorbesPipeline/internal/persistence"
	core "github.com/G226Medi03/PipeLineCore"
	sourceentity "github.com/G226Medi03/SourceEntity"
	cllmcall "github.com/G226Medi03/llm-client/call"
)

type Dependencies struct {
	Store *store.Store
	LLM   *morbesllm.Service

	Solver     *cllmcall.Client
	ExtractLLM *cllmcall.Client

	QuestionStore QuestionStore
}

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
	if deps.Store == nil {
		return nil, errors.New("store is required")
	}
	if deps.QuestionStore == nil {
		return nil, errors.New("question store is required")
	}
	llmService := deps.LLM
	if llmService == nil {
		if deps.Solver == nil && deps.ExtractLLM == nil {
			return nil, errors.New("LLM service or client is required")
		}
		extract := deps.ExtractLLM
		if extract == nil {
			extract = deps.Solver
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
	return self.llm.ExplainResult(ctx, problem, answer, sourceentity.SourceId{}, self.cache_mode)
}
