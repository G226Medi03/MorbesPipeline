package domain

import (
	"fmt"
	"strings"
)

const (
	PromptExplain                = "explain"
	PromptQuestionType           = "question_type"
	PromptCaseExtraction         = "case_extraction"
	PromptActor                  = "actor"
	PromptOtherNames             = "other_names"
	PromptStatus                 = "status"
	PromptAction                 = "action"
	PromptActorAction            = "actor_action"
	PromptPatientDescription     = "patient_description"
	PromptMedicalProcedureOption = "medical_procedure_option"
	PromptRelation               = "relation"
	PromptSituation              = "situation"
)

var promptStages = []string{
	PromptExplain,
	PromptQuestionType,
	PromptActor,
	PromptStatus,
	PromptAction,
	PromptPatientDescription,
	PromptMedicalProcedureOption,
	PromptRelation,
}

var promptPlaceholders = map[string][]string{
	PromptExplain:                {"problem", "answer"},
	PromptQuestionType:           {"problem", "answer", "explanation"},
	PromptCaseExtraction:         {"problem", "answer", "explanation", "patient_description_eligible"},
	PromptActor:                  {"problem", "answer", "explanation"},
	PromptOtherNames:             {"problem", "answer", "explanation", "actor_json"},
	PromptStatus:                 {"problem", "answer", "explanation"},
	PromptAction:                 {"problem", "answer", "explanation"},
	PromptActorAction:            {"problem", "answer", "explanation", "actor_json", "action_json"},
	PromptPatientDescription:     {"problem", "answer", "explanation", "status_json"},
	PromptMedicalProcedureOption: {"problem", "answer", "explanation", "status_json", "patient_description_json"},
	PromptRelation:               {"problem", "answer", "explanation", "status_json", "action_json"},
	PromptSituation:              {"problem", "answer", "explanation"},
}

func PromptStages() []string {
	result := make([]string, len(promptStages))
	copy(result, promptStages)
	return result
}

func PromptPlaceholders(stage string) []string {
	values := promptPlaceholders[stage]
	result := make([]string, len(values))
	copy(result, values)
	return result
}

func ValidatePromptTemplate(stage, template string) error {
	placeholders, ok := promptPlaceholders[stage]
	if !ok {
		return fmt.Errorf("알 수 없는 prompt stage 입니다: %s", stage)
	}
	if strings.TrimSpace(template) == "" {
		return fmt.Errorf("prompt template 이 비어 있습니다")
	}
	for _, placeholder := range placeholders {
		if !strings.Contains(template, "{{"+placeholder+"}}") {
			return fmt.Errorf("필수 placeholder 가 없습니다: {{%s}}", placeholder)
		}
	}
	return nil
}

func RenderPrompt(template string, values map[string]string) string {
	result := template
	for key, value := range values {
		result = strings.ReplaceAll(result, "{{"+key+"}}", strings.TrimSpace(value))
	}
	return strings.TrimSpace(result)
}
