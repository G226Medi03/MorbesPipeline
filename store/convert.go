package store

import (
	"encoding/json"
	"fmt"
	medicalcodec "github.com/G226Medi03/MedicalEntity/codec"

	medicalentity "github.com/G226Medi03/MedicalEntity/entity"
	entity "github.com/G226Medi03/MorbesPipeline/legacy"
	"github.com/G226Medi03/SourceEntity"
)

func typedDiseaseID(id int64) medicalentity.DiseaseID {
	var value medicalentity.DiseaseID
	_ = value.Scan(id)
	return value
}

func typedIndexID(id int64) medicalentity.IndexID {
	var value medicalentity.IndexID
	_ = value.Scan(id)
	return value
}

func typedSymptomID(id int64) medicalentity.SymptomID {
	var value medicalentity.SymptomID
	_ = value.Scan(id)
	return value
}

func entityStatusID(id entity.StatusID) medicalentity.StatusID {
	var value medicalentity.StatusID
	_ = value.Scan(id.ToI64())
	return value
}

func entityActionID(id entity.ActionID) medicalentity.ActionID {
	var value medicalentity.ActionID
	_ = value.Scan(id.ToI64())
	return value
}

func entityActorID(id entity.ActorID) medicalentity.ActorID {
	var value medicalentity.ActorID
	_ = value.Scan(id.ToI64())
	return value
}

func entityProcedureID(id entity.ProcedureID) medicalentity.ProcedureID {
	var value medicalentity.ProcedureID
	_ = value.Scan(id.ToI64())
	return value
}

func localStatusID(id medicalentity.StatusID) entity.StatusID {
	return entity.StatusID(id.ToI64())
}

func localActionID(id medicalentity.ActionID) entity.ActionID {
	return entity.ActionID(id.ToI64())
}

func localActorID(id medicalentity.ActorID) entity.ActorID {
	return entity.ActorID(id.ToI64())
}

func localProcedureID(id medicalentity.ProcedureID) entity.ProcedureID {
	return entity.ProcedureID(id.ToI64())
}

func externalStatusFromLocal(status entity.Status) (medicalentity.IEStatus, error) {
	raw, err := json.Marshal(status.Value)
	if err != nil {
		return nil, err
	}
	localValue, err := entity.UnmarshalStatus(with_kind(status.Variant, string(raw)))
	if err != nil {
		return nil, err
	}
	return entity.ToExternalStatus(localValue), nil
}

func localStatusFromExternal(status medicalentity.Status) (entity.Status, error) {
	value, err := statusValue(status)
	if err != nil {
		return entity.Status{}, err
	}
	kind := ""
	if status.Name() != nil {
		kind = status.Name().Kind()
	}
	result := entity.RestoreStatus(localStatusID(status.Id()), kind, value, status.Sources)
	for _, link := range status.Actions {
		action, err := localActionFromExternal(link.Action)
		if err != nil {
			return entity.Status{}, err
		}
		result.Actions = append(result.Actions, struct {
			Action   entity.Action   `json:"action"`
			SourceID source.SourceId `json:"sourceId"`
		}{Action: action, SourceID: link.SourceID})
	}
	return result, nil
}

