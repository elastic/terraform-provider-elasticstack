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

package rangeslider_test

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/rangeslider"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit/contracttest"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestContract_RangeSliderPanel(t *testing.T) {
	t.Parallel()

	contracttest.Run(t, rangeslider.Handler{}, contracttest.Config{
		// The API's config JSON layout is the same flat Field-branch shape regardless of the
		// restructure (only the Terraform side nests it under `by_field`); required-leaf-presence
		// navigation can't resolve the `by_field` nesting against the flat fixture, so it's omitted
		// here (same rationale as the slo_overview contract test's single/groups branches).
		FullAPIResponse: `{
			"type": "range_slider_control",
			"grid": {"x": 0, "y": 1, "w": 8, "h": 4},
			"id": "rs-contract",
			"config": {
				"data_view_id": "dv-rs",
				"field_name": "bytes",
				"title": "Range",
				"value": ["10", "1000"],
				"step": 5
			}
		}`,
		OmitRequiredLeafPresence: true,
		SkipFields: []string{
			// Optional server flags and value list: API omissions do not reset practitioner-known state.
			"by_field.use_global_filters",
			"by_field.ignore_validations",
			"by_field.value",
			// Fixture only covers the by_field branch; by_esql has no baseline to compare against.
			"by_esql",
			// The provider deliberately omits values_source on by_field writes for backward
			// compatibility with Kibana < 9.5 (see buildFieldConfig); the fixture predates this
			// and includes it.
			"config.values_source",
		},
	})
}

func TestPinned_RangeSliderPinnedRoundtrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ph := rangeslider.Handler{}.PinnedHandler()

	in := models.PinnedPanelModel{
		Type: types.StringValue("range_slider_control"),
		RangeSliderControlConfig: &models.RangeSliderControlConfigModel{
			ByField: &models.RangeSliderControlByFieldModel{
				DataViewID: types.StringValue("dv"),
				FieldName:  types.StringValue("source.bytes"),
				Value: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("100"),
					types.StringValue("500"),
				}),
				Step: types.Float32Value(12),
			},
		},
	}

	raw, d1 := ph.ToAPI(in)
	require.False(t, d1.HasError(), "%v", d1)

	out, d2 := ph.FromAPI(ctx, nil, raw)
	require.False(t, d2.HasError(), "%v", d2)
	require.True(t, out.Type.Equal(types.StringValue("range_slider_control")))
	require.NotNil(t, out.RangeSliderControlConfig)
	require.NotNil(t, out.RangeSliderControlConfig.ByField)
	require.Equal(t, "dv", out.RangeSliderControlConfig.ByField.DataViewID.ValueString())
	require.Equal(t, "source.bytes", out.RangeSliderControlConfig.ByField.FieldName.ValueString())
}

func TestPinned_RangeSliderPinnedRoundtrip_byEsql(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ph := rangeslider.Handler{}.PinnedHandler()

	in := models.PinnedPanelModel{
		Type: types.StringValue("range_slider_control"),
		RangeSliderControlConfig: &models.RangeSliderControlConfigModel{
			ByEsql: &models.RangeSliderControlByEsqlModel{
				EsqlQuery:    types.StringValue("FROM orders | STATS min = MIN(price), max = MAX(price)"),
				ValuesSource: types.StringValue("esql_query"),
			},
		},
	}

	raw, d1 := ph.ToAPI(in)
	require.False(t, d1.HasError(), "%v", d1)

	out, d2 := ph.FromAPI(ctx, nil, raw)
	require.False(t, d2.HasError(), "%v", d2)
	require.True(t, out.Type.Equal(types.StringValue("range_slider_control")))
	require.NotNil(t, out.RangeSliderControlConfig)
	require.NotNil(t, out.RangeSliderControlConfig.ByEsql)
	require.Equal(t, "FROM orders | STATS min = MIN(price), max = MAX(price)", out.RangeSliderControlConfig.ByEsql.EsqlQuery.ValueString())
	require.Equal(t, "esql_query", out.RangeSliderControlConfig.ByEsql.ValuesSource.ValueString())
}
