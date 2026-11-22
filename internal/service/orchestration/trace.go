package csolve

import (
	"github.com/G226Medi03/MorbesPipeline/internal/service/orchestration/prompt"
	"github.com/G226Medi03/MorbesPipeline/internal/service/orchestration/trace"
)

const (
	seqExplain                = 10
	seqQuestionType           = 12
	seqCaseExtraction         = 20
	seqResolve                = 30
	seqActor                  = 15
	seqOtherNames             = 17
	seqStatus                 = 20
	seqAction                 = 30
	seqActorAction            = 35
	seqPatientDescription     = 36
	seqMedicalProcedureOption = 37
	seqRelation               = 40
	seqSituation              = 50
	seqCause                  = 60
)

func (self *Service) load_prompts() (prompt.Set, error) {
	return trace.LoadPrompts(self.store)
}

func (self *Service) record_call(call LLMCallInput) {
	trace.RecordCall(self.store, call)
}

func marshal_json(value any) string {
	return trace.MarshalJSON(value)
}

func error_text(err error) string {
	return trace.ErrorText(err)
}
