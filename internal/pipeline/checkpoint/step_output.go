package checkpoint

import morbesllm "github.com/G226Medi03/MorbesLLMService"

type CaseExtractionStepOutput struct {
	QuestionType   string                   `json:"questionType"`
	CaseExtraction morbesllm.CaseExtraction `json:"caseExtraction"`
}

type ResolveStepOutput struct {
	Resolved ResolvedCaseExtraction `json:"resolved"`
}

type ExplainStepOutput struct {
	Explanation string `json:"explanation"`
}