func statusValue(status medicalentity.Status) (any, error) {
	if status.Name() == nil {
		return nil, nil
	}
	raw, err := medicalcodec.MarshalStatus(status.Name())
	if err != nil {
		return nil, err
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	delete(payload, "kind")
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	var value any
	if err := json.Unmarshal(encoded, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func externalActionFromLocal(action entity.Action) (medicalentity.IEAction, error) {
	raw, err := json.Marshal(action.Value)
	if err != nil {
		return nil, err
	}
	localValue, err := entity.UnmarshalAction(with_kind(action.Variant, string(raw)))
	if err != nil {
		return nil, err
	}
	return entity.ToExternalAction(localValue), nil
}

func localActionFromExternal(action medicalentity.Action) (entity.Action, error) {
	if action.Name() == nil {
		return entity.Action{}, nil
	}
	localValue := entity.FromExternalAction(action.Name())
	raw, err := entity.MarshalAction(localValue)
	if err != nil {
		return entity.Action{}, err
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(raw, &payload); err != nil {
		return entity.Action{}, err
	}
	delete(payload, "kind")
	encoded, err := json.Marshal(payload)
	if err != nil {
		return entity.Action{}, err
	}
	var value any
	if err := json.Unmarshal(encoded, &value); err != nil {
		return entity.Action{}, err
	}
	result := entity.RestoreAction(localActionID(action.Id()), localValue.Kind(), value, action.Sources)
	for _, link := range action.ResultStatuses {
		status, err := localStatusFromExternal(medicalentity.NewStatus(
			medicalentity.StatusID{},
			link.Status,
			nil,
			nil,
		))
		if err != nil {
			return entity.Action{}, err
		}
		result.ResultStatuses = append(result.ResultStatuses, struct {
			Status   entity.Status   `json:"status"`
			SourceID source.SourceId `json:"sourceId"`
		}{Status: status, SourceID: link.SourceID})
	}
	return result, nil
}

func externalProcedureFromLocal(value entity.Procedure) (medicalentity.IMedicalProcedure, error) {
	raw, err := json.Marshal(value.Value)
	if err != nil {
		return nil, err
	}
	localValue, err := entity.UnmarshalProcedure(with_kind(value.Variant, string(raw)))
	if err != nil {
		return nil, err
	}
	return entity.ToExternalProcedure(localValue), nil
}

func localProcedureFromExternal(value medicalentity.Procedure) (entity.Procedure, error) {
	if value.Name() == nil {
		return entity.Procedure{}, nil
	}
	localValue := entity.FromExternalProcedure(value.Name())
	raw, err := medicalcodec.MarshalProcedure(localValue)
	if err != nil {
		return entity.Procedure{}, err
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(raw, &payload); err != nil {
		return entity.Procedure{}, err
	}
	delete(payload, "kind")
	encoded, err := json.Marshal(payload)
	if err != nil {
		return entity.Procedure{}, err
	}
	var data any
	if err := json.Unmarshal(encoded, &data); err != nil {
		return entity.Procedure{}, err
	}
	return entity.RestoreProcedure(localProcedureID(value.Id()), localValue.Kind(), data, value.Sources), nil
}

func localActorRecordFromExternal(value medicalentity.Actor) (entity.ActorRecord, error) {
	meta := entity.ActorMeta{
		ID:         localActorID(value.Id()),
		Kind:       value.Name().Kind(),
		Name:       value.Name().ActorName(),
		OtherNames: value.OtherNames,
		Sources:    value.Sources,
	}
	for _, link := range value.Actions {
		action, err := localActionFromExternal(medicalentity.NewAction(
			medicalentity.ActionID{},
			link.Action,
			nil,
			nil,
			nil,
		))
		if err != nil {
			return entity.ActorRecord{}, err
		}
		meta.Actions = append(meta.Actions, entity.ActorActionLink{
			Action:   action,
			SourceID: link.SourceID,
		})
	}
	return entity.ActorRecord{
		ActorMeta: meta,
		Value:     entity.FromExternalActor(value.Name()),
	}, nil
}

func (s *Store) sourceRefsFor(table, column string, id int64) ([]source.SourceRef, error) {
	rows, err := s.QueryContext(s.ctx,
		`SELECT source_id FROM `+table+` WHERE `+column+` = ? ORDER BY id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck // deferred close
	var ids []source.SourceId
	for rows.Next() {
		var sourceID source.SourceId
		if err := rows.Scan(&sourceID); err != nil {
			return nil, err
		}
		ids = append(ids, sourceID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return s.refs(s.ctx, ids), nil
}

func (s *Store) externalStatus(value entity.Status) (medicalentity.IEStatus, error) {
	raw, err := json.Marshal(value.Value)
	if err != nil {
		return nil, err
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil || payload == nil {
		return externalStatusFromLocal(value)
	}
	switch value.Variant {
	case "have_disease":
		if id, ok := int64Value(payload["disease_id"]); ok && id > 0 {
			item, err := s.DiseaseRepository.Get(typedDiseaseID(id))
			if err != nil {
				return nil, err
			}
			payload["disease"] = item.Name()
			delete(payload, "disease_id")
		}
	case "increased_index", "decreased_index":
		if id, ok := int64Value(payload["index_id"]); ok && id > 0 {
			item, err := s.IndexRepository.Get(typedIndexID(id))
			if err != nil {
				return nil, err
			}
			payload["index"] = map[string]any{"name": item.Name()}
			delete(payload, "index_id")
		} else if text, ok := payload["index"].(string); ok {
			payload["index"] = map[string]any{"name": text}
		}
	case "have_symptom":
		if id, ok := int64Value(payload["symptom_id"]); ok && id > 0 {
			item, err := s.SymptomRepository.Get(typedSymptomID(id))
			if err != nil {
				return nil, err
			}
			payload["symptom"] = map[string]any{"name": item.Name()}
			delete(payload, "symptom_id")
		} else if text, ok := payload["symptom"].(string); ok {
			payload["symptom"] = map[string]any{"name": text}
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
					}
				}
			}
		}
	case "damaged", "inflammation":
		if id, ok := int64Value(payload["actor_id"]); ok && id > 0 {
			item, err := s.ActorRepository.Get(entityActorID(entity.ActorID(id)))
			if err != nil {
				return nil, err
			}
			payload["actor"] = item.Name().ActorName()
			delete(payload, "actor_id")
		}
	case "received_medical_procedure":
		procedureID, ok := int64Value(payload["procedure_id"])
		if !ok {
			procedureID, ok = int64Value(payload["emedical_procedure_id"])
		}
		if ok && procedureID > 0 {
			procedure, found, err := s.GetProcedureByID(entity.ProcedureID(procedureID))
			if err != nil || !found {
				return nil, fmt.Errorf("procedure %d not found", procedureID)
			}
			procedureRaw, err := json.Marshal(procedure.Value)
			if err != nil {
				return nil, err
			}
			localProcedure, err := entity.UnmarshalProcedure(with_kind(procedure.Variant, string(procedureRaw)))
			if err != nil {
				return nil, err
			}
			encoded, err := medicalcodec.MarshalProcedure(localProcedure)
			if err != nil {
				return nil, err
			}
			var procedureValue any
			if err := json.Unmarshal(encoded, &procedureValue); err != nil {
				return nil, err
			}
			payload["medical_procedure"] = procedureValue
			delete(payload, "procedure_id")
			delete(payload, "emedical_procedure_id")
		}
	}
	normalized, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	localValue, err := entity.UnmarshalStatus(with_kind(value.Variant, string(normalized)))
	if err != nil {
		return nil, err
	}
	return entity.ToExternalStatus(localValue), nil
}
