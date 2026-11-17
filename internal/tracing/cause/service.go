package ccause

import (
	"context"
	"fmt"
	"sort"
	"strings"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	entitykey "github.com/G226Medi03/MorbesPipeline/internal/domain/entitykey"
	centitylabel "github.com/G226Medi03/MorbesPipeline/internal/domain/label"
	"github.com/G226Medi03/MorbesPipeline/internal/support/ids"
	cgraph "github.com/G226Medi03/MorbesPipeline/internal/tracing/graph"
	source_entity "github.com/G226Medi03/SourceEntity"
)

type Service struct {
	container *cgraph.Container
}

func New_ContainerService(container *cgraph.Container) *Service {
	return &Service{container: container}
}

// TraceContext is Trace with cancellation/timeout support.
func (s *Service) TraceContext(ctx context.Context, observed []repo.IEStatus, maxDepth int) ([]Candidate, error) {
	if maxDepth <= 0 {
		maxDepth = 3
	}
	seen := make(map[string]Candidate)
	for _, status := range observed {
		statusIDs, err := s.find_status_ids(status)
		if err != nil {
			return nil, err
		}
		for _, id := range statusIDs {
			node, ok := s.status_node(id)
			if !ok {
				continue
			}
			start_path := []TraceNode{node}
			if err := s.trace_from(ctx, id, 0, maxDepth, start_path, map[ids.StatusID]struct{}{id: {}}, map[ids.ActorID]struct{}{}, seen); err != nil {
				return nil, err
			}
		}
	}

	result := make([]Candidate, 0, len(seen))
	for _, candidate := range seen {
		sort.Slice(candidate.SourceIDs, func(i, j int) bool { return candidate.SourceIDs[i].ToI64() < candidate.SourceIDs[j].ToI64() })
		result = append(result, candidate)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Depth != result[j].Depth {
			return result[i].Depth < result[j].Depth
		}
		return result[i].CauseKey < result[j].CauseKey
	})
	return result, nil
}

func (s *Service) trace_from(ctx context.Context, status_id ids.StatusID, depth, max_depth int, path []TraceNode, path_set map[ids.StatusID]struct{}, actor_path_set map[ids.ActorID]struct{}, seen map[string]Candidate) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if depth >= max_depth {
		return nil
	}
	actions, err := s.actions_causing(status_id)
	if err != nil {
		return err
	}
	for _, action := range actions {
		if err := ctx.Err(); err != nil {
			return err
		}
		causes, err := s.statuses_causing(action.ID)
		if err != nil {
			return err
		}
		action_key := entitykey.Action(action.Value)
		action_node := TraceNode{
			Type:  "action",
			ID:    action.ID.ToI64(),
			Key:   action_key,
			Label: s.action_label(action),
		}
		for _, cause := range causes {
			if err := ctx.Err(); err != nil {
				return err
			}
			if _, ok := path_set[cause.ID]; ok {
				continue
			}
			cause_key := entitykey.Status(cause.Value)
			cause_node := TraceNode{
				Type:  "status",
				ID:    cause.ID.ToI64(),
				Key:   cause_key,
				Label: centitylabel.StatusLabelValue(cause.Value),
			}
			candidate_path := make([]TraceNode, 0, len(path)+2)
			candidate_path = append(candidate_path, path...)
			candidate_path = append(candidate_path, action_node, cause_node)

			key := cause_key + ":" + action_key
			candidate := Candidate{
				CauseStatusID:  cause.ID,
				CauseStatusIDs: []ids.StatusID{cause.ID},
				CauseKind:      cause.Value.Kind(),
				CauseKey:       cause_key,
				CauseLabel:     cause_node.Label,
				ViaActionID:    action.ID,
				ViaActionIDs:   []ids.ActionID{action.ID},
				ViaActionKind:  action.Value.Kind(),
				ViaActionKey:   action_key,
				Depth:          depth + 1,
				SourceIDs:      append([]source_entity.SourceId(nil), action.SourceIDs...),
				Path:           candidate_path,
			}
			if existing, ok := seen[key]; ok {
				existing.SourceIDs = append_unique_ids(existing.SourceIDs, candidate.SourceIDs...)
				existing.CauseStatusIDs = append_unique_ids(existing.CauseStatusIDs, candidate.CauseStatusIDs...)
				existing.ViaActionIDs = append_unique_ids(existing.ViaActionIDs, candidate.ViaActionIDs...)
				// For the same cause/action, prefer the shortest path.
				// Existing source/ID aggregates are preserved when replacing
				// a longer candidate with a shorter one.
				if len(candidate.Path) < len(existing.Path) {
					candidate.SourceIDs = existing.SourceIDs
					candidate.CauseStatusIDs = existing.CauseStatusIDs
					candidate.ViaActionIDs = existing.ViaActionIDs
					seen[key] = candidate
				} else {
					seen[key] = existing
				}
			} else {
				seen[key] = candidate
			}
			next_path_set := make(map[ids.StatusID]struct{}, len(path_set)+1)
			for id := range path_set {
				next_path_set[id] = struct{}{}
			}
			next_path_set[cause.ID] = struct{}{}
			if err := s.trace_from(ctx, cause.ID, depth+1, max_depth, candidate_path, next_path_set, actor_path_set, seen); err != nil {
				return err
			}
		}

		for _, actor_id := range s.container.ActionActors[ids.ActionID(action.ID)] {
			if _, ok := actor_path_set[actor_id]; ok {
				continue
			}
			actor_node, ok := s.actor_node(actor_id)
			if !ok {
				continue
			}
			candidate_path := make([]TraceNode, 0, len(path)+2)
			candidate_path = append(candidate_path, path...)
			candidate_path = append(candidate_path, action_node, actor_node)
			cause_key := actor_node.Key
			key := cause_key + ":" + action_key
			candidate := Candidate{
				CauseType:     "actor",
				CauseActorID:  actor_id,
				CauseActorIDs: []ids.ActorID{actor_id},
				CauseKind:     "actor",
				CauseKey:      cause_key,
				CauseLabel:    actor_node.Label,
				ViaActionID:   action.ID,
				ViaActionIDs:  []ids.ActionID{action.ID},
				ViaActionKind: action.Value.Kind(),
				ViaActionKey:  action_key,
				Depth:         depth + 1,
				SourceIDs:     append_unique_ids(append([]source_entity.SourceId(nil), action.SourceIDs...), s.container.ActionActorSourceIDs[ids.ActionID(action.ID)]...),
				Path:          candidate_path,
			}
			if existing, ok := seen[key]; ok {
				existing.SourceIDs = append_unique_ids(existing.SourceIDs, candidate.SourceIDs...)
				existing.CauseActorIDs = append_unique_ids(existing.CauseActorIDs, candidate.CauseActorIDs...)
				existing.ViaActionIDs = append_unique_ids(existing.ViaActionIDs, candidate.ViaActionIDs...)
				if len(candidate.Path) < len(existing.Path) {
					candidate.SourceIDs = existing.SourceIDs
					candidate.CauseActorIDs = existing.CauseActorIDs
					candidate.ViaActionIDs = existing.ViaActionIDs
					seen[key] = candidate
				} else {
					seen[key] = existing
				}
			} else {
				seen[key] = candidate
			}
		}

	}
	return nil
}

