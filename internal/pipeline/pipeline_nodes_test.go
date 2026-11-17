package csolve

import (
	"context"
	"testing"

	"github.com/G226Medi03/PipeLineCore"
)

func TestSolveNodePipelineOrder(t *testing.T) {
	pipeline := newSolveNodePipeline(&Service{})
	if err := pipeline.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	want := []string{
		string(solveLoadNode),
		string(explainNode),
		string(extractionNode),
		string(resolveNode),
		string(causeNode),
	}
	if len(pipeline.Nodes) != len(want) {
		t.Fatalf("nodes = %d, want %d", len(pipeline.Nodes), len(want))
	}
	for index, node := range pipeline.Nodes {
		if string(node.Name) != want[index] {
			t.Fatalf("node[%d] = %q, want %q", index, node.Name, want[index])
		}
	}
}

func TestExtractionPipelineOrder(t *testing.T) {
	pipeline := core.Pipeline{
		Name: "CaseExtractionPipeline",
		Nodes: []core.Node{
			core.NewNode(extractQuestionTypeNode, "", "", core.CacheInherit, func(context.Context, *core.State) error { return nil }, false),
			core.NewNode(extractParallelNode, "", "", core.CacheInherit, func(context.Context, *core.State) error { return nil }, false),
			core.NewNode(extractPatientNode, "", "", core.CacheInherit, func(context.Context, *core.State) error { return nil }, true),
			core.NewNode(extractRelationNode, "", "", core.CacheInherit, func(context.Context, *core.State) error { return nil }, false),
			core.NewNode(extractOptionNode, "", "", core.CacheInherit, func(context.Context, *core.State) error { return nil }, false),
			core.NewNode(extractAssembleNode, "", "", core.CacheInherit, func(context.Context, *core.State) error { return nil }, false),
		},
	}
	if err := pipeline.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	want := []string{
		string(extractQuestionTypeNode),
		string(extractParallelNode),
		string(extractPatientNode),
		string(extractRelationNode),
		string(extractOptionNode),
		string(extractAssembleNode),
	}
	for index, node := range pipeline.Nodes {
		if string(node.Name) != want[index] {
			t.Fatalf("node[%d] = %q, want %q", index, node.Name, want[index])
		}
	}
}

func TestNodeForStage(t *testing.T) {
	tests := map[string]string{
		"":                 "",
		NodeExplain:        string(explainNode),
		NodeCaseExtraction: string(extractionNode),
		NodeResolve:        string(resolveNode),
		NodeCause:          string(causeNode),
	}
	for stage, want := range tests {
		if got := string(nodeForStage(stage)); got != want {
			t.Fatalf("nodeForStage(%q) = %q, want %q", stage, got, want)
		}
	}
}
