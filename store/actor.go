package store

import (
	"fmt"
	"strings"

	entity "github.com/G226Medi03/MorbesPipeline/legacy"
	"github.com/G226Medi03/SourceEntity"
)

func (s *Store) AddActorOtherName(actorID entity.ActorID, name string, sourceID source.SourceId) error {
	unlock := s.lockWrite()
	defer unlock()
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	_, err := s.ExecContext(s.ctx,
		`INSERT OR IGNORE INTO actor_other_name (eactor_id, name, is_primary, source_id) VALUES (?, ?, 0, ?)`,
		actorID.ToI64(), name, sourceID.ToI64())
	return err
}

func (s *Store) LinkActorAction(actorID entity.ActorID, actionID entity.ActionID, sourceID source.SourceId) error {
	if sourceID.ToI64() <= 0 {
		return fmt.Errorf("actor-action source가 필요합니다")
	}
	unlock := s.lockWrite()
	defer unlock()
	_, err := s.ExecContext(s.ctx,
		`INSERT OR IGNORE INTO actor_actions (eactor_id, eaction_id, source_id) VALUES (?, ?, ?)`,
		actorID.ToI64(), actionID.ToI64(), sourceID.ToI64())
	return err
}
