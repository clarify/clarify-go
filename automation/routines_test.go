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

package automation_test

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/clarify/clarify-go/automation"
)

func TestRoutinesSubRoutines(t *testing.T) {
	all := automation.Routines{
		"folder1": automation.Routines{
			"folder1": automation.Routines{
				"routine1": automation.LogInfo("OK"),
				"routine2": automation.LogInfo("OK"),
			},
			"folder2": automation.Routines{
				"routine1": automation.LogInfo("OK"),
				"routine2": automation.LogInfo("OK"),
			},
		},
		"folder2": automation.Routines{
			"folder1": automation.Routines{
				"routine1": automation.LogInfo("OK"),
				"routine2": automation.LogInfo("OK"),
			},
		},
		"routine1": automation.LogInfo("OK"),
		"routine2": automation.LogInfo("OK"),
	}
	type testCase struct {
		patterns    []string
		expectLines []string
	}

	test := func(tc testCase) func(t *testing.T) {
		return func(t *testing.T) {
			t.Helper()

			var buf bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
				Level: slog.LevelInfo,
				ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
					if len(groups) == 0 && a.Key == "time" {
						// Remove time attribute for easier log comparison.
						return slog.Attr{}
					}
					return a
				},
			}))

			ctx := context.Background()
			cfg := automation.
				NewConfig(nil).
				WithLogger(logger)

			routines := all.SubRoutines(tc.patterns...)
			if err := routines.Do(ctx, cfg); err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
			if diff := diffLines(tc.expectLines, lines); len(diff) > 0 {
				t.Errorf("Result does not match expectations:\n%s", diff)
			}
		}
	}

	t.Run("wildcard", test(testCase{
		patterns: []string{"*"},
		expectLines: []string{
			`level=INFO msg=OK routine=folder1/folder1/routine1`,
			`level=INFO msg=OK routine=folder1/folder1/routine2`,
			`level=INFO msg=OK routine=folder1/folder2/routine1`,
			`level=INFO msg=OK routine=folder1/folder2/routine2`,
			`level=INFO msg=OK routine=folder2/folder1/routine1`,
			`level=INFO msg=OK routine=folder2/folder1/routine2`,
			`level=INFO msg=OK routine=routine1`,
			`level=INFO msg=OK routine=routine2`,
		},
	}))
	t.Run("wildcard wildcard", test(testCase{
		patterns: []string{"*/*"},
		expectLines: []string{
			`level=INFO msg=OK routine=folder1/folder1/routine1`,
			`level=INFO msg=OK routine=folder1/folder1/routine2`,
			`level=INFO msg=OK routine=folder1/folder2/routine1`,
			`level=INFO msg=OK routine=folder1/folder2/routine2`,
			`level=INFO msg=OK routine=folder2/folder1/routine1`,
			`level=INFO msg=OK routine=folder2/folder1/routine2`,
		},
	}))
	t.Run("folder1,folder2", test(testCase{
		patterns: []string{"folder1", "folder2"},
		expectLines: []string{
			`level=INFO msg=OK routine=folder1/folder1/routine1`,
			`level=INFO msg=OK routine=folder1/folder1/routine2`,
			`level=INFO msg=OK routine=folder1/folder2/routine1`,
			`level=INFO msg=OK routine=folder1/folder2/routine2`,
			`level=INFO msg=OK routine=folder2/folder1/routine1`,
			`level=INFO msg=OK routine=folder2/folder1/routine2`,
		},
	}))
	t.Run("folder1", test(testCase{
		patterns: []string{"folder1"},
		expectLines: []string{
			`level=INFO msg=OK routine=folder1/folder1/routine1`,
			`level=INFO msg=OK routine=folder1/folder1/routine2`,
			`level=INFO msg=OK routine=folder1/folder2/routine1`,
			`level=INFO msg=OK routine=folder1/folder2/routine2`,
		},
	}))
	t.Run("folder1 wildcard routine1", test(testCase{
		patterns: []string{"folder1/*/routine1"},
		expectLines: []string{
			`level=INFO msg=OK routine=folder1/folder1/routine1`,
			`level=INFO msg=OK routine=folder1/folder2/routine1`,
		},
	}))
	t.Run("wildcard wildcard routine2", test(testCase{
		patterns: []string{"*/*/routine2"},
		expectLines: []string{
			`level=INFO msg=OK routine=folder1/folder1/routine2`,
			`level=INFO msg=OK routine=folder1/folder2/routine2`,
			`level=INFO msg=OK routine=folder2/folder1/routine2`,
		},
	}))
	t.Run("routine1", test(testCase{
		patterns: []string{"routine1"},
		expectLines: []string{
			`level=INFO msg=OK routine=routine1`,
		},
	}))
}

func diffLines(expect, result []string) string {
	var buf bytes.Buffer
	for i, e := range expect {
		switch {
		case len(result) < i+1:
			fmt.Fprintf(&buf, "- %s\n", e)
		case result[i] != e:
			fmt.Fprintf(&buf, "- %s\n", e)
			fmt.Fprintf(&buf, "+ %s\n", result[i])
		}
	}
	if len(result) > len(expect) {
		for _, r := range result[len(expect):] {
			fmt.Fprintf(&buf, "+ %s\n", r)
		}
	}
	return buf.String()
}
