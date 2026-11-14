package domain

import "github.com/G226Medi03/MorbesPipeline/internal/support/ids"

type SolveStepMeta struct {
	CaseID        CaseID        `json:"caseId"`
	PromptVersion int           `json:"promptVersion"`
	OutputJSON    string        `json:"outputJson"`
	CallID        ids.LLMCallID `json:"callId"`
}
