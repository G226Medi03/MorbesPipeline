package cservice

import (
	"bytes"
	"context"
	"fmt"
	medicalcodec "github.com/G226Medi03/MedicalEntity/codec"

	"github.com/G226Medi03/MedicalEntity/ename/eaction"
	"github.com/G226Medi03/MedicalEntity/ename/emedical_procedure"
	"github.com/G226Medi03/MedicalEntity/ename/estatus"
	"github.com/G226Medi03/MedicalEntity/ename/esymptom"
	cgraph "github.com/G226Medi03/MorbesPipeline/graph"
	entity "github.com/G226Medi03/MorbesPipeline/legacy"
	"github.com/G226Medi03/MorbesPipeline/ports"
	source "github.com/G226Medi03/SourceEntity"
)

const RejectionStatusName = "급성면역반응"

// Postprocess loads the graph, applies graph rules, and returns the updated
// in-memory graph.
func Postprocess(ctx context.Context, st ports.Store) (*cgraph.Container, error) {
	c, err := st.LoadGraph()
	if err != nil {
		return nil, err
	}
	rules := []func(context.Context, ports.Store, *cgraph.Container) error{
		transplantRejectionRule,
		increasedIndexActionRule,
		causeStatusResultRule,
	}
	for _, rule := range rules {
		if err := rule(ctx, st, c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

func transplantRejectionRule(ctx context.Context, st ports.Store, c *cgraph.Container) error {
	transplantStatusIDs := make([]entity.StatusID, 0)
	for statusID, statusValue := range c.Statuses {
		received, ok := statusValue.(estatus.ReceivedMedicalProcedure)
		if !ok {
			continue
		}
		if _, ok := received.MedicalProcedure.(emedical_procedure.Transplant); !ok {
			continue
		}
		transplantStatusIDs = append(transplantStatusIDs, statusID)
	}
	if len(transplantStatusIDs) == 0 {
		return nil
	}

	sourceID := source.SourceId{}
	for _, statusID := range transplantStatusIDs {
		if value := firstPostprocessSourceID(c.StatusSourceIDs[statusID]); value.ToI64() > 0 {
			sourceID = value
			break
		}
	}
	if sourceID.ToI64() <= 0 {
		return fmt.Errorf("transplant status source is missing")
	}

	effect := estatus.HaveSymptom{
		Symptom: esymptom.EOtherSymptom{Name: RejectionStatusName},
	}
	effectID, err := st.EnsureStatusIn(c, effect, sourceID)
	if err != nil {
		return err
	}
	actionID, err := st.EnsureActionIn(c, entity.CauseStatusAction{StatusID: int64(effectID)}, sourceID)
	if err != nil {
		return err
	}
	for _, statusID := range transplantStatusIDs {
		if err := st.LinkStatusActionIn(c, statusID, actionID); err != nil {
			return err
		}
		if err := st.LinkActionResultStatusIn(c, actionID, effectID); err != nil {
			return err
		}
	}
	return nil
}

// increasedIndexActionRule ensures every increased_index status is produced by
// exactly one canonical increase_index action. Any other action -> status edge
// is removed, and predecessor statuses are rewired through the canonical action.
func increasedIndexActionRule(ctx context.Context, st ports.Store, c *cgraph.Container) error {
	for statusID, statusValue := range c.Statuses {
		if statusValue.Kind() != "increased_index" {
			continue
		}
		actionValue := entity.CauseStatusAction{StatusID: int64(statusID), Status: statusValue}
		encoded, err := eactionMarshal(actionValue)
		if err != nil {
			continue
		}
		var actionID entity.ActionID
		for candidateID, candidateValue := range c.Actions {
			candidateEncoded, err := eactionMarshal(candidateValue)
			if err == nil && bytes.Equal(candidateEncoded, encoded) {
				actionID = candidateID
				break
			}
		}
		if actionID == 0 {
			sourceID := firstPostprocessSourceID(c.StatusSourceIDs[statusID])
			if sourceID.ToI64() <= 0 {
				return fmt.Errorf("increased_index status source is missing: status=%d", statusID)
			}
			actionID, err = st.EnsureActionIn(c, actionValue, sourceID)
			if err != nil {
				return err
			}
		}

		// An increased_index status may only use the canonical increase_index
		// action as its outgoing action.
		for _, otherActionID := range append([]entity.ActionID(nil), c.StatusActions[statusID]...) {
			if otherActionID == actionID {
				continue
			}
			if err := st.DeleteStatusAction(statusID, otherActionID); err != nil {
				return err
			}
			c.RemoveStatusAction(statusID, otherActionID)
		}

		// Remove every other action that still points at this index status and
		// preserve the predecessor links by moving them to the canonical action.
		for otherActionID, resultStatusIDs := range c.ActionResultStatuses {
			if otherActionID == actionID || !containsStatusID(resultStatusIDs, statusID) {
				continue
			}
			for predecessorStatusID, actionIDs := range c.StatusActions {
				if !containsActionID(actionIDs, otherActionID) {
					continue
				}
				if err := st.LinkStatusActionIn(c, predecessorStatusID, actionID); err != nil {
					return err
				}
				if predecessorStatusID != statusID {
					if err := st.DeleteStatusAction(predecessorStatusID, otherActionID); err != nil {
						return err
					}
					c.RemoveStatusAction(predecessorStatusID, otherActionID)
				}
			}
			if err := st.DeleteActionResultStatus(otherActionID, statusID); err != nil {
				return err
			}
			c.RemoveActionResultStatus(otherActionID, statusID)
		}
		if err := st.LinkActionResultStatusIn(c, actionID, statusID); err != nil {
			return err
		}
	}
	return nil
}

func eactionMarshal(value eaction.IEAction) ([]byte, error) {
	return entity.MarshalAction(value)
}

func firstPostprocessSourceID(groups ...[]source.SourceId) source.SourceId {
	for _, ids := range groups {
		for _, id := range ids {
			if id.ToI64() > 0 {
				return id
			}
		}
	}
	var zero source.SourceId
	return zero
}

func containsStatusID(values []entity.StatusID, target entity.StatusID) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func containsActionID(values []entity.ActionID, target entity.ActionID) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

// causeStatusResultRule ensures every cause_status action exposes its embedded
// status as a result status.
func findStatusIDByValue(c *cgraph.Container, value estatus.IEStatus) entity.StatusID {
	raw, err := medicalcodec.MarshalStatus(value)
	if err != nil {
		return 0
	}
	for statusID, candidate := range c.Statuses {
		if candidate.Kind() != value.Kind() {
			continue
		}
		candidateRaw, err := medicalcodec.MarshalStatus(candidate)
		if err == nil && bytes.Equal(raw, candidateRaw) {
			return statusID
		}
	}
	return 0
}

func causeStatusResultRule(ctx context.Context, st ports.Store, c *cgraph.Container) error {
	for actionID, actionValue := range c.Actions {
		var statusID entity.StatusID
		switch cause := actionValue.(type) {
		case entity.CauseStatusAction:
			if cause.StatusID <= 0 {
				continue
			}
			statusID = entity.StatusID(cause.StatusID)
		case eaction.CauseStatus:
			if cause.Status == nil {
				continue
			}
			statusID = findStatusIDByValue(c, cause.Status)
			if statusID == 0 {
				continue
			}
		default:
			continue
		}
		if _, found := c.Statuses[statusID]; !found {
			continue
		}
		// A cause_status action may only expose its embedded status as a result.
		for _, resultStatusID := range append([]entity.StatusID(nil), c.ActionResultStatuses[actionID]...) {
			if resultStatusID == statusID {
				continue
			}
			if err := st.DeleteActionResultStatus(actionID, resultStatusID); err != nil {
				return err
			}
			c.RemoveActionResultStatus(actionID, resultStatusID)
		}
		if err := st.LinkActionResultStatusIn(c, actionID, statusID); err != nil {
			return err
		}
	}
	return nil
}
