package domain

import (
	"github.com/G226Medi03/MorbesPipeline/internal/domain/trait"
	source "github.com/G226Medi03/MorbesPipeline/internal/external/sourceentity"
	"github.com/G226Medi03/MorbesPipeline/internal/support/ids"
)

// MedicalProcedureOptionGraph is the candidate graph used for procedure
// choice questions.
type MedicalProcedureOptionGraph struct {
	Options                         []MedicalProcedureOptionGraphItem `json:"options"`
	SelectedOptionID                ids.MedicalProcedureOptionID      `json:"selectedOptionId,omitempty"`
	Selected                        *MedicalProcedureOptionGraphItem  `json:"selected,omitempty"`
	AnswerInMedicalProcedureOptions bool                              `json:"answerInMedicalProcedureOptions"`
}

type MedicalProcedureOptionGraphItem struct {
	ID                  ids.MedicalProcedureOptionID       `json:"id"`
	ProcedureID         ids.ProcedureID                    `json:"procedureId"`
	Category            string                             `json:"category"`
	ProcedureTraits     []trait.Trait                      `json:"procedureTraits,omitempty"`
	ContextTraits       []trait.Trait                      `json:"contextTraits,omitempty"`
	PriorityTraits      []trait.Trait                      `json:"priorityTraits,omitempty"`
	DescriptiveTraits   []trait.Trait                      `json:"descriptiveTraits,omitempty"`
	TreatOptionPriority int                                `json:"treatOptionPriority"`
	PriorityReason      string                             `json:"priorityReason,omitempty"`
	ProcedureLabel      string                             `json:"procedureLabel"`
	Action              string                             `json:"action,omitempty"`
	Indications         []MedicalProcedureOptionIndication `json:"indications"`
	ExpectedEffects     []MedicalProcedureOptionIndication `json:"expectedEffects,omitempty"`
	Outcomes            []MedicalProcedureOptionOutcome    `json:"outcomes,omitempty"`
	Sources             []source.SourceRef                 `json:"sources"`
}

type MedicalProcedureOptionOutcome struct {
	StatusID ids.StatusID `json:"statusId"`
	Label    string       `json:"label"`
	Kind     string       `json:"kind,omitempty"`
}

type MedicalProcedureOptionIndication struct {
	StatusID ids.StatusID `json:"statusId"`
	Label    string       `json:"label"`
	Observed bool         `json:"observed"`
	Cause    bool         `json:"cause,omitempty"`
}
