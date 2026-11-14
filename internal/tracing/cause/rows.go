package ccause

import (
	source_entity "github.com/G226Medi03/MorbesPipeline/internal/external/sourceentity"
	"github.com/G226Medi03/MorbesPipeline/internal/support/ids"
)

type actionRow struct {
	ID        ids.ActionID
	Kind      string
	Data      string
	SourceIDs []source_entity.SourceId
}

type statusRow struct {
	ID   ids.StatusID
	Kind string
	Data string
}
