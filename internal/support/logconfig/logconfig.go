// Package logconfig contains the application's log database configuration.
package logconfig

import (
	"os"
	"strings"

	clog "github.com/G226Medi03/log-sqlite"
)

const default_log_db_path = "./db/logs.db"

// Logging is the application logger backed by the configured log database.
type Logging struct {
	*clog.App
}

// Open opens the log database described by opts.
func Open(opts clog.Config) (*Logging, error) {
	app, err := clog.Open(opts)
	if err != nil {
		return nil, err
	}
	return &Logging{App: app}, nil
}

// OptionsFromEnv reads log settings from the environment.
// LOG_DB falls back to the application's db directory.
func OptionsFromEnv(component string) clog.Config {
	opts := clog.OpenFromEnv(component)
	if strings.TrimSpace(os.Getenv("LOG_DB")) == "" {
		opts.Path = default_log_db_path
	}
	return opts
}
