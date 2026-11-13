// Package core re-exports the generic pipeline model.
package core

import (
	pipelinecore "github.com/G226Medi03/PipeLineCore"
)

type Pipeline = pipelinecore.Pipeline
type Node = pipelinecore.Node
type State = pipelinecore.State
type PipelineName = pipelinecore.PipelineName
type NodeName = pipelinecore.NodeName
type NodeKind = pipelinecore.NodeKind
type NodeState = pipelinecore.NodeState
type Event = pipelinecore.Event

const (
	KindPrepare   = pipelinecore.KindPrepare
	KindLLM       = pipelinecore.KindLLM
	KindTransform = pipelinecore.KindTransform
	KindResolve   = pipelinecore.KindResolve
	KindPersist   = pipelinecore.KindPersist
	KindGraph     = pipelinecore.KindGraph
	KindCheck     = pipelinecore.KindCheck
	KindFinalize  = pipelinecore.KindFinalize

	NodePending   = pipelinecore.NodePending
	NodeRunning   = pipelinecore.NodeRunning
	NodeSucceeded = pipelinecore.NodeSucceeded
	NodeFailed    = pipelinecore.NodeFailed
	NodeSkipped   = pipelinecore.NodeSkipped
)

var NewState = pipelinecore.NewState
