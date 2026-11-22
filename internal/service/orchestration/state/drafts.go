package state

import (
	"github.com/G226Medi03/MedicalEntityParser/llmdto"
	morbesllm "github.com/G226Medi03/MorbesLLMService"
)

// Extraction draft types are owned by MorbesLLMService.
type (
	CaseExtraction          = morbesllm.CaseExtraction
	ActorDraft              = llmdto.ActorItem
	StatusDraft             = morbesllm.StatusDraft
	ActionDraft             = morbesllm.ActionDraft
	RelationDraft           = morbesllm.RelationDraft
	OptionDraft             = morbesllm.OptionDraft
	PatientDescriptionDraft = morbesllm.PatientDescriptionDraft
)
