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

package clarify_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/clarify/clarify-go/fields"
	"github.com/clarify/clarify-go/views"
)

func BenchmarkMergeDataFrame(b *testing.B) {
	cols := []int{1, 2, 4, 8, 40, 50}
	rows := []int{10, 100, 1000, 10000, 100000}
	now := fields.AsTimestamp(time.Now())

	for _, colCount := range cols {
		for _, rowCount := range rows {
			name := fmt.Sprintf("cols:%d,rows:%d", colCount, rowCount)
			df2 := make(views.DataFrame, colCount)
			for ci := 0; ci < colCount; ci++ {
				s := make(views.DataSeries, rowCount)
				for ri := 0; ri < rowCount; ri++ {
					s[now+fields.Timestamp(ri)] = float64(ri)
				}
				df2["sig"+strconv.Itoa(ci)] = s
			}
			b.Run(name, func(b *testing.B) {
				for n := 0; n < b.N; n++ {
					df1 := make(views.DataFrame)
					for ci := 0; ci < colCount; ci++ {
						s := make(views.DataSeries)
						for ri := 0; ri < rowCount; ri++ {
							s[now+fields.Timestamp(ri)] = float64(ri)
						}
						df1["sig"+strconv.Itoa(ci)] = s
					}

					b.ResetTimer()
					for k, s2 := range df2 {
						s1 := df1[k]
						if len(s1) == 0 {
							df1[k] = s2
						} else {
							for t, v2 := range s2 {
								df1[k][t] = v2
							}
						}
					}
				}
			})
		}
	}
}
