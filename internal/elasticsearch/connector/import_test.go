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

package connector

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseConnectorImportID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		importID        string
		wantConnectorID string
		wantErr         bool
	}{
		{
			name:            "bare id",
			importID:        "music",
			wantConnectorID: "music",
		},
		{
			name:            "whitespace trimmed",
			importID:        "  music  ",
			wantConnectorID: "music",
		},
		{
			name:            "composite id",
			importID:        "cluster-uuid/music",
			wantConnectorID: "music",
		},
		{
			name:            "composite with slash in resource id",
			importID:        "cluster-uuid/foo/bar",
			wantConnectorID: "foo/bar",
		},
		{
			name:            "trailing slash on bare id",
			importID:        "music/",
			wantConnectorID: "music",
		},
		{
			name:            "leading slash on bare id",
			importID:        "/music",
			wantConnectorID: "music",
		},
		{
			name:     "empty",
			importID: "  ",
			wantErr:  true,
		},
		{
			name:     "only slashes",
			importID: "///",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			connectorID, diags := parseConnectorImportID(tt.importID)
			if tt.wantErr {
				require.True(t, diags.HasError())
				return
			}
			require.False(t, diags.HasError())
			require.Equal(t, tt.wantConnectorID, connectorID)
		})
	}
}
