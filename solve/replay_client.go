package csolve

import (
	"context"
	"fmt"
	"strings"

	cllmcall "github.com/G226Medi03/llm-client/call"
)

func newReplayLLM(calls []LLMCall, live *cllmcall.Client, liveStages map[string]bool) *cllmcall.Client {
	outputs := make(map[string]string, len(calls))
	byStage := make(map[string]string, len(calls))
	for _, call := range calls {
		if call.Stage == "" {
			continue
		}
		byStage[call.Stage] = call.RawOutput
		if call.Prompt == "" {
			continue
		}
		outputs[call.Prompt] = call.RawOutput
	}
	model := "replay"
	if live != nil {
		model = live.Model()
	}
	return cllmcall.NewFunc(model, func(ctx context.Context, prompt string) (string, error) {
		stage := replay_stage_from_prompt(prompt)
		if stage != "" && live != nil && liveStages[stage] {
			return live.Generate(ctx, prompt)
		}
		if output, ok := outputs[prompt]; ok {
			return output, nil
		}
		if stage != "" {
			if output, ok := byStage[stage]; ok {
				return output, nil
			}
		}
		return "", fmt.Errorf("replay raw output 이 없습니다")
	})
}

func replay_stage_from_prompt(prompt string) string {
	lower := strings.ToLower(prompt)
	switch {
	case strings.Contains(lower, "biological entity") || strings.Contains(lower, "actor_extract"):
		return "actor_extract"
	case strings.Contains(prompt, "다른 이름") || strings.Contains(lower, "other_names"):
		return "other_names_extract"
	case strings.Contains(lower, "actor-action") || strings.Contains(lower, "actor/action") || strings.Contains(lower, "actor_action"):
		return "actor_action_extract"
	case strings.Contains(prompt, "CaseExtraction") || strings.Contains(lower, "case_extraction"):
		return "case_extraction"
	case strings.Contains(prompt, "환자가 서술") || strings.Contains(lower, "patient_description"):
		return "patient_description_extract"
	case strings.Contains(prompt, "치료/중재 선택지") || strings.Contains(lower, "medical_procedure_option"):
		return "medical_procedure_option_extract"
	case strings.Contains(prompt, "action만 추출") || strings.Contains(prompt, "행동(action)") || strings.Contains(lower, "action_extract"):
		return "action_extract"
	case strings.Contains(prompt, "s1, s2") || strings.Contains(prompt, "status만 추출") || strings.Contains(prompt, "상태(status)") || strings.Contains(lower, "status_extract"):
		return "status_extract"
	case strings.Contains(prompt, "원인-결과") || strings.Contains(lower, "relation_extract"):
		return "relation_extract"
	case strings.Contains(prompt, "상황") || strings.Contains(lower, "situation"):
		return "situation"
	case strings.Contains(lower, "explanation") || strings.Contains(prompt, "해설"):
		return "solver_explain"
	default:
		return ""
	}
}
