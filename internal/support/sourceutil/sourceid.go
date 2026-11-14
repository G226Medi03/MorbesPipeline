// Package sourceutil contains conversions for SourceEntity opaque IDs.
package sourceutil

import (
	source "github.com/G226Medi03/MorbesPipeline/internal/external/sourceentity"
)

// SourceID restores a SourceId from its persisted int64 value.
func SourceID(value int64) source.SourceId {
	var id source.SourceId
	_ = id.Scan(value)
	return id
}

// RespID restores a RespId from its persisted int64 value.
func RespID(value int64) source.RespId {
	var id source.RespId
	_ = id.Scan(value)
	return id
}
