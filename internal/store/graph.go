package store

import (
	"encoding/json"
	"fmt"
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"

	medicalprocedureoption "github.com/G226Medi03/MorbesPipeline/internal/domain/procedureoption"
	"github.com/G226Medi03/MorbesPipeline/internal/domain/trait"
	"github.com/G226Medi03/MorbesPipeline/internal/support/ids"
	"github.com/G226Medi03/MorbesPipeline/internal/support/sliceutil"
	cgraph "github.com/G226Medi03/MorbesPipeline/internal/tracing/graph"
	source "github.com/G226Medi03/SourceEntity"
)

func (s *StoreContainer) LoadGraph() (*cgraph.Container, error) {
	rows, err := s.Graph.Load()
	if err != nil {
		return nil, err
	}
	c := cgraph.New_Container()
	for _, row := range rows.Actors {
		value, err := actorValueByKind(row.Kind, row.Name)
		if err != nil {
			return nil, fmt.Errorf("load actor %d: %w", row.ID.ToI64(), err)
		}
		c.Actors[row.ID] = value
	}
	for _, row := range rows.Statuses {
		value, err := s.statusValueByKindPtr(row.Kind, row.Ptr)
		if err != nil {
			continue
		}
		c.Statuses[row.ID] = value
	}
	for _, row := range rows.Actions {
		value, err := s.actionValueByKindPtr(row.Kind, row.Ptr)
		if err != nil {
			continue
		}
		c.Actions[row.ID] = value
	}
	procedureTraitsByID := make(map[ids.ProcedureID][]trait.Trait, len(rows.Procedures))
	for _, procedureRow := range rows.Procedures {
		var traits []trait.Trait
		_ = json.Unmarshal([]byte(procedureRow.TraitsJSON), &traits)
		procedureTraitsByID[procedureRow.ID] = traits
	}
	for _, row := range rows.Options {
		procedureValue, err := s.procedureValueByID(row.ProcedureID.ToI64())
		if err != nil {
			continue
		}
		indicationValues := make([]repo.IEStatus, 0, len(row.IndicationStatusIDs))
		for _, statusID := range row.IndicationStatusIDs {
			if statusValue, ok := c.Statuses[statusID]; ok {
				indicationValues = append(indicationValues, statusValue)
			}
		}
		var optionTraits []trait.Trait
		_ = json.Unmarshal([]byte(row.TraitsJSON), &optionTraits)
		procedureTraits := mergeTraits(procedureTraitsByID[row.ProcedureID], optionTraits)
		var contextTraits []trait.Trait
		_ = json.Unmarshal([]byte(row.ContextTraitsJSON), &contextTraits)
		option := medicalprocedureoption.NewMedicalProcedureOption(
			row.ID,
			row.ProcedureID,
			row.Category,
			procedureTraits,
			contextTraits,
			row.IndicationStatusIDs,
			row.ExpectedEffectStatusIDs,
			nil,
			procedureValue,
			indicationValues,
		)
		c.MedicalProcedureOptionsAll = append(c.MedicalProcedureOptionsAll, option)
		for _, statusID := range row.IndicationStatusIDs {
			if _, ok := c.Statuses[statusID]; ok {
				c.MedicalProcedureOptions[statusID] = append(c.MedicalProcedureOptions[statusID], option)
			}
		}
	}
	for _, row := range rows.StatusSources {
		c.StatusSourceIDs[ids.StatusID(row.EntityID)] = sliceutil.AppendUnique(c.StatusSourceIDs[ids.StatusID(row.EntityID)], row.SourceID)
	}
	for _, row := range rows.ActionSources {
		c.ActionSourceIDs[ids.ActionID(row.EntityID)] = sliceutil.AppendUnique(c.ActionSourceIDs[ids.ActionID(row.EntityID)], row.SourceID)
	}
	for _, row := range rows.ActorActions {
		c.ActionActors[row.ActionID] = sliceutil.AppendUnique(c.ActionActors[row.ActionID], row.ActorID)
		c.ActionActorSourceIDs[row.ActionID] = sliceutil.AppendUnique(c.ActionActorSourceIDs[row.ActionID], row.SourceID)
	}
	for _, row := range rows.StatusActions {
		c.StatusActions[row.StatusID] = sliceutil.AppendUnique(c.StatusActions[row.StatusID], row.ActionID)
	}
	for _, row := range rows.ActionResultStatuses {
		c.ActionResultStatuses[row.ActionID] = sliceutil.AppendUnique(c.ActionResultStatuses[row.ActionID], row.StatusID)
	}
	return c, nil
}

