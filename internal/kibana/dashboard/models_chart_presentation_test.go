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

package dashboard

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_lensDrilldownItemFromAPIJSON_dispatch_and_trigger_defaults(t *testing.T) {
	t.Run("dashboard trigger omitted defaults", func(t *testing.T) {
		raw := []byte(`{"type":"dashboard_drilldown","dashboard_id":"d1","label":"L"}`)
		item, diags := lensDrilldownItemFromAPIJSON(raw)
		require.False(t, diags.HasError())
		require.NotNil(t, item.DashboardDrilldown)
		assert.Nil(t, item.DiscoverDrilldown)
		assert.Nil(t, item.URLDrilldown)
		assert.Equal(t, "d1", item.DashboardDrilldown.DashboardID.ValueString())
		assert.Equal(t, "L", item.DashboardDrilldown.Label.ValueString())
		assert.Equal(t, lensDrilldownTriggerOnApplyFilter, item.DashboardDrilldown.Trigger.ValueString())
	})

	t.Run("discover trigger omitted defaults", func(t *testing.T) {
		raw := []byte(`{"type":"discover_drilldown","label":"D"}`)
		item, diags := lensDrilldownItemFromAPIJSON(raw)
		require.False(t, diags.HasError())
		require.NotNil(t, item.DiscoverDrilldown)
		assert.Equal(t, "D", item.DiscoverDrilldown.Label.ValueString())
		assert.Equal(t, lensDrilldownTriggerOnApplyFilter, item.DiscoverDrilldown.Trigger.ValueString())
	})

	t.Run("url requires explicit trigger in payload", func(t *testing.T) {
		raw := []byte(`{"type":"url_drilldown","url":"https://x","label":"U","trigger":"on_open_panel_menu"}`)
		item, diags := lensDrilldownItemFromAPIJSON(raw)
		require.False(t, diags.HasError())
		require.NotNil(t, item.URLDrilldown)
		assert.Equal(t, "https://x", item.URLDrilldown.URL.ValueString())
		assert.Equal(t, "on_open_panel_menu", item.URLDrilldown.Trigger.ValueString())
	})
}

func Test_lensDrilldownsToRawJSON_variantCountErrors(t *testing.T) {
	t.Run("multiple variants set", func(t *testing.T) {
		item := models.LensDrilldownItemTFModel{
			DashboardDrilldown: &models.LensDashboardDrilldownTFModel{
				DashboardID: types.StringValue("d1"),
				Label:       types.StringValue("x"),
			},
			URLDrilldown: &models.LensURLDrilldownTFModel{
				URL:     types.StringValue("https://x"),
				Label:   types.StringValue("y"),
				Trigger: types.StringValue("on_click_row"),
			},
		}
		_, diags := lensDrilldownsToRawJSON([]models.LensDrilldownItemTFModel{item})
		require.True(t, diags.HasError())
	})

	t.Run("zero variants set", func(t *testing.T) {
		_, diags := lensDrilldownsToRawJSON([]models.LensDrilldownItemTFModel{{}})
		require.True(t, diags.HasError())
	})
}
