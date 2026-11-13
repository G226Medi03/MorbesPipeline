package store

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/G226Medi03/MedicalEntity/ename/eactor"
	medicalentity "github.com/G226Medi03/MedicalEntity/entity"
	entity "github.com/G226Medi03/MorbesPipeline/legacy"
	"github.com/G226Medi03/MorbesPipeline/logs"
	"github.com/G226Medi03/SourceEntity"
)

func (s *Store) GetActorRecordByID(id entity.ActorID) (entity.ActorRecord, bool, error) {
	value, err := s.ActorRepository.Get(entityActorID(id))
	if err != nil {
		if err == medicalentity.ErrNotFound {
			return entity.ActorRecord{}, false, nil
		}
		return entity.ActorRecord{}, false, err
	}
	record, err := localActorRecordFromExternal(value)
	if err != nil {
		return entity.ActorRecord{}, false, err
	}
	refs, err := s.sourceRefsFor("actor_sources", "eactor_id", id.ToI64())
	if err != nil {
		return entity.ActorRecord{}, false, err
	}
	record.Sources = refs
	links, err := s.ActionRepository.ActorActions(entityActorID(id))
	if err != nil {
		return entity.ActorRecord{}, false, err
	}
	for _, link := range links {
		action, err := s.getActionShallow(localActionID(link.ActionID))
		if err != nil {
			continue
		}
		action.ResultStatuses = nil
		record.Actions = append(record.Actions, entity.ActorActionLink{Action: action, SourceID: link.SourceID})
	}
	return record, true, nil
}

func (s *Store) GetActorRecordByName(name string) (entity.ActorRecord, bool, error) {
	id, found, err := s.actor_id_by_name(name)
	if err != nil || !found {
		return entity.ActorRecord{}, found, err
	}
	return s.GetActorRecordByID(id)
}

func (s *Store) actor_id_by_name(name string) (entity.ActorID, bool, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, false, nil
	}
	var id entity.ActorID
	err := s.QueryRowContext(s.ctx,
		`SELECT eactor_id FROM actor_other_name WHERE name = ?`, name).Scan(&id)
	if err == nil {
		return id, true, nil
	}
	if err != sql.ErrNoRows {
		return 0, false, err
	}
	err = s.QueryRowContext(s.ctx,
		`SELECT eactor_id FROM eactor WHERE name = ? AND kind <> 'procedure'`, name).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return id, true, nil
}

func (s *Store) saveActorValue(value eactor.IEActor, sources []source.SourceRef) (entity.ActorMeta, error) {
	unlock := s.lockWrite()
	defer unlock()
	if value == nil {
		return entity.ActorMeta{}, fmt.Errorf("actor value is nil")
	}
	if err := entity.ValidateSources(sources); err != nil {
		return entity.ActorMeta{}, err
	}
	name := strings.TrimSpace(value.ActorName())
	if name == "" {
		return entity.ActorMeta{}, fmt.Errorf("actor name이 비어 있습니다")
	}
	kind, err := actorKind(value)
	if err != nil {
		return entity.ActorMeta{}, err
	}
	if existingID, found, err := s.actor_id_by_name(name); err != nil {
		return entity.ActorMeta{}, err
	} else if found {
		for _, ref := range sources {
			if ref.SourceID.ToI64() <= 0 {
				continue
			}
			if _, err := s.ExecContext(s.ctx,
				`INSERT OR IGNORE INTO actor_sources (eactor_id, source_id) VALUES (?, ?)`,
				existingID.ToI64(), ref.SourceID.ToI64()); err != nil {
				return entity.ActorMeta{}, err
			}
		}
		record, found, err := s.GetActorRecordByID(existingID)
		if err != nil || !found {
			return entity.ActorMeta{}, fmt.Errorf("actor reload 실패: %d", existingID)
		}
		if err := s.validateKindPtr("eactor"); err != nil {
			return entity.ActorMeta{}, err
		}
		return record.ActorMeta, nil
	}
	external := entity.ToExternalActor(value)
	saved, err := s.ActorRepository.Insert(external, nil, sourceIDs(sources))
	if err != nil {
		return entity.ActorMeta{}, err
	}
	if err := s.validateKindPtr("eactor"); err != nil {
		return entity.ActorMeta{}, err
	}
	record, err := localActorRecordFromExternal(saved)
	if err != nil {
		return entity.ActorMeta{}, err
	}
	if record.Kind != kind {
		logs.Debug("actor kind mismatch; using existing canonical actor",
			logs.AttrOf("name", name),
			logs.AttrOf("existing_kind", record.Kind),
			logs.AttrOf("requested_kind", kind),
			logs.AttrOf("actor_id", record.ID))
	}
	return record.ActorMeta, nil
}

func actorKind(value eactor.IEActor) (string, error) {
	if value == nil {
		return "", fmt.Errorf("actor value is nil")
	}
	return value.Kind(), nil
}

func (s *Store) SaveOrgan(value eactor.UnknownActor, sources []source.SourceRef) (entity.Organ, error) {
	meta, err := s.saveActorValue(value, sources)
	return entity.Organ{ActorMeta: meta, Value: value}, err
}

func (s *Store) SaveTissue(value eactor.UnknownActor, sources []source.SourceRef) (entity.Tissue, error) {
	meta, err := s.saveActorValue(value, sources)
	return entity.Tissue{ActorMeta: meta, Value: value}, err
}

func (s *Store) SaveMat(value eactor.UnknownActor, sources []source.SourceRef) (entity.Mat, error) {
	meta, err := s.saveActorValue(value, sources)
	return entity.Mat{ActorMeta: meta, Value: value}, err
}

func (s *Store) SaveDevice(value eactor.Device, sources []source.SourceRef) (entity.Device, error) {
	meta, err := s.saveActorValue(value, sources)
	return entity.Device{ActorMeta: meta, Value: value}, err
}

func (s *Store) SaveUnknownActor(value eactor.UnknownActor, sources []source.SourceRef) (entity.UnknownActor, error) {
	meta, err := s.saveActorValue(value, sources)
	return entity.UnknownActor{ActorMeta: meta, Value: value}, err
}

func (s *Store) SaveOtherActor(value eactor.UnknownActor, sources []source.SourceRef) (entity.UnknownActor, error) {
	return s.SaveUnknownActor(value, sources)
}
