// Package logs is the application logging facade on top of github.com/G226Medi03/log-sqlite.
package logs

import (
	"io"
	"os"
	"strings"

	clog "github.com/G226Medi03/log-sqlite"
)

type (
	Level   = clog.Level
	Attr    = clog.Attr
	Entry   = clog.Entry
	ISink   = clog.Sink
	Logger  = clog.Logger
	Options = clog.Config
)

const (
	LevelDebug = clog.LevelDebug
	LevelInfo  = clog.LevelInfo
	LevelWarn  = clog.LevelWarn
	LevelError = clog.LevelError
)

// Logging is the SQLite-backed application logger.
type Logging struct {
	*clog.App
}

// Open opens the log database described by opts.
func Open(opts Options) (*Logging, error) {
	app, ce := clog.Open(opts)
	if ce != nil {
		return nil, ce
	}
	return &Logging{App: app}, nil
}

const default_log_db_path = "./db/logs.db"

// OptionsFromEnv reads log settings from the environment.
// LOG_DB falls back to the application's db directory.
func OptionsFromEnv(component string) Options {
	opts := clog.OpenFromEnv(component)
	if strings.TrimSpace(os.Getenv("LOG_DB")) == "" {
		opts.Path = default_log_db_path
	}
	return opts
}

// AttrOf builds an Attr from any value.
func AttrOf(key string, value any) Attr {
	return clog.AttrOf(key, value)
}

func Debug(msg string, attrs ...Attr) { clog.Debug(msg, attrs...) }
func Info(msg string, attrs ...Attr)  { clog.Info(msg, attrs...) }
func Warn(msg string, attrs ...Attr)  { clog.Warn(msg, attrs...) }
func Error(msg string, attrs ...Attr) { clog.Error(msg, attrs...) }
func Fatal(msg string, attrs ...Attr) { clog.Fatal(msg, attrs...) }

// SetDefault replaces the package-level logger.
func SetDefault(logger *Logger) {
	clog.SetDefault(logger)
}

// InstallStdlogBridge forwards standard library log output to target and the sink.
func InstallStdlogBridge(target io.Writer) *Logger {
	return clog.InstallStdlogBridge(target)
}
