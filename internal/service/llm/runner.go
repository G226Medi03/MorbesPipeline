package llm

import (
	"context"

	morbesllm "github.com/G226Medi03/MorbesLLMService"
	domain "github.com/G226Medi03/MorbesPipeline/internal/service/solve"
)

func Run(ctx context.Context, runner *morbesllm.ExtractionRunner, request morbesllm.ExtractionRequest) (morbesllm.ExtractionResult, domain.QuestionType, error) {
	result, err := runner.Run(ctx, request)
	if err != nil {
		return result, "", err
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
	return result, questionType, nil
}
