package csolve

import (
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	cservice "github.com/G226Medi03/MorbesPipeline/internal/llm"
)

func build_relation_prompt_from_drafts(
	template string,
	problem string,
	statuses []StatusDraft,
	actions []ActionDraft,
) (string, error) {
	status_map := make(map[string]repo.IEStatus, len(statuses))
	for _, draft := range statuses {
		status, err := parseStatusDraft(draft)
		if err != nil {
			return "", err
		}
		status_map[draft.Key] = status
	}
	action_map := make(map[string]repo.IEAction, len(actions))
	for _, draft := range actions {
		action, err := parseActionDraft(draft)
		if err != nil {
			return "", err
		}
		action_map[draft.Key] = action
	}
	return cservice.Build_RelationPromptFromMaps(
		template,
		problem,
		"",
		"",
		status_map,
		action_map,
	), nil
}
