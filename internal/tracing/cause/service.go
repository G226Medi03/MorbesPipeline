package ccause

import (
	"context"
	"sort"
	"strings"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	entitykey "github.com/G226Medi03/MorbesPipeline/internal/domain/entitykey"
	centitylabel "github.com/G226Medi03/MorbesPipeline/internal/domain/label"
	source_entity "github.com/G226Medi03/SourceEntity"
)

// EntityReader is the typed repository surface used by cause tracing.
type EntityReader interface {
	ActionsResultingInStatus(status repo.IEStatus) ([]repo.ActionStatusLink, error)
	StatusesCausingAction(action repo.IEAction) ([]repo.ActionStatusLink, error)
	ActorsCausingAction(action repo.IEAction) ([]repo.ActionActorLink, error)
}

type Service struct {
	reader EntityReader
}

func New_Service(reader EntityReader) *Service {
	return &Service{reader: reader}
}

// TraceContext is Trace with cancellation/timeout support.
func (s *Service) TraceContext(ctx context.Context, observed []repo.IEStatus, maxDepth int) ([]Candidate, error) {
	if maxDepth <= 0 {
		maxDepth = 3
	}
	seen := make(map[string]Candidate)
	for _, status := range observed {
		status_key := entitykey.Status(status)
		node := status_node(status)
		start_path := []TraceNode{node}
		if err := s.trace_from(ctx, status, 0, maxDepth, start_path, map[string]struct{}{status_key: {}}, map[string]struct{}{}, seen); err != nil {
			return nil, err
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

func (s *Service) trace_from(
	ctx context.Context,
	status repo.IEStatus,
	depth, max_depth int,
	path []TraceNode,
	path_set map[string]struct{},
	actor_path_set map[string]struct{},
	seen map[string]Candidate,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if depth >= max_depth {
		return nil
	}
	actions, err := s.reader.ActionsResultingInStatus(status)
	if err != nil {
		return err
	}
	for _, actionLink := range actions {
		if err := ctx.Err(); err != nil {
			return err
		}
		action := actionLink.Action
		action_key := entitykey.Action(action)
		action_node := TraceNode{
			Type:  "action",
			Key:   action_key,
			Label: action_label(action),
		}
		causes, err := s.reader.StatusesCausingAction(action)
		if err != nil {
			return err
		}
		for _, cause_link := range causes {
			if err := ctx.Err(); err != nil {
				return err
			}
			cause := cause_link.Status
			cause_key := entitykey.Status(cause)
			if _, ok := path_set[cause_key]; ok {
				continue
			}
			cause_node := status_node(cause)
			candidate_path := make([]TraceNode, 0, len(path)+2)
			candidate_path = append(candidate_path, path...)
			candidate_path = append(candidate_path, action_node, cause_node)

			key := cause_key + ":" + action_key
			candidate := Candidate{
				CauseKey:      cause_key,
				CauseKind:     cause.Kind(),
				CauseLabel:    cause_node.Label,
				ViaActionKey:  action_key,
				ViaActionKind: action.Kind(),
				Depth:         depth + 1,
				SourceIDs:     append_unique_ids([]source_entity.SourceId{actionLink.SourceID}, cause_link.SourceID),
				Path:          candidate_path,
			}
			if existing, ok := seen[key]; ok {
				existing.SourceIDs = append_unique_ids(existing.SourceIDs, candidate.SourceIDs...)
				if len(candidate.Path) < len(existing.Path) {
					candidate.SourceIDs = existing.SourceIDs
					seen[key] = candidate
				} else {
					seen[key] = existing
				}
			} else {
				seen[key] = candidate
			}
			next_path_set := make(map[string]struct{}, len(path_set)+1)
			for id := range path_set {
				next_path_set[id] = struct{}{}
			}
			next_path_set[cause_key] = struct{}{}
			if err := s.trace_from(ctx, cause, depth+1, max_depth, candidate_path, next_path_set, actor_path_set, seen); err != nil {
				return err
			}
		}

		actors, err := s.reader.ActorsCausingAction(action)
		if err != nil {
			return err
		}
		for _, actor_link := range actors {
			actor := actor_link.Actor
			actor_key := entitykey.Actor(actor)
			if _, ok := actor_path_set[actor_key]; ok {
				continue
			}
			actor_node, ok := actor_node(actor)
			if !ok {
				continue
			}
			candidate_path := make([]TraceNode, 0, len(path)+2)
			candidate_path = append(candidate_path, path...)
			candidate_path = append(candidate_path, action_node, actor_node)
			key := actor_key + ":" + action_key
			candidate := Candidate{
				CauseType:     "actor",
				CauseKey:      actor_key,
				CauseKind:     "actor",
				CauseLabel:    actor_node.Label,
				ViaActionKey:  action_key,
				ViaActionKind: action.Kind(),
				Depth:         depth + 1,
				SourceIDs:     append_unique_ids([]source_entity.SourceId{actionLink.SourceID}, actor_link.SourceID),
				Path:          candidate_path,
			}
			if existing, ok := seen[key]; ok {
				existing.SourceIDs = append_unique_ids(existing.SourceIDs, candidate.SourceIDs...)
				if len(candidate.Path) < len(existing.Path) {
					candidate.SourceIDs = existing.SourceIDs
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

func status_node(status repo.IEStatus) TraceNode {
	return TraceNode{
		Type:  "status",
		Key:   entitykey.Status(status),
		Label: centitylabel.StatusLabelValue(status),
	}
}

func action_label(action repo.IEAction) string {
	if cause, ok := action.(repo.CauseStatus); ok {
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
	return action.Kind() + ":" + centitylabel.ActionLabelValue(action)
}

func actor_node(actor repo.IEActor) (TraceNode, bool) {
	name := strings.TrimSpace(actor.ActorName())
	if name == "" {
		return TraceNode{}, false
	}
	node_type := "actor"
	prefix := "actor:"
	if _, isDevice := actor.(repo.Device); isDevice {
		node_type = "device"
		prefix = "device:"
	}
	return TraceNode{Type: node_type, Key: prefix + name, Label: prefix + name}, true
}
