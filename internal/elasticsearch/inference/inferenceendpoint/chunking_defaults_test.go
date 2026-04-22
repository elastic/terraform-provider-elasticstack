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

package inferenceendpoint

import (
	"encoding/json"
	"reflect"
	"testing"
)

func Test_populateChunkingSettingsDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input map[string]any
		want  map[string]any
	}{
		{
			name:  "nil becomes sentence defaults",
			input: nil,
			want: map[string]any{
				"strategy":         "sentence",
				"max_chunk_size":   float64(250),
				"sentence_overlap": float64(1),
			},
		},
		{
			name:  "empty map gets sentence defaults",
			input: map[string]any{},
			want: map[string]any{
				"strategy":         "sentence",
				"max_chunk_size":   float64(250),
				"sentence_overlap": float64(1),
			},
		},
		{
			name: "sentence partial only fills missing defaults",
			input: map[string]any{
				"strategy":       "sentence",
				"max_chunk_size": float64(100),
			},
			want: map[string]any{
				"strategy":         "sentence",
				"max_chunk_size":   float64(100),
				"sentence_overlap": float64(1),
			},
		},
		{
			name: "word strategy defaults",
			input: map[string]any{
				"strategy": "word",
			},
			want: map[string]any{
				"strategy":       "word",
				"max_chunk_size": float64(250),
				"overlap":        float64(100),
			},
		},
		{
			name: "recursive unchanged",
			input: map[string]any{
				"strategy":       "recursive",
				"max_chunk_size": float64(200),
				"separators":     []any{"\n"},
			},
			want: map[string]any{
				"strategy":       "recursive",
				"max_chunk_size": float64(200),
				"separators":     []any{"\n"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var before []byte
			if tt.input != nil {
				before, _ = json.Marshal(tt.input)
			}
			got := populateChunkingSettingsDefaults(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("populateChunkingSettingsDefaults() = %#v, want %#v", got, tt.want)
			}
			if tt.input != nil {
				after, _ := json.Marshal(tt.input)
				if string(before) != string(after) {
					t.Fatalf("input map was mutated: before %s after %s", before, after)
				}
			}
		})
	}
}
