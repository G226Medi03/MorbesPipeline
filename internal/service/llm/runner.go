package llm

import (
	"context"

	morbesllm "github.com/G226Medi03/MorbesLLMService"
	solve "github.com/G226Medi03/MorbesPipeline/internal/entity/solve"
)

func Run(ctx context.Context, runner *morbesllm.ExtractionRunner, request morbesllm.ExtractionRequest) (morbesllm.ExtractionResult, solve.QuestionType, error) {
	result, err := runner.Run(ctx, request)
	if err != nil {
		return result, "", err
	}
	questionType := solve.NormalizeQuestionType(result.RawQuestionType)
	policy := solve.InferQuestionPolicy(request.ProblemStem)
	switch {
	case policy.Type == solve.QuestionTypeDeviceAction:
		questionType = policy.Type
	case questionType == solve.QuestionTypeStatus && policy.Type == solve.QuestionTypeProcedure:
		questionType = policy.Type
	case policy.Type == solve.QuestionTypeMechanism && questionType != solve.QuestionTypeMechanism:
		questionType = policy.Type
	case questionType == "" && policy.Type != "":
		questionType = policy.Type
	}
	if questionType == "" {
		questionType = solve.QuestionTypeProcedure
	}
	return result, questionType, nil
}
