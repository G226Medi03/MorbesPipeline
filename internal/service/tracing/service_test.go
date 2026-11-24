package tracing

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	label "github.com/G226Medi03/MorbesPipeline/internal/entity/label"
	source "github.com/G226Medi03/SourceEntity"
)

type fakeTraceReader struct {
	actions map[[32]byte][]repo.ActionStatusLink
	causes  map[[32]byte][]repo.ActionStatusLink
	calls   map[[32]byte]int
	err     error

	nextSourceID int64
}

func newFakeTraceReader() *fakeTraceReader {
	return &fakeTraceReader{
		actions: make(map[[32]byte][]repo.ActionStatusLink),
		causes:  make(map[[32]byte][]repo.ActionStatusLink),
		calls:   make(map[[32]byte]int),
	}
}

func (r *fakeTraceReader) setCauses(from repo.IEStatus, causes ...repo.IEStatus) {
	fromHash := from.Hash()
	r.nextSourceID++
	r.actions[fromHash] = []repo.ActionStatusLink{{
		Action:   repo.CauseStatus{Status: from},
		SourceID: source.NewSourceId(r.nextSourceID),
	}}
	for _, cause := range causes {
		r.nextSourceID++
		r.causes[fromHash] = append(r.causes[fromHash], repo.ActionStatusLink{
			Status:   cause,
			SourceID: source.NewSourceId(r.nextSourceID),
		})
	}
}

func (r *fakeTraceReader) ActionsResultingInStatus(status repo.IEStatus) ([]repo.ActionStatusLink, error) {
	hash := status.Hash()
	r.calls[hash]++
	if r.err != nil {
		return nil, r.err
	}
	return r.actions[hash], nil
}

func (r *fakeTraceReader) StatusesCausingAction(action repo.IEAction) ([]repo.ActionStatusLink, error) {
	cause, ok := action.(repo.CauseStatus)
	if !ok || cause.Status == nil {
		return nil, nil
	}
	return r.causes[cause.Status.Hash()], nil
}

func (r *fakeTraceReader) ActorsCausingAction(repo.IEAction) ([]repo.ActionActorLink, error) {
	return nil, nil
}

func TestTraceContextExpandsSameLabelDifferentPayload(t *testing.T) {
	ctx := context.Background()
	reader := newFakeTraceReader()

	root := repo.HaveDisease{Disease: "root"}
	c1 := repo.HaveSymptom{Symptom: repo.EOtherSymptom{Name: "통증"}}
	c2 := repo.HaveSymptom{Symptom: repo.EPain{}}
	d1 := repo.HaveDisease{Disease: "d1"}
	d2 := repo.HaveDisease{Disease: "d2"}

	if got, want := label.StatusLabelValue(c1), label.StatusLabelValue(c2); got != want {
		t.Fatalf("test status labels = %q and %q, want same label", got, want)
	}
	if c1.Hash() == c2.Hash() {
		t.Fatal("test statuses have the same hash, want different payload hashes")
	}

	reader.setCauses(root, c1, c2)
	reader.setCauses(c1, d1)
	reader.setCauses(c2, d2)

	candidates, err := New_Service(reader).TraceContext(ctx, []repo.IEStatus{root}, 3)
	if err != nil {
		t.Fatalf("TraceContext: %v", err)
	}
	for _, want := range []repo.IEStatus{c1, c2, d1, d2} {
		if !hasCause(candidates, want) {
			t.Fatalf("candidate with cause %q was not traced", label.StatusLabelValue(want))
		}
	}
	if reader.calls[c1.Hash()] != 1 || reader.calls[c2.Hash()] != 1 {
		t.Fatalf("expansion calls = %v, want one call for each distinct status", reader.calls)
	}
}

func TestTraceContextDeduplicatesWhitespaceOnlyDifference(t *testing.T) {
	ctx := context.Background()
	reader := newFakeTraceReader()

	root := repo.HaveDisease{Disease: "root"}
	c1 := repo.HaveDisease{Disease: " 고칼륨혈증 "}
	c2 := repo.HaveDisease{Disease: "고칼륨혈증"}
	d1 := repo.HaveDisease{Disease: "d1"}

	if c1.Hash() != c2.Hash() {
		t.Fatal("whitespace variants have different hashes, want same hash")
	}

	reader.setCauses(root, c1, c2)
	reader.setCauses(c1, d1)
	reader.setCauses(c2, d1)

	candidates, err := New_Service(reader).TraceContext(ctx, []repo.IEStatus{root}, 3)
	if err != nil {
		t.Fatalf("TraceContext: %v", err)
	}
	if !hasCause(candidates, c1) {
		t.Fatal("semantic status was not traced")
	}
	if !hasCause(candidates, d1) {
		t.Fatal("semantic status was not expanded")
	}
	if got := reader.calls[c1.Hash()]; got != 1 {
		t.Fatalf("semantic status expansion calls = %d, want 1", got)
	}
}

