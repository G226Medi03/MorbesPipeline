package store

import (
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
)

// Graph is the ename-based medical graph snapshot used by read-side tools.
type Graph struct {
	MedicalProcedureOptionsAll []repo.MedicalProcedureOption
}

// LoadGraph loads the option portion of the medical graph without exposing
// numeric row identifiers to callers.
func (s *StoreContainer) LoadGraph() (Graph, error) {
	rows, err := s.QueryContext(s.ctx, `SELECT key FROM medical_procedure_options ORDER BY key`)
	if err != nil {
		return Graph{}, err
	}
	keys := make([]string, 0)
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			_ = rows.Close()
			return Graph{}, err
		}
		keys = append(keys, key)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return Graph{}, err
	}
	_ = rows.Close()
	entities, err := s.Entities()
	if err != nil {
		return Graph{}, err
	}
	options, err := entities.ListMedicalProcedureOptionsByKeys(keys)
	if err != nil {
		return Graph{}, err
	}
	return Graph{MedicalProcedureOptionsAll: options}, nil
}

// GetMedicalProcedureOptionByKey returns one option by its stable ename key.
func (s *StoreContainer) GetMedicalProcedureOptionByKey(key string) (repo.MedicalProcedureOption, bool, error) {
	entities, err := s.Entities()
	if err != nil {
		return repo.MedicalProcedureOption{}, false, err
	}
	value, found, err := entities.GetMedicalProcedureOptionByKey(key)
	if err != nil || !found {
		return repo.MedicalProcedureOption{}, found, err
	}
	return value, true, nil
}
