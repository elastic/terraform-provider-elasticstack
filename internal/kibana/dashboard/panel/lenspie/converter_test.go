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

package lenspie

import (
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubResolver struct{}

func (stubResolver) ResolveChartTimeRange(chartLevel *models.TimeRangeModel) kbapi.KbnEsQueryServerTimeRangeSchema {
	_ = chartLevel
	return kbapi.KbnEsQueryServerTimeRangeSchema{}
}

func (stubResolver) DashboardLensComparableTimeRange() (kbapi.KbnEsQueryServerTimeRangeSchema, bool) {
	return kbapi.KbnEsQueryServerTimeRangeSchema{}, false
}

func TestConverter_VizType(t *testing.T) {
	var c converter
	require.Equal(t, string(kbapi.PieNoESQLTypePie), c.VizType())
}

func TestConverter_HandlesBlocks(t *testing.T) {
	var c converter
	require.False(t, c.HandlesBlocks(nil))
	require.False(t, c.HandlesBlocks(&models.LensByValueChartBlocks{}))
	require.True(t, c.HandlesBlocks(&models.LensByValueChartBlocks{
		PieChartConfig: &models.PieChartConfigModel{},
	}))
}

func TestConverter_roundTrip_NoESQL(t *testing.T) {
	ctx := t.Context()
	var c converter
	resolver := stubResolver{}

	nested := true
	truncate := int64(3)
	cfg := &models.PieChartConfigModel{
		Title:               types.StringValue("Pie RT"),
		Description:         types.StringValue("d"),
		IgnoreGlobalFilters: types.BoolValue(true),
		Sampling:            types.Float64Value(0.75),
		DonutHole:           types.StringValue(string(kbapi.PieStylingDonutHoleS)),
		LabelPosition:       types.StringValue(string(kbapi.PieStylingLabelsPositionInside)),
		DataSourceJSON:      jsontypes.NewNormalizedValue(`{"type":"data_view_spec","index_pattern":"logs-*"}`),
		Query: &models.FilterSimpleModel{
			Language:   types.StringValue("kql"),
			Expression: types.StringValue("response:200"),
		},
		Legend: &models.PartitionLegendModel{
			Size:              types.StringValue("auto"),
			Nested:            types.BoolValue(nested),
			TruncateAfterLine: types.Int64Value(truncate),
			Visible:           types.StringValue(string(kbapi.PieLegendVisibilityVisible)),
		},
		Metrics: []models.PieMetricModel{
			{
				Config: customtypes.NewJSONWithDefaultsValue(`{"operation":"count"}`, lenscommon.PopulatePieChartMetricDefaults),
			},
		},
	}

	in := &models.LensByValueChartBlocks{PieChartConfig: cfg}
	attrs, diags := c.BuildAttributes(in, resolver)
	require.False(t, diags.HasError())

	out := &models.LensByValueChartBlocks{}
	diags = c.PopulateFromAttributes(ctx, resolver, out, attrs)
	require.False(t, diags.HasError())
	require.NotNil(t, out.PieChartConfig)

	assert.Equal(t, cfg.Title, out.PieChartConfig.Title)
	assert.Equal(t, cfg.Description, out.PieChartConfig.Description)
	assert.Equal(t, cfg.Query.Expression, out.PieChartConfig.Query.Expression)
	eq, d := cfg.DataSourceJSON.StringSemanticEquals(ctx, out.PieChartConfig.DataSourceJSON)
	require.False(t, d.HasError())
	assert.True(t, eq)
	require.NotNil(t, out.PieChartConfig.Legend)
	assert.Equal(t, cfg.Legend.Size, out.PieChartConfig.Legend.Size)
	assert.Equal(t, cfg.Legend.Visible, out.PieChartConfig.Legend.Visible)
	mEq, md := cfg.Metrics[0].Config.StringSemanticEquals(ctx, out.PieChartConfig.Metrics[0].Config)
	require.False(t, md.HasError())
	assert.True(t, mEq)
}

func TestConverter_roundTrip_ESQL(t *testing.T) {
	ctx := t.Context()
	var c converter
	resolver := stubResolver{}

	apiJSON := `{
		"type": "pie",
		"title": "ESQL Pie Chart",
		"description": "ESQL pie description",
		"data_source": {"type":"esql","query":"FROM logs-* | LIMIT 10"},
		"sampling": 0.5,
		"ignore_global_filters": true,
		"legend": {"size":"auto","visibility":"visible"},
		"metrics": [{"operation":"value","column":"bytes","color":{"type":"static","color":"#54B399"},"format":{"type":"number"}}],
		"group_by": [{"operation":"value","column":"host.name","collapse_by":"avg","color":{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}}}]
	}`
	var apiESQL kbapi.PieESQL
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &apiESQL))

	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, attrs.FromPieESQL(apiESQL))

	out := &models.LensByValueChartBlocks{}
	diags := c.PopulateFromAttributes(ctx, resolver, out, attrs)
	require.False(t, diags.HasError())
	require.NotNil(t, out.PieChartConfig)

	attrs2, diags := c.BuildAttributes(out, resolver)
	require.False(t, diags.HasError())

	p2, err := attrs2.AsPieESQL()
	require.NoError(t, err)
	assert.Equal(t, "ESQL Pie Chart", *p2.Title)
	assert.Len(t, p2.Metrics, 1)
	require.NotNil(t, p2.GroupBy)
	assert.Len(t, *p2.GroupBy, 1)
}
