// Package logsqlite re-exports log-sqlite types and functions.
package logsqlite

import ext "github.com/G226Medi03/log-sqlite"

type (
	App     = ext.App
	Attr    = ext.Attr
	Config  = ext.Config
	DB      = ext.DB
	Entry   = ext.Entry
	Level   = ext.Level
	Logger  = ext.Logger
	Options = ext.Options
	Sink    = ext.Sink
)

const (
	DefaultPath = ext.DefaultPath
	LevelDebug  = ext.LevelDebug
	LevelInfo   = ext.LevelInfo
	LevelWarn   = ext.LevelWarn
	LevelError  = ext.LevelError
)

var (
	AttrOf              = ext.AttrOf
	Debug               = ext.Debug
	Error               = ext.Error
	Fatal               = ext.Fatal
	Info                = ext.Info
	Warn                = ext.Warn
	SetDefault          = ext.SetDefault
	Open                = ext.Open
	OpenFromEnv         = ext.OpenFromEnv
	OpenSQLite          = ext.OpenSQLite
	Default             = ext.Default
	InstallStdlogBridge = ext.InstallStdlogBridge
	Named               = ext.Named
	NewLogger           = ext.NewLogger
	ParseLevel          = ext.ParseLevel
)
