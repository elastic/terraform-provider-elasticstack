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

package slo

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestTfModel_GetVersionRequirements(t *testing.T) {
	t.Parallel()

	preventInitialBackfillMessage := "The 'prevent_initial_backfill' setting requires Elastic Stack version " +
		SLOSupportsPreventInitialBackfillMinVersion.String() + " or higher."
	dataViewIDMessage := "data_view_id is not supported on Elastic Stack versions < " +
		SLOSupportsDataViewIDMinVersion.String()

	settingsWithPreventInitialBackfill := func(t *testing.T) types.Object {
		t.Helper()
		obj, diags := types.ObjectValue(tfSettingsAttrTypes, map[string]attr.Value{
			"sync_delay":               types.StringNull(),
			"frequency":                types.StringNull(),
			"sync_field":               types.StringNull(),
			"prevent_initial_backfill": types.BoolValue(true),
		})
		require.False(t, diags.HasError())
		return obj
	}

	modelWithDataViewID := tfModel{
		KqlCustomIndicator: []tfKqlCustomIndicator{{
			DataViewID: types.StringValue("dv-1"),
		}},
	}

	assertRequirements := func(t *testing.T, reqs []entitycore.VersionRequirement, wantCount int, wantMessages ...string) {
		t.Helper()
		require.Len(t, reqs, wantCount)
		gotMessages := make(map[string]struct{}, len(reqs))
		for _, req := range reqs {
			gotMessages[req.ErrorMessage] = struct{}{}
		}
		for _, msg := range wantMessages {
			require.Contains(t, gotMessages, msg)
		}
	}

	t.Run("neither condition set", func(t *testing.T) {
		t.Parallel()
		reqs, diags := (tfModel{}).GetVersionRequirements()
		require.False(t, diags.HasError())
		assertRequirements(t, reqs, 0)
	})

	t.Run("only prevent_initial_backfill", func(t *testing.T) {
		t.Parallel()
		m := tfModel{Settings: settingsWithPreventInitialBackfill(t)}
		reqs, diags := m.GetVersionRequirements()
		require.False(t, diags.HasError())
		require.Len(t, reqs, 1)
		require.Equal(t, *SLOSupportsPreventInitialBackfillMinVersion, reqs[0].MinVersion)
		require.Equal(t, preventInitialBackfillMessage, reqs[0].ErrorMessage)
	})

	t.Run("only data_view_id", func(t *testing.T) {
		t.Parallel()
		reqs, diags := modelWithDataViewID.GetVersionRequirements()
		require.False(t, diags.HasError())
		require.Len(t, reqs, 1)
		require.Equal(t, *SLOSupportsDataViewIDMinVersion, reqs[0].MinVersion)
		require.Equal(t, dataViewIDMessage, reqs[0].ErrorMessage)
	})

	t.Run("both conditions set", func(t *testing.T) {
		t.Parallel()
		m := modelWithDataViewID
		m.Settings = settingsWithPreventInitialBackfill(t)
		reqs, diags := m.GetVersionRequirements()
		require.False(t, diags.HasError())
		assertRequirements(t, reqs, 2, preventInitialBackfillMessage, dataViewIDMessage)
	})
}
