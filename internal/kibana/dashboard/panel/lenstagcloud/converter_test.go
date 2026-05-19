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

package lenstagcloud

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
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

func TestConverter_VizType(t *testing.T) {
	var c converter
	require.Equal(t, string(kbapi.TagcloudNoESQLTypeTagCloud), c.VizType())
}

func TestConverter_HandlesBlocks(t *testing.T) {
	var c converter
	require.False(t, c.HandlesBlocks(nil))
	require.False(t, c.HandlesBlocks(&models.LensByValueChartBlocks{}))
	require.True(t, c.HandlesBlocks(&models.LensByValueChartBlocks{
		TagcloudConfig: &models.TagcloudConfigModel{},
	}))
}

func TestConverter_roundTrip_NoESQL(t *testing.T) {
	ctx := t.Context()
	var c converter
	resolver := stubResolver{}

	cfg := &models.TagcloudConfigModel{
		Title:               types.StringValue("TC RT"),
		Description:         types.StringValue("desc"),
		IgnoreGlobalFilters: types.BoolValue(true),
		Sampling:            types.Float64Value(0.9),
		DataSourceJSON:      jsontypes.NewNormalizedValue(`{"type":"data_view_spec","index_pattern":"logs-*"}`),
		Query: &models.FilterSimpleModel{
			Language:   types.StringValue("kql"),
			Expression: types.StringValue(`status:200`),
		},
		MetricJSON: customtypes.NewJSONWithDefaultsValue[map[string]any](
			`{"operation":{"operation_type":"count"}}`,
			lenscommon.PopulateTagcloudMetricDefaults,
		),
		TagByJSON: customtypes.NewJSONWithDefaultsValue[map[string]any](
			`{"operation":{"operation_type":"terms"},"field":"user.keyword","size":20}`,
			lenscommon.PopulateTagcloudTagByDefaults,
		),
	}

	blocks := &models.LensByValueChartBlocks{TagcloudConfig: cfg}
	attrs, diags := c.BuildAttributes(blocks, resolver)
	require.False(t, diags.HasError(), "%v", diags)

	out := &models.LensByValueChartBlocks{TagcloudConfig: &models.TagcloudConfigModel{}}
	diags = c.PopulateFromAttributes(ctx, resolver, out, attrs)
	require.False(t, diags.HasError(), "%v", diags)

	require.Equal(t, cfg.Title.ValueString(), out.TagcloudConfig.Title.ValueString())
	require.Equal(t, cfg.Query.Expression.ValueString(), out.TagcloudConfig.Query.Expression.ValueString())
	require.False(t, out.TagcloudConfig.MetricJSON.IsNull())
	require.False(t, out.TagcloudConfig.TagByJSON.IsNull())
}
