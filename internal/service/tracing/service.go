package tracing

import (
	"context"
	"sort"
	"strings"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	label "github.com/G226Medi03/MorbesPipeline/internal/entity/label"
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
	seen := make([]Candidate, 0)
	for _, status := range observed {
		node := status_node(status)
		start_path := []TraceNode{node}
		if err := s.trace_from(ctx, status, 0, maxDepth, start_path, []repo.IEStatus{status}, nil, &seen); err != nil {
			return nil, err
		}
	}

	result := make([]Candidate, len(seen))
	copy(result, seen)
	for index := range result {
		sort.Slice(result[index].SourceIDs, func(i, j int) bool {
			return result[index].SourceIDs[i].ToI64() < result[index].SourceIDs[j].ToI64()
		})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Depth != result[j].Depth {
			return result[i].Depth < result[j].Depth
		}
		return result[i].CauseLabel < result[j].CauseLabel
	})
	return result, nil
}

func (s *Service) trace_from(
	ctx context.Context,
	status repo.IEStatus,
	depth, max_depth int,
	path []TraceNode,
	path_set []repo.IEStatus,
	actor_path_set []repo.IEActor,
	seen *[]Candidate,
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
		action_node := TraceNode{
			Type:  "action",
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
			if containsStatus(path_set, cause) {
				continue
			}
			cause_node := status_node(cause)
			candidate_path := make([]TraceNode, 0, len(path)+2)
			candidate_path = append(candidate_path, path...)
			candidate_path = append(candidate_path, action_node, cause_node)

			candidate := Candidate{
				CauseKind:     cause.Kind(),
				CauseLabel:    cause_node.Label,
				ViaActionKind: action.Kind(),
				Depth:         depth + 1,
				SourceIDs:     append_unique_ids([]source_entity.SourceId{actionLink.SourceID}, cause_link.SourceID),
				Path:          candidate_path,
				causeStatus:   cause,
				viaAction:     action,
			}
			mergeCandidate(seen, candidate)
			next_path_set := make([]repo.IEStatus, 0, len(path_set)+1)
			next_path_set = append(next_path_set, path_set...)
			next_path_set = append(next_path_set, cause)
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
			if containsActor(actor_path_set, actor) {
				continue
			}
			actor_node, ok := actor_node(actor)
			if !ok {
				continue
			}
			candidate_path := make([]TraceNode, 0, len(path)+2)
			candidate_path = append(candidate_path, path...)
			candidate_path = append(candidate_path, action_node, actor_node)
			candidate := Candidate{
				CauseType:     "actor",
				CauseKind:     "actor",
				CauseLabel:    actor_node.Label,
				ViaActionKind: action.Kind(),
				Depth:         depth + 1,
				SourceIDs:     append_unique_ids([]source_entity.SourceId{actionLink.SourceID}, actor_link.SourceID),
				Path:          candidate_path,
				causeActor:    actor,
				viaAction:     action,
			}
			mergeCandidate(seen, candidate)
		}
	}
	return nil
}

func mergeCandidate(seen *[]Candidate, candidate Candidate) {
	for index := range *seen {
		existing := (*seen)[index]
		if !sameCandidate(existing, candidate) {
			continue
		}
		existing.SourceIDs = append_unique_ids(existing.SourceIDs, candidate.SourceIDs...)
		if len(candidate.Path) < len(existing.Path) {
			candidate.SourceIDs = existing.SourceIDs
			(*seen)[index] = candidate
		} else {
			(*seen)[index] = existing
		}
		return
	}
	*seen = append(*seen, candidate)
}

func sameCandidate(left, right Candidate) bool {
	if left.CauseType != right.CauseType {
		return false
	}
	if left.causeStatus != nil || right.causeStatus != nil {
		return equalStatus(left.causeStatus, right.causeStatus) &&
			equalAction(left.viaAction, right.viaAction)
	}
	return equalActor(left.causeActor, right.causeActor) &&
		equalAction(left.viaAction, right.viaAction)
}

func containsStatus(values []repo.IEStatus, target repo.IEStatus) bool {
	for _, value := range values {
		if equalStatus(value, target) {
			return true
		}
	}
	return false
}

func containsActor(values []repo.IEActor, target repo.IEActor) bool {
	for _, value := range values {
		if equalActor(value, target) {
			return true
		}
	}
	return false
}

func equalStatus(a, b repo.IEStatus) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.Equal(b)
}

func equalAction(a, b repo.IEAction) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.Equal(b)
}

func equalActor(a, b repo.IEActor) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.Equal(b)
}

func status_node(status repo.IEStatus) TraceNode {
	return TraceNode{
		Type:  "status",
		Label: label.StatusLabelValue(status),
	}
}

func action_label(action repo.IEAction) string {
	if cause, ok := action.(repo.CauseStatus); ok {
		if cause.Status == nil {
			return "원인 상태 정보 없음"
		}
		label := label.StatusLabelValue(cause.Status)
		switch cause.Status.Kind() {
		case "increased_index", "decreased_index":
			name := strings.TrimSpace(strings.TrimSuffix(strings.TrimSuffix(label, " 증가"), " 감소"))
			return "cause_status: " + cause.Status.Kind() + " (" + name + ")"
		default:
			return "cause_status: " + label
		}
	}
	return action.Kind() + ":" + label.ActionLabelValue(action)
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
	return TraceNode{Type: node_type, Label: prefix + name}, true
}
