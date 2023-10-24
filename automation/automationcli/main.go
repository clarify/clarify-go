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
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/clarify/clarify-go/automation"
)

// ParseAndRun parses command-line options and runs content from routines. On
// completion, the function return an exit status that should be passed on to
// os.Exit.
func ParseAndRun(routines automation.Routines) int {
	cfg, err := ParseArguments(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s", os.Args[0], err.Error())
		return 2
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	err = cfg.Run(ctx, routines)
	switch {
	case errors.Is(err, context.Canceled):
		fmt.Fprintf(os.Stderr, "%s: interrupt", os.Args[0])
		return 127
	case err != nil:
		fmt.Fprintf(os.Stderr, "%s: %s", os.Args[0], err.Error())
		return 1
	}
	return 0
}
