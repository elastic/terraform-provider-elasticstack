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

func TestContract_OptionsListPanel_ByField(t *testing.T) {
	t.Parallel()

	contracttest.Run(t, optionslist.Handler{}, contracttest.Config{
		// The FullAPIResponse fixture uses Kibana's flat wire format (unaffected by the TF
		// schema restructure), while the TF schema now nests these attributes under `by_field`.
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
			// This fixture discriminates to the Field branch; the ES|QL branch is exercised by
			// TestContract_OptionsListPanel_ByEsql instead.
			"by_esql",
			// The provider deliberately omits values_source on by_field writes for backward
			// compatibility with Kibana < 9.5 (see buildFieldConfig); it is intentionally absent
			// from the read fixture and not a user-facing attribute for by_field.
			"config.values_source",
			// Optional `sort` block has required inner leaves; minimal API fixtures omit it.
			"by_field.sort",
			"by_field.display_settings",
			"by_field.use_global_filters",
			"by_field.ignore_validations",
			"by_field.single_select",
			"by_field.exclude",
			"by_field.exists_selected",
			"by_field.run_past_timeout",
			"by_field.search_technique",
			"by_field.selected_options",
		},
	})
}

func TestContract_OptionsListPanel_ByEsql(t *testing.T) {
	t.Parallel()

	contracttest.Run(t, optionslist.Handler{}, contracttest.Config{
		OmitRequiredLeafPresence: true,
		FullAPIResponse: `{
			"type": "options_list_control",
			"grid": {"x": 2, "y": 0, "w": 10, "h": 5},
			"id": "ol-contract-esql",
			"config": {
				"esql_query": "FROM logs | STATS BY host.name",
				"values_source": "esql",
				"title": "Hosts"
			}
		}`,
		SkipFields: []string{
			// This fixture discriminates to the ES|QL branch; the Field branch is exercised by
			// TestContract_OptionsListPanel_ByField instead.
			"by_field",
			"by_esql.sort",
			"by_esql.display_settings",
			"by_esql.use_global_filters",
			"by_esql.ignore_validations",
			"by_esql.single_select",
			"by_esql.exclude",
			"by_esql.exists_selected",
			"by_esql.run_past_timeout",
			"by_esql.search_technique",
			"by_esql.selected_options",
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
			ByField: &models.OptionsListControlByFieldModel{
				DataViewID: types.StringValue("pinned-dv"),
				FieldName:  types.StringValue("svc"),
			},
		},
	}
	raw, d1 := ph.ToAPI(in)
	require.False(t, d1.HasError(), "%v", d1)

	out, d2 := ph.FromAPI(ctx, nil, raw)
	require.False(t, d2.HasError(), "%v", d2)
	require.True(t, out.Type.Equal(types.StringValue("options_list_control")))
	require.NotNil(t, out.OptionsListControlConfig)
	require.NotNil(t, out.OptionsListControlConfig.ByField)
	require.Equal(t, "pinned-dv", out.OptionsListControlConfig.ByField.DataViewID.ValueString())
	require.Equal(t, "svc", out.OptionsListControlConfig.ByField.FieldName.ValueString())
}
