package solve

import (
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	source "github.com/G226Medi03/SourceEntity"
)

// MedicalProcedureOptionGraph is the candidate graph used for procedure
// choice questions.
type MedicalProcedureOptionGraph struct {
	Options                         []MedicalProcedureOptionGraphItem `json:"options"`
	SelectedOptionIndex             int                               `json:"selectedOptionIndex,omitempty"`
	Selected                        *MedicalProcedureOptionGraphItem  `json:"selected,omitempty"`
	AnswerInMedicalProcedureOptions bool                              `json:"answerInMedicalProcedureOptions"`
}

type MedicalProcedureOptionGraphItem struct {
	Key             string                             `json:"key,omitempty"`
	Procedure       repo.IMedicalProcedure             `json:"procedure"`
	Category        string                             `json:"category"`
	ProcedureLabel  string                             `json:"procedureLabel"`
	Action          string                             `json:"action,omitempty"`
	Indications     []MedicalProcedureOptionIndication `json:"indications"`
	ExpectedEffects []MedicalProcedureOptionIndication `json:"expectedEffects,omitempty"`
	Outcomes        []MedicalProcedureOptionOutcome    `json:"outcomes,omitempty"`
	Sources         []source.Source                    `json:"sources"`
}

type MedicalProcedureOptionOutcome struct {
	Status repo.IEStatus `json:"status"`
	Label  string        `json:"label"`
	Kind   string        `json:"kind,omitempty"`
}

type MedicalProcedureOptionIndication struct {
	Status   repo.IEStatus `json:"status"`
	Label    string        `json:"label"`
	Observed bool          `json:"observed"`
	Cause    bool          `json:"cause,omitempty"`
}
