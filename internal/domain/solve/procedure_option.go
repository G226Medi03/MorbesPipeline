package domain

import (
	"github.com/G226Medi03/MedicalEntity/ename/emedical_procedure"
	"github.com/G226Medi03/MedicalEntity/ename/estatus"
	"github.com/G226Medi03/MorbesPipeline/internal/domain/trait"
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
	Procedure           emedical_procedure.IMedicalProcedure `json:"procedure"`
	Category            string                               `json:"category"`
	ProcedureTraits     []trait.Trait                        `json:"procedureTraits,omitempty"`
	ContextTraits       []trait.Trait                        `json:"contextTraits,omitempty"`
	PriorityTraits      []trait.Trait                        `json:"priorityTraits,omitempty"`
	DescriptiveTraits   []trait.Trait                        `json:"descriptiveTraits,omitempty"`
	TreatOptionPriority int                                  `json:"treatOptionPriority"`
	PriorityReason      string                               `json:"priorityReason,omitempty"`
	ProcedureLabel      string                               `json:"procedureLabel"`
	Action              string                               `json:"action,omitempty"`
	Indications         []MedicalProcedureOptionIndication   `json:"indications"`
	ExpectedEffects     []MedicalProcedureOptionIndication   `json:"expectedEffects,omitempty"`
	Outcomes            []MedicalProcedureOptionOutcome      `json:"outcomes,omitempty"`
	Sources             []source.Source                      `json:"sources"`
}

type MedicalProcedureOptionOutcome struct {
	Status estatus.IEStatus `json:"status"`
	Label  string           `json:"label"`
	Kind   string           `json:"kind,omitempty"`
}

type MedicalProcedureOptionIndication struct {
	Status   estatus.IEStatus `json:"status"`
	Label    string           `json:"label"`
	Observed bool             `json:"observed"`
	Cause    bool             `json:"cause,omitempty"`
}
