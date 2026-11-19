package csolve

import ccause "github.com/G226Medi03/MorbesPipeline/internal/tracing"

type caseExtractionStepOutput struct {
	QuestionType   string         `json:"questionType"`
	CaseExtraction CaseExtraction `json:"caseExtraction"`
}

type resolveStepOutput struct {
	Resolved ResolvedCaseExtraction `json:"resolved"`
}

type causeStepOutput struct {
	Observed      []string           `json:"observed"`
	Candidates    []ccause.Candidate `json:"candidates"`
	Missing       []string           `json:"missing"`
	CoverageAdded []string           `json:"coverageAdded,omitempty"`
}

type explainStepOutput struct {
	Explanation string `json:"explanation"`
}
