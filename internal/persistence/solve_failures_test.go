package store

import (
	"context"
	"github.com/G226Medi03/MorbesPipeline/internal/support/sourceutil"
	"path/filepath"
	"testing"

	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
	testdb "github.com/G226Medi03/MorbesPipeline/internal/testdb"
)

func TestCountAndListFailedCases(t *testing.T) {
	ctx := context.Background()
	db, err := testdb.Open(filepath.Join(t.TempDir(), "solve-failures.db"))
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	defer db.Close() //nolint:errcheck // deferred close
	st := New(ctx, db)
	if err := st.InitSchema(); err != nil {
		t.Fatalf("InitSchema: %v", err)
	}

	failedID, err := st.CreateForExamQuestionWithSource(101, sourceutil.SourceID(1))
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if err := st.SetFailed(failedID, context.Canceled); err != nil {
		t.Fatalf("SetFailed: %v", err)
	}

	notFoundID, err := st.CreateForExamQuestionWithSource(102, sourceutil.SourceID(1))
	if err != nil {
		t.Fatalf("Create not found: %v", err)
	}
	if err := st.SetAnswer(notFoundID, "{}"); err != nil {
		t.Fatalf("SetAnswer: %v", err)
	}
	if err := st.SaveAnswerCheck(domain.AnswerCheck{
		CaseID:         notFoundID,
		ContainsAnswer: false,
		MatchedType:    "none",
	}); err != nil {
		t.Fatalf("SaveAnswerCheck: %v", err)
	}

	successID, err := st.CreateForExamQuestionWithSource(103, sourceutil.SourceID(1))
	if err != nil {
		t.Fatalf("Create success: %v", err)
	}
	if err := st.SetAnswer(successID, "{}"); err != nil {
		t.Fatalf("SetAnswer success: %v", err)
	}
	if err := st.SaveAnswerCheck(domain.AnswerCheck{
		CaseID:         successID,
		ContainsAnswer: true,
		MatchedType:    "cause",
	}); err != nil {
		t.Fatalf("SaveAnswerCheck success: %v", err)
	}

	count, err := st.CountFailedCases()
	if err != nil {
		t.Fatalf("CountFailedCases: %v", err)
	}
	if count != 2 {
		t.Fatalf("CountFailedCases = %d, want 2", count)
	}
	items, err := st.ListFailedCases(10)
	if err != nil {
		t.Fatalf("ListFailedCases: %v", err)
	}
	if len(items) != 2 || items[0].ID != failedID || items[1].ID != notFoundID {
		t.Fatalf("ListFailedCases = %#v", items)
	}
}
