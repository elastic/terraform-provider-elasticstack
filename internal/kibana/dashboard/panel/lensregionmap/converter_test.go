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

package lensregionmap

import (
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
	require.Equal(t, string(kbapi.RegionMapNoESQLTypeRegionMap), c.VizType())
}

func TestConverter_HandlesBlocks(t *testing.T) {
	var c converter
	require.False(t, c.HandlesBlocks(nil))
	require.False(t, c.HandlesBlocks(&models.LensByValueChartBlocks{}))
	require.True(t, c.HandlesBlocks(&models.LensByValueChartBlocks{
		RegionMapConfig: &models.RegionMapConfigModel{},
	}))
}

func TestConverter_roundTrip_NoESQL(t *testing.T) {
	ctx := t.Context()
	var c converter
	resolver := stubResolver{}

	cfg := &models.RegionMapConfigModel{
		Title:               types.StringValue("RM"),
		Description:         types.StringValue("d"),
		IgnoreGlobalFilters: types.BoolValue(true),
		Sampling:            types.Float64Value(0.5),
		DataSourceJSON:      jsontypes.NewNormalizedValue(`{"type":"data_view_spec","index_pattern":"metrics-*"}`),
		Query: &models.FilterSimpleModel{
			Language:   types.StringValue("kql"),
			Expression: types.StringValue("*"),
		},
		MetricJSON: customtypes.NewJSONWithDefaultsValue(`{"operation":"count"}`, lenscommon.PopulateRegionMapMetricDefaults),
		RegionJSON: jsontypes.NewNormalizedValue(`{"operation":"terms","field":"country"}`),
	}

	in := &models.LensByValueChartBlocks{RegionMapConfig: cfg}
	attrs, diags := c.BuildAttributes(in, resolver)
	require.False(t, diags.HasError())

	out := &models.LensByValueChartBlocks{}
	diags = c.PopulateFromAttributes(ctx, resolver, out, attrs)
	require.False(t, diags.HasError())
	require.NotNil(t, out.RegionMapConfig)

	assert.Equal(t, cfg.Title, out.RegionMapConfig.Title)
	assert.Equal(t, cfg.Description, out.RegionMapConfig.Description)
	eq, d := cfg.DataSourceJSON.StringSemanticEquals(ctx, out.RegionMapConfig.DataSourceJSON)
	require.False(t, d.HasError())
	assert.True(t, eq)
	assert.Equal(t, cfg.Query.Expression, out.RegionMapConfig.Query.Expression)
	mEq, md := cfg.MetricJSON.StringSemanticEquals(ctx, out.RegionMapConfig.MetricJSON)
	require.False(t, md.HasError())
	assert.True(t, mEq)
	rEq, rd := cfg.RegionJSON.StringSemanticEquals(ctx, out.RegionMapConfig.RegionJSON)
	require.False(t, rd.HasError())
	assert.True(t, rEq)
}
