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
	"encoding"
	"flag"
	"fmt"
	"os"
	"strings"
)

type flagSetAdder struct {
	envPrefix string
	set       *flag.FlagSet
}

func (set flagSetAdder) StringVar(target *string, name string, fallback string, usage string) {
	k := envKey(set.envPrefix, name)
	usage = fmt.Sprintf("%s (env: %s)", usage, k)
	if v := os.Getenv(k); v != "" {
		fallback = v
	}
	set.set.StringVar(target, name, fallback, usage)
}

func (set flagSetAdder) StringSliceVar(target *[]string, name string, fallback []string, usage string) {
	k := envKey(set.envPrefix, name)
	usage = fmt.Sprintf("%s (env: %s)", usage, k)
	if v := os.Getenv(k); v != "" {
		fallback = strings.Split(v, ",")
	}
	*target = fallback
	set.set.Var(stringSlice{target: target}, name, usage)
}

func (set flagSetAdder) BoolVar(target *bool, name string, fallback bool, usage string) {
	k := envKey(set.envPrefix, name)
	usage = fmt.Sprintf("%s (env: %s)", usage, k)
	if v := strings.ToLower(os.Getenv(k)); v != "" {
		fallback = (v == "true" || v == "1")
	}
	set.set.BoolVar(target, name, fallback, usage)
}

func envKey(prefix, name string) string {
	return prefix + strings.ReplaceAll(strings.ToUpper(name), "-", "_")
}

type stringSlice struct {
	target *[]string
}

var _ flag.Value = stringSlice{}

func (tf stringSlice) String() string {
	if tf.target == nil {
		return ""
	}
	return strings.Join(*tf.target, ",")
}

func (tf stringSlice) Set(v string) error {
	if v == "" {
		return nil
	}
	*tf.target = strings.Split(v, ",")
	return nil
}

// Password helps prevent a string value from being exposed in logs or
// marshalled as text or JSON.
type Password struct {
	value string
}

var (
	_ encoding.TextMarshaler   = Password{}
	_ encoding.TextUnmarshaler = &Password{}
)

func NewPassword(value string) Password {
	return Password{value: value}
}

func (p Password) String() string {
	return "****"
}

func (p Password) GoString() string {
	return `NewPassword("****")`
}

func (p *Password) UnmarshalText(data []byte) error {
	p.value = string(data)
	return nil
}

func (p Password) MarshalText() ([]byte, error) {
	return []byte("****"), nil
}
