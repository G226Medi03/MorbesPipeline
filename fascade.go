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
	IEActor           = repo.IEActor
	IEStatus          = repo.IEStatus
	IMedicalProcedure = repo.IMedicalProcedure
	IESymptom         = repo.IESymptom
	IAspect           = repo.IAspect

	Block                   = repo.Block
	CauseStatus             = repo.CauseStatus
	Relax                   = repo.Relax
	Device                  = repo.Device
	ActorProcedure          = repo.ActorProcedure
	UnknownActor            = repo.UnknownActor
	EDisease                = repo.EDisease
	EIndex                  = repo.EIndex
	Cleaning                = repo.Cleaning
	DiagnosticTest          = repo.DiagnosticTest
	Education               = repo.Education
	Infuse                  = repo.Infuse
	Insertion               = repo.Insertion
	Instruction             = repo.Instruction
	Surgery                 = repo.Surgery
	Transplant              = repo.Transplant
	UnknownMedicalProcedure = repo.UnknownMedicalProcedure
	Damaged                 = repo.Damaged
	DecreasedIndex          = repo.DecreasedIndex
	HaveDisease             = repo.HaveDisease
	HaveSymptom             = repo.HaveSymptom
	IncreasedIndex          = repo.IncreasedIndex
	Inflammation            = repo.Inflammation
	ReceivedProcedure       = repo.ReceivedMedicalProcedure
	EOtherSymptom           = repo.EOtherSymptom
	EPain                   = repo.EPain
)

type (
	CaseID           = domain.CaseID
	CaseState        = domain.CaseState
	Stage            = domain.Stage
	QuestionType     = domain.QuestionType
	QuestionIntent   = domain.QuestionIntent
	QuestionPolarity = domain.QuestionPolarity
	QuestionPolicy   = domain.QuestionPolicy
	QuestionInput    = domain.QuestionInput
	QuestionStore    = domain.QuestionStore
	AnswerCheck      = domain.AnswerCheck

	MedicalProcedureOption   = procedureoption.MedicalProcedureOption
	MedicalProcedureOptionID = ids.MedicalProcedureOptionID
	PatientDescription       = patient.PatientDescription
	Trait                    = trait.Trait
)

const (
	CasePending = domain.CasePending
	CaseRunning = domain.CaseRunning
	CaseDone    = domain.CaseDone
	CaseFailed  = domain.CaseFailed

	StageExplain        = domain.StageExplain
	StageCaseExtraction = domain.StageCaseExtraction
	StageResolve        = domain.StageResolve
	StageCause          = domain.StageCause

	QuestionTypeCause        = domain.QuestionTypeCause
	QuestionTypeStatus       = domain.QuestionTypeStatus
	QuestionTypeProcedure    = domain.QuestionTypeProcedure
	QuestionTypePurpose      = domain.QuestionTypePurpose
	QuestionTypeMechanism    = domain.QuestionTypeMechanism
	QuestionTypeDeviceAction = domain.QuestionTypeDeviceAction

	IntentProcedure       = domain.IntentProcedure
	IntentEducation       = domain.IntentEducation
	IntentNursing         = domain.IntentNursing
	IntentDeviceAction    = domain.IntentDeviceAction
	IntentPurpose         = domain.IntentPurpose
	IntentMechanism       = domain.IntentMechanism
	IntentKnowledge       = domain.IntentKnowledge
	IntentStatementChoice = domain.IntentStatementChoice
	IntentEmergencyCare   = domain.IntentEmergencyCare

	PolarityPositive = domain.PolarityPositive
	PolarityNegative = domain.PolarityNegative

	TraitResuscitative = trait.TraitResuscitative
	TraitAcute         = trait.TraitAcute
	TraitDefinitive    = trait.TraitDefinitive
	TraitFirstLine     = trait.TraitFirstLine
	TraitRescue        = trait.TraitRescue
	TraitInvasive      = trait.TraitInvasive
	TraitNonInvasive   = trait.TraitNonInvasive
	TraitSurgical      = trait.TraitSurgical
	TraitPharmacologic = trait.TraitPharmacologic
	TraitDiagnostic    = trait.TraitDiagnostic
	TraitMonitoring    = trait.TraitMonitoring
	TraitSupportive    = trait.TraitSupportive
	TraitChronic       = trait.TraitChronic
	TraitMaintenance   = trait.TraitMaintenance
	TraitProphylactic  = trait.TraitProphylactic
	TraitEducational   = trait.TraitEducational
	TraitNursing       = trait.TraitNursing
)

var (
	NewMedicalProcedureOption = procedureoption.NewMedicalProcedureOption
	ErrSourceRequired         = procedureoption.ErrSourceRequired
	ParseStatus               = patient.ParseStatus

	MarshalAction    = repo.MarshalAction
	MarshalActor     = repo.MarshalActor
	MarshalProcedure = repo.MarshalProcedure
	MarshalStatus    = repo.MarshalStatus
	MarshalSymptom   = repo.MarshalSymptom

	UnmarshalAction           = repo.UnmarshalAction
	UnmarshalActor            = repo.UnmarshalActor
	UnmarshalOrParseProcedure = repo.UnmarshalOrParseProcedure
	UnmarshalProcedure        = repo.UnmarshalProcedure
	UnmarshalStatus           = repo.UnmarshalStatus
	UnmarshalSymptom          = repo.UnmarshalSymptom

	InferQuestionPolicy           = domain.InferQuestionPolicy
	InferQuestionType             = domain.InferQuestionType
	InferQuestionIntent           = domain.InferQuestionIntent
	InferQuestionPolarity         = domain.InferQuestionPolarity
	InferDesiredProcedureCategory = domain.InferDesiredProcedureCategory
	NormalizeQuestionType         = domain.NormalizeQuestionType
	IsStage                       = domain.IsStage
	Stages                        = domain.Stages
	Downstream                    = domain.Downstream
	Sequence                      = domain.Sequence
)
