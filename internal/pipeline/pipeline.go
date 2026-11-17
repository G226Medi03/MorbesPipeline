package csolve

import (
	"context"
	"encoding/json"
	"fmt"
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	ccause "github.com/G226Medi03/MorbesPipeline/internal/tracing/cause"
	"strings"
	"time"

	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
	core "github.com/G226Medi03/PipeLineCore"
	source "github.com/G226Medi03/SourceEntity"
	clog "github.com/G226Medi03/log-sqlite"
)

func (self *Service) Run(ctx context.Context, id domain.CaseID) error {
	return self.RunWithCache(ctx, id, true)
}

func (self *Service) RunWithCache(ctx context.Context, id domain.CaseID, useCache bool) error {
	run := *self
	run.ctx = ctx
	run.cache_mode = core.CacheUse
	if !useCache {
		run.cache_mode = core.CacheRefresh
	}
	return run.run_v2_from(ctx, id, "")
}

func (self *Service) record_provided_explanation(id domain.CaseID, state *pipelineState) error {
	if err := self.set_stage(id, "solver_explain"); err != nil {
		return err
	}
	step_id, err := self.store.StartStep(id, NodeExplain, 0)
	if err != nil {
		return err
	}
	state.Explanation = strings.TrimSpace(state.Case.Explanation)
	state.ExplanationSourceID = state.Case.SourceID
	if state.ExplanationSourceID.ToI64() <= 0 {
		state.ExplanationSourceID = state.Case.SituationSourceID
	}
	if state.ExplanationSourceID.ToI64() > 0 {
		if err := self.store.SetSource(id, state.ExplanationSourceID); err != nil {
			return err
		}
	}
	return self.store.FinishStep(
		step_id,
		"succeeded",
		encode_json(explainStepOutput{Explanation: state.Explanation}),
		"",
		0,
	)
}

func (self *Service) run_explain_node(ctx context.Context, id domain.CaseID, state *pipelineState) error {
	if err := self.set_stage(id, "solver_explain"); err != nil {
		return err
	}
	step_id, err := self.store.StartStep(id, NodeExplain, state.Prompts.versions[PromptExplain])
	if err != nil {
		return err
	}
	explain_started := time.Now()
	clog.Info("explain 시작", clog.AttrOf("case", id))
	result, err := self.explain(ctx, state.Case.Problem, state.Case.Answer)
	explain_model := ""
	prompt := ""
	raw := ""
	explanation := ""
	prompt_version := state.Prompts.versions[PromptExplain]
	if result != nil {
		explanation = result.Explanation
		raw = result.Raw
		prompt_version = result.PromptVersion
		if result.Response != nil {
			explain_model = result.Response.Model
			prompt = result.Response.Prompt
		}
	}
	self.record_call(LLMCallInput{
		CaseID:        id,
		Seq:           seqExplain,
		Stage:         "solver_explain",
		Role:          "solver",
		Model:         explain_model,
		PromptVersion: prompt_version,
		Prompt:        prompt,
		RawOutput:     raw,
		ParsedOutput:  encode_json(map[string]string{"explanation": explanation}),
		State:         state_text(err),
		Error:         error_text(err),
	})
	if err != nil {
		clog.Error("explain 실패", clog.AttrOf("case", id), clog.AttrOf("duration_ms", time.Since(explain_started).Milliseconds()), clog.AttrOf("error", err))
		_ = self.store.FinishStep(step_id, "failed", "", err.Error(), 0)
		return err
	}
	clog.Info("explain 완료", clog.AttrOf("case", id), clog.AttrOf("duration_ms", time.Since(explain_started).Milliseconds()))
	state.Explanation = explanation
	response, err := self.store.GetOrInsertLlmResponse(explain_model, "solve:"+string(id)+":explanation", source.SourceId{}, explanation)
	if err != nil {
		_ = self.store.FinishStep(step_id, "failed", "", err.Error(), 0)
		return err
	}
	state.ExplanationSourceID = response.SourceID
	if err := self.store.SetSource(id, state.ExplanationSourceID); err != nil {
		_ = self.store.FinishStep(step_id, "failed", "", err.Error(), 0)
		return err
	}
	return self.store.FinishStep(step_id, "succeeded", encode_json(explainStepOutput{Explanation: explanation}), "", 0)
}

