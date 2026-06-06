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

package clients_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestResolveCompositeSpaceAndID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		configSpaceID  types.String
		rawID          string
		wantSpaceID    string
		wantResourceID string
	}{
		{
			name:           "bare id, no explicit space: uses default space",
			configSpaceID:  types.StringNull(),
			rawID:          "my-resource",
			wantSpaceID:    clients.DefaultSpaceID,
			wantResourceID: "my-resource",
		},
		{
			name:           "composite id, no explicit space: extracts space and resource from composite",
			configSpaceID:  types.StringNull(),
			rawID:          "my-space/my-resource",
			wantSpaceID:    "my-space",
			wantResourceID: "my-resource",
		},
		{
			name:           "explicit space overrides composite space",
			configSpaceID:  types.StringValue("explicit-space"),
			rawID:          "other-space/my-resource",
			wantSpaceID:    "explicit-space",
			wantResourceID: "my-resource",
		},
		{
			name:           "explicit space with bare id",
			configSpaceID:  types.StringValue("my-space"),
			rawID:          "my-resource",
			wantSpaceID:    "my-space",
			wantResourceID: "my-resource",
		},
		{
			name:           "unknown space with bare id: uses default space",
			configSpaceID:  types.StringUnknown(),
			rawID:          "my-resource",
			wantSpaceID:    clients.DefaultSpaceID,
			wantResourceID: "my-resource",
		},
		{
			name:           "empty string space treated as not explicit: uses default space",
			configSpaceID:  types.StringValue(""),
			rawID:          "my-resource",
			wantSpaceID:    clients.DefaultSpaceID,
			wantResourceID: "my-resource",
		},
		{
			name:           "empty string space with composite id: uses composite space",
			configSpaceID:  types.StringValue(""),
			rawID:          "my-space/my-resource",
			wantSpaceID:    "my-space",
			wantResourceID: "my-resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotSpace, gotResource := clients.ResolveCompositeSpaceAndID(tt.configSpaceID, tt.rawID)
			require.Equal(t, tt.wantSpaceID, gotSpace)
			require.Equal(t, tt.wantResourceID, gotResource)
		})
	}
}
