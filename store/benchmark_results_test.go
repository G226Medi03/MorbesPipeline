package store

import (
	"context"
	"path/filepath"
	"testing"

	testdb "github.com/G226Medi03/MorbesPipeline/internal/testdb"
)

func TestListProblemResultsFiltersAndOrder(t *testing.T) {
	ctx := context.Background()
	db, err := testdb.Open(filepath.Join(t.TempDir(), "results.db"))
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	defer db.Close() //nolint:errcheck // deferred close
	store := New(ctx, db)
	if err := store.InitSchema(); err != nil {
		t.Fatalf("InitSchema: %v", err)
	}

	insert := func(model, output string, correct int) {
		t.Helper()
		result, err := db.Exec(`INSERT INTO llm_responses (model, prompt, refered_source_id, output) VALUES (?, ?, 0, ?)`, model, "prompt-"+output, output)
		if err != nil {
			t.Fatalf("insert response: %v", err)
		}
		responseID, err := result.LastInsertId()
		if err != nil {
			t.Fatalf("LastInsertId: %v", err)
		}
		if _, err := db.Exec(`INSERT INTO llm_problem_results (exam_question_id, llm_response_id, correct, created_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)`, responseID, responseID, correct); err != nil {
			t.Fatalf("insert result: %v", err)
		}
	}
	insert("model-a", "①", 1)
	insert("model-a", "②", 0)
	insert("model-b", "③", 1)

	all, err := store.ListProblemResults(ctx, ProblemResultFilter{Limit: 10})
	if err != nil {
		t.Fatalf("ListProblemResults: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("all = %d, want 3", len(all))
	}

	correct := false
	wrong, err := store.ListProblemResults(ctx, ProblemResultFilter{Correct: &correct, Limit: 10})
	if err != nil {
		t.Fatalf("wrong filter: %v", err)
	}
	if len(wrong) != 1 || wrong[0].Model != "model-a" || wrong[0].Correct {
		t.Fatalf("wrong = %#v", wrong)
	}

	modelA, err := store.ListProblemResults(ctx, ProblemResultFilter{Model: "model-a", Limit: 10})
	if err != nil {
		t.Fatalf("model filter: %v", err)
	}
	if len(modelA) != 2 {
		t.Fatalf("model-a = %d, want 2", len(modelA))
	}

	count, err := store.CountProblemResults(ctx, ProblemResultFilter{Model: "model-a"})
	if err != nil || count != 2 {
		t.Fatalf("count = %d, err=%v", count, err)
	}
	models, err := store.ListProblemResultModels(ctx)
	if err != nil {
		t.Fatalf("models: %v", err)
	}
	if len(models) != 2 || models[0] != "model-a" || models[1] != "model-b" {
		t.Fatalf("models = %#v", models)
	}

	page, err := store.ListProblemResults(ctx, ProblemResultFilter{Limit: 1, Offset: 1})
	if err != nil {
		t.Fatalf("page: %v", err)
	}
	if len(page) != 1 {
		t.Fatalf("page = %d, want 1", len(page))
	}
}
