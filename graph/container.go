package cgraph

import (
	"sort"

	"github.com/G226Medi03/MedicalEntity/ename/eaction"
	"github.com/G226Medi03/MedicalEntity/ename/eactor"
	"github.com/G226Medi03/MedicalEntity/ename/estatus"
	entity "github.com/G226Medi03/MorbesPipeline/legacy"
	medicalprocedureoption "github.com/G226Medi03/MorbesPipeline/medicalprocedureoption"
	"github.com/G226Medi03/MorbesPipeline/sliceutil"
	source "github.com/G226Medi03/SourceEntity"
)

type Container struct {
	Actors                     map[entity.ActorID]eactor.IEActor
	Statuses                   map[entity.StatusID]estatus.IEStatus
	StatusSourceIDs            map[entity.StatusID][]source.SourceId
	Actions                    map[entity.ActionID]eaction.IEAction
	ActionSourceIDs            map[entity.ActionID][]source.SourceId
	StatusActions              map[entity.StatusID][]entity.ActionID
	ActionResultStatuses       map[entity.ActionID][]entity.StatusID
	ActionActors               map[entity.ActionID][]entity.ActorID
	ActionActorSourceIDs       map[entity.ActionID][]source.SourceId
	MedicalProcedureOptions    map[entity.StatusID][]medicalprocedureoption.MedicalProcedureOption
	MedicalProcedureOptionsAll []medicalprocedureoption.MedicalProcedureOption
}

// New_Container creates an empty in-memory graph.
func New_Container() *Container {
	return &Container{
		Actors:                  map[entity.ActorID]eactor.IEActor{},
		Statuses:                map[entity.StatusID]estatus.IEStatus{},
		StatusSourceIDs:         map[entity.StatusID][]source.SourceId{},
		Actions:                 map[entity.ActionID]eaction.IEAction{},
		ActionSourceIDs:         map[entity.ActionID][]source.SourceId{},
		StatusActions:           map[entity.StatusID][]entity.ActionID{},
		ActionResultStatuses:    map[entity.ActionID][]entity.StatusID{},
		ActionActors:            map[entity.ActionID][]entity.ActorID{},
		ActionActorSourceIDs:    map[entity.ActionID][]source.SourceId{},
		MedicalProcedureOptions: map[entity.StatusID][]medicalprocedureoption.MedicalProcedureOption{},
	}
}

func (c *Container) FindStatusById(id entity.StatusID) (estatus.IEStatus, bool) {
	value, ok := c.Statuses[id]
	return value, ok
}

func (c *Container) FindActionById(id entity.ActionID) (eaction.IEAction, bool) {
	value, ok := c.Actions[id]
	return value, ok
}

func (c *Container) Get_StatusCauseActions(statusID entity.StatusID) []entity.ActionID {
	var result []entity.ActionID
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

func (c *Container) StatusesCausingAction(actionID entity.ActionID) []entity.StatusID {
	var result []entity.StatusID
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

func (c *Container) AddStatus(id entity.StatusID, value estatus.IEStatus, sourceID source.SourceId) {
	c.Statuses[id] = value
	c.StatusSourceIDs[id] = sliceutil.AppendUnique(c.StatusSourceIDs[id], sourceID)
}

func (c *Container) AddAction(id entity.ActionID, value eaction.IEAction, sourceID source.SourceId) {
	c.Actions[id] = value
	c.ActionSourceIDs[id] = sliceutil.AppendUnique(c.ActionSourceIDs[id], sourceID)
}

func (c *Container) AddStatusAction(statusID entity.StatusID, actionID entity.ActionID) {
	c.StatusActions[statusID] = sliceutil.AppendUnique(c.StatusActions[statusID], actionID)
}

func (c *Container) RemoveStatusAction(statusID entity.StatusID, actionID entity.ActionID) {
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

func (c *Container) AddActionResultStatus(actionID entity.ActionID, statusID entity.StatusID) {
	c.ActionResultStatuses[actionID] = sliceutil.AppendUnique(c.ActionResultStatuses[actionID], statusID)
}

func (c *Container) RemoveActionResultStatus(actionID entity.ActionID, statusID entity.StatusID) {
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
