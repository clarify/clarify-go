// Copyright 2023-2024 Searis AS
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

package automationcli

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/clarify/clarify-go"
	"github.com/clarify/clarify-go/automation"
	"github.com/clarify/clarify-go/internal/logging"
	"github.com/clarify/clarify-go/jsonrpc"
)

var defaultProgName string

func init() {
	if len(os.Args) > 0 {
		defaultProgName = filepath.Base(os.Args[0])
	}
}

const (
	usageCredentials = "Specify the path to your Clarify Integration's credentials file."
	usageUsername    = "Clarify integration ID to use as username; alternative to providing -credentials."
	usagePassword    = "Clarify integration password; required when username is set, ignored otherwise."
	usageVerbose     = "Set to true for printing logs at level DEBUG (the default is to log at INFO level)."
	usageJSON        = "Set to true to output logs in compact JSON format."
	usageDryRun      = "Signal to routines that they should mot write or persist changes."
	usageEarlyOut    = "Signal to routines that they should abort at the first error."
)

const usageFmt = `Usage: %[1]s [OPTIONS] [PATTERNS...]

PATTERNS are expected to match routine or sub-routine names. Sub-routines are
matched via the slash character (/). The asterisk (*) can be used for wildcard
matching of a single path level.

Example: Given routines "a/b/b", "a/b/c" and "b/b/c", then:
- "a/b" will match "a/b/b" and "a/b/c"
- "*/b/c" will match "a/b/c" and "b/b/c"
`

// Config describe a set of command-line options.
type Config struct {
	// AppName holds the app-name to use in automation. The default is set to
	// match the Go main module import path.
	AppName string

	// CredentialsFile describes the path to a valid Clarify integration JSON
	// credentials file. This property is required if Username is set.
	CredentialsFile string

	// Username can be set to use Basic Authentication ot authorize against
	// Clarify. The user-name is the same as the integration ID. When set,
	// the content of the CredentialsFile config is ignored.
	Username string

	// Password can be set to use Basic Authentication to authorize against
	// Clarify. This property is required if Username is set.
	Password Password

	// Available routines for patterns to match.
	Routines automation.Routines

	// Patterns describes which sub-routines to run. If empty, all sub-routines
	// are run.
	Patterns []string

	// Verbose, if set, turns on DEBUG logging. The default log level is INFO.
	Verbose bool

	// Use a JSON log format.
	JSON bool

	// DryRun, if set, signals routines and actions to not persist changes.
	DryRun bool

	// EarlyOut, if set, signals the program to abort at the first routine
	// error. The default is to continue to the next routine.
	EarlyOut bool
}

// ParseArguments parses command-line arguments into a Config structure using
// the provided or prints usage information to os.Stderr. When using this
// method, the Patterns property is set from the remaining command-line
// arguments after all flags (options) have been parsed.
//
// This function can be used by users who need to customize the configuration
// before it's run, but do not need to customize command-line flags.
func ParseArguments(routines automation.Routines, arguments []string) (*Config, error) {
	cfg := Config{
		Routines: routines,
	}
	set := cfg.FlagSet(defaultProgName, flag.ContinueOnError)
	err := set.Parse(arguments)
	if err != nil {
		return nil, err
	}
	cfg.Patterns = append(cfg.Patterns, set.Args()...)
	return &cfg, nil
}

// FlagSet returns a named flag-set for parsing command line configuration flags
// where fallback values are fetched from environment variables. Note that there
// is no flag for the AppName and Patterns property.
//
// This method can be used by users who need to customize the which command-line
// flags are available in their application.
func (cfg *Config) FlagSet(progName string, errorHandling flag.ErrorHandling) *flag.FlagSet {
	if progName == "" {
		progName = defaultProgName
	}
	adder := flagSetAdder{
		envPrefix: "CLARIFY_",
		set:       flag.NewFlagSet(progName, errorHandling),
	}
	adder.set.Usage = func() {
		out := adder.set.Output()
		fmt.Fprintf(out, usageFmt, progName)
		fmt.Fprintln(out, "\nAvailable routines:")
		cfg.Routines.Print(out, "  ")
		fmt.Fprintln(out, "\nOptions:")
		adder.set.PrintDefaults()
	}

	adder.StringVar(&cfg.CredentialsFile, "credentials", "credentials.json", usageCredentials)
	adder.StringVar(&cfg.Username, "username", "", usageUsername)
	adder.StringVar(&cfg.Password.value, "password", "", usagePassword)
	adder.BoolVar(&cfg.Verbose, "v", false, usageVerbose)
	adder.BoolVar(&cfg.JSON, "json", false, usageJSON)
	adder.BoolVar(&cfg.DryRun, "dry-run", false, usageDryRun)
	adder.BoolVar(&cfg.EarlyOut, "early-out", false, usageEarlyOut)
	return adder.set
}

// Run runs configuration from routines using configuration from cfg in
// an arbitrary order.
func (cfg *Config) Run(ctx context.Context) error {
	opts := &slog.HandlerOptions{}
	if cfg.Verbose {
		opts.Level = slog.LevelDebug
	} else {
		opts.Level = slog.LevelInfo
	}

	var h slog.Handler
	if cfg.JSON {
		h = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		var shutdown func()
		h, shutdown = logging.NewPrettyHandler(os.Stderr, opts)
		defer shutdown()
	}
	logger := slog.New(h)

	client, err := cfg.client(ctx, logger)
	if err != nil {
		return err
	}

	runCfg := automation.NewConfig(client).
		WithLogger(logger).
		WithDryRun(cfg.DryRun).
		WithEarlyOut(cfg.EarlyOut)
	if cfg.AppName != "" {
		runCfg = runCfg.WithAppName(cfg.AppName).WithLogger(logger)
	}

	var routines automation.Routines
	if len(cfg.Patterns) == 0 {
		routines = cfg.Routines
	} else {
		routines = cfg.Routines.SubRoutines(cfg.Patterns...)
	}
	return routines.Do(ctx, runCfg)
}

func (cfg *Config) client(ctx context.Context, logger *slog.Logger) (*clarify.Client, error) {
	var creds *clarify.Credentials
	switch {
	case cfg.Username != "" && cfg.Password.value == "":
		return nil, fmt.Errorf("-password: required when -username is specified")
	case cfg.CredentialsFile == "":
		return nil, fmt.Errorf("-credentials: required when -username is not specified")
	case cfg.Username != "":
		creds = clarify.BasicAuthCredentials(cfg.Username, cfg.Password.value)
	default:
		var err error
		creds, err = clarify.CredentialsFromFile(cfg.CredentialsFile)
		if err != nil {
			return nil, err
		}
	}

	h, err := creds.HTTPHandler(ctx)
	if err != nil {
		return nil, err
	}
	if cfg.Verbose && logger != nil {
		h.RequestLogger = func(request jsonrpc.Request, trace string, latency time.Duration, err error) {
			var b bytes.Buffer
			enc := json.NewEncoder(&b)
			_ = enc.Encode(request)
			logger.Debug("Performing JSON RPC request", "trace", trace, "latency", latency, "err", err, "body", json.RawMessage(b.Bytes()))
		}
	}

	return clarify.NewClient(creds.Integration, h), nil
}
