package domain

import (
	entity "github.com/G226Medi03/MorbesPipeline/legacy"
	"github.com/G226Medi03/MorbesPipeline/trait"
	source "github.com/G226Medi03/SourceEntity"
)

// MedicalProcedureOptionGraph is the candidate graph used for procedure
// choice questions.
type MedicalProcedureOptionGraph struct {
	Options                         []MedicalProcedureOptionGraphItem `json:"options"`
	SelectedOptionID                entity.MedicalProcedureOptionID   `json:"selectedOptionId,omitempty"`
	Selected                        *MedicalProcedureOptionGraphItem  `json:"selected,omitempty"`
	AnswerInMedicalProcedureOptions bool                              `json:"answerInMedicalProcedureOptions"`
}

type MedicalProcedureOptionGraphItem struct {
	ID                  entity.MedicalProcedureOptionID    `json:"id"`
	ProcedureID         entity.ProcedureID                 `json:"procedureId"`
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
	StatusID entity.StatusID `json:"statusId"`
	Label    string          `json:"label"`
	Kind     string          `json:"kind,omitempty"`
}

type MedicalProcedureOptionIndication struct {
	StatusID entity.StatusID `json:"statusId"`
	Label    string          `json:"label"`
	Observed bool            `json:"observed"`
	Cause    bool            `json:"cause,omitempty"`
}
