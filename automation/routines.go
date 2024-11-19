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

package automation

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"slices"
	"strings"
)

// Routines describe a set of named (sub-)routines. Routines can be nested by
// letting the value be a Routines entry.
//
// For usability reasons, keys are recommended to only contain ASCII
// alphanumerical characters (0-9, A-Z, a-z), dash (-) and underscore (_).
//
// Keys must not contain the slash (/) or asterisk (*) characters as they hold
// special meaning during matching. The question mark (?) character should be
// considered reserved. Keys must also not be empty strings. Failing to follow
// these restrictions will result in undefined behavior for the SubRoutines
// method.
type Routines map[string]Routine

func (routines Routines) Print(w io.Writer, indent string) {
	for _, k := range slices.Sorted(maps.Keys(routines)) {
		if sub, ok := routines[k].(Routines); ok {
			fmt.Fprintf(w, "%s%s/\n", indent, k)
			sub.Print(w, indent+"  ")
		} else {
			fmt.Fprintf(w, "%s%s\n", indent, k)
		}
	}
}

// SubRoutines returns a sub-set composed of routines that matches the passed in
// patterns. When routines are nested, the slash character (/) can be used to
// match nested entries. The asterisk (*) character will match all entries at
// the given level.
//
// Examples:
//   - "*", "*/": matches all entries.
//   - "a" or "a/": Match sub-routine "a" with sub-routines.
//   - "a/*/b": Match sub-routine "b" for all sub routines of sub-routine "a".
func (routines Routines) SubRoutines(patterns ...string) Routines {
	// Early out if there is nothing to filter.
	if len(routines) == 0 {
		return routines
	}

	// Construct a nested lookup map without duplicates, or early out on a match
	// all condition.
	//
	// The map uses the first element of the path as a key. As a special case
	// "*" will match all.
	var matchAll bool
	lookup := make(map[string][]string, len(patterns))
LOOKUP:
	for _, path := range patterns {
		name, nestedPath, _ := strings.Cut(path, "/")

		var found bool
		if name == "*" {
			found = true
		} else {
			_, found = routines[name]
		}

		switch {
		case !found:
			// Entry not found; nothing to do.
		case name == "*" && nestedPath == "":
			// Match all or end of path; early out.
			matchAll = true
			break LOOKUP
			// Routine not found; continue.
		case len(lookup[name]) == 1 && lookup[name][0] == "":
			// Path already match with an end of-path criteria; nothing to do.
		case nestedPath == "":
			// End of path; replace existing lookup as the end-of-path criteria
			// match all cases.
			lookup[name] = []string{""}
		case len(lookup[name]) == 1 && lookup[name][0] == "*":
			// Path already match all sub-routines; nothing to do.
		case nestedPath == "*":
			// Match all sub-routines; replace existing lookup with a wildcard
			// criteria.
			lookup[name] = []string{"*"}
		default:
			// Append nested lookup path.
			lookup[name] = append(lookup[name], nestedPath)
		}
	}

	if matchAll {
		return routines
	}

	// Filter routines based on the lookup map.
	filtered := make(Routines, len(patterns))
	var nestedPath []string
	for name, r := range routines {
		// Reset subPatterns before use.
		nestedPath = nestedPath[:0]
		// Add all patterns that apply to name.
		nestedPath = append(nestedPath, lookup["*"]...)
		nestedPath = append(nestedPath, lookup[name]...)

		slices.Sort(nestedPath)
		rs, canNest := r.(Routines)
		switch {
		case len(nestedPath) == 0:
			// No lookup matching the routine; skip entry.
		case slices.Contains(nestedPath, ""):
			// End of path; add routine.
			filtered[name] = r
		case !canNest:
			// Remaining matchers require the canNest property.
		case slices.Contains(nestedPath, "*"):
			// Match all sub-routines.
			filtered[name] = rs
		default:
			// Match named sub-routines.
			filtered[name] = rs.SubRoutines(nestedPath...)
		}
	}

	return filtered
}

// Do runs the member routines in an alphanumerical order and assigns correct
// sub-routine names. If cfg.EarlyOut() returns true, return at the first error.
// Otherwise log the error and continue.
func (routines Routines) Do(ctx context.Context, cfg *Config) error {
	earlyOut := cfg.EarlyOut()

	keys := make([]string, 0, len(routines))
	for k := range routines {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	var errCnt int
	for _, k := range keys {
		r := routines[k]
		cfg := cfg.WithSubRoutineName(k)
		logger := cfg.Logger()
		if r == nil {
			cfg.Logger().LogAttrs(ctx, slog.LevelWarn, "Routine is nil")
			continue
		}
		logger.LogAttrs(ctx, slog.LevelDebug, "Routine started")
		if err := r.Do(ctx, cfg); err != nil {
			if earlyOut {
				return fmt.Errorf("%s: %w", k, err)
			}
			cfg.Logger().LogAttrs(ctx, slog.LevelError, "Failed", AttrError(err))
			errCnt++
		} else {
			logger.LogAttrs(ctx, slog.LevelDebug, "OK")
		}
	}
	if errCnt > 0 {
		return fmt.Errorf("%d/%d routines failed", errCnt, len(routines))
	}

	return nil
}