func (s *StoreContainer) EnsureStatusIn(c *cgraph.Container, value repo.IEStatus, sourceID source.SourceId) (ids.StatusID, error) {
	if sourceID.ToI64() <= 0 {
		return 0, fmt.Errorf("status source is required: kind=%s", value.Kind())
	}
	unlock := s.lockWrite()
	defer unlock()
	entityStore, err := s.entities()
	if err != nil {
		return 0, err
	}
	if _, err := entityStore.InsertStatus(value, []source.SourceId{sourceID}); err != nil {
		return 0, err
	}
	id, err := s.statusIDForValue(value)
	if err != nil {
		return 0, err
	}
	c.AddStatus(id, value, sourceID)
	return id, nil
}

func (s *StoreContainer) EnsureActionIn(c *cgraph.Container, value repo.IEAction, sourceID source.SourceId) (ids.ActionID, error) {
	if sourceID.ToI64() <= 0 {
		return 0, fmt.Errorf("action source is required: kind=%s", value.Kind())
	}
	unlock := s.lockWrite()
	defer unlock()
	entityStore, err := s.entities()
	if err != nil {
		return 0, err
	}
	if _, err := entityStore.InsertAction(value, nil, []source.SourceId{sourceID}); err != nil {
		return 0, err
	}
	id, err := s.actionIDForValue(value)
	if err != nil {
		return 0, err
	}
	c.AddAction(id, value, sourceID)
	return id, nil
}

func (s *StoreContainer) LinkStatusActionIn(c *cgraph.Container, statusID ids.StatusID, actionID ids.ActionID) error {
	sourceID := firstSourceIDFromIDs(c.StatusSourceIDs[statusID], c.ActionSourceIDs[actionID])
	if sourceID.ToI64() <= 0 {
		return fmt.Errorf("status-action source is missing: status=%d action=%d", statusID, actionID)
	}
	if err := s.LinkStatusAction(statusID, actionID, sourceID); err != nil {
		return err
	}
	c.AddStatusAction(statusID, actionID)
	return nil
}

func (s *StoreContainer) LinkActionResultStatusIn(c *cgraph.Container, actionID ids.ActionID, statusID ids.StatusID) error {
	sourceID := firstSourceIDFromIDs(c.ActionSourceIDs[actionID], c.StatusSourceIDs[statusID])
	if sourceID.ToI64() <= 0 {
		return fmt.Errorf("action-result source is missing: action=%d status=%d", actionID, statusID)
	}
	if err := s.LinkActionResultStatus(actionID, statusID, sourceID); err != nil {
		return err
	}
	c.AddActionResultStatus(actionID, statusID)
	return nil
}

func (s *StoreContainer) LinkStatusAction(statusID ids.StatusID, actionID ids.ActionID, sourceID source.SourceId) error {
	if sourceID.ToI64() <= 0 {
		return fmt.Errorf("status-action source가 필요합니다")
	}
	unlock := s.lockWrite()
	defer unlock()
	_, err := s.ExecContext(s.ctx,
		`INSERT OR IGNORE INTO status_actions (estatus_id, eaction_id, source_id) VALUES (?, ?, ?)`,
		statusID.ToI64(), actionID.ToI64(), sourceID.ToI64())
	return err
}

func (s *StoreContainer) LinkActionResultStatus(actionID ids.ActionID, statusID ids.StatusID, sourceID source.SourceId) error {
	if sourceID.ToI64() <= 0 {
		return fmt.Errorf("action-result source가 필요합니다")
	}
	unlock := s.lockWrite()
	defer unlock()
	_, err := s.ExecContext(s.ctx,
		`INSERT OR IGNORE INTO action_result_statuses (eaction_id, estatus_id, source_id) VALUES (?, ?, ?)`,
		actionID.ToI64(), statusID.ToI64(), sourceID.ToI64())
	return err
}

func (s *StoreContainer) DeleteStatusAction(statusID ids.StatusID, actionID ids.ActionID) error {
	unlock := s.lockWrite()
	defer unlock()
	_, err := s.ExecContext(s.ctx,
		`DELETE FROM status_actions WHERE estatus_id = ? AND eaction_id = ?`,
		statusID.ToI64(), actionID.ToI64())
	return err
}

func (s *StoreContainer) DeleteActionResultStatus(actionID ids.ActionID, statusID ids.StatusID) error {
	unlock := s.lockWrite()
	defer unlock()
	_, err := s.ExecContext(s.ctx,
		`DELETE FROM action_result_statuses WHERE eaction_id = ? AND estatus_id = ?`,
		actionID.ToI64(), statusID.ToI64())
	return err
}

func firstSourceIDFromIDs(groups ...[]source.SourceId) source.SourceId {
	for _, ids := range groups {
		for _, id := range ids {
			if id.ToI64() > 0 {
				return id
			}
		}
	}
	return source.SourceId{}
}

func mergeTraits(groups ...[]trait.Trait) []trait.Trait {
	seen := make(map[trait.Trait]struct{})
	var result []trait.Trait
	for _, group := range groups {
		for _, value := range group {
			if _, ok := seen[value]; ok {
				continue
			}
			seen[value] = struct{}{}
			result = append(result, value)
		}
	}
	return result
}
