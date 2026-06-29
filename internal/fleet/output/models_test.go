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

package output

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildCommonNewOutput_OutputID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		value  types.String
		wantID *string
	}{
		{
			name:  "null output_id",
			value: types.StringNull(),
		},
		{
			name:  "unknown output_id",
			value: types.StringUnknown(),
		},
		{
			name:   "explicit output_id",
			value:  types.StringValue("my-output-id"),
			wantID: new("my-output-id"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var diags diag.Diagnostics
			body := outputModel{
				OutputID: tc.value,
				Name:     types.StringValue("test-output"),
				Hosts:    types.ListNull(types.StringType),
				Ssl:      types.ObjectNull(outputSSLAttrTypes()),
			}.buildCommonNewOutput(t.Context(), &diags)

			require.False(t, diags.HasError())
			assert.Equal(t, tc.wantID, body.ID)
		})
	}
}

func outputSSLAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"certificate_authorities": types.ListType{ElemType: types.StringType},
		"certificate":             types.StringType,
		"key":                     types.StringType,
		"verification_mode":       types.StringType,
	}
}
