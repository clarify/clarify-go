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
	"log/slog"
)

// LogError return a routine that logs the provided message and attributes
// at error level.
//
// This routine can be useful for early testing.
func LogError(msg string, attrs ...slog.Attr) RoutineFunc {
	return logAttr(slog.LevelError, msg, attrs...)
}

// LogWarm return a routine that logs the provided message and attributes
// at warning level.
//
// This routine can be useful for early testing.
func LogWarm(msg string, attrs ...slog.Attr) RoutineFunc {
	return logAttr(slog.LevelWarn, msg, attrs...)
}

// LogInfo return a routine that logs the provided message and attributes
// at info level.
//
// This routine can be useful for early testing.
func LogInfo(msg string, attrs ...slog.Attr) RoutineFunc {
	return logAttr(slog.LevelInfo, msg, attrs...)
}

// LogDebug return a routine that logs the provided message and attributes at
// debug level.
//
// This routine can be useful for early testing.
func LogDebug(msg string, attrs ...slog.Attr) RoutineFunc {
	return logAttr(slog.LevelDebug, msg, attrs...)
}

// logAttr return a routine that logs the provided message and attributes
// at the provided level.
//
// This routine can be useful for early testing.
func logAttr(level slog.Level, msg string, attrs ...slog.Attr) RoutineFunc {
	return func(ctx context.Context, cfg *Config) error {
		logger := cfg.Logger()
		logger.LogAttrs(ctx, level, msg, attrs...)
		return nil
	}
}