func looks_like_patient_description(problem string) bool {
	for _, token := range []string{"진술", "서술", "표현", "이해한", "말한", "호소"} {
		if strings.Contains(problem, token) {
			return true
		}
	}
	return strings.ContainsAny(problem, "“”‘’\"'")
}

func trim_string_list(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		result = append(result, value)
	}
	return result
}

const cause_trace_timeout = 90 * time.Second

func (self *Service) run_cause_node(ctx context.Context, id domain.CaseID, state *pipelineState) error {
	step_id, err := self.store.StartStep(id, NodeCause, 0)
	if err != nil {
		return err
	}
	if err := self.set_stage(id, "cause_trace"); err != nil {
		return err
	}
	statuses := make([]repo.IEStatus, 0, len(state.Statuses))
	for _, status := range state.Statuses {
		statuses = append(statuses, status)
	}
	if len(statuses) == 0 {
		statuses, err = state.Situation.ToStatuses()
		if err != nil {
			_ = self.store.FinishStep(step_id, "failed", "", err.Error(), 0)
			return err
		}
	}
	statuses, coverageAdded := coverageObservedStatuses(state, statuses)
	cause_ctx, cancel := context.WithTimeout(ctx, cause_trace_timeout)
	defer cancel()
	container, err := Postprocess(cause_ctx, self.store)
	if err != nil {
		_ = self.store.FinishStep(step_id, "failed", "", err.Error(), 0)
		return err
	}
	cause_service := ccause.New_ContainerService(container)
	const cause_trace_max_depth = 100
	candidates, err := cause_service.TraceContext(cause_ctx, statuses, cause_trace_max_depth)
	self.record_call(LLMCallInput{
		CaseID: id, Seq: seqCause, Stage: "cause_trace", Role: "trace",
		ParsedOutput: encode_json(candidates), State: state_text(err), Error: error_text(err),
	})
	if err != nil {
		_ = self.store.FinishStep(step_id, "failed", "", err.Error(), 0)
		return err
	}
	result := causeStepOutput{
		Observed:      ccause.ObservedLabels(statuses),
		Candidates:    candidates,
		Missing:       []string{},
		CoverageAdded: coverageAdded,
	}
	if len(candidates) == 0 {
		result.Missing = []string{"Morbes graph 에서 원인 후보를 찾지 못했습니다"}
	}
	encoded := encode_json(result)
	if err := self.store.SetAnswer(id, encoded); err != nil {
		_ = self.store.FinishStep(step_id, "failed", "", err.Error(), 0)
		return err
	}
	state.Candidates = candidates
	return self.store.FinishStep(step_id, "succeeded", encoded, "", 0)
}

func (self *Service) load_explain(id domain.CaseID, state *pipelineState) error {
	var out explainStepOutput
	if err := self.load_step_output(id, NodeExplain, &out); err != nil {
		return err
	}
	state.Explanation = out.Explanation
	item, found, err := self.store.Get(id)
	if err != nil || !found {
		return fmt.Errorf("explain checkpoint 를 찾지 못했습니다")
	}
	state.ExplanationSourceID = item.SourceID
	return nil
}

func (self *Service) load_cause(id domain.CaseID, state *pipelineState) error {
	var out causeStepOutput
	if err := self.load_step_output(id, NodeCause, &out); err != nil {
		return err
	}
	state.Candidates = out.Candidates
	return nil
}

func (self *Service) load_step_output(case_id domain.CaseID, node string, target any) error {
	step, found, err := self.store.GetStep(case_id, node)
	if err != nil {
		return err
	}
	if !found || step.State != "succeeded" || strings.TrimSpace(step.Meta.OutputJSON) == "" {
		return fmt.Errorf("%s checkpoint 가 없습니다", node)
	}
	if err := json.Unmarshal([]byte(step.Meta.OutputJSON), target); err != nil {
		return fmt.Errorf("%s checkpoint 파싱 실패: %w", node, err)
	}
	return nil
}

func state_text(err error) string {
	if err != nil {
		return "failed"
	}
	return "ok"
}
