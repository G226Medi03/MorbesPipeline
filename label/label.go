// Package centitylabel builds short, human-readable labels for entity nodes.
package centitylabel

import (
	"encoding/json"
	"fmt"
	medicalcodec "github.com/G226Medi03/MedicalEntity/codec"
	"strings"

	eprocedure "github.com/G226Medi03/MedicalEntity/ename/emedical_procedure"
	"github.com/G226Medi03/MedicalEntity/ename/esymptom"
	codec "github.com/G226Medi03/MorbesPipeline/legacy"
	"github.com/G226Medi03/MorbesPipeline/trait"
	cduration "github.com/G226Medi03/go-duration"
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
			if value, err := codec.UnmarshalProcedure(raw); err == nil {
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
		if value, err := codec.UnmarshalProcedure([]byte(data)); err == nil {
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

func ProcedureLabel(value eprocedure.IMedicalProcedure) string {
	return procedure_label(value)
}

func procedure_label(value eprocedure.IMedicalProcedure) string {
	switch typed := value.(type) {
	case eprocedure.Transplant:
		target := strings.TrimSpace(typed.Target.ActorName())
		if target == "" {
			return "이식"
		}
		return target + " 이식"
	case eprocedure.Infuse:
		return first_nonempty(strings.TrimSpace(typed.Material.ActorName()), "주입")
	case eprocedure.Surgery:
		return first_nonempty(strings.TrimSpace(typed.Name), "수술")
	case eprocedure.Insertion:
		device := ""
		if typed.Device.Name != "" {
			device = strings.TrimSpace(typed.Device.Name)
		}
		return first_nonempty(device, "삽입")
	case eprocedure.Cleaning:
		target := ""
		if typed.Target != nil {
			target = strings.TrimSpace(typed.Target.ActorName())
		}
		if target == "" {
			return "세척"
		}
		return "세척 (" + target + ")"
	case eprocedure.DiagnosticTest:
		return first_nonempty(strings.TrimSpace(typed.Name), "검사")
	case eprocedure.Education:
		return first_nonempty(strings.TrimSpace(typed.Name), strings.TrimSpace(typed.Topic), "교육")
	case eprocedure.Instruction:
		return first_nonempty(strings.TrimSpace(typed.Name), strings.TrimSpace(typed.Action), "지시")
	case eprocedure.UnknownMedicalProcedure:
		return first_nonempty(strings.TrimSpace(typed.Name), "알 수 없는 시술")
	default:
		return value.Kind()
	}
}

func is_procedure_kind(kind string) bool {
	switch trait.ProcedureType(kind) {
	case trait.KindSurgery, trait.KindTransplant,
		trait.KindInfuse, trait.KindCleaning, trait.KindDiagnosticTest,
		trait.KindEducation, trait.KindInstruction, trait.KindInsertion, trait.KindUnknown:
		return true
	}
	return false
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
	value, err := medicalcodec.UnmarshalSymptom(raw)
	if err != nil {
		return value_label(raw)
	}
	switch typed := value.(type) {
	case esymptom.EOtherSymptom:
		return strings.TrimSpace(typed.Name)
	case esymptom.EPain:
		return pain_label(typed)
	}
	return value_label(raw)
}

func pain_label(value esymptom.EPain) string {
	site := ""
	details := make([]string, 0, len(value.Aspects))
	for _, aspect := range value.Aspects {
		switch typed := aspect.(type) {
		case esymptom.Site:
			site = strings.TrimSpace(typed.Site.ActorName())
		case esymptom.Duration:
			if text := strings.TrimSpace(strings.TrimSpace(typed.Value) + strings.TrimSpace(typed.Measure)); text != "" {
				details = append(details, text)
			}
		case esymptom.Intensity:
			if text := strings.TrimSpace(strings.TrimSpace(typed.Value) + strings.TrimSpace(typed.Measure)); text != "" {
				details = append(details, text)
			}
		case esymptom.Quality:
			if text := strings.TrimSpace(typed.Name); text != "" {
				details = append(details, text)
			}
		case esymptom.Radiation:
			if text := strings.TrimSpace(typed.Name); text != "" {
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
