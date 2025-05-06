// Copyright 2025 Searis AS
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

package clarifyx

import (
	"github.com/clarify/clarify-go/internal/request"
	"github.com/clarify/clarify-go/views"
)

type AdminSignals struct {
	AdminNamespace
}

func (ns AdminNamespace) Signals() AdminSignals {
	return AdminSignals{AdminNamespace: ns}
}

// Annotate returns a new request for publishing signals as items.
func (s AdminSignals) Annotate(integration string, data []views.SignalAnnotate) SignalsAnnotateRequest {
	return methodSignalsAnnotate.NewRequest(s.Handler(),
		paramIntegration.Value(integration),
		paramData.Value(data),
		paramFormat.Value(views.SelectionFormat{
			DataAsArray:         true,
			GroupIncludedByType: true,
		}),
	)
}

type (
	// SignalsAnnotateRequest describe an initialized admin.signals.annotate RPC
	// request with access to a request handler.
	SignalsAnnotateRequest = request.Request[SignalsAnnotateResult]

	// SignalsAnnotateResult describe the result format for a SignalsAnnotateRequest.
	SignalsAnnotateResult = views.Selection[[]views.Signal, views.SignalInclude]
)

var methodSignalsAnnotate = request.Method[SignalsAnnotateResult]{
	APIVersion: apiVersionExperimental,
	Method:     "admin.signals.annotate",
}
