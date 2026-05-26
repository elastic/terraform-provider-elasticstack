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

package privatelocation

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestModel_GetVersionRequirements(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		model       Model
		wantReqs    int
		wantVersion bool
	}{
		{
			name:     "default space empty SpaceID null ID",
			model:    Model{},
			wantReqs: 0,
		},
		{
			name: "SpaceID default",
			model: Model{
				SpaceID: types.StringValue("default"),
			},
			wantReqs: 0,
		},
		{
			name: "SpaceID production",
			model: Model{
				SpaceID: types.StringValue("production"),
			},
			wantReqs:    1,
			wantVersion: true,
		},
		{
			name: "composite production import empty SpaceID",
			model: Model{
				ID: types.StringValue("production/uuid-123"),
			},
			wantReqs:    1,
			wantVersion: true,
		},
		{
			name: "composite default import empty SpaceID",
			model: Model{
				ID: types.StringValue("default/uuid-123"),
			},
			wantReqs: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reqs, diags := tt.model.GetVersionRequirements()
			require.False(t, diags.HasError())
			require.Len(t, reqs, tt.wantReqs)

			if tt.wantVersion {
				require.Equal(t, *MinVersionSpaceID, reqs[0].MinVersion)
				require.NotEmpty(t, reqs[0].ErrorMessage)
			}
		})
	}
}

func TestVersionGateSpaceID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		model Model
		want  string
	}{
		{
			name:  "empty state",
			model: Model{},
			want:  "",
		},
		{
			name: "composite import",
			model: Model{
				ID: types.StringValue("production/uuid-123"),
			},
			want: "production",
		},
		{
			name: "SpaceID wins over composite",
			model: Model{
				ID:      types.StringValue("production/uuid-123"),
				SpaceID: types.StringValue("staging"),
			},
			want: "staging",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, versionGateSpaceID(tt.model))
		})
	}
}
