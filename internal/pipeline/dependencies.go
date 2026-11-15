package csolve

import (
	morbesllm "github.com/G226Medi03/MorbesLLmPipeline"
	"github.com/G226Medi03/MorbesPipeline/internal/persistence"
)

type Dependencies struct {
	Store *store.Store
	LLM   *morbesllm.Service

	Solver        morbesllm.Generator
	ExtractLLM    morbesllm.Generator
	SituationLLM  morbesllm.Generator

	QuestionStore QuestionStore
}