func (s *Service) find_status_ids(status repo.IEStatus) ([]ids.StatusID, error) {
	if s.container == nil {
		return nil, fmt.Errorf("morbes graph container 가 없습니다")
	}
	expected := entitykey.Status(status)

	var result []ids.StatusID
	for id, value := range s.container.Statuses {
		if value.Kind() != status.Kind() {
			continue
		}
		if entitykey.Status(value) == expected {
			result = append(result, id)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result, nil
}

func (s *Service) status_node(id ids.StatusID) (TraceNode, bool) {
	if s.container == nil {
		return TraceNode{}, false
	}
	value, ok := s.container.FindStatusById(id)
	if !ok {
		return TraceNode{}, false
	}
	return TraceNode{
		Type:  "status",
		ID:    id.ToI64(),
		Key:   entitykey.Status(value),
		Label: centitylabel.StatusLabelValue(value),
	}, true
}

func (s *Service) action_label(action actionRow) string {
	if cause, ok := action.Value.(repo.CauseStatus); ok {
		if cause.Status == nil {
			return "원인 상태 정보 없음"
		}
		label := centitylabel.StatusLabelValue(cause.Status)
		switch cause.Status.Kind() {
		case "increased_index", "decreased_index":
			name := strings.TrimSpace(strings.TrimSuffix(strings.TrimSuffix(label, " 증가"), " 감소"))
			return "cause_status: " + cause.Status.Kind() + " (" + name + ")"
		default:
			return "cause_status: " + label
		}
	}
	return action.Value.Kind() + ":" + centitylabel.ActionLabelValue(action.Value)
}

func (s *Service) actor_node(id ids.ActorID) (TraceNode, bool) {
	if s.container == nil {
		return TraceNode{}, false
	}
	value, ok := s.container.Actors[id]
	if !ok {
		return TraceNode{}, false
	}
	name := strings.TrimSpace(value.ActorName())
	if name == "" {
		return TraceNode{}, false
	}
	node_type := "actor"
	prefix := "actor:"
	if _, isDevice := value.(repo.Device); isDevice {
		node_type = "device"
		prefix = "device:"
	}
	return TraceNode{Type: node_type, ID: int64(id), Key: prefix + name, Label: prefix + name}, true
}
