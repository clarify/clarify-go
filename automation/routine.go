// Copyright 2023 Searis AS
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package automation

import (
	"context"
	"io"
	"log/slog"
	"runtime/debug"

	"github.com/clarify/clarify-go"
)

var defaultAppName string

func init() {
	// Set default app name.
	if info, ok := debug.ReadBuildInfo(); ok {
		defaultAppName = info.Main.Path
	}
}

// Routine describe the interface for an automation routine.
type Routine interface {
	Do(context.Context, *Config) error
}

// The RoutineFunc type is an adapter to allow the use of ordinary functions as
// automation routines. If f is a function with the appropriate signature,
// RoutineFunc(f) is a Routine that calls f.
type RoutineFunc func(context.Context, *Config) error

func (f RoutineFunc) Do(ctx context.Context, params *Config) error {
	return f(ctx, params)
}

// Config contain configuration for running routines, including a reference to
// a Clarify Client.
type Config struct {
	appName     string
	routinePath string
	logger      *slog.Logger
	client      *clarify.Client
	dryRun      bool
	earlyOut    bool
}

// NewConfig returns a new configuration for the passed in clients, using
// sensible defaults for all optional configuration. This means using slog
// package default logger, and setting the application name based on the main
// module's import path.
func NewConfig(client *clarify.Client) *Config {
	return &Config{
		appName: defaultAppName,
		logger:  slog.Default(),
		client:  client,
	}
}

// WithAppName returns a new configuration with the specified application name.
// When set, this property is added to the logger. The default app name is the
// main Go module's declared import path.
func (cfg Config) WithAppName(name string) *Config {
	cfg.appName = name
	return &cfg
}

// WithSubRoutineName returns a new configuration where name is appended to the
// existing routine path property.
//
// Special cases:
//   - If name is an empty string then no changes is performed.
func (cfg Config) WithSubRoutineName(name string) *Config {
	switch {
	case name == "":
		// No changes.
	case cfg.routinePath == "":
		// Start a new path.
		cfg.routinePath = name
	default:
		// Append name to existing path.
		cfg.routinePath += "/" + name
	}
	return &cfg
}

// WithDryRun returns a new configuration where the dry-run property is set to
// the specified value.
//
// Dry-run signals routines and actions to not perform any write or persist
// operations, but instead log the action as if it has been performed.
func (cfg Config) WithDryRun(dryRun bool) *Config {
	cfg.dryRun = dryRun
	return &cfg
}

// WithEarlyOut returns a new configuration where the early-out property is set
// to the specified value.
//
// Early-out signals routines with sub-routines to abort at the first error.
func (cfg Config) WithEarlyOut(value bool) *Config {
	cfg.earlyOut = true
	return &cfg
}

// WithLogger returns a new configuration with the specified logger. If a
// nil logger is passed in, all logs will be omitted. The default logger is the
// default logger from the slog package.
func (cfg Config) WithLogger(l *slog.Logger) *Config {
	cfg.logger = l
	return &cfg
}

// Client returns the Clarify client contained within options.
func (cfg Config) Client() *clarify.Client {
	return cfg.client
}

// AppName returns the app name.
func (cfg *Config) AppName() string {
	if cfg == nil {
		return defaultAppName
	}
	return cfg.appName
}

// RoutinePath returns the current routine name path joined by slash (/).
func (cfg *Config) RoutinePath() string {
	if cfg == nil {
		return ""
	}
	return cfg.routinePath
}

// EarlyOut returns the value of the early-out option. When true, routines with
// sub-routines should abort at the first error.
func (cfg *Config) EarlyOut() bool {
	return cfg.dryRun
}

// DryRun returns the value of the dry-run option. When true, routines and
// actions should not perform write and persist operations.
func (cfg *Config) DryRun() bool {
	return cfg.dryRun
}

// Logger returns a structured logger instance.
func (cfg *Config) Logger() *slog.Logger {
	logger := cfg.logger
	if logger == nil {
		// Return a logger that discards all entries.
		return slog.New(slog.NewJSONHandler(io.Discard, nil))
	}

	if cfg.appName != "" {
		logger = logger.With(attrAppName(cfg.appName))
	}
	if cfg.routinePath != "" {
		logger = logger.With(attrRoutineName(cfg.routinePath))
	}
	if cfg.dryRun {
		logger = logger.With(attrDryRun())
	}
	return logger
}
