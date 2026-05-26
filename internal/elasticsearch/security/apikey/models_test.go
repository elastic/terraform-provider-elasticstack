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

package apikey

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestTfModelGetReadResourceID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		model TfModel
		want  string
	}{
		{
			name: "returns_key_id_when_known_nonempty",
			model: TfModel{
				KeyID: types.StringValue("U-abc"),
				Name:  types.StringValue("friendly-name"),
				ID:    types.StringValue("cluster-uuid/U-other"),
			},
			want: "U-abc",
		},
		{
			name: "parses_resource_id_from_composite_when_key_id_empty",
			model: TfModel{
				KeyID: types.StringValue(""),
				Name:  types.StringValue("friendly-name"),
				ID:    types.StringValue("cluster-uuid/the-key-id-segment"),
			},
			want: "the-key-id-segment",
		},
		{
			name: "parses_resource_id_from_composite_when_key_id_unknown",
			model: TfModel{
				KeyID: types.StringUnknown(),
				ID:    types.StringValue("cluster-uuid/k-from-composite"),
			},
			want: "k-from-composite",
		},
		{
			name: "empty_when_no_key_id_and_no_id",
			model: TfModel{
				KeyID: types.StringNull(),
				Name:  types.StringValue("only-name"),
				ID:    types.StringNull(),
			},
			want: "",
		},
		{
			name: "empty_when_composite_id_invalid",
			model: TfModel{
				KeyID: types.StringValue(""),
				ID:    types.StringValue("not-a-composite-id"),
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.model.GetReadResourceID(); got != tt.want {
				t.Fatalf("GetReadResourceID() = %q, want %q", got, tt.want)
			}
		})
	}
}
