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

package securityuser

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/stretchr/testify/assert"
)

func TestIsEmptyJSONObject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value jsontypes.Normalized
		want  bool
	}{
		{
			name:  "null value returns false",
			value: jsontypes.NewNormalizedNull(),
			want:  false,
		},
		{
			name:  "unknown value returns false",
			value: jsontypes.NewNormalizedUnknown(),
			want:  false,
		},
		{
			name:  "empty JSON object returns true",
			value: jsontypes.NewNormalizedValue("{}"),
			want:  true,
		},
		{
			name:  "empty JSON object with whitespace returns true",
			value: jsontypes.NewNormalizedValue("{ }"),
			want:  true,
		},
		{
			name:  "non-empty JSON object returns false",
			value: jsontypes.NewNormalizedValue(`{"k":"v"}`),
			want:  false,
		},
		{
			name:  "JSON null literal returns false",
			value: jsontypes.NewNormalizedValue("null"),
			want:  false,
		},
		{
			name:  "invalid JSON returns false",
			value: jsontypes.NewNormalizedValue("not json"),
			want:  false,
		},
		{
			name:  "JSON array returns false",
			value: jsontypes.NewNormalizedValue("[]"),
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, isEmptyJSONObject(tt.value))
		})
	}
}
