package store

import (
	"encoding/json"

	centitylabel "github.com/G226Medi03/MorbesPipeline/label"
	entity "github.com/G226Medi03/MorbesPipeline/legacy"
)

// ProcedureLabelByID hydrates a procedure row and returns its display label.
func (s *Store) ProcedureLabelByID(id entity.ProcedureID) (string, bool) {
	data, ok := s.hydrateProcedureData(int64(id))
	if !ok {
		return "", false
	}
	encoded, err := json.Marshal(data)
	if err != nil {
		return "", false
	}
	value, err := entity.UnmarshalProcedure(encoded)
	if err != nil {
		return "", false
	}
	return centitylabel.ProcedureLabel(value), true
}
