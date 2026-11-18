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

package integration

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

// TestIntegrationModel_GetResourceID_pinsHashedFormat locks in the current
// hashed resource ID format so the typeutils.HashedPackageID consolidation
// cannot silently change existing state IDs.
func TestIntegrationModel_GetResourceID_pinsHashedFormat(t *testing.T) {
	t.Parallel()

	m := integrationModel{
		Name:    types.StringValue("nginx"),
		Version: types.StringValue("1.2.3"),
	}

	want := typeutils.HashedPackageID("nginx", "1.2.3")
	require.Equal(t, want, m.GetResourceID().ValueString())
	// Pin the literal hash too, so a future change to HashedPackageID's
	// own implementation is also caught here.
	require.Equal(t, "697d06a88f7147a4e6a5ea127ffd8f5b2cacbc4c", m.GetResourceID().ValueString())
}

func TestIntegrationModel_GetVersionRequirements_allowOutdatedVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     types.Bool
		wantGated bool
	}{
		{name: "null is not gated", value: types.BoolNull()},
		{name: "unknown is not gated", value: types.BoolUnknown()},
		{name: "explicit false is gated", value: types.BoolValue(false), wantGated: true},
		{name: "explicit true is gated", value: types.BoolValue(true), wantGated: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reqs, diags := integrationModel{AllowOutdatedVersion: tt.value}.GetVersionRequirements(t.Context())
			require.False(t, diags.HasError())
			if !tt.wantGated {
				require.Empty(t, reqs)
				return
			}
			require.Len(t, reqs, 1)
			require.True(t, reqs[0].MinVersion.Equal(MinVersionAllowOutdatedVersion))
		})
	}
}
