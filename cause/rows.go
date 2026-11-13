package ccause

import (
	entity "github.com/G226Medi03/MorbesPipeline/legacy"
	source_entity "github.com/G226Medi03/SourceEntity"
)

type actionRow struct {
	ID        entity.ActionID
	Kind      string
	Data      string
	SourceIDs []source_entity.SourceId
}

type statusRow struct {
	ID   entity.StatusID
	Kind string
	Data string
}
