package store

import (
	"context"
	"path/filepath"
	"testing"

	testdb "github.com/G226Medi03/MorbesPipeline/internal/testdb"
	source "github.com/G226Medi03/SourceEntity"
)

func TestTraceSourceFollowsLlmReferedChainToOriginal(t *testing.T) {
	ctx := context.Background()
	db, err := testdb.Open(filepath.Join(t.TempDir(), "source-trace.db"))
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	defer db.Close() //nolint:errcheck // deferred close
	st := New(ctx, db)
	if err := st.InitSchema(); err != nil {
		t.Fatalf("InitSchema: %v", err)
	}
	for _, query := range []string{
		`INSERT INTO sources (source_id, path, source_type, id_ptr) VALUES (1, 'original.pdf', 'PdfFile', 0)`,
		`INSERT INTO sources (source_id, path, source_type, id_ptr) VALUES (2, '', 'LlmResponse', 10)`,
		`INSERT INTO sources (source_id, path, source_type, id_ptr) VALUES (3, '', 'LlmResponse', 11)`,
		`INSERT INTO llm_responses (resp_id, model, prompt, refered_source_id, output) VALUES (10, 'test', 'p1', 1, 'o1')`,
		`INSERT INTO llm_responses (resp_id, model, prompt, refered_source_id, output) VALUES (11, 'test', 'p2', 2, 'o2')`,
	} {
		if _, err := db.ExecContext(ctx, query); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}
	trace, err := st.TraceSource(source.NewSourceId(3))
	if err != nil {
		t.Fatalf("TraceSource: %v", err)
	}
	if len(trace.Chain) != 3 {
		t.Fatalf("chain = %#v", trace.Chain)
	}
	if trace.Origin == nil || trace.Origin.SourceID.ToI64() != 1 || trace.Origin.Path != "original.pdf" {
		t.Fatalf("origin = %#v", trace.Origin)
	}
	if trace.Immediate.SourceID.ToI64() != 3 || trace.Immediate.SourceType != source.SourceTypeLlmResponse {
		t.Fatalf("immediate = %#v", trace.Immediate)
	}
}

func TestTraceSourceStopsAtUnresolvedLlmSource(t *testing.T) {
	ctx := context.Background()
	db, err := testdb.Open(filepath.Join(t.TempDir(), "source-trace-empty.db"))
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	defer db.Close() //nolint:errcheck // deferred close
	st := New(ctx, db)
	if err := st.InitSchema(); err != nil {
		t.Fatalf("InitSchema: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO sources (source_id, path, source_type, id_ptr) VALUES (1, '', 'LlmResponse', 10)`); err != nil {
		t.Fatalf("insert source: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO llm_responses (resp_id, model, prompt, refered_source_id, output) VALUES (10, 'test', 'p', 0, 'o')`); err != nil {
		t.Fatalf("insert response: %v", err)
	}
	trace, err := st.TraceSource(source.NewSourceId(1))
	if err != nil {
		t.Fatalf("TraceSource: %v", err)
	}
	if trace.Origin != nil {
		t.Fatalf("origin = %#v, want nil", trace.Origin)
	}
	if len(trace.Chain) != 1 {
		t.Fatalf("chain = %#v", trace.Chain)
	}
}
