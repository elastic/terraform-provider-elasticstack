// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package alertingactions

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDaysFromAPI(t *testing.T) {
	tests := []struct {
		name    string
		raw     any
		want    []int64
		wantErr bool
	}{
		{name: "typed int64 slice", raw: []int64{1, 2, 3}, want: []int64{1, 2, 3}},
		{name: "typed int32 slice", raw: []int32{1, 2, 3}, want: []int64{1, 2, 3}},
		{name: "typed int slice", raw: []int{1, 2, 3}, want: []int64{1, 2, 3}},
		{name: "empty any slice", raw: []any{}, want: []int64{}},
		{name: "any slice of float64 (encoding/json default)", raw: []any{float64(1), float64(2)}, want: []int64{1, 2}},
		{name: "any slice of json.Number (UseNumber decode)", raw: []any{json.Number("1"), json.Number("2")}, want: []int64{1, 2}},
		{name: "any slice of int", raw: []any{1, 2}, want: []int64{1, 2}},
		{name: "any slice of int64", raw: []any{int64(1), int64(2)}, want: []int64{1, 2}},
		{name: "any slice of mixed numeric types", raw: []any{float64(1), json.Number("2"), 3, int64(4)}, want: []int64{1, 2, 3, 4}},
		{name: "not an array", raw: "not-an-array", wantErr: true},
		{name: "any slice with invalid json.Number", raw: []any{json.Number("not-a-number")}, wantErr: true},
		{name: "any slice with unsupported element type", raw: []any{"monday"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DaysFromAPI(tt.raw)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestInt32FromInt64(t *testing.T) {
	require.Equal(t, []int32{1, 2, 3}, Int32FromInt64([]int64{1, 2, 3}))
	require.Equal(t, []int32{}, Int32FromInt64([]int64{}))
}

func TestIntFromInt64(t *testing.T) {
	require.Equal(t, []int{1, 2, 3}, IntFromInt64([]int64{1, 2, 3}))
	require.Equal(t, []int{}, IntFromInt64([]int64{}))
}

func TestInt32FromInt(t *testing.T) {
	require.Equal(t, []int32{1, 2, 3}, Int32FromInt([]int{1, 2, 3}))
	require.Equal(t, []int32{}, Int32FromInt([]int{}))
}

func TestIntFromInt32(t *testing.T) {
	require.Equal(t, []int{1, 2, 3}, IntFromInt32([]int32{1, 2, 3}))
	require.Equal(t, []int{}, IntFromInt32([]int32{}))
}
