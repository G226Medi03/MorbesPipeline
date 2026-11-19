package trace

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve/model"
	"github.com/G226Medi03/MorbesPipeline/internal/pipeline/orchestration/prompt"
	store "github.com/G226Medi03/MorbesPipeline/internal/store"
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

func LoadPrompts(st *store.StoreContainer) (prompt.Set, error) {
	result := prompt.Set{
		Present:  make(map[string]bool, len(prompt.PromptStages())),
		Versions: make(map[string]int, len(prompt.PromptStages())),
	}
	for _, stage := range prompt.PromptStages() {
		template, version, found, err := st.Prompt(stage)
		if err != nil {
			if stage == prompt.PromptActor || stage == prompt.PromptOtherNames || stage == prompt.PromptPatientDescription || stage == prompt.PromptQuestionType {
				continue
			}
			return prompt.Set{}, err
		}
		if !found {
			if stage == prompt.PromptActor || stage == prompt.PromptOtherNames || stage == prompt.PromptPatientDescription || stage == prompt.PromptQuestionType {
				continue
			}
			return prompt.Set{}, fmt.Errorf("prompt %q is not configured", stage)
		}
		result.Present[stage] = strings.TrimSpace(template) != ""
		result.Versions[stage] = version
	}
	return result, nil
}

func RecordCall(st *store.StoreContainer, call model.LLMCallInput) {
	if call.State == "" {
		call.State = "ok"
	}
	// cause_trace is deterministic graph work, not an LLM call. Replays should
	// replace the previous trace record instead of accumulating one panel per run.
	if call.Stage == "cause_trace" {
		if err := st.DeleteLLMCallsByStage(call.CaseID, call.Stage); err != nil {
			clog.Warn("solve llm call 이전 trace 삭제 실패",
				clog.AttrOf("case", call.CaseID),
				clog.AttrOf("stage", call.Stage),
				clog.AttrOf("error", err))
		}
	}
	if _, err := st.AppendLLMCall(call); err != nil {
		clog.Warn("solve llm call 저장 실패",
			clog.AttrOf("case", call.CaseID),
			clog.AttrOf("stage", call.Stage),
			clog.AttrOf("error", err))
	}
}

func MarshalJSON(value any) string {
	if value == nil {
		return ""
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(encoded)
}

func ErrorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
