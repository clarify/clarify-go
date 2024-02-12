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

func getDefaultTimeRange() (time.Time, time.Time) {
	tidenesMorgen := time.Unix(0, 0).UTC().Truncate(time.Hour)
	tidenesKveld := tidenesMorgen.Add(10 * time.Hour)

	return tidenesMorgen, tidenesKveld
}

func testName() string {
	pc := make([]uintptr, 10)
	n := runtime.Callers(3, pc)
	frame, _ := runtime.CallersFrames(pc[:n]).Next()
	funks := strings.Split(frame.Function, ".")
	funk := funks[len(funks)-1]

	return funk
}

func createPrefix() string {
	test := testName()

	return test + "/"
}

func getCredentials(t *testing.T) *clarify.Credentials {
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

func jsonEncode[v any](t *testing.T, a v) {
	enc := json.NewEncoder(os.Stdout)

	enc.SetIndent("", "  ")

	err := enc.Encode(a)
	if err != nil {
		t.Errorf("%v", err)
	}
}

func createAnnotationQuery(prefix string) fields.ResourceQuery {
	return fields.Query().
		Where(fields.Comparisons{"annotations." + prefix + AnnotationKey: fields.Equal(AnnotationValue)}).
		Limit(10)
}

func onlyError[R any](f func(TestArgs) (R, error)) func(TestArgs) error {
	return func(a TestArgs) error {
		_, err := f(a)

		return err
	}
}

func applyTestArgs(a TestArgs, fs ...func(a TestArgs) error) {
	for _, f := range fs {
		if err := f(a); err != nil {
			panic(err)
		}
	}
}

type TestArgs struct {
	ctx         context.Context
	integration string
	client      *clarify.Client
	prefix      string
}

func Map[A any, B any](f func(a A) B, as []A) []B {
	g := func(index int, a A) B {
		return f(a)
	}

	return MapIndex(g, as)
}

func MapIndex[A any, B any](f func(i int, a A) B, as []A) []B {
	bs := make([]B, len(as))

	for i := range as {
		bs[i] = f(i, as[i])
	}

	return bs
}
