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

package discoversession_test

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/discoversession"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/iface"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func panelModelBase() models.PanelModel {
	return models.PanelModel{
		Grid: models.PanelGridModel{
			X: types.Int64Value(0),
			Y: types.Int64Value(0),
			W: types.Int64Value(24),
			H: types.Int64Value(15),
		},
		ID: types.StringValue("panel-id"),
	}
}

func TestHandler_roundTrip_byValue_dsl(t *testing.T) {
	ctx := iface.WithEnclosingDashboard(context.Background(), &models.DashboardModel{})
	rawFilter := `{"condition":{"field":"host.name","operator":"is","value":"web"},"type":"condition"}`
	pm := panelModelBase()
	pm.DiscoverSessionConfig = &models.DiscoverSessionPanelConfigModel{
		ByValue: &models.DiscoverSessionPanelByValueModel{
			TimeRange: &models.TimeRangeModel{
				From: types.StringValue("now-30m"),
				To:   types.StringValue("now"),
				Mode: types.StringNull(),
			},
			Tab: models.DiscoverSessionTabModel{
				DSL: &models.DiscoverSessionDSLTabModel{
					ColumnOrder: types.ListValueMust(types.StringType, []attr.Value{
						types.StringValue("@timestamp"),
						types.StringValue("message"),
					}),
					Query: &models.FilterSimpleModel{
						Language:   types.StringValue("kql"),
						Expression: types.StringValue(`host.name : "web-01"`),
					},
					DataSourceJSON: jsontypes.NewNormalizedValue(`{"id":"logs-*","type":"data_view_reference"}`),
					Filters: []models.ChartFilterJSONModel{
						{FilterJSON: jsontypes.NewNormalizedValue(rawFilter)},
					},
				},
			},
		},
	}

	item, diags := discoversession.Handler{}.ToAPI(pm, nil)
	require.False(t, diags.HasError(), "%s", diags)

	next := pm
	d2 := discoversession.Handler{}.FromAPI(ctx, &next, &pm, item)
	require.False(t, d2.HasError(), "%s", d2)

	require.Nil(t, next.DiscoverSessionConfig.ByValue.Tab.ESQL)
	require.NotNil(t, next.DiscoverSessionConfig.ByValue.Tab.DSL)
	dsl := next.DiscoverSessionConfig.ByValue.Tab.DSL
	ctxSE := context.Background()
	if assert.Len(t, dsl.Filters, 1) {
		eq, d := dsl.Filters[0].FilterJSON.StringSemanticEquals(ctxSE, jsontypes.NewNormalizedValue(rawFilter))
		require.False(t, d.HasError())
		assert.True(t, eq)
	}
	co := dsl.ColumnOrder.Elements()
	require.Len(t, co, 2)
	assert.Equal(t, "@timestamp", co[0].(types.String).ValueString())
	assert.Equal(t, "message", co[1].(types.String).ValueString())
	assert.Equal(t, `host.name : "web-01"`, dsl.Query.Expression.ValueString())
	dsJSONEq, d := dsl.DataSourceJSON.StringSemanticEquals(ctxSE, jsontypes.NewNormalizedValue(`{"type":"data_view_reference","id":"logs-*"}`))
	require.False(t, d.HasError())
	assert.True(t, dsJSONEq)
}

func TestHandler_roundTrip_byValue_esql(t *testing.T) {
	ctx := iface.WithEnclosingDashboard(context.Background(), &models.DashboardModel{})
	pm := panelModelBase()
	pm.DiscoverSessionConfig = &models.DiscoverSessionPanelConfigModel{
		ByValue: &models.DiscoverSessionPanelByValueModel{
			TimeRange: &models.TimeRangeModel{
				From: types.StringValue("now-30m"),
				To:   types.StringValue("now"),
				Mode: types.StringNull(),
			},
			Tab: models.DiscoverSessionTabModel{
				ESQL: &models.DiscoverSessionESQLTabModel{
					DataSourceJSON: jsontypes.NewNormalizedValue(`{"query":"FROM logs-*","type":"esql"}`),
				},
			},
		},
	}

	item, diags := discoversession.Handler{}.ToAPI(pm, nil)
	require.False(t, diags.HasError(), "%s", diags)

	next := pm
	d2 := discoversession.Handler{}.FromAPI(ctx, &next, &pm, item)
	require.False(t, d2.HasError(), "%s", d2)

	require.Nil(t, next.DiscoverSessionConfig.ByValue.Tab.DSL)
	require.NotNil(t, next.DiscoverSessionConfig.ByValue.Tab.ESQL)
	esql := next.DiscoverSessionConfig.ByValue.Tab.ESQL
	dsJSONEq, d := esql.DataSourceJSON.StringSemanticEquals(ctx, jsontypes.NewNormalizedValue(`{"type":"esql","query":"FROM logs-*"}`))
	require.False(t, d.HasError())
	assert.True(t, dsJSONEq)
}

