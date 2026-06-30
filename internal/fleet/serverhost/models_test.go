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

package serverhost

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToAPICreateModel_HostID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		value  types.String
		wantID *string
	}{
		{
			name:  "null host_id",
			value: types.StringNull(),
		},
		{
			name:  "unknown host_id",
			value: types.StringUnknown(),
		},
		{
			name:   "explicit host_id",
			value:  types.StringValue("my-host-id"),
			wantID: new("my-host-id"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			body, diags := serverHostModel{
				HostID: tc.value,
				Name:   types.StringValue("test-host"),
				Hosts: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("https://fleet-server:8220"),
				}),
			}.toAPICreateModel(t.Context())

			require.False(t, diags.HasError())
			assert.Equal(t, tc.wantID, body.Id)
		})
	}
}
