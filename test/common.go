// Copyright 2024 Searis AS
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

package test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/clarify/clarify-go"

	"github.com/clarify/clarify-go/fields"
)

const (
	AnnotationKey   = "there-is-always-money"
	AnnotationValue = "in-the-banana-stand"
)

// timeRange returns a default time range used for test queries.
func timeRange() (time.Time, time.Time) {
	start := time.Unix(0, 0).UTC().Truncate(time.Hour)
	end := start.Add(10 * time.Hour)

	return start, end
}

// TestName returns the name of the calling test function.
func testName() string {
	pc := make([]uintptr, 10)
	n := runtime.Callers(3, pc)
	frame, _ := runtime.CallersFrames(pc[:n]).Next()
	funks := strings.Split(frame.Function, ".")
	funk := funks[len(funks)-1]

	return funk
}

// prefixFromTestName creates a prefix based on the test name.
func prefixFromTestName() string {
	test := testName()

	return test + "/"
}

// credentialsFromEnv reads Clarify credentials from environment variables.
func credentialsFromEnv(t *testing.T) *clarify.Credentials {
	t.Helper()

	var creds *clarify.Credentials

	username := os.Getenv("CLARIFY_USERNAME")
	password := os.Getenv("CLARIFY_PASSWORD")
	endpoint := os.Getenv("CLARIFY_ENDPOINT")
	credentials := os.Getenv("CLARIFY_CREDENTIALS")

	switch {
	case username != "" && password != "":
		creds = clarify.BasicAuthCredentials(username, password)

		if endpoint != "" {
			creds.APIURL = endpoint
		}
	case credentials != "":
		var err error
		creds, err = clarify.CredentialsFromFile(credentials)
		if err != nil {
			t.Fatalf("failed to parse credentials file: %s", err)
		}
	default:
		t.Skip("no credentials found, skipping integration tests")
	}

	if err := creds.Validate(); err != nil {
		t.Fatalf("invalid credentials: %s", err)
	}

	return creds
}

// mustPrintJSON pretty-prints a value to stdout for test diagnostics.
func mustPrintJSON[v any](t *testing.T, a v) {
	t.Helper()
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(a); err != nil {
		t.Errorf("json encoding failed: %v", err)
	}
}

// annotationQuery builds a filter for test resources based on a known annotation.
func annotationQuery(prefix string) fields.ResourceQuery {
	return fields.Query().
		Where(fields.Comparisons{"annotations." + prefix + AnnotationKey: fields.Equal(AnnotationValue)}).
		Limit(10)
}

// onlyError wraps a function that returns a result and error,
// and discards the result.
func onlyError[R any](f func(TestArgs) (R, error)) func(TestArgs) error {
	return func(a TestArgs) error {
		_, err := f(a)

		return err
	}
}

// mustApplyTestArgs applies a series of test functions, panicking on error.
func mustApplyTestArgs(a TestArgs, fs ...func(a TestArgs) error) {
	for i, f := range fs {
		if err := f(a); err != nil {
			panic(fmt.Errorf("applyTestArgs fs[%d]: %w", i, err))
		}
	}
}

// TestArgs bundles common arguments for test operations.
type TestArgs struct {
	ctx         context.Context
	integration string
	client      *clarify.Client
	prefix      string
}

// Map transforms a slice using the provided function.
func Map[A any, B any](f func(a A) B, as []A) []B {
	g := func(index int, a A) B {
		return f(a)
	}

	return MapIndex(g, as)
}

// MapIndex transforms a slice using the provided index-aware function.
func MapIndex[A any, B any](f func(i int, a A) B, as []A) []B {
	bs := make([]B, len(as))

	for i := range as {
		bs[i] = f(i, as[i])
	}

	return bs
}
