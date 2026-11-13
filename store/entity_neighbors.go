package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// ErrUnknownEntityType is returned when neighbor lookup receives an entity type
// that has no table mapping.
var ErrUnknownEntityType = errors.New("지원하지 않는 entity type")

// EntityNeighbors is the previous/next existing ID in an entity table.
// Zero means that there is no neighbor in that direction.
type EntityNeighbors struct {
	PreviousID int64
	NextID     int64
}

type entityTableInfo struct {
	table    string
	idColumn string
}

// Keep this mapping in the persistence layer. HTTP handlers should only pass
// the normalized entity type; table and column identifiers never come from
// request input.
var entityTables = map[string]entityTableInfo{
	"actor":                    {"actors", "actor_id"},
	"device":                   {"actors", "actor_id"},
	"action":                   {"e_actions", "action_id"},
	"status":                   {"e_statuses", "status_id"},
	"disease":                  {"edisease", "disease_id"},
	"index":                    {"eindex", "index_id"},
	"symptom":                  {"esymptom", "symptom_id"},
	"procedure":                {"eprocedure", "procedure_id"},
	"medical-procedure-option": {"medical_procedure_options", "medical_procedure_option_id"},
	"medical_procedure_option": {"medical_procedure_options", "medical_procedure_option_id"},
}

// GetEntityNeighbors returns the nearest existing IDs before and after id.
// IDs are not assumed to be contiguous.
func (s *Store) GetEntityNeighbors(entityType string, id int64) (EntityNeighbors, bool, error) {
	entityType = strings.ToLower(strings.TrimSpace(entityType))
	info, ok := entityTables[entityType]
	if !ok {
		return EntityNeighbors{}, false, fmt.Errorf("%w: %s", ErrUnknownEntityType, entityType)
	}
	if id <= 0 {
		return EntityNeighbors{}, false, fmt.Errorf("잘못된 entity id: %d", id)
	}

	var existingID int64
	if err := s.QueryRowContext(
		s.ctx,
		fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", info.idColumn, info.table, info.idColumn),
		id,
	).Scan(&existingID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return EntityNeighbors{}, false, nil
		}
		return EntityNeighbors{}, false, fmt.Errorf("%s neighbor 대상 조회 실패: %w", entityType, err)
	}

	var previousID, nextID sql.NullInt64
	query := fmt.Sprintf(`
		SELECT
			(SELECT %[2]s FROM %[1]s WHERE %[2]s < ? ORDER BY %[2]s DESC LIMIT 1),
			(SELECT %[2]s FROM %[1]s WHERE %[2]s > ? ORDER BY %[2]s ASC LIMIT 1)`,
		info.table, info.idColumn,
	)
	if err := s.QueryRowContext(s.ctx, query, id, id).Scan(&previousID, &nextID); err != nil {
		return EntityNeighbors{}, false, fmt.Errorf("%s neighbor 조회 실패: %w", entityType, err)
	}

	return EntityNeighbors{
		PreviousID: previousID.Int64,
		NextID:     nextID.Int64,
	}, true, nil
}
