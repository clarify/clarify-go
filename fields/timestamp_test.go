// Copyright 2022 Searis AS
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

package fields_test

import (
	"testing"
	"time"

	data "github.com/clarify/clarify-go/fields"
)

func TestOriginTime(t *testing.T) {
	expectTime := time.Date(2000, 01, 03, 0, 0, 0, 0, time.UTC)
	if result := data.OriginTime.Time(); !result.Equal(expectTime) {
		t.Errorf("expected OriginZeroTime.Time() equal %v, got %v", expectTime, result)
	}
	if result := int64(data.OriginTime); result != expectTime.UnixMicro() {
		t.Errorf("expected OriginZeroTime equal %v, got %v", expectTime.UnixMicro(), result)
	}
}

func TestZeroTime(t *testing.T) {
	expectTime := time.Date(1970, 01, 01, 0, 0, 0, 0, time.UTC)
	if result := data.Timestamp(0).Time(); !result.Equal(expectTime) {
		t.Errorf("expected Timestamp(0).Time() equal %v, got %v", expectTime, result)
	}
	if result := int64(data.Timestamp(0)); result != expectTime.UnixMicro() {
		t.Errorf("expected Timestamp(0) equal %v, got %v", expectTime.UnixMicro(), result)
	}
}
