package store

import (
	"context"
	"testing"

	"github.com/G226Medi03/MorbesPipeline/internal/entity/solve/model"
	source "github.com/G226Medi03/SourceEntity"
)

func TestListLLMCallsHydratesFromSource(t *testing.T) {
	ctx := context.Background()
	medicalDB, sourceDB, solveDB := openServiceSplitDBs(t)
	st := NewWithDBs(ctx, medicalDB, sourceDB, solveDB)
	if err := st.InitSchema(); err != nil {
		t.Fatalf("init schema: %v", err)
	}

	response, err := st.ResponseStore().GetOrInsertLlmResponse(
		"test-model",
		"test-prompt",
		source.NewSourceId(42),
		"test-output",
	)
	if err != nil {
		t.Fatalf("insert response: %v", err)
	}
	caseID, err := st.SolveRepo().Case.CreateForExamQuestion(
		ctx, st.SolveExecutor(), 1, source.NewSourceId(42))
	if err != nil {
		t.Fatalf("create case: %v", err)
	}
	if _, err := st.SolveRepo().Call.Append(ctx, st.SolveExecutor(), model.LLMCallInput{
		CaseID:   caseID,
		Seq:      1,
		Stage:    "status_extract",
		Role:     "status",
		SourceID: source.NewSourceId(42),
		RespID:   response.RespID,
	}); err != nil {
		t.Fatalf("append call: %v", err)
	}

	calls, err := st.SolveService.ListLLMCalls(caseID)
	if err != nil {
		t.Fatalf("list calls: %v", err)
	}
	if len(calls) != 1 {
		t.Fatalf("calls = %d, want 1", len(calls))
	}
	if calls[0].Model != "test-model" || calls[0].Prompt != "test-prompt" || calls[0].RawOutput != "test-output" {
		t.Fatalf("hydrated call = %#v", calls[0])
	}
}
