package csolve

import (
	"context"

	domain "github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
	morbesllm "github.com/G226Medi03/MorbesLLMService"
	source "github.com/G226Medi03/SourceEntity"
)

func extraction_refered_source(state *pipelineState) source.SourceId {
	if state.ExplanationSourceID.ToI64() > 0 {
		return state.ExplanationSourceID
	}
	return state.Case.SourceID
}

func (self *Service) runLLMExtraction(ctx context.Context, id domain.CaseID, state *pipelineState) error {
	request := morbesllm.ExtractionRequest{
		CaseID:                     string(id),
		Problem:                    state.Case.Problem,
		ProblemStem:                problem_stem(state.Case.Problem),
		Answer:                     state.Case.Answer,
		Explanation:                state.Explanation,
		SourceID:                   extraction_refered_source(state),
		PatientDescriptionEligible: looks_like_patient_description(state.Case.Problem),
		CacheMode:                  self.cache_mode,
	}
	result, err := self.llmExtraction.Run(ctx, request)
	if err != nil {
		return err
	}
	questionType := domain.NormalizeQuestionType(result.RawQuestionType)
	policy := domain.InferQuestionPolicy(request.ProblemStem)
	switch {
	case policy.Type == domain.QuestionTypeDeviceAction:
		questionType = policy.Type
	case questionType == domain.QuestionTypeStatus && policy.Type == domain.QuestionTypeProcedure:
		questionType = policy.Type
	case policy.Type == domain.QuestionTypeMechanism && questionType != domain.QuestionTypeMechanism:
		questionType = policy.Type
	case questionType == "" && policy.Type != "":
		questionType = policy.Type
	}
	if questionType == "" {
		questionType = domain.QuestionTypeProcedure
	}
	state.QuestionType = questionType
	state.Extraction = result.Extraction
	return self.store.SetQuestionType(id, string(questionType))
}
