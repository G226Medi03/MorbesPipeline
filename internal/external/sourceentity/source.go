// Package source re-exports SourceEntity types used by the pipeline.
package source

import ext "github.com/G226Medi03/SourceEntity"

type (
	SourceId       = ext.SourceId
	RespId         = ext.RespId
	SourceRef      = ext.SourceRef
	SourceRow      = ext.SourceRow
	LlmResponse    = ext.LlmResponse
	LlmResponseRow = ext.LlmResponseRow
	SourceTrace    = ext.SourceTrace
)

const (
	CanonicalDeepSeekFlashModel = ext.CanonicalDeepSeekFlashModel
	SourceTypeLlmResponse       = ext.SourceTypeLlmResponse
)

var (
	NewSourceId    = ext.NewSourceId
	NewRespId      = ext.NewRespId
	CanonicalModel = ext.CanonicalModel
)
