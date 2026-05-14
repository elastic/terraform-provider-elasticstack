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

package lensxy

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

func minimalXYNoESQLChartForRoundTrip() *models.XYChartConfigModel {
	return &models.XYChartConfigModel{
		Title: types.StringValue("XY RT"),
		Axis: &models.XYAxisModel{
			X: &models.XYAxisConfigModel{},
			Y: &models.YAxisConfigModel{},
		},
		Decorations: &models.XYDecorationsModel{},
		Fitting:     &models.XYFittingModel{Type: types.StringValue("none")},
		Layers: []models.XYLayerModel{
			{
				Type: types.StringValue("area"),
				DataLayer: &models.DataLayerModel{
					DataSourceJSON: jsontypes.NewNormalizedValue(`{"type":"dataView","id":"logs-*"}`),
					Y: []models.YMetricModel{
						{ConfigJSON: jsontypes.NewNormalizedValue(`{"operation":"count","axis":"left"}`)},
					},
				},
			},
		},
		Legend: &models.XYLegendModel{
			Inside:     types.BoolValue(false),
			Visibility: types.StringValue("visible"),
		},
		Query: &models.FilterSimpleModel{
			Expression: types.StringValue("*"),
			Language:   types.StringValue("kql"),
		},
	}
}

func TestConverter_VizType(t *testing.T) {
	var c converter
	require.Equal(t, string(kbapi.XyChartNoESQLTypeXy), c.VizType())
}

func TestConverter_HandlesBlocks(t *testing.T) {
	var c converter
	require.False(t, c.HandlesBlocks(nil))
	require.False(t, c.HandlesBlocks(&models.LensByValueChartBlocks{}))
	require.True(t, c.HandlesBlocks(&models.LensByValueChartBlocks{
		XYChartConfig: &models.XYChartConfigModel{},
	}))
}

func TestConverter_roundTrip_NoESQL(t *testing.T) {
	ctx := t.Context()
	var c converter
	resolver := stubResolver{}

	cfg := minimalXYNoESQLChartForRoundTrip()
	require.NotNil(t, cfg.Query)
	wantTitle := cfg.Title.ValueString()
	wantExpr := cfg.Query.Expression.ValueString()
	in := &models.LensByValueChartBlocks{XYChartConfig: cfg}

	attrs, diags := c.BuildAttributes(in, resolver)
	require.False(t, diags.HasError(), "%v", diags)

	out := &models.LensByValueChartBlocks{}
	diags = c.PopulateFromAttributes(ctx, resolver, out, attrs)
	require.False(t, diags.HasError(), "%v", diags)
	require.NotNil(t, out.XYChartConfig)

	require.Equal(t, wantTitle, out.XYChartConfig.Title.ValueString())
	require.Len(t, out.XYChartConfig.Layers, 1)
	require.NotNil(t, out.XYChartConfig.Layers[0].DataLayer)
	require.Len(t, out.XYChartConfig.Layers[0].DataLayer.Y, 1)
	require.NotNil(t, out.XYChartConfig.Query)
	require.Equal(t, wantExpr, out.XYChartConfig.Query.Expression.ValueString())
}
