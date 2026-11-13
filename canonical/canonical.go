package canonical

import (
	"encoding/json"
	"fmt"
	"strings"

	medicalcodec "github.com/G226Medi03/MedicalEntity/codec"
	"github.com/G226Medi03/MedicalEntity/ename/eactor"
	eprocedure "github.com/G226Medi03/MedicalEntity/ename/emedical_procedure"
	"github.com/G226Medi03/MedicalEntityParser/parser"
	"github.com/G226Medi03/MorbesPipeline/entityconv"
	entity "github.com/G226Medi03/MorbesPipeline/legacy"
	"github.com/G226Medi03/MorbesPipeline/ports"
	"github.com/G226Medi03/MorbesPipeline/trait"
	source "github.com/G226Medi03/SourceEntity"
)

// Canonical converts extracted values into persisted canonical values.
type Canonical struct {
	store ports.Store
}

func New(st ports.Store) *Canonical {
	return &Canonical{store: st}
}

// CanonicalizeStatus replaces entity names inside a status value with
// canonical IDs. The persisted data therefore contains IDs only; value fields
// such as pain aspects and post-procedure periods stay as values.
func (c *Canonical) CanonicalizeStatus(kind string, value any, sources []source.SourceRef) (any, error) {
	payload, err := statusValueMap(value)
	if err != nil || len(payload) == 0 {
		return value, err
	}
	switch kind {
	case "have_disease":
		if _, ok := payload["disease_id"]; ok {
			return value, nil
		}
		name := statusFieldText(payload["disease"], "name")
		if name == "" {
			return value, nil
		}
		saved, err := c.store.SaveDisease(name, sources)
		if err != nil {
			return nil, err
		}
		return map[string]any{"disease_id": saved.Id().ToI64()}, nil
	case "increased_index", "decreased_index":
		if _, ok := payload["index_id"]; ok {
			return value, nil
		}
		name := statusFieldText(payload["index"], "name")
		if name == "" {
			return value, nil
		}
		saved, err := c.store.SaveIndex(name, sources)
		if err != nil {
			return nil, err
		}
		return map[string]any{"index_id": saved.Id().ToI64()}, nil
	case "have_symptom":
		if _, ok := payload["symptom_id"]; ok {
			return value, nil
		}
		symptomPayload, _ := payload["symptom"].(map[string]any)
		if symptomPayload == nil {
			if raw, ok := payload["symptom"]; ok {
				symptomPayload, _ = statusValueMap(raw)
			}
		}
		if symptomPayload == nil {
			return value, nil
		}
		symptomKind, _ := symptomPayload["kind"].(string)
		if symptomKind == "symptom_code" {
			name := statusFieldText(symptomPayload["code"], "name")
			if name == "" {
				return value, nil
			}
			saved, err := c.store.SaveSymptom(name, sources)
			if err != nil {
				return nil, err
			}
			return map[string]any{"symptom_id": saved.Id().ToI64()}, nil
		}
		if symptomKind == "pain" {
			saved, err := c.store.SaveSymptom("??", sources)
			if err != nil {
				return nil, err
			}
			return map[string]any{"symptom_id": saved.Id().ToI64(), "pain": symptomPayload}, nil
		}
		return value, nil
	case "received_medical_procedure":
		if _, ok := payload["procedure_id"]; ok {
			return value, nil
		}
		procedureID, err := c.canonicalizeMedicalProcedure(payload["medical_procedure"], nil, sources)
		if err != nil {
			return nil, err
		}
		result := map[string]any{"procedure_id": procedureID}
		if period, ok := payload["post_procedure_period"]; ok {
			result["post_procedure_period"] = period
		}
		return result, nil
	case "damaged", "inflammation":
		if _, ok := payload["actor_id"]; ok {
			return value, nil
		}
		name := statusFieldText(payload["actor"], "name")
		if name == "" {
			return value, nil
		}
		actorValue, err := parser.ParseActor(name)
		if err != nil {
			return nil, err
		}
		actorID, err := saveActorValue(c.store, actorValue, sources)
		if err != nil {
			return nil, err
		}
		return map[string]any{"actor_id": actorID}, nil
	default:
		return value, nil
	}
}

func (c *Canonical) CanonicalizeStatusEntity(kind string, value any, sources []source.SourceRef) (entity.Status, error) {
	canonical, err := c.CanonicalizeStatus(kind, value, sources)
	if err != nil {
		return entity.NewStatus("", nil, nil), err
	}
	return c.store.SaveStatus(entity.NewStatus(kind, canonical, sources))
}

func (c *Canonical) CanonicalizeProcedure(raw any, traits []trait.Trait, sources []source.SourceRef) (entity.ProcedureID, error) {
	return c.canonicalizeMedicalProcedure(raw, traits, sources)
}

