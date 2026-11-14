package csolve

import (
	cllmcall "github.com/G226Medi03/MorbesPipeline/internal/external/llmclient/call"
	"github.com/G226Medi03/MorbesPipeline/internal/persistence"
)

type Dependencies struct {
	Store         *store.Store
	Solver        *cllmcall.Client
	ExtractLLM    *cllmcall.Client
	SituationLLM  *cllmcall.Client
	QuestionStore QuestionStore
}
