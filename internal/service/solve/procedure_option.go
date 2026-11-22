package domain

import (
	source "github.com/G226Medi03/SourceEntity"
)

// MedicalProcedureOptionGraph is the candidate graph used for procedure
// choice questions.
type MedicalProcedureOptionGraph struct {
	Options                         []MedicalProcedureOptionGraphItem `json:"options"`
	SelectedOptionID                int64                             `json:"selectedOptionId,omitempty"`
	Selected                        *MedicalProcedureOptionGraphItem  `json:"selected,omitempty"`
	AnswerInMedicalProcedureOptions bool                              `json:"answerInMedicalProcedureOptions"`
}

type MedicalProcedureOptionGraphItem struct {
	ID              int64                              `json:"id"`
	ProcedureID     int64                              `json:"procedureId"`
	Category        string                             `json:"category"`
	ProcedureLabel  string                             `json:"procedureLabel"`
	Action          string                             `json:"action,omitempty"`
	Indications     []MedicalProcedureOptionIndication `json:"indications"`
	ExpectedEffects []MedicalProcedureOptionIndication `json:"expectedEffects,omitempty"`
	Outcomes        []MedicalProcedureOptionOutcome    `json:"outcomes,omitempty"`
	Sources         []source.Source                    `json:"sources"`
}

type MedicalProcedureOptionOutcome struct {
	StatusID int64  `json:"statusId"`
	Label    string `json:"label"`
	Kind     string `json:"kind,omitempty"`
}

type MedicalProcedureOptionIndication struct {
	StatusID int64  `json:"statusId"`
	Label    string `json:"label"`
	Observed bool   `json:"observed"`
	Cause    bool   `json:"cause,omitempty"`
}
