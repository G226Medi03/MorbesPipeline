// Package codec adapts the MedicalEntity codec to the application's local
// ename compatibility types. Keeping the conversion here lets the domain model
// stay stable while the persistence library evolves.
package legacy

import (
	"encoding/json"
	"fmt"
	"strings"

	medicalcodec "github.com/G226Medi03/MedicalEntity/codec"
	extaction "github.com/G226Medi03/MedicalEntity/ename/eaction"
	extactor "github.com/G226Medi03/MedicalEntity/ename/eactor"
	"github.com/G226Medi03/MedicalEntity/ename/einfuse_route"
	extprocedure "github.com/G226Medi03/MedicalEntity/ename/emedical_procedure"
	"github.com/G226Medi03/MedicalEntity/ename/estatus"
)

func MarshalAction(value extaction.IEAction) ([]byte, error) {
	if cause, ok := value.(CauseStatusAction); ok && cause.Status == nil && cause.StatusID > 0 {
		return json.Marshal(map[string]any{"kind": "cause_status", "status_id": cause.StatusID})
	}
	return medicalcodec.MarshalAction(toExternalAction(value))
}

func UnmarshalAction(data []byte) (extaction.IEAction, error) {
	value, err := medicalcodec.UnmarshalAction(data)
	if err == nil {
		if cause, ok := value.(extaction.CauseStatus); ok && cause.Status == nil {
			if statusID := legacyInt64Field(data, "status_id"); statusID > 0 {
				return CauseStatusAction{StatusID: statusID}, nil
			}
		}
		return value, nil
	}
	return unmarshalLegacyAction(data)
}

