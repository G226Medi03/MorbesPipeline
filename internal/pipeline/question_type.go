package csolve

import (
	"context"
	"fmt"
	"time"

	morbesllm "github.com/G226Medi03/MorbesLLMService"
	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
)

func (self *Service) run_question_type(ctx context.Context, id domain.CaseID, state *pipelineState) domain.QuestionType {
	problem := problem_stem(state.Case.Problem)
	started := time.Now()
	var result morbesllm.QuestionTypeResult
	var call_err error
	if self.llm == nil {
		call_err = fmt.Errorf("extraction LLM is not configured")
	} else if !state.Prompts.present[PromptQuestionType] {
		call_err = fmt.Errorf("question_type prompt가 DB에 없습니다")
	} else {
		result, call_err = self.llm.ExtractQuestionType(
			ctx,
			problem,
			"",
			"",
			extraction_refered_source(state),
			self.cache_mode,
		)
	}
	raw := result.Raw
	response := result.Response
	question_type := domain.QuestionType("")
	if call_err == nil {
		question_type = domain.NormalizeQuestionType(result.QuestionType)
	}
	prompt, _ := response_texts(response)
	version := prompt_version_or(state.Prompts.versions[PromptQuestionType], result.PromptVersion)
	policy := domain.InferQuestionPolicy(problem)
	inferred := policy.Type
	switch {
	case inferred == domain.QuestionTypeDeviceAction:
		question_type = inferred
	case question_type == domain.QuestionTypeStatus && inferred == domain.QuestionTypeProcedure:
		question_type = inferred
	case inferred == domain.QuestionTypeMechanism && question_type != domain.QuestionTypeMechanism:
		question_type = inferred
	case question_type == "" && inferred != "":
		question_type = inferred
	}
	if question_type == "" {
		question_type = domain.QuestionTypeProcedure
	}
	state.QuestionType = question_type
	_ = self.store.SetQuestionType(id, string(question_type))
	self.record_v2_extraction_call(
		id, seqQuestionType, "question_type", "classify",
		version, prompt, raw, response,
		map[string]any{
			"questionType":    question_type,
			"intent":          policy.Intent,
			"polarity":        policy.Polarity,
			"desiredCategory": policy.DesiredCategory,
		}, started, call_err,
	)
	return question_type
}
