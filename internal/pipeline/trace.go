package csolve

import (
	"encoding/json"
	"fmt"
	"strings"

	clog "github.com/G226Medi03/log-sqlite"
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

type promptSet struct {
	present  map[string]bool
	versions map[string]int
}

func (self *Service) load_prompts() (promptSet, error) {
	result := promptSet{
		present:  make(map[string]bool, len(PromptStages())),
		versions: make(map[string]int, len(PromptStages())),
	}
	for _, stage := range PromptStages() {
		template, version, found, err := self.store.Prompt(stage)
		if err != nil {
			if stage == PromptActor || stage == PromptOtherNames || stage == PromptPatientDescription || stage == PromptQuestionType {
				continue
			}
			return promptSet{}, err
		}
		if !found {
			if stage == PromptActor || stage == PromptOtherNames || stage == PromptPatientDescription || stage == PromptQuestionType {
				continue
			}
			return promptSet{}, fmt.Errorf("prompt %q is not configured", stage)
		}
		result.present[stage] = strings.TrimSpace(template) != ""
		result.versions[stage] = version
	}
	return result, nil
}

func (self *Service) record_call(call LLMCallInput) {
	if call.State == "" {
		call.State = "ok"
	}
	// cause_trace is deterministic graph work, not an LLM call. Replays should
	// replace the previous trace record instead of accumulating one panel per run.
	if call.Stage == "cause_trace" {
		if err := self.store.DeleteLLMCallsByStage(call.CaseID, call.Stage); err != nil {
			clog.Warn("solve llm call 이전 trace 삭제 실패",
				clog.AttrOf("case", call.CaseID),
				clog.AttrOf("stage", call.Stage),
				clog.AttrOf("error", err))
		}
	}
	if _, err := self.store.AppendLLMCall(call); err != nil {
		clog.Warn("solve llm call 저장 실패",
			clog.AttrOf("case", call.CaseID),
			clog.AttrOf("stage", call.Stage),
			clog.AttrOf("error", err))
	}
}

func encode_json(value any) string {
	if value == nil {
		return ""
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(encoded)
}

func error_text(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
