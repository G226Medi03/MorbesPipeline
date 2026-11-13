package store

import (
	"encoding/json"
	"fmt"

	"github.com/G226Medi03/MedicalEntity/ename/eaction"
	"github.com/G226Medi03/MedicalEntity/ename/estatus"
	cgraph "github.com/G226Medi03/MorbesPipeline/graph"
	entity "github.com/G226Medi03/MorbesPipeline/legacy"
	"github.com/G226Medi03/MorbesPipeline/sliceutil"
	medicalprocedureoption "github.com/G226Medi03/MorbesPipeline/store/medicalprocedureoption"
	"github.com/G226Medi03/SourceEntity"
)

func (s *Store) LoadGraph() (*cgraph.Container, error) {
	if err := s.validateKindPtr(); err != nil {
		return nil, err
	}
	rows, err := s.GraphRepository.Load()
	if err != nil {
		return nil, err
	}
	c := cgraph.New_Container()
	for _, row := range rows.Actors {
		value, err := s.ActorRepository.Get(entityActorID(row.ID))
		if err != nil {
			return nil, fmt.Errorf("load actor %d: %w", row.ID.ToI64(), err)
		}
		c.Actors[row.ID] = entity.FromExternalActor(value.Name())
	}
	for _, row := range rows.Statuses {
		value, err := s.StatusRepository.Get(entityStatusID(row.ID))
		if err != nil {
			continue
		}
		c.Statuses[row.ID] = value.Name()
	}
	for _, row := range rows.Actions {
		value, err := s.ActionRepository.Get(entityActionID(row.ID))
		if err != nil {
			continue
		}
		c.Actions[row.ID] = s.localGraphAction(row.ID, value.Name())
	}
	procedureTraitsByID := make(map[entity.ProcedureID][]entity.Trait, len(rows.Procedures))
	for _, procedureRow := range rows.Procedures {
		var traits []entity.Trait
		_ = json.Unmarshal([]byte(procedureRow.TraitsJSON), &traits)
		procedureTraitsByID[procedureRow.ID] = traits
	}
	for _, row := range rows.Options {
		procedureValue, ok, procedureErr := s.GetProcedureByID(row.ProcedureID)
		if procedureErr != nil || !ok {
			continue
		}
		encodedProcedure, _ := json.Marshal(procedureValue)
		procedure, err := entity.UnmarshalProcedure(encodedProcedure)
		if err != nil {
			continue
		}
		indicationValues := make([]estatus.IEStatus, 0, len(row.IndicationStatusIDs))
		for _, statusID := range row.IndicationStatusIDs {
			if statusValue, ok := c.Statuses[statusID]; ok {
				indicationValues = append(indicationValues, statusValue)
			}
		}
		var optionTraits []entity.Trait
		_ = json.Unmarshal([]byte(row.TraitsJSON), &optionTraits)
		procedureTraits := mergeEntityTraits(procedureTraitsByID[row.ProcedureID], optionTraits)
		var contextTraits []entity.Trait
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
			procedure,
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
		c.StatusSourceIDs[entity.StatusID(row.EntityID)] = sliceutil.AppendUnique(c.StatusSourceIDs[entity.StatusID(row.EntityID)], row.SourceID)
	}
	for _, row := range rows.ActionSources {
		c.ActionSourceIDs[entity.ActionID(row.EntityID)] = sliceutil.AppendUnique(c.ActionSourceIDs[entity.ActionID(row.EntityID)], row.SourceID)
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

func (s *Store) EnsureStatus(value estatus.IEStatus, sourceID source.SourceId) (entity.StatusID, error) {
	if sourceID.ToI64() <= 0 {
		return 0, fmt.Errorf("status source is required: kind=%s", value.Kind())
	}
	unlock := s.lockWrite()
	defer unlock()
	saved, err := s.StatusRepository.Insert(entity.ToExternalStatus(value), []source.SourceId{sourceID})
	if err != nil {
		return 0, err
	}
	if err := s.validateKindPtr("estatus"); err != nil {
		return 0, err
	}
	return localStatusID(saved.Id()), nil
}

func (s *Store) EnsureAction(value eaction.IEAction, sourceID source.SourceId) (entity.ActionID, error) {
	if sourceID.ToI64() <= 0 {
		return 0, fmt.Errorf("action source is required: kind=%s", value.Kind())
	}
	unlock := s.lockWrite()
	defer unlock()
	if cause, ok := value.(entity.CauseStatusAction); ok && cause.StatusID > 0 {
		action, err := s.saveCauseStatusAction(
			entity.RestoreAction(0, "cause_status", map[string]any{"status_id": cause.StatusID}, []source.SourceRef{{SourceID: sourceID}}),
			entity.StatusID(cause.StatusID),
		)
		if err != nil {
			return 0, err
		}
		if err := s.validateKindPtr("eaction"); err != nil {
			return 0, err
		}
		return action.Id(), nil
	}
	saved, err := s.ActionRepository.Insert(entity.ToExternalAction(value), nil, []source.SourceId{sourceID})
	if err != nil {
		return 0, err
	}
	if err := s.validateKindPtr("eaction"); err != nil {
		return 0, err
	}
	return localActionID(saved.Id()), nil
}

func (s *Store) EnsureStatusIn(c *cgraph.Container, value estatus.IEStatus, sourceID source.SourceId) (entity.StatusID, error) {
	id, err := s.EnsureStatus(value, sourceID)
	if err != nil {
		return 0, err
	}
	c.AddStatus(id, value, sourceID)
	return id, nil
}

func (s *Store) EnsureActionIn(c *cgraph.Container, value eaction.IEAction, sourceID source.SourceId) (entity.ActionID, error) {
	id, err := s.EnsureAction(value, sourceID)
	if err != nil {
		return 0, err
	}
	c.AddAction(id, value, sourceID)
	return id, nil
}

func (s *Store) LinkStatusActionIn(c *cgraph.Container, statusID entity.StatusID, actionID entity.ActionID) error {
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

func (s *Store) LinkActionResultStatusIn(c *cgraph.Container, actionID entity.ActionID, statusID entity.StatusID) error {
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

func with_kind(kind, data string) []byte {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal([]byte(data), &obj); err != nil {
		return []byte(data)
	}
	kindJSON, err := json.Marshal(kind)
	if err != nil {
		return []byte(data)
	}
	obj["kind"] = kindJSON
	payload, err := json.Marshal(obj)
	if err != nil {
		return []byte(data)
	}
	return payload
}

func (s *Store) localGraphAction(id entity.ActionID, external eaction.IEAction) eaction.IEAction {
	local := entity.FromExternalAction(external)
	cause, ok := local.(eaction.CauseStatus)
	if !ok {
		return local
	}
	var ptr int64
	if err := s.QueryRowContext(s.ctx,
		`SELECT ptr FROM eaction WHERE eaction_id = ? AND kind = 'cause_status'`, id.ToI64()).Scan(&ptr); err != nil {
		return local
	}
	var statusID int64
	if err := s.QueryRowContext(s.ctx,
		`SELECT estatus_id FROM action_cause_status WHERE id = ?`, ptr).Scan(&statusID); err != nil {
		return local
	}
	return entity.CauseStatusAction{StatusID: statusID, Status: cause.Status}
}

func mergeEntityTraits(groups ...[]entity.Trait) []entity.Trait {
	seen := make(map[entity.Trait]struct{})
	var result []entity.Trait
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
