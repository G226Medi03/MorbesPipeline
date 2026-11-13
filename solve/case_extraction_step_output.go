package csolve

type caseExtractionStepOutput struct {
	QuestionType   string         `json:"questionType"`
	CaseExtraction CaseExtraction `json:"caseExtraction"`
}

type resolveStepOutput struct {
	Resolved ResolvedCaseExtraction `json:"resolved"`
}