func UnmarshalProcedure(data []byte) (extprocedure.IMedicalProcedure, error) {
	if value, ok := unmarshalLegacyProcedure(data); ok {
		return value, nil
	}
	value, err := medicalcodec.UnmarshalProcedure(data)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func UnmarshalOrParseProcedure(data []byte) (extprocedure.IMedicalProcedure, error) {
	if value, ok := unmarshalLegacyProcedure(data); ok {
		return value, nil
	}
	value, err := medicalcodec.UnmarshalOrParseProcedure(data)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func UnmarshalStatus(data []byte) (estatus.IEStatus, error) {
	data = normalizeLegacyStatusJSON(data)
	value, err := medicalcodec.UnmarshalStatus(data)
	if err != nil {
		return nil, err
	}
	return fromExternalStatus(value), nil
}

func toExternalActor(value extactor.IEActor) extactor.IEActor {
	if value == nil {
		return nil
	}
	switch value.Kind() {
	case string(extactor.KindDevice):
		return extactor.NewDevice(value.ActorName())
	case string(extactor.KindProcedure):
		return extactor.NewProcedure(value.ActorName())
	default:
		return extactor.NewUnknownActor(value.ActorName())
	}
}

func fromExternalActor(value extactor.IEActor) extactor.IEActor {
	if value == nil {
		return nil
	}
	switch value.Kind() {
	case string(extactor.KindDevice):
		return extactor.Device{Name: value.ActorName()}
	case string(extactor.KindProcedure):
		return extactor.Procedure{Name: value.ActorName()}
	default:
		return extactor.UnknownActor{Name: value.ActorName()}
	}
}

func toExternalAction(value extaction.IEAction) extaction.IEAction {
	if value == nil {
		return nil
	}
	switch typed := value.(type) {
	case extaction.Block:
		return extaction.Block{Actor: toExternalActor(typed.Actor)}
	case extaction.Relax:
		return extaction.Relax{Actor: toExternalActor(typed.Actor)}
	case extaction.CauseStatus:
		return extaction.CauseStatus{Status: typed.Status}
	case CauseStatusAction:
		return extaction.CauseStatus{Status: typed.Status}
	case extaction.IEAction:
		return typed
	default:
		return nil
	}
}

func fromExternalAction(value extaction.IEAction) extaction.IEAction {
	if value == nil {
		return nil
	}
	switch typed := value.(type) {
	case extaction.Block:
		return extaction.Block{Actor: fromExternalActor(typed.Actor)}
	case extaction.Relax:
		return extaction.Relax{Actor: fromExternalActor(typed.Actor)}
	case extaction.CauseStatus:
		return extaction.CauseStatus{Status: fromExternalStatus(typed.Status)}
	default:
		return nil
	}
}

func unmarshalLegacyActor(raw json.RawMessage) extactor.IEActor {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return extactor.NewUnknownActor(text)
	}
	var value struct {
		Kind string `json:"kind"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(raw, &value); err != nil {
		return extactor.NewUnknownActor("")
	}
	return extactor.New_EActorFromKind(value.Kind, value.Name)
}

func unmarshalLegacyAction(data []byte) (extaction.IEAction, error) {
	var value struct {
		Kind     string          `json:"kind"`
		Actor    json.RawMessage `json:"actor"`
		Status   json.RawMessage `json:"status"`
		StatusID int64           `json:"status_id"`
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, fmt.Errorf("action JSON 파싱 실패: %w", err)
	}
	switch strings.ToLower(strings.TrimSpace(value.Kind)) {
	case "block":
		return extaction.Block{Actor: unmarshalLegacyActor(value.Actor)}, nil
	case "relax":
		return extaction.Relax{Actor: unmarshalLegacyActor(value.Actor)}, nil
	case "cause_status", "cause", "status":
		var status estatus.IEStatus
		if len(value.Status) > 0 && string(value.Status) != "null" {
			statusData := normalizeLegacyStatusJSON(value.Status)
			if parsed, err := medicalcodec.UnmarshalStatus(statusData); err == nil {
				status = parsed
			}
		}
		if value.StatusID > 0 {
			return CauseStatusAction{StatusID: value.StatusID, Status: status}, nil
		}
		return extaction.CauseStatus{Status: status}, nil
	default:
		return nil, fmt.Errorf("지원하지 않는 action kind 입니다: %s", value.Kind)
	}
}

func unmarshalLegacyProcedure(data []byte) (extprocedure.IMedicalProcedure, bool) {
	var value map[string]json.RawMessage
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, false
	}
	kind := legacyString(value["kind"])
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "transplant":
		raw, ok := legacyNestedPayload(value, "transplant")
		if !ok {
			return nil, false
		}
		var payload struct {
			Target json.RawMessage `json:"target"`
		}
		_ = json.Unmarshal(raw, &payload)
		actor := unmarshalLegacyActor(payload.Target)
		name := ""
		if actor != nil {
			name = actor.ActorName()
		}
		return extprocedure.Transplant{Target: extactor.UnknownActor{Name: name}}, true
	case "cleaning":
		raw, ok := legacyNestedPayload(value, "cleaning")
		if !ok {
			if _, hasName := value["target_name"]; !hasName {
				return nil, false
			}
		}
		var payload struct {
			Target json.RawMessage `json:"target"`
			Device json.RawMessage `json:"device"`
		}
		_ = json.Unmarshal(raw, &payload)
		actor := unmarshalLegacyActor(payload.Target)
		if actor == nil {
			actor = unmarshalLegacyActor(payload.Device)
		}
		name := ""
		if actor != nil {
			name = actor.ActorName()
		}
		if name == "" {
			name = legacyString(value["target_name"])
		}
		return extprocedure.Cleaning{Target: extactor.UnknownActor{Name: name}}, true
	case "infuse":
		raw, ok := legacyNestedPayload(value, "infuse")
		if !ok {
			return nil, false
		}
		var payload struct {
			Material json.RawMessage `json:"material"`
			Route    string          `json:"route"`
		}
		_ = json.Unmarshal(raw, &payload)
		return extprocedure.Infuse{
			Material: unmarshalLegacyActor(payload.Material),
			Route:    einfuse_route.InfuseRoute(payload.Route),
		}, true
	case "surgery":
		raw, ok := legacyNestedPayload(value, "surgery")
		if !ok {
			return nil, false
		}
		var payload struct {
			Name    string `json:"name"`
			RawName string `json:"raw_name"`
		}
		_ = json.Unmarshal(raw, &payload)
		name := payload.Name
		if name == "" {
			name = payload.RawName
		}
		return extprocedure.Surgery{Name: name}, true
	case "insertion":
		raw, ok := legacyNestedPayload(value, "insertion")
		if !ok {
			return nil, false
		}
		var payload struct {
			Name string `json:"name"`
			Tube string `json:"tube"`
			Site string `json:"site"`
		}
		_ = json.Unmarshal(raw, &payload)
		name := payload.Name
		if name == "" {
			name = payload.Tube
		}
		return extprocedure.Insertion{
			Device: extactor.NewDevice(name),
			Site:   extactor.NewUnknownActor(payload.Site),
		}, true
	case "procedure", "generic":
		raw, ok := legacyNestedPayload(value, "procedure")
		if !ok {
			if _, hasName := value["name"]; !hasName {
				return nil, false
			}
		}
		var payload struct {
			Name string `json:"name"`
		}
		_ = json.Unmarshal(raw, &payload)
		if payload.Name == "" {
			payload.Name = legacyString(value["name"])
		}
		return extprocedure.UnknownMedicalProcedure{Name: payload.Name}, true
	case "diagnostic_test", "diagnostictest":
		raw, ok := legacyNestedPayload(value, "diagnostic_test")
		if !ok {
			return nil, false
		}
		var payload struct {
			Name string `json:"name"`
		}
		_ = json.Unmarshal(raw, &payload)
		return extprocedure.DiagnosticTest{Name: payload.Name}, true
	case "education":
		raw, ok := legacyNestedPayload(value, "education")
		if !ok {
			return nil, false
		}
		var payload struct {
			Name  string `json:"name"`
			Topic string `json:"topic"`
		}
		_ = json.Unmarshal(raw, &payload)
		return extprocedure.Education{Name: payload.Name, Topic: payload.Topic}, true
	case "instruction":
		raw, ok := legacyNestedPayload(value, "instruction")
		if !ok {
			return nil, false
		}
		var payload struct {
			Name     string `json:"name"`
			Action   string `json:"action"`
			Duration string `json:"duration"`
			Position string `json:"position"`
		}
		_ = json.Unmarshal(raw, &payload)
		return extprocedure.Instruction{Name: payload.Name, Action: payload.Action, Duration: payload.Duration, Position: payload.Position}, true
	case "unknown", "unknown_medical_procedure":
		raw, ok := legacyNestedPayload(value, "unknown")
		if !ok {
			if _, hasName := value["name"]; !hasName {
				return nil, false
			}
		}
		var payload struct {
			Name string `json:"name"`
		}
		_ = json.Unmarshal(raw, &payload)
		if payload.Name == "" {
			payload.Name = legacyString(value["name"])
		}
		return extprocedure.UnknownMedicalProcedure{Name: payload.Name}, true
	}
	return nil, false
}

func legacyNestedPayload(value map[string]json.RawMessage, key string) (json.RawMessage, bool) {
	raw, ok := value[key]
	if !ok || len(raw) == 0 || string(raw) == "null" {
		return nil, false
	}
	return raw, true
}

func legacyString(raw json.RawMessage) string {
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return ""
	}
	return value
}

func legacyInt64Field(data []byte, key string) int64 {
	var value map[string]json.RawMessage
	if err := json.Unmarshal(data, &value); err != nil {
		return 0
	}
	var number int64
	if err := json.Unmarshal(value[key], &number); err != nil {
		return 0
	}
	return number
}

func fromExternalStatus(value estatus.IEStatus) estatus.IEStatus {
	switch typed := value.(type) {
	case estatus.ReceivedMedicalProcedure:
		return typed
	case estatus.Damaged:
		typed.Actor = fromExternalActor(typed.Actor)
		return typed
	case estatus.Inflammation:
		typed.Actor = fromExternalActor(typed.Actor)
		return typed
	default:
		return value
	}
}

// ToExternalStatus converts local nested procedure/actor values to the
// MedicalEntity persistence types.
func ToExternalStatus(value estatus.IEStatus) estatus.IEStatus {
	switch typed := value.(type) {
	case estatus.ReceivedMedicalProcedure:

		return typed
	case estatus.Damaged:
		typed.Actor = toExternalActor(typed.Actor)
		return typed
	case estatus.Inflammation:
		typed.Actor = toExternalActor(typed.Actor)
		return typed
	default:
		return value
	}
}

// ToExternalAction converts a local action to a MedicalEntity action.
func ToExternalAction(value extaction.IEAction) extaction.IEAction {
	return toExternalAction(value)
}

// FromExternalAction converts a MedicalEntity action to a local action.
func FromExternalAction(value extaction.IEAction) extaction.IEAction {
	return fromExternalAction(value)
}

// ToExternalProcedure converts a local procedure to a MedicalEntity procedure.
func ToExternalProcedure(value extprocedure.IMedicalProcedure) extprocedure.IMedicalProcedure {
	return value
}

// FromExternalProcedure converts a MedicalEntity procedure to a local procedure.
func FromExternalProcedure(value extprocedure.IMedicalProcedure) extprocedure.IMedicalProcedure {
	return value
}

// ToExternalActor converts a local actor to a MedicalEntity actor.
func ToExternalActor(value extactor.IEActor) extactor.IEActor {
	return toExternalActor(value)
}

// FromExternalActor converts a MedicalEntity actor to a local actor.
func FromExternalActor(value extactor.IEActor) extactor.IEActor {
	return fromExternalActor(value)
}

func normalizeLegacyStatusJSON(data []byte) []byte {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return data
	}
	changed := false
	if text, ok := payload["actor"].(string); ok {
		payload["actor"] = map[string]any{"kind": "unknown", "name": text}
		changed = true
	}
	if text, ok := payload["index"].(string); ok {
		payload["index"] = map[string]any{"name": text}
		changed = true
	}
	if text, ok := payload["symptom"].(string); ok {
		payload["symptom"] = map[string]any{"name": text}
		changed = true
	}
	if symptom, ok := payload["symptom"].(map[string]any); ok {
		if aspects, ok := symptom["aspects"].([]any); ok {
			for _, rawAspect := range aspects {
				aspect, ok := rawAspect.(map[string]any)
				if !ok || aspect["kind"] != "site" {
					continue
				}
				if text, ok := aspect["site"].(string); ok {
					aspect["site"] = map[string]any{"kind": "unknown", "name": text}
					changed = true
				}
			}
		}
	}
	if !changed {
		return data
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return data
	}
	return encoded
}

// MarshalStatusValue serializes a status payload without the top-level kind.
func MarshalStatusValue(value estatus.IEStatus) ([]byte, error) {
	payload, err := medicalcodec.MarshalStatus(value)
	if err != nil {
		return nil, err
	}
	return stripTopLevelKind(payload)
}

// UnmarshalStatusValue parses a status payload using the kind stored in the column.
func UnmarshalStatusValue(kind string, raw []byte) (estatus.IEStatus, error) {
	payload, err := addTopLevelKind(raw, kind)
	if err != nil {
		return nil, err
	}
	return UnmarshalStatus(payload)
}

// MarshalActionValue serializes an action payload without the top-level kind.
func MarshalActionValue(value extaction.IEAction) ([]byte, error) {
	payload, err := MarshalAction(value)
	if err != nil {
		return nil, err
	}
	return stripTopLevelKind(payload)
}

// UnmarshalActionValue parses an action payload using the kind stored in the column.
func UnmarshalActionValue(kind string, raw []byte) (extaction.IEAction, error) {
	payload, err := addTopLevelKind(raw, kind)
	if err != nil {
		return nil, err
	}
	return UnmarshalAction(payload)
}

// CanonicalActionValue normalizes a DTO value through the typed action serializer.
func CanonicalActionValue(kind string, value any) (any, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	parsed, err := UnmarshalActionValue(kind, raw)
	if err != nil {
		return nil, err
	}
	canonical, err := MarshalActionValue(parsed)
	if err != nil {
		return nil, err
	}
	var result any
	if err := json.Unmarshal(canonical, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func stripTopLevelKind(payload []byte) ([]byte, error) {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(payload, &obj); err != nil {
		return nil, err
	}
	delete(obj, "kind")
	return json.Marshal(obj)
}

func addTopLevelKind(raw []byte, kind string) ([]byte, error) {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, err
	}
	kindJSON, err := json.Marshal(kind)
	if err != nil {
		return nil, err
	}
	obj["kind"] = kindJSON
	return json.Marshal(obj)
}
