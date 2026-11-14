package cservice

import (
	"context"
	"encoding/json"
	"fmt"
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"sort"
	"strings"

	cllmparser "github.com/G226Medi03/MedicalEntityParser/llmparser"
	cllmcall "github.com/G226Medi03/MorbesPipeline/internal/external/llmclient/call"
	source "github.com/G226Medi03/MorbesPipeline/internal/external/sourceentity"
	"github.com/G226Medi03/MorbesPipeline/internal/persistence"
	"github.com/G226Medi03/MorbesPipeline/internal/support/logs"
)

// Build_RelationPromptFromMaps renders a relation prompt template with status/action JSON.
func Build_RelationPromptFromMaps(template, problem, answer, explanation string, statuses map[string]repo.IEStatus, actions map[string]repo.IEAction) string {
	status_json, _ := status_prompt_json(statuses)
	action_json, _ := action_prompt_json(actions)
	result := strings.ReplaceAll(template, "{{problem}}", strings.TrimSpace(problem))
	result = strings.ReplaceAll(result, "{{answer}}", strings.TrimSpace(answer))
	result = strings.ReplaceAll(result, "{{explanation}}", strings.TrimSpace(explanation))
	result = strings.ReplaceAll(result, "{{status_json}}", status_json)
	result = strings.ReplaceAll(result, "{{action_json}}", action_json)
	return strings.TrimSpace(result)
}

// ExtractRelationsWithPrompt runs relation extraction with a caller-provided prompt.
func ExtractRelationsWithPrompt(
	ctx context.Context,
	llm *cllmcall.Client,
	st *store.Store,
	refered_source_id source.SourceId,
	log_scope string,
	prompt string,
	cache_mode LLMCacheMode,
) ([][]string, source.SourceRef, *source.LlmResponse, error) {
	if strings.TrimSpace(prompt) == "" {
		return nil, source.SourceRef{}, nil, fmt.Errorf("relation prompt가 비어 있습니다")
	}
	ask, ce := ask_llm_with_mode(ctx, llm, st, prompt, refered_source_id, cache_mode)
	if ce != nil {
		return nil, source.SourceRef{}, nil, ce
	}
	relations, ce := cllmparser.ParseRelations(ask.Output)
	if ce != nil {
		logs.LLMError(log_scope, "relation_parse", ask, ce)
		return nil, source.SourceRef{}, ask, ce
	}
	return relations, source.SourceRef{SourceID: source.SourceId(ask.SourceID)}, ask, nil
}

func status_prompt_json(statuses map[string]repo.IEStatus) (string, error) {
	keys := sorted_keys(statuses)
	items := make([]map[string]any, 0, len(keys))
	for _, key := range keys {
		encoded, err := repo.MarshalStatus(statuses[key])
		if err != nil {
			return "", err
		}
		payload := map[string]any{}
		if err := json.Unmarshal(encoded, &payload); err != nil {
			return "", err
		}
		payload["id"] = key
		payload["kind"] = statuses[key].Kind()
		items = append(items, payload)
	}
	encoded, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func action_prompt_json(actions map[string]repo.IEAction) (string, error) {
	keys := sorted_keys(actions)
	items := make([]map[string]any, 0, len(keys))
	for _, key := range keys {
		encoded, err := repo.MarshalAction(actions[key])
		if err != nil {
			return "", err
		}
		payload := map[string]any{}
		if err := json.Unmarshal(encoded, &payload); err != nil {
			return "", err
		}
		payload["id"] = key
		payload["kind"] = actions[key].Kind()
		items = append(items, payload)
	}
	encoded, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func sorted_keys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
