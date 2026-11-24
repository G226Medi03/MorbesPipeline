package tracing

import (
	"context"
	"fmt"
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
	for index, status := range observed {
		if status == nil {
			return nil, fmt.Errorf("cause trace: observed status[%d] is nil", index)
		}
	}
	seen := make([]Candidate, 0)
	// A graph traversal must expand each status at most once. Without this,
	// dense cause graphs produce an exponential number of paths that reach the
	// same nodes, which is the main cause of cause_trace timeouts.
	expanded := make(map[[32]byte]bool)
	for _, status := range observed {
		node := status_node(status)
		start_path := []TraceNode{node}
		start_hashes := [][32]byte{status.Hash()}
		if err := s.trace_from(ctx, status, 0, maxDepth, start_path, start_hashes, nil, &seen, expanded); err != nil {
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
	path_hashes [][32]byte,
	actor_path_set []repo.IEActor,
	seen *[]Candidate,
	expanded map[[32]byte]bool,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if depth >= max_depth {
		return nil
	}
	statusHash := status.Hash()
	if expanded[statusHash] {
		return nil
	}
	expanded[statusHash] = true
	actions, err := s.reader.ActionsResultingInStatus(status)
	if err != nil {
		return fmt.Errorf("cause trace: actions resulting in %q (hash %x): %w", label.StatusLabelValue(status), statusHash, err)
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
			return fmt.Errorf("cause trace: causes of action %q: %w", action_label(action), err)
		}
		for _, cause_link := range causes {
			if err := ctx.Err(); err != nil {
				return err
			}
			cause := cause_link.Status
			if cause == nil {
				continue
			}
			causeHash := cause.Hash()
			if containsHash(path_hashes, causeHash) {
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
			next_path_hashes := make([][32]byte, 0, len(path_hashes)+1)
			next_path_hashes = append(next_path_hashes, path_hashes...)
			next_path_hashes = append(next_path_hashes, causeHash)
			if err := s.trace_from(ctx, cause, depth+1, max_depth, candidate_path, next_path_hashes, actor_path_set, seen, expanded); err != nil {
				return err
			}
		}

		actors, err := s.reader.ActorsCausingAction(action)
		if err != nil {
			return fmt.Errorf("cause trace: actors of action %q: %w", action_label(action), err)
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

func containsHash(values [][32]byte, target [32]byte) bool {
	for _, value := range values {
		if value == target {
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
