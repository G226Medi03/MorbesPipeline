package domain

import entity "github.com/G226Medi03/MorbesPipeline/legacy"

type SolveStepMeta struct {
	CaseID        CaseID           `json:"caseId"`
	PromptVersion int              `json:"promptVersion"`
	OutputJSON    string           `json:"outputJson"`
	CallID        entity.LLMCallID `json:"callId"`
}
