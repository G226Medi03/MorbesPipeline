package ccause

import (
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"github.com/G226Medi03/MorbesPipeline/internal/support/ids"
	source_entity "github.com/G226Medi03/SourceEntity"
)

type actionRow struct {
	ID        ids.ActionID
	Value     repo.IEAction
	SourceIDs []source_entity.SourceId
}

type statusRow struct {
	ID    ids.StatusID
	Value repo.IEStatus
}
