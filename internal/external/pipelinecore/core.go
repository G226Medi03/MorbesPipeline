// Package core re-exports the PipeLineCore pipeline model.
package core

import ext "github.com/G226Medi03/PipeLineCore"

type (
	Pipeline     = ext.Pipeline
	Node         = ext.Node
	State        = ext.State
	PipelineName = ext.PipelineName
	NodeName     = ext.NodeName
	NodeKind     = ext.NodeKind
	NodeState    = ext.NodeState
	CacheMode    = ext.CacheMode
	ReplayCall   = ext.ReplayCall
	Event        = ext.Event
)

const (
	KindPrepare   = ext.KindPrepare
	KindLLM       = ext.KindLLM
	KindTransform = ext.KindTransform
	KindResolve   = ext.KindResolve
	KindPersist   = ext.KindPersist
	KindGraph     = ext.KindGraph
	KindCheck     = ext.KindCheck
	KindFinalize  = ext.KindFinalize

	NodePending   = ext.NodePending
	NodeRunning   = ext.NodeRunning
	NodeSucceeded = ext.NodeSucceeded
	NodeFailed    = ext.NodeFailed
	NodeSkipped   = ext.NodeSkipped

	CacheInherit = ext.CacheInherit
	CacheUse     = ext.CacheUse
	CacheBypass  = ext.CacheBypass
	CacheRefresh = ext.CacheRefresh
)

var (
	NewState = ext.NewState
	NewNode  = ext.NewNode
)
