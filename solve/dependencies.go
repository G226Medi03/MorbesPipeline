package csolve

import (
	"github.com/G226Medi03/MorbesPipeline/canonical"
	"github.com/G226Medi03/MorbesPipeline/ports"
	cllmcall "github.com/G226Medi03/llm-client/call"
)

type Dependencies struct {
	Store         ports.Store
	Canonical     *canonical.Canonical
	Solver        *cllmcall.Client
	ExtractLLM    *cllmcall.Client
	SituationLLM  *cllmcall.Client
	QuestionStore QuestionStore
}
