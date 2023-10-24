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

package automationcli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/clarify/clarify-go"
	"github.com/clarify/clarify-go/automation"
)

var (
	defaultAppName  string
	defaultProgName string
)

func init() {
	// Set default app name.
	if info, ok := debug.ReadBuildInfo(); ok {
		defaultAppName = info.Main.Path
	}
	if len(os.Args) > 0 {
		defaultProgName = os.Args[0]
	}
}

const (
	usageCredentials = "Specify the path to your Clarify Integration's credentials file."
	usageUsername    = "Clarify integration ID to use as username; alternative to providing -credentials."
	usagePassword    = "Clarify integration password; required when username is set, ignored otherwise."
	usageVerbose     = "Set to true for printing logs at level DEBUG (the default is to log at INFO level)."
	usageDryRun      = "Signal to routines that they should mot write or persist changes."
	usageEarlyOut    = "Signal to routines that they should mot write or persist changes."
	usageArgs        = `
PATTERNS are expected to match routine or sub-routine names. Use slash (/) to
match specific sub-paths or asterisk (*) to match all sub-routines as a given
path level.
`
)

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

	// Patterns describes which sub-routines to run. If empty, all sub-routines
	// are run.
	Patterns []string

	// Verbose, if set, turns on DEBUG logging. The default log level is INFO.
	Verbose bool

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
func ParseArguments(arguments []string) (*Config, error) {
	var cfg Config
	cfg.AppName = defaultAppName
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
func (cfg *Config) FlagSet(name string, errorHandling flag.ErrorHandling) *flag.FlagSet {
	if name == "" {
		name = defaultProgName
	}
	adder := flagSetAdder{
		envPrefix: "CLARIFY_",
		set:       flag.NewFlagSet(name, errorHandling),
	}
	adder.set.Usage = func() {
		out := adder.set.Output()
		fmt.Fprintf(out, "Usage: %s [OPTIONS] [PATTERNS, ...]\n", name)
		fmt.Fprint(out, usageArgs+"\n")
		flag.PrintDefaults()
	}

	adder.StringVar(&cfg.CredentialsFile, "credentials", "credentials.json", usageCredentials)
	adder.StringVar(&cfg.Username, "username", "", usageUsername)
	adder.StringVar(&cfg.Password.value, "password", "", usagePassword)
	adder.BoolVar(&cfg.Verbose, "v", false, usageVerbose)
	adder.BoolVar(&cfg.DryRun, "dry-run", false, usageDryRun)
	adder.BoolVar(&cfg.EarlyOut, "early-out", false, usageEarlyOut)
	return adder.set
}

// Client returns an http.Client according to the configuration or an error in
// case of invalid configuration. Note that invalid or expired credentials is
// not considered a configuration error, and will still result in a client being
// returned.
func (p Config) Client(ctx context.Context) (*clarify.Client, error) {
	var err error
	var creds *clarify.Credentials
	switch {
	case p.Username != "" && p.Password.value == "":
		return nil, fmt.Errorf("-password: required when -username is specified")
	case p.CredentialsFile == "":
		return nil, fmt.Errorf("-credentials: required when -username is not specified")
	case p.Username != "":
		creds = clarify.BasicAuthCredentials(p.Username, p.Password.value)
	default:
		creds, err = clarify.CredentialsFromFile(p.CredentialsFile)
	}
	if err != nil {
		return nil, err
	}
	return creds.Client(ctx), nil
}

// Run runs configuration from routines using configuration from cfg in
// an arbitrary order.
func (cfg *Config) Run(ctx context.Context, routines automation.Routines) error {
	client, err := cfg.Client(ctx)
	if err != nil {
		return err
	}
	if len(cfg.Patterns) > 0 {
		routines = routines.SubRoutines(cfg.Patterns...)

	}
	routineCfg := automation.NewConfig(client).
		WithDryRun(cfg.DryRun).
		WithEarlyOut(cfg.EarlyOut)

	if cfg.AppName != "" {
		routineCfg = routineCfg.WithAppName(cfg.AppName)
	}
	return routines.Do(ctx, routineCfg)
}
