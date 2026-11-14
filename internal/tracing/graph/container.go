package cgraph

import (
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"sort"

	medicalprocedureoption "github.com/G226Medi03/MorbesPipeline/internal/domain/procedureoption"
	source "github.com/G226Medi03/MorbesPipeline/internal/external/sourceentity"
	"github.com/G226Medi03/MorbesPipeline/internal/support/ids"
	"github.com/G226Medi03/MorbesPipeline/internal/support/sliceutil"
)

type Container struct {
	Actors                     map[ids.ActorID]repo.IEActor
	Statuses                   map[ids.StatusID]repo.IEStatus
	StatusSourceIDs            map[ids.StatusID][]source.SourceId
	Actions                    map[ids.ActionID]repo.IEAction
	ActionSourceIDs            map[ids.ActionID][]source.SourceId
	StatusActions              map[ids.StatusID][]ids.ActionID
	ActionResultStatuses       map[ids.ActionID][]ids.StatusID
	ActionActors               map[ids.ActionID][]ids.ActorID
	ActionActorSourceIDs       map[ids.ActionID][]source.SourceId
	MedicalProcedureOptions    map[ids.StatusID][]medicalprocedureoption.MedicalProcedureOption
	MedicalProcedureOptionsAll []medicalprocedureoption.MedicalProcedureOption
}

// New_Container creates an empty in-memory graph.
func New_Container() *Container {
	return &Container{
		Actors:                  map[ids.ActorID]repo.IEActor{},
		Statuses:                map[ids.StatusID]repo.IEStatus{},
		StatusSourceIDs:         map[ids.StatusID][]source.SourceId{},
		Actions:                 map[ids.ActionID]repo.IEAction{},
		ActionSourceIDs:         map[ids.ActionID][]source.SourceId{},
		StatusActions:           map[ids.StatusID][]ids.ActionID{},
		ActionResultStatuses:    map[ids.ActionID][]ids.StatusID{},
		ActionActors:            map[ids.ActionID][]ids.ActorID{},
		ActionActorSourceIDs:    map[ids.ActionID][]source.SourceId{},
		MedicalProcedureOptions: map[ids.StatusID][]medicalprocedureoption.MedicalProcedureOption{},
	}
}

func (c *Container) FindStatusById(id ids.StatusID) (repo.IEStatus, bool) {
	value, ok := c.Statuses[id]
	return value, ok
}

func (c *Container) FindActionById(id ids.ActionID) (repo.IEAction, bool) {
	value, ok := c.Actions[id]
	return value, ok
}

func (c *Container) Get_StatusCauseActions(statusID ids.StatusID) []ids.ActionID {
	var result []ids.ActionID
	for actionID, statusIDs := range c.ActionResultStatuses {
		for _, id := range statusIDs {
			if id == statusID {
				result = append(result, actionID)
				break
			}
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func (c *Container) StatusesCausingAction(actionID ids.ActionID) []ids.StatusID {
	var result []ids.StatusID
	for statusID, actionIDs := range c.StatusActions {
		for _, id := range actionIDs {
			if id == actionID {
				result = append(result, statusID)
				break
			}
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func (c *Container) AddStatus(id ids.StatusID, value repo.IEStatus, sourceID source.SourceId) {
	c.Statuses[id] = value
	c.StatusSourceIDs[id] = sliceutil.AppendUnique(c.StatusSourceIDs[id], sourceID)
}

func (c *Container) AddAction(id ids.ActionID, value repo.IEAction, sourceID source.SourceId) {
	c.Actions[id] = value
	c.ActionSourceIDs[id] = sliceutil.AppendUnique(c.ActionSourceIDs[id], sourceID)
}

func (c *Container) AddStatusAction(statusID ids.StatusID, actionID ids.ActionID) {
	c.StatusActions[statusID] = sliceutil.AppendUnique(c.StatusActions[statusID], actionID)
}

func (c *Container) RemoveStatusAction(statusID ids.StatusID, actionID ids.ActionID) {
	values := c.StatusActions[statusID]
	filtered := values[:0]
	for _, value := range values {
		if value != actionID {
			filtered = append(filtered, value)
		}
	}
	if len(filtered) == 0 {
		delete(c.StatusActions, statusID)
		return
	}
	c.StatusActions[statusID] = filtered
}

func (c *Container) AddActionResultStatus(actionID ids.ActionID, statusID ids.StatusID) {
	c.ActionResultStatuses[actionID] = sliceutil.AppendUnique(c.ActionResultStatuses[actionID], statusID)
}

func (c *Container) RemoveActionResultStatus(actionID ids.ActionID, statusID ids.StatusID) {
	values := c.ActionResultStatuses[actionID]
	filtered := values[:0]
	for _, value := range values {
		if value != statusID {
			filtered = append(filtered, value)
		}
	}
	if len(filtered) == 0 {
		delete(c.ActionResultStatuses, actionID)
		return
	}
	c.ActionResultStatuses[actionID] = filtered
}