func TestHandler_roundTrip_byReference(t *testing.T) {
	ctx := iface.WithEnclosingDashboard(context.Background(), &models.DashboardModel{})
	pm := panelModelBase()
	pm.DiscoverSessionConfig = &models.DiscoverSessionPanelConfigModel{
		Title:       types.StringValue("Discover link"),
		Description: types.StringValue("linked panel"),
		ByReference: &models.DiscoverSessionPanelByRefModel{
			TimeRange: &models.TimeRangeModel{
				From: types.StringValue("now-1h"),
				To:   types.StringValue("now"),
				Mode: types.StringNull(),
			},
			RefID:         types.StringValue("saved-discover-abc"),
			SelectedTabID: types.StringValue("tab-explicit"),
			Overrides: &models.DiscoverSessionOverridesModel{
				Density:     types.StringValue("compact"),
				RowsPerPage: types.Int64Value(50),
				SampleSize:  types.Int64Value(500),
				Sort: []models.DiscoverSessionSortModel{
					{Name: types.StringValue("@timestamp"), Direction: types.StringValue("desc")},
				},
			},
		},
	}

	item1, diags := discoversession.Handler{}.ToAPI(pm, nil)
	require.False(t, diags.HasError(), "%s", diags)

	next := pm
	d2 := discoversession.Handler{}.FromAPI(ctx, &next, &pm, item1)
	require.False(t, d2.HasError(), "%s", d2)

	require.Nil(t, next.DiscoverSessionConfig.ByValue)
	br := next.DiscoverSessionConfig.ByReference
	require.NotNil(t, br)
	assert.Equal(t, "saved-discover-abc", br.RefID.ValueString())
	assert.Equal(t, "tab-explicit", br.SelectedTabID.ValueString())
	require.NotNil(t, br.Overrides)
	assert.Equal(t, "compact", br.Overrides.Density.ValueString())

	item2, diags2 := discoversession.Handler{}.ToAPI(next, nil)
	require.False(t, diags2.HasError(), "%s", diags2)

	raw1, err := item1.MarshalJSON()
	require.NoError(t, err)
	raw2, err := item2.MarshalJSON()
	require.NoError(t, err)
	require.JSONEq(t, string(raw1), string(raw2))
}

func TestHandler_optionalEnvelopeFieldsStayNull_afterRoundTrip(t *testing.T) {
	ctx := iface.WithEnclosingDashboard(context.Background(), &models.DashboardModel{})
	pm := panelModelBase()
	pm.DiscoverSessionConfig = &models.DiscoverSessionPanelConfigModel{
		Title:       types.StringValue("titled"),
		Description: types.StringNull(),
		HideTitle:   types.BoolNull(),
		HideBorder:  types.BoolNull(),
		ByValue: &models.DiscoverSessionPanelByValueModel{
			TimeRange: &models.TimeRangeModel{
				From: types.StringValue("now-30m"),
				To:   types.StringValue("now"),
				Mode: types.StringNull(),
			},
			Tab: models.DiscoverSessionTabModel{
				DSL: &models.DiscoverSessionDSLTabModel{
					Query: &models.FilterSimpleModel{
						Language:   types.StringValue("kql"),
						Expression: types.StringValue("*"),
					},
					DataSourceJSON: jsontypes.NewNormalizedValue(`{"type":"data_view_reference","id":"logs-*"}`),
				},
			},
		},
	}

	item, diags := discoversession.Handler{}.ToAPI(pm, nil)
	require.False(t, diags.HasError(), "%s", diags)

	next := pm
	d2 := discoversession.Handler{}.FromAPI(ctx, &next, &pm, item)
	require.False(t, d2.HasError(), "%s", d2)

	cfg := next.DiscoverSessionConfig
	require.NotNil(t, cfg)
	assert.Equal(t, "titled", cfg.Title.ValueString())
	assert.True(t, cfg.Description.IsNull())
	assert.True(t, cfg.HideTitle.IsNull())
	assert.True(t, cfg.HideBorder.IsNull())
}
