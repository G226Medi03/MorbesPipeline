// Package sqlite re-exports SQLCore sqlite helpers.
package sqlite

import ext "github.com/G226Medi03/SQLCore/sqlite"

type Executor = ext.Executor

var (
	EnsureRenamedColumn = ext.EnsureRenamedColumn
	TableExists         = ext.TableExists
)
