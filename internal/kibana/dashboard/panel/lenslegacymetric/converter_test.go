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

package lenslegacymetric

import (
	"context"
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
	require.Equal(t, string(kbapi.LegacyMetric), c.VizType())
}

func TestConverter_HandlesBlocks(t *testing.T) {
	var c converter
	require.False(t, c.HandlesBlocks(nil))
	require.False(t, c.HandlesBlocks(&models.LensByValueChartBlocks{}))
	require.True(t, c.HandlesBlocks(&models.LensByValueChartBlocks{
		LegacyMetricConfig: &models.LegacyMetricConfigModel{},
	}))
}

func assertLegacyMetricRoundTripEqual(ctx context.Context, t *testing.T, a, b *models.LegacyMetricConfigModel) {
	t.Helper()
	require.NotNil(t, a)
	require.NotNil(t, b)
	assert.Equal(t, a.Title, b.Title)
	assert.Equal(t, a.Description, b.Description)
	assert.Equal(t, a.IgnoreGlobalFilters, b.IgnoreGlobalFilters)
	assert.Equal(t, a.Sampling, b.Sampling)
	if a.DataSourceJSON.IsNull() != b.DataSourceJSON.IsNull() || a.DataSourceJSON.IsUnknown() != b.DataSourceJSON.IsUnknown() {
		assert.Fail(t, "dataset null/unknown state mismatch")
		return
	}
	if !a.DataSourceJSON.IsNull() && !a.DataSourceJSON.IsUnknown() {
		eq, d := a.DataSourceJSON.StringSemanticEquals(ctx, b.DataSourceJSON)
		require.False(t, d.HasError())
		assert.True(t, eq, "dataset should be semantically equal")
	}
	if (a.Query == nil) != (b.Query == nil) {
		assert.Fail(t, "query nil mismatch")
		return
	}
	if a.Query != nil {
		assert.Equal(t, a.Query.Language, b.Query.Language)
		assert.Equal(t, a.Query.Expression, b.Query.Expression)
	}
	require.Len(t, b.Filters, len(a.Filters))
	for i := range a.Filters {
		eq, d := a.Filters[i].FilterJSON.StringSemanticEquals(ctx, b.Filters[i].FilterJSON)
		require.False(t, d.HasError())
		assert.True(t, eq)
	}
	if a.MetricJSON.IsNull() != b.MetricJSON.IsNull() || a.MetricJSON.IsUnknown() != b.MetricJSON.IsUnknown() {
		assert.Fail(t, "metric null/unknown state mismatch")
		return
	}
	if !a.MetricJSON.IsNull() && !a.MetricJSON.IsUnknown() {
		eq, d := a.MetricJSON.StringSemanticEquals(ctx, b.MetricJSON)
		require.False(t, d.HasError())
		assert.True(t, eq)
	}
}

func TestConverter_BuildAttributes_PopulateFromAttributes_roundTrip_NoESQL(t *testing.T) {
	ctx := t.Context()
	var c converter
	resolver := stubResolver{}

	cfg := &models.LegacyMetricConfigModel{
		Title:               types.StringValue("Hello Legacy"),
		Description:         types.StringValue("desc"),
		IgnoreGlobalFilters: types.BoolValue(true),
		Sampling:            types.Float64Value(0.25),
		DataSourceJSON:      jsontypes.NewNormalizedValue(`{"type":"data_view_spec","index_pattern":"metrics-*"}`),
		Query: &models.FilterSimpleModel{
			Language:   types.StringValue("kql"),
			Expression: types.StringValue("*"),
		},
		MetricJSON: customtypes.NewJSONWithDefaultsValue(`{"operation":"count","format":{"type":"number"}}`, lenscommon.PopulateLegacyMetricMetricDefaults),
	}

	in := &models.LensByValueChartBlocks{LegacyMetricConfig: cfg}

	attrs, diags := c.BuildAttributes(in, resolver)
	require.False(t, diags.HasError())

	out := &models.LensByValueChartBlocks{}
	diags = c.PopulateFromAttributes(ctx, resolver, out, attrs)
	require.False(t, diags.HasError())
	require.NotNil(t, out.LegacyMetricConfig)

	assertLegacyMetricRoundTripEqual(ctx, t, cfg, out.LegacyMetricConfig)
}
