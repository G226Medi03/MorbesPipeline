package store

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/G226Medi03/MedicalEntity/ename/eactor"
	eprocedure "github.com/G226Medi03/MedicalEntity/ename/emedical_procedure"
	"github.com/G226Medi03/MorbesPipeline/entityconv"
	codec "github.com/G226Medi03/MorbesPipeline/legacy"
	"github.com/G226Medi03/MorbesPipeline/trait"
	"github.com/G226Medi03/SourceEntity"
)

// SaveProcedureValue canonicalizes a raw medical_procedure object from the
// extraction layer and persists the resulting canonical procedure. The status
// store uses this when a direct caller provides a nested procedure value.
func (s *Store) SaveProcedureValue(raw any, traits []trait.Trait, sources []source.SourceRef) (int64, error) {
	if raw == nil {
		return 0, fmt.Errorf("medical_procedure가 비어 있습니다")
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return 0, err
	}
	value, err := codec.UnmarshalOrParseProcedure(data)
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
			return 0, fmt.Errorf("transplant target가 비어 있습니다")
		}
		switch target := typed.Target.(type) {
		case eactor.Device:
			saved, err := s.SaveDevice(target, sources)
			if err != nil {
				return 0, err
			}
			payload = map[string]any{"target_kind": "device", "target_id": saved.ID}
		case eactor.UnknownActor:
			saved, err := s.SaveUnknownActor(target, sources)
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
			return 0, fmt.Errorf("infuse material이 비어 있습니다")
		}
		meta, err := s.saveActorValue(typed.Material, sources)
		if err != nil {
			return 0, err
		}
		payload = map[string]any{"material_id": meta.ID, "route": string(typed.Route)}
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
			saved, err := s.SaveDevice(target, sources)
			if err != nil {
				return 0, err
			}
			payload = map[string]any{"target_id": saved.ID, "target_kind": "device"}
		case eactor.UnknownActor:
			saved, err := s.SaveUnknownActor(target, sources)
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
	id, err := s.SaveProcedureWithTraits(value.Kind(), string(encoded), entityconv.EntityTraits(intrinsicTraits), sources)
	return int64(id), err
}
