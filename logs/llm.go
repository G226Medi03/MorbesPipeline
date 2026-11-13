package logs

import (
	source "github.com/G226Medi03/SourceEntity"
)

// LLMError records which stored LLM response caused a parse or canonicalization error.
// The raw prompt/output is intentionally not logged; the source detail page shows it.
func LLMError(scope, stage string, ask *source.LlmResponse, err error) {
	attrs := []Attr{
		AttrOf("stage", stage),
		AttrOf("error", err),
	}
	if scope != "" {
		attrs = append(attrs, AttrOf("llm_scope", scope))
	}
	if ask != nil {
		attrs = append(attrs,
			AttrOf("llm_resp_id", ask.RespID),
			AttrOf("llm_source_id", ask.SourceID))
	}
	Error("LLM 응답 처리 실패", attrs...)
}

// PreprocessChange describes one structural rule applied to an LLM response.
type PreprocessChange struct {
	Rule string
	Path string
}

// PreprocessChanges records which structural rules were applied to an LLM response.
func PreprocessChanges(scope, stage string, ask *source.LlmResponse, changes []PreprocessChange) {
	for _, change := range changes {
		attrs := []Attr{
			AttrOf("stage", stage),
			AttrOf("rule", change.Rule),
			AttrOf("path", change.Path),
		}
		if scope != "" {
			attrs = append(attrs, AttrOf("llm_scope", scope))
		}
		if ask != nil {
			attrs = append(attrs,
				AttrOf("llm_resp_id", ask.RespID),
				AttrOf("llm_source_id", ask.SourceID))
		}
		Warn("LLM 출력 보정", attrs...)
	}
}
