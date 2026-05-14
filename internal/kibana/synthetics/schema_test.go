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

package synthetics

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTryReadCompositeID(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		wantClusterID string
		wantResource  string
		wantError     bool
	}{
		{
			name: "bare ID",
			id:   "resource-id",
		},
		{
			name:          "composite ID",
			id:            "space-a/resource-id",
			wantClusterID: "space-a",
			wantResource:  "resource-id",
		},
		{
			name:      "malformed composite ID",
			id:        "space-a/resource-id/extra",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compositeID, diags := TryReadCompositeID(tt.id)

			if tt.wantError {
				require.True(t, diags.HasError())
				assert.Nil(t, compositeID)
				return
			}

			require.False(t, diags.HasError())
			if tt.wantClusterID == "" {
				assert.Nil(t, compositeID)
				return
			}

			require.NotNil(t, compositeID)
			assert.Equal(t, tt.wantClusterID, compositeID.ClusterID)
			assert.Equal(t, tt.wantResource, compositeID.ResourceID)
		})
	}
}
