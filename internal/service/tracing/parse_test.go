package tracing

import (
	"encoding/json"
	"testing"
)

func TestParseCauseTraceJSONKeepsOnlyOwnedFields(t *testing.T) {
	raw := `{
		"observed": ["발열"],
		"missing": ["기침"],
		"debug": "discard",
		"candidates": [
			{
				"causeType": "infection",
				"causeKind": "status",
				"causeLabel": "폐렴",
				"viaActionKind": "cause",
				"depth": 1,
				"sourceIds": [10],
				"unexpected": "discard",
				"path": [
					{"type": "status", "label": "폐렴", "unexpected": "discard"},
					{"type": "action", "label": "원인"}
				]
			}
		]
	}`
	var result CauseTraceResult
	if err := ParseCauseTraceJSON(raw, &result); err != nil {
		t.Fatalf("ParseCauseTraceJSON: %v", err)
	}
	if len(result.Observed) != 1 || result.Observed[0] != "발열" {
		t.Fatalf("observed = %#v", result.Observed)
	}
	if len(result.Missing) != 1 || result.Missing[0] != "기침" {
		t.Fatalf("missing = %#v", result.Missing)
	}
	if len(result.Candidates) != 1 {
		t.Fatalf("candidates = %#v", result.Candidates)
	}
	candidate := result.Candidates[0]
	if candidate.CauseLabel != "폐렴" || candidate.Depth != 1 || len(candidate.Path) != 2 {
		t.Fatalf("candidate = %#v", candidate)
	}
	if len(candidate.SourceIDs) != 1 || candidate.SourceIDs[0].ToI64() != 10 {
		t.Fatalf("source IDs = %#v", candidate.SourceIDs)
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(encoded) == "" {
		t.Fatal("empty result")
	}
}

func TestParseCauseTraceJSONEmpty(t *testing.T) {
	var result CauseTraceResult
	if err := ParseCauseTraceJSON("  ", &result); err != nil {
		t.Fatalf("ParseCauseTraceJSON: %v", err)
	}
	if len(result.Observed) != 0 || len(result.Candidates) != 0 {
		t.Fatalf("result = %#v", result)
	}
}
