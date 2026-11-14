// Package duration re-exports go-duration types and functions.
package duration

import ext "github.com/G226Medi03/go-duration"

type (
	Duration = ext.Duration
	TimeUnit = ext.TimeUnit
)

const (
	Days = ext.Days
)

var (
	Marshal   = ext.Marshal
	Unmarshal = ext.Unmarshal
	Parse     = ext.Parse
)
