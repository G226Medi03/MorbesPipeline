package morbes_pipeline

import (
	label "github.com/G226Medi03/MorbesPipeline/internal/domain/label"
	"github.com/G226Medi03/MorbesPipeline/internal/domain/trait"
	"github.com/G226Medi03/MorbesPipeline/internal/support/ids"
	"github.com/G226Medi03/MorbesPipeline/internal/support/logs"
	"github.com/G226Medi03/MorbesPipeline/internal/support/sourceutil"
)

type (
	StatusID           = ids.StatusID
	ActionID           = ids.ActionID
	ActorID            = ids.ActorID
	ProcedureID        = ids.ProcedureID
	DeviceID           = ids.DeviceID
	ExamQuestionID     = ids.ExamQuestionID
	LLMProblemResultID = ids.LLMProblemResultID
	LLMCallID          = ids.LLMCallID
	StatusEntityID     = ids.StatusEntityID

	Attr    = logs.Attr
	Logger  = logs.Logger
	Options = logs.Options
	Logging = logs.Logging
)

var (
	Open                = logs.Open
	OptionsFromEnv      = logs.OptionsFromEnv
	AttrOf              = logs.AttrOf
	SetDefault          = logs.SetDefault
	InstallStdlogBridge = logs.InstallStdlogBridge

	SourceID = sourceutil.SourceID
	RespID   = sourceutil.RespID

	ProcedureLabel = label.ProcedureLabel

	MergeTraits       = trait.MergeTraits
	PriorityTraits    = trait.PriorityTraits
	DescriptiveTraits = trait.DescriptiveTraits
	ComparePriority   = trait.ComparePriority
	CategoryRank      = trait.CategoryRank
	PriorityReason    = trait.PriorityReason
)

func Info(msg string, attrs ...Attr)  { logs.Info(msg, attrs...) }
func Warn(msg string, attrs ...Attr)  { logs.Warn(msg, attrs...) }
func Error(msg string, attrs ...Attr) { logs.Error(msg, attrs...) }
func Fatal(msg string, attrs ...Attr) { logs.Fatal(msg, attrs...) }
