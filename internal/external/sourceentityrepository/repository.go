// Package sourcerepository re-exports SourceEntityRepository types.
package sourcerepository

import ext "github.com/G226Medi03/SourceEntityRepository"

type (
	Repository       = ext.Repository
	LlmResponseTable = ext.LlmResponseTable
	SourceTable      = ext.SourceTable
	Executor         = ext.Executor
)

var (
	AddSourceIDs    = ext.AddSourceIDs
	ExecIndexes     = ext.ExecIndexes
	InitSourceTable = ext.InitSourceTable
	ListSourceIDs   = ext.ListSourceIDs
)
