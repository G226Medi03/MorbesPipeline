package morbes_pipeline

import (
	"context"
	"path/filepath"
	"testing"

	morbesllm "github.com/G226Medi03/MorbesLLMService"
	source "github.com/G226Medi03/SourceEntity"
	testdb "github.com/G226Medi03/MorbesPipeline/internal/repository/testdb"
)

type testQuestionStore struct{}

func (testQuestionStore) GetQuestion(context.Context, int64) (QuestionInput, bool, error) {
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
	entityStore, err := store.Entities()
	if err != nil {
		t.Fatalf("Entities: %v", err)
	}
	status := HaveSymptom{Symptom: EOtherSymptom{Name: "고혈압"}}
	if _, err := entityStore.InsertStatus(status, []source.SourceId{source.NewSourceId(1)}); err != nil {
		t.Fatalf("InsertStatus: %v", err)
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
