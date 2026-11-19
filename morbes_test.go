package morbes_pipeline

import (
	"context"
	"path/filepath"
	"testing"

	morbesllm "github.com/G226Medi03/MorbesLLMService"
	source "github.com/G226Medi03/SourceEntity"

	"github.com/G226Medi03/MorbesPipeline/internal/support/ids"
	testdb "github.com/G226Medi03/MorbesPipeline/internal/testdb"
	graph "github.com/G226Medi03/MorbesPipeline/internal/tracing/graph"
)

type testQuestionStore struct{}

func (testQuestionStore) GetQuestion(context.Context, ids.ExamQuestionID) (QuestionInput, bool, error) {
	return QuestionInput{}, false, nil
}

func TestPublicFacade(t *testing.T) {
	ctx := context.Background()
	db, err := testdb.Open(filepath.Join(t.TempDir(), "public-facade.db"))
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	defer db.Close() //nolint:errcheck // deferred close

	store := New(ctx, db)
	if err := store.InitSchema(); err != nil {
		t.Fatalf("InitSchema: %v", err)
	}
	container := graph.New_Container()
	status := HaveSymptom{Symptom: EOtherSymptom{Name: "고혈압"}}
	if _, err := store.EnsureStatusIn(container, status, source.NewSourceId(1)); err != nil {
		t.Fatalf("EnsureStatusIn: %v", err)
	}
	if _, err := store.LoadGraph(); err != nil {
		t.Fatalf("LoadGraph: %v", err)
	}
	if _, err := NewService(ctx, Dependencies{}); err == nil {
		t.Fatal("NewService accepted missing dependencies")
	}
	service, serviceErr := NewService(ctx, Dependencies{
		Store:         store,
		LLM:           &morbesllm.Service{},
		QuestionStore: testQuestionStore{},
	})
	if serviceErr != nil {
		t.Fatalf("NewService: %v", serviceErr)
	}
	if service == nil {
		t.Fatal("NewService returned nil")
	}
}
