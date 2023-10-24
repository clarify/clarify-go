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
	"fmt"
	"log/slog"

	"github.com/clarify/clarify-go/views"
)

// Clarify annotations used by automation routines.
const (
	AnnotationPrefix = "clarify/clarify-go/"

	AnnotationPublisherName             = AnnotationPrefix + "publisher/name"
	AnnotationPublisherTransformVersion = AnnotationPrefix + "publisher/transform-version"
	AnnotationPublisherSignalID         = AnnotationPrefix + "publisher/signal-id"
	AnnotationPublisherSignalAttributes = AnnotationPrefix + "publisher/signal-attributes"
)

// AttrError return a log attribute for err.
func AttrError(err error) slog.Attr {
	return slog.Group("error",
		slog.String("message", err.Error()),
		slog.String("type", fmt.Sprintf("%T", err)),
	)
}

// AttrDataFrame returns a log attribute for data.
func AttrDataFrame(data views.DataFrame) slog.Attr {
	return slog.Any("data_frame", data)
}

func attrDryRun() slog.Attr {
	return slog.Bool("dry_run", true)
}

func attrAppName(name string) slog.Attr {
	return slog.String("app", name)
}

func attrRoutineName(name string) slog.Attr {
	return slog.String("routine", name)
}
