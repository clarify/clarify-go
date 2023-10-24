// Copyright 2022-2023 Searis AS
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

// Package automation offers tools for building customized Clarify routines. Use
// the Routines type to build a three structure of named routines, and use the
// automationcli sub-package for setting up a command-line tool that can match
// and run these routines by name.
//
// While you could write all your routines from scratch, we provide a few
// pre-implemented types that you might find useful:
//   - Routines: Build a three-structure of sub-routines.
//   - PublishSignals: Define filters that inspect signals in your organization,
//     tracks changes to signal input, and publish them as new or existing items.
//     Apply custom transforms to improve your item meta-data before save.
//   - EvaluateActions: Run the powerful evaluate method against your Clarify
//     instance to detect conditions and trigger custom actions.
//   - LogDebug,LogInfo,LogWarn,LogError: Log a message to the console; useful
//     for debugging and testing.
package automation
