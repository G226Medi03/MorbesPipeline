package morbes_pipeline

import (
	"encoding/json"
	"fmt"
	"strings"

	entityjson "github.com/G226Medi03/MorbesPipeline/internal/domain/entityjson"
	label "github.com/G226Medi03/MorbesPipeline/internal/domain/label"
)

// MarshalAction encodes a typed action value.
func MarshalAction(value IEAction) ([]byte, error) {
	return entityjson.EncodeAction(value)
}

// MarshalProcedure encodes a typed medical procedure value.
func MarshalProcedure(value IMedicalProcedure) ([]byte, error) {
	return entityjson.EncodeProcedure(value)
}

// MarshalStatus encodes a typed status value.
func MarshalStatus(value IEStatus) ([]byte, error) {
	return entityjson.EncodeStatus(value)
}

// IsStatusKind reports whether kind is a persisted status variant.
func IsStatusKind(kind string) bool {
	switch strings.TrimSpace(kind) {
	case "increased_index", "decreased_index", "have_disease", "have_symptom",
		"damaged", "inflammation", "received_medical_procedure":
		return true
	}
	return false
}

// StatusLabel returns a label for a kind-tagged status payload.
func StatusLabel(kind, data string) string {
	raw, err := withJSONKind(kind, data)
	if err != nil {
		return strings.TrimSpace(data)
	}
	value, err := entityjson.DecodeStatus(raw)
	if err != nil {
		return strings.TrimSpace(data)
	}
	return label.StatusLabelValue(value)
}

// NodeLabel returns a label for a kind-tagged entity payload.
func NodeLabel(kind, data string) string {
	if IsStatusKind(kind) {
		return StatusLabel(kind, data)
	}
	raw, err := withJSONKind(kind, data)
	if err == nil {
		if value, err := entityjson.DecodeAction(raw); err == nil {
			return label.ActionLabelValue(value)
		}
		if value, err := entityjson.DecodeProcedure(raw); err == nil {
			return label.ProcedureLabel(value)
		}
	}
	return strings.TrimSpace(kind) + ":" + strings.TrimSpace(data)
}

// ExtractJSONElements extracts JSON array elements from an LLM response.
func ExtractJSONElements(rawText string) ([]json.RawMessage, error) {
	trimmed := strings.TrimSpace(rawText)
	if trimmed == "" {
		return nil, fmt.Errorf("응답에서 JSON 을 찾지 못했습니다")
	}
	if start := strings.Index(trimmed, "["); start != -1 {
		if end := strings.LastIndex(trimmed, "]"); end > start {
			var array []json.RawMessage
			if err := json.Unmarshal([]byte(trimmed[start:end+1]), &array); err == nil {
				return array, nil
			}
		}
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal([]byte(trimmed), &object); err != nil {
		return nil, fmt.Errorf("응답에서 JSON 을 찾지 못했습니다: %w", err)
	}
	for _, key := range []string{"nodes", "chains", "items", "results"} {
		raw, ok := object[key]
		if !ok {
			continue
		}
		var array []json.RawMessage
		if err := json.Unmarshal(raw, &array); err == nil {
			return array, nil
		}
	}
	return []json.RawMessage{json.RawMessage(trimmed)}, nil
}

func withJSONKind(kind, data string) ([]byte, error) {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		return nil, err
	}
	encodedKind, err := json.Marshal(kind)
	if err != nil {
		return nil, err
	}
	payload["kind"] = encodedKind
	return json.Marshal(payload)
}
