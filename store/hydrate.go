package store

import (
	"encoding/json"
	medicalcodec "github.com/G226Medi03/MedicalEntity/codec"

	entity "github.com/G226Medi03/MorbesPipeline/legacy"
)

func (s *Store) hydrateProcedureData(procedureID int64) (map[string]any, bool) {
	value, found, err := s.GetProcedureByID(entity.ProcedureID(procedureID))
	if err != nil || !found {
		return nil, false
	}
	rawValue, err := json.Marshal(value.Value)
	if err != nil {
		return nil, false
	}
	localProcedure, err := entity.UnmarshalProcedure(with_kind(value.Variant, string(rawValue)))
	if err != nil {
		return nil, false
	}
	encoded, err := medicalcodec.MarshalProcedure(localProcedure)
	if err != nil {
		return nil, false
	}
	var payload map[string]any
	if err := json.Unmarshal(encoded, &payload); err != nil {
		return nil, false
	}
	return payload, true
}
func int64Value(value any) (int64, bool) {
	switch typed := value.(type) {
	case float64:
		return int64(typed), true
	case int64:
		return typed, true
	case int:
		return int64(typed), true
	case json.Number:
		value, err := typed.Int64()
		return value, err == nil
	default:
		return 0, false
	}
}

func decode_data_value(raw string) any {
	var value any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return raw
	}
	return value
}
