package store

import (
	"database/sql"

	"github.com/G226Medi03/MedicalEntity/ename/eactor"
	entity "github.com/G226Medi03/MorbesPipeline/legacy"
	"github.com/G226Medi03/SourceEntity"
)

func (s *Store) GetDeviceByID(id entity.DeviceID) (entity.Device, error) {
	record, found, err := s.GetActorRecordByID(entity.ActorID(id))
	if err != nil {
		return entity.Device{}, err
	}
	if !found {
		return entity.Device{}, sql.ErrNoRows
	}
	value, ok := record.Value.(eactor.Device)
	if !ok {
		return entity.Device{}, sql.ErrNoRows
	}
	return entity.Device{ActorMeta: record.ActorMeta, Value: value}, nil
}

func (s *Store) LinkDeviceAction(deviceID entity.DeviceID, actionID entity.ActionID, sourceID source.SourceId) error {
	return s.LinkActorAction(entity.ActorID(deviceID), actionID, sourceID)
}