func TestTraceContextDeduplicatesPainAspectOrder(t *testing.T) {
	ctx := context.Background()
	reader := newFakeTraceReader()

	root := repo.HaveDisease{Disease: "root"}
	quality := repo.Quality{Name: "날카로운"}
	intensity := repo.Intensity{Value: "5", Measure: "점"}
	c1 := repo.HaveSymptom{Symptom: repo.EPain{
		Site:    repo.NewUnknownActor("우상복부"),
		Aspects: []repo.IPainAspect{quality, intensity},
	}}
	c2 := repo.HaveSymptom{Symptom: repo.EPain{
		Site:    repo.NewUnknownActor("우상복부"),
		Aspects: []repo.IPainAspect{intensity, quality},
	}}
	d1 := repo.HaveDisease{Disease: "d1"}

	if c1.Hash() != c2.Hash() {
		t.Fatal("pain aspect order variants have different hashes, want same hash")
	}
	if label.StatusLabelValue(c1) == label.StatusLabelValue(c2) {
		t.Fatal("test labels are equal, want different display labels")
	}

	reader.setCauses(root, c1, c2)
	reader.setCauses(c1, d1)
	reader.setCauses(c2, d1)

	candidates, err := New_Service(reader).TraceContext(ctx, []repo.IEStatus{root}, 3)
	if err != nil {
		t.Fatalf("TraceContext: %v", err)
	}
	if !hasCause(candidates, d1) {
		t.Fatal("pain status was not expanded")
	}
	if got := reader.calls[c1.Hash()]; got != 1 {
		t.Fatalf("pain status expansion calls = %d, want 1", got)
	}
}

func TestTraceContextDenseCycleTerminates(t *testing.T) {
	const statusCount = 8
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	statuses := make([]repo.IEStatus, 0, statusCount)
	for index := 0; index < statusCount; index++ {
		statuses = append(statuses, repo.HaveDisease{Disease: repo.EDisease(fmt.Sprintf("s%d", index))})
	}
	reader := newFakeTraceReader()
	for _, status := range statuses {
		reader.setCauses(status, statuses...)
	}

	candidates, err := New_Service(reader).TraceContext(ctx, statuses[:1], 100)
	if err != nil {
		t.Fatalf("TraceContext: %v", err)
	}
	if len(candidates) == 0 {
		t.Fatal("dense cycle produced no candidates")
	}
	if len(reader.calls) != statusCount {
		t.Fatalf("expanded status count = %d, want %d", len(reader.calls), statusCount)
	}
	for status, calls := range reader.calls {
		if calls != 1 {
			t.Fatalf("status %x expansion calls = %d, want 1", status, calls)
		}
	}
}

func TestTraceContextReturnsReaderErrorWithLabelAndHash(t *testing.T) {
	sentinel := errors.New("reader failed")
	reader := newFakeTraceReader()
	reader.err = sentinel
	root := repo.HaveDisease{Disease: "root"}

	_, err := New_Service(reader).TraceContext(context.Background(), []repo.IEStatus{root}, 3)
	if !errors.Is(err, sentinel) {
		t.Fatalf("TraceContext error = %v, want wrapped sentinel", err)
	}
	if !strings.Contains(err.Error(), label.StatusLabelValue(root)) {
		t.Fatalf("TraceContext error = %q, want status label", err)
	}
	if !strings.Contains(err.Error(), fmt.Sprintf("%x", root.Hash())) {
		t.Fatalf("TraceContext error = %q, want status hash", err)
	}
}

func TestTraceContextRejectsNilObservedStatus(t *testing.T) {
	_, err := New_Service(newFakeTraceReader()).TraceContext(context.Background(), []repo.IEStatus{nil}, 3)
	if err == nil {
		t.Fatal("TraceContext(nil status) = nil error, want error")
	}
}

func hasCause(candidates []Candidate, target repo.IEStatus) bool {
	for _, candidate := range candidates {
		if equalStatus(candidate.causeStatus, target) {
			return true
		}
	}
	return false
}
