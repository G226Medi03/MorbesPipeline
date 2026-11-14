// Package centitylabel builds short, human-readable labels for entity nodes.
package centitylabel

import (
	"encoding/json"
	"fmt"
	"strings"

	cduration "github.com/G226Medi03/MorbesPipeline/internal/external/duration"
)

// IsStatusKind reports whether kind is a persisted status variant.
func IsStatusKind(kind string) bool {
	switch strings.TrimSpace(kind) {
	case "increased_index", "decreased_index", "have_disease", "have_symptom",
		"damaged", "inflammation", "received_medical_procedure":
		return true
	}
	return false
}

// StatusLabel returns a human-readable status label without a kind prefix.
func StatusLabel(kind, data string) string {
	kind = strings.TrimSpace(kind)
	var obj map[string]json.RawMessage
	if err := json.Unmarshal([]byte(data), &obj); err != nil {
		return strings.TrimSpace(data)
	}

	switch kind {
	case "cause_status":
		if raw, ok := obj["status_id"]; ok {
			var status_id float64
			if err := json.Unmarshal(raw, &status_id); err == nil {
				return fmt.Sprintf("%d", int64(status_id))
			}
		}
		return "원인 상태 정보 없음"
	case "increased_index":
		return state_suffix(index_label(obj["index"], obj["name"]), "증가")
	case "decreased_index":
		return state_suffix(index_label(obj["index"], obj["name"]), "감소")
	case "have_disease":
		return first_nonempty(value_label(obj["disease"]), value_label(obj["name"]))
	case "have_symptom":
		return symptom_label(obj["symptom"], obj["name"])
	case "damaged":
		return first_nonempty(value_label(obj["actor"]), value_label(obj["name"]), "손상")
	case "inflammation":
		return first_nonempty(value_label(obj["actor"]), value_label(obj["name"]), "염증")
	case "received_medical_procedure":
		label := ""
		if raw, ok := obj["medical_procedure"]; ok {
			if value, err := unmarshalProcedure(raw); err == nil {
				label = procedure_label(value)
			}
		}
		if label == "" {
			label = value_label(obj["name"])
		}
		label = state_suffix(label, "받음")
		if period := period_label(obj["post_procedure_period"]); period != "" {
			label += " (" + period + ")"
		}
		return strings.TrimSpace(label)
	default:
		return first_nonempty(value_label(obj["name"]), strings.TrimSpace(data))
	}
}

// NodeLabel - kind 와 JSON data 로부터 "kind:value" 형태의 짧은 라벨을 만든다.
// 알려진 필드(index/disease/symptom/actor)를 우선 사용하고, 없으면 raw data 를 붙인다.
func NodeLabel(kind, data string) string {
	kind = strings.TrimSpace(kind)
	var obj map[string]json.RawMessage
	if err := json.Unmarshal([]byte(data), &obj); err != nil {
		return kind + ":" + data
	}

	if kind == "cause_status" {
		return kind + ":" + StatusLabel(kind, data)
	}
	if IsStatusKind(kind) {
		return kind + ":" + StatusLabel(kind, data)
	}
	if is_procedure_kind(kind) {
		if value, err := unmarshalProcedure([]byte(data)); err == nil {
			return kind + ":" + procedure_label(value)
		}
	}

	for _, key := range []string{"index", "disease", "symptom", "actor"} {
		raw, ok := obj[key]
		if !ok {
			continue
		}
		if value := value_label(raw); value != "" {
			return kind + ":" + value
		}
	}

	return kind + ":" + data
}

func value_label(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return normalize_index_text(text)
	}
	var nested map[string]json.RawMessage
	if err := json.Unmarshal(raw, &nested); err != nil {
		return ""
	}
	for _, key := range []string{"name", "code", "value"} {
		if value := value_label(nested[key]); value != "" {
			return value
		}
	}
	return ""
}

func index_label(raw json.RawMessage, fallback json.RawMessage) string {
	return first_nonempty(normalize_index_text(value_label(raw)), normalize_index_text(value_label(fallback)))
}

func normalize_index_text(value string) string {
	text := strings.TrimSpace(value)
	lower := strings.ToLower(text)
	if strings.HasPrefix(lower, "index(") && strings.HasSuffix(text, ")") {
		return strings.TrimSpace(text[len("index(") : len(text)-1])
	}
	return text
}

func state_suffix(value, suffix string) string {
	text := strings.TrimSpace(value)
	if text == "" {
		return suffix
	}
	if strings.HasSuffix(text, suffix) {
		return text
	}
	return text + " " + suffix
}

func symptom_label(raw json.RawMessage, fallback json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return value_label(fallback)
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(raw, &payload); err != nil {
		return value_label(raw)
	}
	kind := strings.ToLower(raw_string(payload["kind"]))
	switch kind {
	case "pain":
		return pain_label(payload["aspects"])
	case "other_symptom", "symptom", "symptom_code", "code", "":
		return first_nonempty(raw_string(payload["name"]), raw_string(payload["code"]), value_label(raw), first_nonempty(value_label(fallback)))
	default:
		return value_label(raw)
	}
}

func pain_label(raw json.RawMessage) string {
	var aspects []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &aspects); err != nil {
		return "통증"
	}
	site := ""
	details := make([]string, 0, len(aspects))
	for _, aspect := range aspects {
		switch strings.ToLower(raw_string(aspect["kind"])) {
		case "site":
			site = first_nonempty(value_label(aspect["site"]), raw_string(aspect["name"]))
		case "duration", "intensity":
			if text := strings.TrimSpace(raw_string(aspect["value"]) + raw_string(aspect["measure"])); text != "" {
				details = append(details, text)
			}
		case "quality":
			if text := strings.TrimSpace(raw_string(aspect["name"])); text != "" {
				details = append(details, text)
			}
		case "radiation":
			if text := strings.TrimSpace(raw_string(aspect["name"])); text != "" {
				details = append(details, text+" 방사")
			}
		}
	}
	label := "통증"
	if site != "" {
		label = site + " 통증"
	}
	if len(details) > 0 {
		label += " (" + strings.Join(details, ", ") + ")"
	}
	return label
}

func raw_string(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return ""
	}
	return strings.TrimSpace(value)
}

func first_nonempty(values ...string) string {
	for _, value := range values {
		if text := strings.TrimSpace(value); text != "" {
			return text
		}
	}
	return ""
}

func period_label(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return strings.TrimSpace(text)
	}
	value, err := cduration.Unmarshal(raw)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(value.Format())
}
