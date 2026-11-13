// Package ports defines the persistence boundary used by the medical core.
package ports

import (
	"github.com/G226Medi03/MedicalEntity/ename/eaction"
	"github.com/G226Medi03/MedicalEntity/ename/eactor"
	"github.com/G226Medi03/MedicalEntity/ename/estatus"
	medicalentity "github.com/G226Medi03/MedicalEntity/entity"
	cgraph "github.com/G226Medi03/MorbesPipeline/graph"
	entity "github.com/G226Medi03/MorbesPipeline/legacy"
	"github.com/G226Medi03/MorbesPipeline/medicalprocedureoption"
	"github.com/G226Medi03/MorbesPipeline/solve/domain"
	"github.com/G226Medi03/MorbesPipeline/solve/model"
	source "github.com/G226Medi03/SourceEntity"
)

// Store is the subset of the application store used by the medical core.
type Store interface {
	Get(id domain.CaseID) (domain.Case, bool, error)
	SetStage(id domain.CaseID, stage string) error
	SetRunning(id domain.CaseID) error
	SetFailed(id domain.CaseID, ce error) error
	SetAnswer(id domain.CaseID, answerJSON string) error
	SetQuestionType(id domain.CaseID, questionType string) error
	SetSource(id domain.CaseID, sourceID source.SourceId) error
	ResetCaseForReplay(id domain.CaseID, mode, fromNode string, keepCalls bool) error

	StartStep(caseID domain.CaseID, node string, promptVersion int) (entity.SolveStepID, error)
	FinishStep(id entity.SolveStepID, state, outputJSON, message string, callID entity.LLMCallID) error
	GetStep(caseID domain.CaseID, node string) (domain.Step, bool, error)
	ResetSteps(caseID domain.CaseID) error
	MarkDownstreamStale(caseID domain.CaseID, fromNode string) error

	AppendLLMCall(call model.LLMCallInput) (entity.LLMCallID, error)
	ListLLMCalls(caseID domain.CaseID) ([]model.LLMCall, error)
	DeleteLLMCallsByStage(caseID domain.CaseID, stage string) error
	FindLlmResponse(modelName, prompt string, referedSourceID source.SourceId) (*source.LlmResponse, bool, error)
	GetOrInsertLlmResponse(modelName, prompt string, referedSourceID source.SourceId, output string) (*source.LlmResponse, error)
	RefreshLlmResponse(modelName, prompt string, referedSourceID source.SourceId, output string) (*source.LlmResponse, error)

	SaveAction(value entity.Action) (entity.Action, error)
	SaveDevice(value eactor.Device, sources []source.SourceRef) (entity.Device, error)
	SaveUnknownActor(value eactor.UnknownActor, sources []source.SourceRef) (entity.UnknownActor, error)
	SaveMedicalProcedureOption(option medicalprocedureoption.MedicalProcedureOption) (medicalprocedureoption.MedicalProcedureOption, error)
	LinkStatusAction(statusID entity.StatusID, actionID entity.ActionID, sourceID source.SourceId) error
	LinkActionResultStatus(actionID entity.ActionID, statusID entity.StatusID, sourceID source.SourceId) error

	SaveDisease(name string, sources []source.SourceRef) (medicalentity.Disease, error)
	SaveIndex(name string, sources []source.SourceRef) (medicalentity.Index, error)
	SaveSymptom(name string, sources []source.SourceRef) (medicalentity.Symptom, error)
	SaveStatus(value entity.Status) (entity.Status, error)
	SaveProcedureWithTraits(kind, data string, traits []entity.Trait, sources []source.SourceRef) (entity.ProcedureID, error)

	GetPrompt(stage string) (string, int, error)

	LoadGraph() (*cgraph.Container, error)
	EnsureStatusIn(c *cgraph.Container, value estatus.IEStatus, sourceID source.SourceId) (entity.StatusID, error)
	EnsureActionIn(c *cgraph.Container, value eaction.IEAction, sourceID source.SourceId) (entity.ActionID, error)
	LinkStatusActionIn(c *cgraph.Container, statusID entity.StatusID, actionID entity.ActionID) error
	LinkActionResultStatusIn(c *cgraph.Container, actionID entity.ActionID, statusID entity.StatusID) error
	DeleteStatusAction(statusID entity.StatusID, actionID entity.ActionID) error
	DeleteActionResultStatus(actionID entity.ActionID, statusID entity.StatusID) error
}
