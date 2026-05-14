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

package optionslist_test

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/optionslist"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit/contracttest"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestContract_OptionsListPanel(t *testing.T) {
	t.Parallel()

	contracttest.Run(t, optionslist.Handler{}, contracttest.Config{
		OmitRequiredLeafPresence: true,
		FullAPIResponse: `{
			"type": "options_list_control",
			"grid": {"x": 2, "y": 0, "w": 10, "h": 5},
			"id": "ol-contract",
			"config": {
				"data_view_id": "dv-fixed",
				"field_name": "host.name",
				"title": "Hosts"
			}
		}`,
		SkipFields: []string{
			// Optional `sort` block has required inner leaves; minimal API fixtures omit it.
			"sort",
			"display_settings",
			"use_global_filters",
			"ignore_validations",
			"single_select",
			"exclude",
			"exists_selected",
			"run_past_timeout",
			"search_technique",
			"selected_options",
		},
	})
}

func TestPinned_OptionsListPinnedRoundtrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ph := optionslist.Handler{}.PinnedHandler()

	in := models.PinnedPanelModel{
		Type: types.StringValue("options_list_control"),
		OptionsListControlConfig: &models.OptionsListControlConfigModel{
			DataViewID: types.StringValue("pinned-dv"),
			FieldName:  types.StringValue("svc"),
		},
	}
	raw, d1 := ph.ToAPI(in)
	require.False(t, d1.HasError(), "%v", d1)

	out, d2 := ph.FromAPI(ctx, nil, raw)
	require.False(t, d2.HasError(), "%v", d2)
	require.True(t, out.Type.Equal(types.StringValue("options_list_control")))
	require.NotNil(t, out.OptionsListControlConfig)
	require.Equal(t, "pinned-dv", out.OptionsListControlConfig.DataViewID.ValueString())
	require.Equal(t, "svc", out.OptionsListControlConfig.FieldName.ValueString())
}
