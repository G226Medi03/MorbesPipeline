package orchestration

import (
	"context"

	morbesllm "github.com/G226Medi03/MorbesLLMService"
	solve "github.com/G226Medi03/MorbesPipeline/internal/entity/solve"
	"github.com/G226Medi03/MorbesPipeline/internal/entity/solve/input"
	pllm "github.com/G226Medi03/MorbesPipeline/internal/service/llm"
	source "github.com/G226Medi03/SourceEntity"
)

func extraction_refered_source(state *pipelineState) source.SourceId {
	if state.ExplanationSourceID.ToI64() > 0 {
		return state.ExplanationSourceID
	}
	return state.Case.SourceID
}

func (self *Service) runLLMExtraction(ctx context.Context, id solve.CaseID, state *pipelineState) error {
	request := morbesllm.ExtractionRequest{
		CaseID:                     string(id),
		Problem:                    state.Case.Problem,
		ProblemStem:                input.Stem(state.Case.Problem),
		Answer:                     state.Case.Answer,
		Explanation:                state.Explanation,
		SourceID:                   extraction_refered_source(state),
		PatientDescriptionEligible: looks_like_patient_description(state.Case.Problem),
		CacheMode:                  self.cache_mode,
	}
	result, questionType, err := pllm.Run(ctx, self.llmExtraction, request)
	if err != nil {
		return err
	}
	state.QuestionType = questionType
	state.Extraction = result.Extraction
	return self.store.SetQuestionType(id, string(questionType))
}
