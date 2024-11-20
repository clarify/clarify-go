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

package logging

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"slices"
	"strings"
	"sync"
)

const (
	colorReset   = "\033[0m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorGray    = "\033[37m"
	colorWhite   = "\033[97m"

	colorBrightRed     = "\033[31;1m"
	colorBrightGreen   = "\033[32;1m"
	colorBrightYellow  = "\033[33;1m"
	colorBrightBlue    = "\033[34;1m"
	colorBrightMagenta = "\033[35;1m"
	colorBrightCyan    = "\033[36;1m"
	colorBrightGray    = "\033[37;1m"
	colorBrightWhite   = "\033[97;1m"
)

// NewPrettyHandler creates a background go routine that will write prettified
// output to logs written using the returned [slog.JSONHandler] instance.
//
// To shut-down the background worker and wait for it's return, the caller
// should defer a call to the returned shutdown function.
func NewPrettyHandler(out io.Writer, opts *slog.HandlerOptions) (h *slog.JSONHandler, shutdown func()) {
	in, shutdown := newPrettyWriter(out)
	h = slog.NewJSONHandler(in, opts)
	return h, shutdown
}

// newPrettyWriter starts a background goroutine that reads log output from the
// returned in and writing formatted log output to out.
func newPrettyWriter(out io.Writer) (in io.Writer, shutdown func()) {
	pr, pw := io.Pipe()
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		prettyLog(pr, out)
	}()

	shutdown = func() {
		_ = pw.Close()
		wg.Wait()
		_ = pr.Close()
	}
	return pw, shutdown
}

func prettyLog(r io.Reader, w io.Writer) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		b := scanner.Bytes()
		m := make(map[string]json.RawMessage)

		if err := json.Unmarshal(b, &m); err != nil {
			// Fallback to print line.
			fmt.Fprintln(w, string(b))
			continue
		}

		level := strings.ToUpper(formatString(m["level"], ""))
		levelColor := colorBrightGray
		switch level {
		case "INFO":
			levelColor = colorBrightWhite
		case "ERROR":
			levelColor = colorBrightRed
		}

		fmt.Fprintln(w,
			colorGreen+formatString(m["time"], "\t")+colorReset,
			levelColor+"["+level+"]"+colorReset,
			formatString(m["msg"], "\t"),
		)
		delete(m, "time")
		delete(m, "level")
		delete(m, "msg")
		for _, k := range slices.Sorted(maps.Keys(m)) {
			data := m[k]
			switch data[0] {
			case '"':
				// Format single-line or multi-line string value.
				fmt.Fprintf(w, "\t%s%s:%s %s\n", colorGray, k, colorReset, formatString(data, "\t\t"))
			default:
				// Format all other JSON types inline JSON.
				fmt.Printf("\t%s%s:%s %s\n", colorGray, k, colorReset, data)
			}
		}
	}
}

func formatString(data json.RawMessage, indent string) string {
	var s string
	_ = json.Unmarshal(data, &s)
	if strings.ContainsRune(s, '\n') {
		s = strings.ReplaceAll(s, "\n", "\n"+indent)
		return fmt.Sprintf("\n%s%s\n", indent, s)
	}
	return s
}
