package morbes_pipeline

import (
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"

	"github.com/G226Medi03/MorbesPipeline/internal/domain/patient"
	procedureoption "github.com/G226Medi03/MorbesPipeline/internal/domain/procedureoption"
	domain "github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
	"github.com/G226Medi03/MorbesPipeline/internal/domain/trait"
	"github.com/G226Medi03/MorbesPipeline/internal/support/ids"
)

type (
	IEAction          = repo.IEAction
	IEStatus          = repo.IEStatus
	IMedicalProcedure = repo.IMedicalProcedure

	Cleaning                = repo.Cleaning
	DiagnosticTest          = repo.DiagnosticTest
	Education               = repo.Education
	EOtherSymptom           = repo.EOtherSymptom
	HaveSymptom             = repo.HaveSymptom
	Infuse                  = repo.Infuse
	Instruction             = repo.Instruction
	Surgery                 = repo.Surgery
	Transplant              = repo.Transplant
	UnknownActor            = repo.UnknownActor
	UnknownMedicalProcedure = repo.UnknownMedicalProcedure
)

type (
	CaseID        = domain.CaseID
	QuestionType  = domain.QuestionType
	QuestionInput = domain.QuestionInput
	QuestionStore = domain.QuestionStore
	AnswerCheck   = domain.AnswerCheck

	MedicalProcedureOption   = procedureoption.MedicalProcedureOption
	MedicalProcedureOptionID = ids.MedicalProcedureOptionID
	PatientDescription       = patient.PatientDescription
	Trait                    = trait.Trait
)

const (
	QuestionTypeCause        = domain.QuestionTypeCause
	QuestionTypeStatus       = domain.QuestionTypeStatus
	QuestionTypeProcedure    = domain.QuestionTypeProcedure
	QuestionTypePurpose      = domain.QuestionTypePurpose
	QuestionTypeMechanism    = domain.QuestionTypeMechanism
	QuestionTypeDeviceAction = domain.QuestionTypeDeviceAction

	IntentEducation       = domain.IntentEducation
	IntentNursing         = domain.IntentNursing
	IntentStatementChoice = domain.IntentStatementChoice
	IntentEmergencyCare   = domain.IntentEmergencyCare

	TraitAcute      = trait.TraitAcute
	TraitDefinitive = trait.TraitDefinitive
)

var (
	MarshalAction    = repo.MarshalAction
	MarshalProcedure = repo.MarshalProcedure
	MarshalStatus    = repo.MarshalStatus

	InferQuestionPolicy   = domain.InferQuestionPolicy
	NormalizeQuestionType = domain.NormalizeQuestionType
)