func (c *Canonical) canonicalizeMedicalProcedure(raw any, traits []trait.Trait, sources []source.SourceRef) (entity.ProcedureID, error) {
	if c.store == nil {
		return 0, fmt.Errorf("procedure store is not configured")
	}
	if raw == nil {
		return 0, medicalcodec.ErrTransplantTargetMissing
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return 0, err
	}
	value, err := entity.UnmarshalOrParseProcedure(data)
	if err != nil {
		return 0, err
	}
	intrinsicTraits := trait.IntrinsicTraits(traits)
	if len(intrinsicTraits) == 0 {
		intrinsicTraits = trait.DefaultIntrinsicTraits(trait.ProcedureType(value.Kind()))
	}
	var payload map[string]any
	switch typed := value.(type) {
	case eprocedure.Transplant:
		targetName := strings.TrimSpace(typed.Target.ActorName())
		if targetName == "" {
			return 0, medicalcodec.ErrTransplantTargetMissing
		}
		switch target := typed.Target.(type) {
		case eactor.Device:
			saved, err := c.store.SaveDevice(target, sources)
			if err != nil {
				return 0, err
			}
			payload = map[string]any{"target_kind": "device", "target_id": saved.ID}
		case eactor.UnknownActor:
			saved, err := c.store.SaveUnknownActor(target, sources)
			if err != nil {
				return 0, err
			}
			payload = map[string]any{"target_kind": "unknown_actor", "target_id": saved.ID}
		default:
			payload = map[string]any{"target_kind": "unknown", "target_name": targetName}
		}
	case eprocedure.Infuse:
		material := strings.TrimSpace(typed.Material.ActorName())
		if material == "" {
			return 0, medicalcodec.ErrInfuseMaterialMissing
		}
		materialID, err := saveActorValue(c.store, typed.Material, sources)
		if err != nil {
			return 0, err
		}
		payload = map[string]any{"material_id": materialID, "route": string(typed.Route)}
	case eprocedure.Surgery:
		name := strings.TrimSpace(typed.Name)
		if name == "" {
			name = "수술"
		}
		payload = map[string]any{"name": name}
	case eprocedure.Insertion:
		device := strings.TrimSpace(typed.Device.Name)
		site := ""
		if typed.Site != nil {
			site = strings.TrimSpace(typed.Site.ActorName())
		}
		payload = map[string]any{"name": device, "tube": device, "site": site}
	case eprocedure.Cleaning:
		switch target := typed.Target.(type) {
		case eactor.Device:
			saved, err := c.store.SaveDevice(target, sources)
			if err != nil {
				return 0, err
			}
			payload = map[string]any{"target_id": saved.ID, "target_kind": "device"}
		case eactor.UnknownActor:
			saved, err := c.store.SaveUnknownActor(target, sources)
			if err != nil {
				return 0, err
			}
			payload = map[string]any{"target_id": saved.ID, "target_kind": "unknown_actor"}
		default:
			payload = map[string]any{"target_kind": "unknown"}
		}
	case eprocedure.DiagnosticTest:
		payload = map[string]any{"name": strings.TrimSpace(typed.Name)}
	case eprocedure.Education:
		payload = map[string]any{"name": strings.TrimSpace(typed.Name), "topic": strings.TrimSpace(typed.Topic)}
	case eprocedure.Instruction:
		payload = map[string]any{"name": strings.TrimSpace(typed.Name), "action": strings.TrimSpace(typed.Action), "duration": strings.TrimSpace(typed.Duration), "position": strings.TrimSpace(typed.Position)}
	case eprocedure.UnknownMedicalProcedure:
		payload = map[string]any{"name": strings.TrimSpace(typed.Name)}
	default:
		return 0, fmt.Errorf("unsupported procedure type: %T", value)
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}
	return c.store.SaveProcedureWithTraits(value.Kind(), string(encoded), entityconv.EntityTraits(intrinsicTraits), sources)
}

func (c *Canonical) CanonicalizeCauseStatusValue(value any, sources []source.SourceRef) (any, error) {
	payload, ok := value.(map[string]any)
	if !ok {
		return value, nil
	}
	if _, ok := payload["status_id"]; ok {
		return value, nil
	}
	statusValue, ok := payload["status"].(map[string]any)
	if !ok {
		return value, nil
	}
	kind, _ := statusValue["kind"].(string)
	if kind == "" {
		return value, nil
	}
	copyValue := make(map[string]any, len(statusValue))
	for key, item := range statusValue {
		if key != "kind" {
			copyValue[key] = item
		}
	}
	saved, err := c.CanonicalizeStatusEntity(kind, copyValue, sources)
	if err != nil {
		return nil, err
	}
	return map[string]any{"status_id": int64(saved.Id())}, nil
}

func statusValueMap(value any) (map[string]any, error) {
	if value == nil {
		return nil, nil
	}
	if payload, ok := value.(map[string]any); ok {
		return payload, nil
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var payload map[string]any
	if err := json.Unmarshal(encoded, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func statusFieldText(value any, key string) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case map[string]any:
		if key != "" {
			if text, ok := typed[key].(string); ok {
				return strings.TrimSpace(text)
			}
		}
		for _, candidate := range []string{"name", "code", "value"} {
			if text, ok := typed[candidate].(string); ok {
				return strings.TrimSpace(text)
			}
		}
	}
	return ""
}

func saveActorValue(st ports.Store, value eactor.IEActor, sources []source.SourceRef) (entity.ActorID, error) {
	switch value.Kind() {
	case "device":
		saved, err := st.SaveDevice(eactor.Device{Name: value.ActorName()}, sources)
		return saved.ID, err
	default:
		saved, err := st.SaveUnknownActor(eactor.UnknownActor{Name: value.ActorName()}, sources)
		return saved.ID, err
	}
}
