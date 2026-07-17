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

package lensgauge

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

func TestConverter_VizType(t *testing.T) {
	var c converter
	require.Equal(t, string(kbapi.KibanaHTTPAPIsGaugeNoESQLByValuePanelTypeGauge), c.VizType())
}

func TestConverter_HandlesBlocks(t *testing.T) {
	var c converter
	require.False(t, c.HandlesBlocks(nil))
	require.False(t, c.HandlesBlocks(&models.LensByValueChartBlocks{}))
	require.True(t, c.HandlesBlocks(&models.LensByValueChartBlocks{
		GaugeConfig: &models.GaugeConfigModel{},
	}))
}

func TestConverter_roundTrip_NoESQL(t *testing.T) {
	ctx := t.Context()
	var c converter
	cfg := &models.GaugeConfigModel{
		LensChartBaseTFModel: models.LensChartBaseTFModel{
			Title:          types.StringValue("Gauge RT"),
			DataSourceJSON: jsontypes.NewNormalizedValue(`{"type":"data_view_spec","index_pattern":"metrics-*"}`),
		},
		Query: &models.FilterSimpleModel{
			Language:   types.StringValue("kql"),
			Expression: types.StringValue(`*`),
		},
		MetricJSON: customtypes.NewJSONWithDefaultsValue[map[string]any](
			`{"operation":"count"}`,
			lenscommon.PopulateGaugeMetricDefaults,
		),
		Styling: &models.GaugeStylingModel{
			ShapeJSON: jsontypes.NewNormalizedValue(`{"type":"circle"}`),
		},
	}

	blocks := &models.LensByValueChartBlocks{GaugeConfig: cfg}
	attrs, diags := c.BuildAttributes(blocks)
	require.False(t, diags.HasError(), "%v", diags)

	out := &models.LensByValueChartBlocks{GaugeConfig: &models.GaugeConfigModel{}}
	diags = c.PopulateFromAttributes(ctx, out, attrs)
	require.False(t, diags.HasError(), "%v", diags)

	require.Equal(t, cfg.Title.ValueString(), out.GaugeConfig.Title.ValueString())
	require.False(t, out.GaugeConfig.MetricJSON.IsNull())
	require.NotNil(t, out.GaugeConfig.Styling)
}
