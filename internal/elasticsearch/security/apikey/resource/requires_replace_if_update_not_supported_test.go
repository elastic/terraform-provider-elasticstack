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

package resource

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequiresReplaceIfUpdateNotSupported(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name            string
		private         testPrivateData
		wantReplace     bool
		wantDiagError   bool
		wantDiagSummary string
	}{
		{
			name:        "empty private state",
			private:     testPrivateData{},
			wantReplace: false,
		},
		{
			name: "new shape supports update",
			private: testPrivateData{
				clusterVersionPrivateDataKey: []byte(`{"SupportsUpdate":true,"SupportsRoleDescriptors":true,"SupportsRestriction":true}`),
			},
			wantReplace: false,
		},
		{
			name: "new shape all false requires replace",
			private: testPrivateData{
				clusterVersionPrivateDataKey: []byte(`{"SupportsUpdate":false,"SupportsRoleDescriptors":false,"SupportsRestriction":false}`),
			},
			wantReplace: true,
		},
		{
			name: "new shape update false with other flags true requires replace",
			private: testPrivateData{
				clusterVersionPrivateDataKey: []byte(`{"SupportsUpdate":false,"SupportsRoleDescriptors":true,"SupportsRestriction":true}`),
			},
			wantReplace: true,
		},
		{
			name: "legacy 7.0.0 requires replace",
			private: testPrivateData{
				clusterVersionPrivateDataKey: []byte(`{"Version":"7.0.0"}`),
			},
			wantReplace: true,
		},
		{
			name: "legacy 8.20.0 does not require replace",
			private: testPrivateData{
				clusterVersionPrivateDataKey: []byte(`{"Version":"8.20.0"}`),
			},
			wantReplace: false,
		},
		{
			name:        "legacy empty version treated as no data",
			private:     testPrivateData{clusterVersionPrivateDataKey: []byte(`{"Version":""}`)},
			wantReplace: false,
		},
		{
			name:            "malformed json surfaces diagnostics",
			private:         testPrivateData{clusterVersionPrivateDataKey: []byte(`{invalid`)},
			wantReplace:     false,
			wantDiagError:   true,
			wantDiagSummary: "failed to parse private data json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requiresReplace, diags := requiresReplaceBecauseUpdateNotSupported(ctx, tt.private)

			assert.Equal(t, tt.wantReplace, requiresReplace)
			assert.Equal(t, tt.wantDiagError, diags.HasError())
			if tt.wantDiagSummary != "" {
				require.True(t, diags.HasError())
				assert.Contains(t, diags.Errors()[0].Summary(), tt.wantDiagSummary)
			}
		})
	}
}
