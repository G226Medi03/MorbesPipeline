package csolve

import (
	morbesllm "github.com/G226Medi03/MorbesLLMService"
	"github.com/G226Medi03/MorbesPipeline/internal/persistence"
	cllmcall "github.com/G226Medi03/llm-client/call"
)

type Dependencies struct {
	Store *store.Store
	LLM   *morbesllm.Service

	Solver       *cllmcall.Client
	ExtractLLM   *cllmcall.Client
	SituationLLM *cllmcall.Client

	QuestionStore QuestionStore
}
