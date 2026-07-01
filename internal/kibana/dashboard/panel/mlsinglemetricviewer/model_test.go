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

package mlsinglemetricviewer_test

import (
	"context"
	"math"
	"math/big"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/mlsinglemetricviewer"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func entityObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"string_value":  types.StringType,
			"numeric_value": types.NumberType,
		},
	}
}

func makeSelectedEntitiesMap(t *testing.T, elems map[string]models.MlSingleMetricViewerEntityModel) types.Map {
	t.Helper()
	ctx := context.Background()
	var diags diag.Diagnostics
	m := typeutils.MapValueFrom(ctx, elems, entityObjectType(), path.Empty(), &diags)
	require.False(t, diags.HasError(), "%v", diags)
	return m
}

func makeStringEntityProp(s string) kbapi.KibanaHTTPAPIsMlSingleMetricViewer_SelectedEntities_AdditionalProperties {
	var prop kbapi.KibanaHTTPAPIsMlSingleMetricViewer_SelectedEntities_AdditionalProperties
	if err := prop.FromKibanaHTTPAPIsMlSingleMetricViewerSelectedEntities0(s); err != nil {
		panic(err)
	}
	return prop
}

func makeNumericEntityProp(n float32) kbapi.KibanaHTTPAPIsMlSingleMetricViewer_SelectedEntities_AdditionalProperties {
	var prop kbapi.KibanaHTTPAPIsMlSingleMetricViewer_SelectedEntities_AdditionalProperties
	if err := prop.FromKibanaHTTPAPIsMlSingleMetricViewerSelectedEntities1(n); err != nil {
		panic(err)
	}
	return prop
}

func makeAPIConfig(opts ...func(*kbapi.KibanaHTTPAPIsMlSingleMetricViewer)) kbapi.KibanaHTTPAPIsMlSingleMetricViewer {
	cfg := kbapi.KibanaHTTPAPIsMlSingleMetricViewer{
		JobIds: []string{"job-a"},
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

func withSelectedEntities(m map[string]kbapi.KibanaHTTPAPIsMlSingleMetricViewer_SelectedEntities_AdditionalProperties) func(*kbapi.KibanaHTTPAPIsMlSingleMetricViewer) {
	return func(c *kbapi.KibanaHTTPAPIsMlSingleMetricViewer) {
		c.SelectedEntities = &m
	}
}

func withSelectedDetectorIndex(v float32) func(*kbapi.KibanaHTTPAPIsMlSingleMetricViewer) {
	return func(c *kbapi.KibanaHTTPAPIsMlSingleMetricViewer) { c.SelectedDetectorIndex = &v }
}

func withForecastID(id string) func(*kbapi.KibanaHTTPAPIsMlSingleMetricViewer) {
	return func(c *kbapi.KibanaHTTPAPIsMlSingleMetricViewer) { c.ForecastId = &id }
}

func withFunctionDescription(fn string) func(*kbapi.KibanaHTTPAPIsMlSingleMetricViewer) {
	return func(c *kbapi.KibanaHTTPAPIsMlSingleMetricViewer) { c.FunctionDescription = &fn }
}

func withTitle(title string) func(*kbapi.KibanaHTTPAPIsMlSingleMetricViewer) {
	return func(c *kbapi.KibanaHTTPAPIsMlSingleMetricViewer) { c.Title = &title }
}

func withDescription(desc string) func(*kbapi.KibanaHTTPAPIsMlSingleMetricViewer) {
	return func(c *kbapi.KibanaHTTPAPIsMlSingleMetricViewer) { c.Description = &desc }
}

func withHideTitle(v bool) func(*kbapi.KibanaHTTPAPIsMlSingleMetricViewer) {
	return func(c *kbapi.KibanaHTTPAPIsMlSingleMetricViewer) { c.HideTitle = &v }
}

func withHideBorder(v bool) func(*kbapi.KibanaHTTPAPIsMlSingleMetricViewer) {
	return func(c *kbapi.KibanaHTTPAPIsMlSingleMetricViewer) { c.HideBorder = &v }
}

func withTimeRange(from, to string, mode *kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchemaMode) func(*kbapi.KibanaHTTPAPIsMlSingleMetricViewer) {
	return func(c *kbapi.KibanaHTTPAPIsMlSingleMetricViewer) {
		c.TimeRange = &kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema{
			From: from,
			To:   to,
			Mode: mode,
		}
	}
}

func TestBuildConfig_minimal(t *testing.T) {
	ctx := context.Background()
	pm := models.PanelModel{
		MlSingleMetricViewerConfig: &models.MlSingleMetricViewerConfigModel{
			JobIDs: []types.String{types.StringValue("job-a")},
		},
	}
	var panel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeMlSingleMetricViewer
	diags := mlsinglemetricviewer.BuildConfig(ctx, pm, &panel)
	require.False(t, diags.HasError(), "%v", diags)

	assert.Equal(t, []string{"job-a"}, panel.Config.JobIds)
	assert.Nil(t, panel.Config.SelectedEntities)
	assert.Nil(t, panel.Config.SelectedDetectorIndex)
}

func TestBuildConfig_selectedEntities_stringAndNumeric(t *testing.T) {
	ctx := context.Background()
	pm := models.PanelModel{
		MlSingleMetricViewerConfig: &models.MlSingleMetricViewerConfigModel{
			JobIDs: []types.String{types.StringValue("job-a")},
			SelectedEntities: makeSelectedEntitiesMap(t, map[string]models.MlSingleMetricViewerEntityModel{
				"airline": {
					StringValue:  types.StringValue("AAL"),
					NumericValue: types.NumberNull(),
				},
				"region_code": {
					StringValue:  types.StringNull(),
					NumericValue: types.NumberValue(big.NewFloat(4)),
				},
			}),
		},
	}
	var panel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeMlSingleMetricViewer
	diags := mlsinglemetricviewer.BuildConfig(ctx, pm, &panel)
	require.False(t, diags.HasError(), "%v", diags)

	require.NotNil(t, panel.Config.SelectedEntities)
	entities := *panel.Config.SelectedEntities
	require.Len(t, entities, 2)

	airline, err := entities["airline"].AsKibanaHTTPAPIsMlSingleMetricViewerSelectedEntities0()
	require.NoError(t, err)
	assert.Equal(t, "AAL", airline)

	region, err := entities["region_code"].AsKibanaHTTPAPIsMlSingleMetricViewerSelectedEntities1()
	require.NoError(t, err)
	require.Equal(t, math.Float32bits(float32(4)), math.Float32bits(region))
}

func TestBuildConfig_selectedEntities_numericValueOutOfFloat32Range(t *testing.T) {
	ctx := context.Background()
	pm := models.PanelModel{
		MlSingleMetricViewerConfig: &models.MlSingleMetricViewerConfigModel{
			JobIDs: []types.String{types.StringValue("job-a")},
			SelectedEntities: makeSelectedEntitiesMap(t, map[string]models.MlSingleMetricViewerEntityModel{
				"region_code": {
					StringValue:  types.StringNull(),
					NumericValue: types.NumberValue(big.NewFloat(float64(math.MaxFloat32) * 2)),
				},
			}),
		},
	}
	var panel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeMlSingleMetricViewer
	diags := mlsinglemetricviewer.BuildConfig(ctx, pm, &panel)
	require.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary(), "Invalid ML single metric viewer configuration")
	assert.Contains(t, diags[0].Detail(), "numeric_value is out of float32 range")
}

func TestBuildConfig_withOptionalFields(t *testing.T) {
	ctx := context.Background()
	pm := models.PanelModel{
		MlSingleMetricViewerConfig: &models.MlSingleMetricViewerConfigModel{
			JobIDs:                []types.String{types.StringValue("job-a")},
			SelectedDetectorIndex: types.Float32Value(2),
			ForecastID:            types.StringValue("forecast-1"),
			FunctionDescription:   types.StringValue("mean"),
			Title:                 types.StringValue("SMV Panel"),
			Description:           types.StringValue("desc"),
			HideTitle:             types.BoolValue(true),
			HideBorder:            types.BoolValue(false),
			TimeRange: &models.TimeRangeModel{
				From: types.StringValue("now-7d"),
				To:   types.StringValue("now"),
				Mode: types.StringValue("relative"),
			},
		},
	}
	var panel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeMlSingleMetricViewer
	diags := mlsinglemetricviewer.BuildConfig(ctx, pm, &panel)
	require.False(t, diags.HasError(), "%v", diags)

	cfg := panel.Config
	require.NotNil(t, cfg.SelectedDetectorIndex)
	require.Equal(t, math.Float32bits(float32(2)), math.Float32bits(*cfg.SelectedDetectorIndex))
	require.NotNil(t, cfg.ForecastId)
	assert.Equal(t, "forecast-1", *cfg.ForecastId)
	require.NotNil(t, cfg.FunctionDescription)
	assert.Equal(t, "mean", *cfg.FunctionDescription)
	require.NotNil(t, cfg.Title)
	assert.Equal(t, "SMV Panel", *cfg.Title)
	require.NotNil(t, cfg.TimeRange)
	assert.Equal(t, "now-7d", cfg.TimeRange.From)
}

func TestBuildConfig_missingConfigBlock(t *testing.T) {
	ctx := context.Background()
	var panel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeMlSingleMetricViewer
	diags := mlsinglemetricviewer.BuildConfig(ctx, models.PanelModel{}, &panel)
	require.True(t, diags.HasError())
	assert.Equal(t, "Missing ML single metric viewer panel configuration", diags[0].Summary())
}

func TestPopulateFromAPI_import_withSelectedEntities(t *testing.T) {
	ctx := context.Background()
	apiCfg := makeAPIConfig(withSelectedEntities(map[string]kbapi.KibanaHTTPAPIsMlSingleMetricViewer_SelectedEntities_AdditionalProperties{
		"host": makeStringEntityProp("web-01"),
	}))

	pm := &models.PanelModel{}
	diags := mlsinglemetricviewer.PopulateFromAPI(ctx, pm, nil, apiCfg)
	require.False(t, diags.HasError(), "%v", diags)

	cfg := pm.MlSingleMetricViewerConfig
	require.NotNil(t, cfg)
	require.True(t, typeutils.IsKnown(cfg.SelectedEntities))
	raw := typeutils.MapTypeAs[models.MlSingleMetricViewerEntityModel](ctx, cfg.SelectedEntities, path.Empty(), &diags)
	require.False(t, diags.HasError(), "%v", diags)
	require.Equal(t, "web-01", raw["host"].StringValue.ValueString())
	assert.True(t, raw["host"].NumericValue.IsNull())
}

func TestPopulateFromAPI_import_numericEntity(t *testing.T) {
	ctx := context.Background()
	apiCfg := makeAPIConfig(withSelectedEntities(map[string]kbapi.KibanaHTTPAPIsMlSingleMetricViewer_SelectedEntities_AdditionalProperties{
		"region_code": makeNumericEntityProp(4),
	}))

	pm := &models.PanelModel{}
	diags := mlsinglemetricviewer.PopulateFromAPI(ctx, pm, nil, apiCfg)
	require.False(t, diags.HasError(), "%v", diags)

	cfg := pm.MlSingleMetricViewerConfig
	raw := typeutils.MapTypeAs[models.MlSingleMetricViewerEntityModel](ctx, cfg.SelectedEntities, path.Empty(), &diags)
	require.False(t, diags.HasError(), "%v", diags)
	assert.True(t, raw["region_code"].StringValue.IsNull())
	f64, acc := raw["region_code"].NumericValue.ValueBigFloat().Float64()
	require.Equal(t, big.Exact, acc)
	assert.InEpsilon(t, float64(4), f64, 1e-6)
}

func TestPopulateFromAPI_nullPreservation(t *testing.T) {
	ctx := context.Background()
	mode := kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchemaModeRelative
	apiCfg := makeAPIConfig(
		withSelectedDetectorIndex(3),
		withForecastID("forecast-api"),
		withFunctionDescription("max"),
		withSelectedEntities(map[string]kbapi.KibanaHTTPAPIsMlSingleMetricViewer_SelectedEntities_AdditionalProperties{
			"host": makeStringEntityProp("web-01"),
		}),
		withTitle("API title"),
		withDescription("API desc"),
		withHideTitle(true),
		withHideBorder(false),
		withTimeRange("now-7d", "now", &mode),
	)

	pm := &models.PanelModel{
		MlSingleMetricViewerConfig: &models.MlSingleMetricViewerConfigModel{
			JobIDs:                []types.String{types.StringValue("job-a")},
			SelectedDetectorIndex: types.Float32Null(),
			ForecastID:            types.StringNull(),
			FunctionDescription:   types.StringNull(),
			SelectedEntities:      types.MapNull(entityObjectType()),
			Title:                 types.StringNull(),
			Description:           types.StringNull(),
			HideTitle:             types.BoolNull(),
			HideBorder:            types.BoolNull(),
			TimeRange:             nil,
		},
	}
	prior := &models.PanelModel{
		MlSingleMetricViewerConfig: &models.MlSingleMetricViewerConfigModel{
			JobIDs:                []types.String{types.StringValue("job-a")},
			SelectedDetectorIndex: types.Float32Null(),
			ForecastID:            types.StringNull(),
			FunctionDescription:   types.StringNull(),
			SelectedEntities:      types.MapNull(entityObjectType()),
			Title:                 types.StringNull(),
			Description:           types.StringNull(),
			HideTitle:             types.BoolNull(),
			HideBorder:            types.BoolNull(),
			TimeRange:             nil,
		},
	}
	diags := mlsinglemetricviewer.PopulateFromAPI(ctx, pm, prior, apiCfg)
	require.False(t, diags.HasError(), "%v", diags)

	cfg := pm.MlSingleMetricViewerConfig
	require.NotNil(t, cfg)
	assert.True(t, cfg.SelectedDetectorIndex.IsNull())
	assert.True(t, cfg.ForecastID.IsNull())
	assert.True(t, cfg.FunctionDescription.IsNull())
	assert.True(t, cfg.SelectedEntities.IsNull())
	assert.True(t, cfg.Title.IsNull())
	assert.True(t, cfg.Description.IsNull())
	assert.True(t, cfg.HideTitle.IsNull())
	assert.True(t, cfg.HideBorder.IsNull())
	assert.Nil(t, cfg.TimeRange)
}

func TestPopulateFromAPI_selectedEntities_roundTrip(t *testing.T) {
	ctx := context.Background()
	selectedEntities := makeSelectedEntitiesMap(t, map[string]models.MlSingleMetricViewerEntityModel{
		"host": {
			StringValue:  types.StringValue("web-01"),
			NumericValue: types.NumberNull(),
		},
	})
	apiCfg := makeAPIConfig(withSelectedEntities(map[string]kbapi.KibanaHTTPAPIsMlSingleMetricViewer_SelectedEntities_AdditionalProperties{
		"host": makeStringEntityProp("web-01"),
	}))

	pm := &models.PanelModel{
		MlSingleMetricViewerConfig: &models.MlSingleMetricViewerConfigModel{
			JobIDs:           []types.String{types.StringValue("job-a")},
			SelectedEntities: selectedEntities,
		},
	}
	prior := &models.PanelModel{
		MlSingleMetricViewerConfig: &models.MlSingleMetricViewerConfigModel{
			JobIDs:           []types.String{types.StringValue("job-a")},
			SelectedEntities: selectedEntities,
		},
	}
	diags := mlsinglemetricviewer.PopulateFromAPI(ctx, pm, prior, apiCfg)
	require.False(t, diags.HasError(), "%v", diags)

	raw := typeutils.MapTypeAs[models.MlSingleMetricViewerEntityModel](ctx, pm.MlSingleMetricViewerConfig.SelectedEntities, path.Empty(), &diags)
	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, "web-01", raw["host"].StringValue.ValueString())
	assert.True(t, raw["host"].NumericValue.IsNull())
}

func TestPopulateFromAPI_timeRangeSubfields_nullPreservation(t *testing.T) {
	ctx := context.Background()
	mode := kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchemaModeRelative
	apiCfg := makeAPIConfig(withTimeRange("now-7d", "now", &mode))

	pm := &models.PanelModel{
		MlSingleMetricViewerConfig: &models.MlSingleMetricViewerConfigModel{
			JobIDs: []types.String{types.StringValue("job-a")},
			TimeRange: &models.TimeRangeModel{
				From: types.StringValue("now-7d"),
				To:   types.StringValue("now"),
				Mode: types.StringNull(),
			},
		},
	}
	prior := &models.PanelModel{
		MlSingleMetricViewerConfig: &models.MlSingleMetricViewerConfigModel{
			JobIDs: []types.String{types.StringValue("job-a")},
			TimeRange: &models.TimeRangeModel{
				From: types.StringValue("now-7d"),
				To:   types.StringValue("now"),
				Mode: types.StringNull(),
			},
		},
	}
	diags := mlsinglemetricviewer.PopulateFromAPI(ctx, pm, prior, apiCfg)
	require.False(t, diags.HasError(), "%v", diags)

	cfg := pm.MlSingleMetricViewerConfig
	require.NotNil(t, cfg.TimeRange)
	assert.Equal(t, "now-7d", cfg.TimeRange.From.ValueString())
	assert.Equal(t, "now", cfg.TimeRange.To.ValueString())
	assert.True(t, cfg.TimeRange.Mode.IsNull())
}

func TestPopulateFromAPI_import_selectedDetectorIndexDefault(t *testing.T) {
	ctx := context.Background()
	apiCfg := makeAPIConfig(withSelectedDetectorIndex(0))

	pm := &models.PanelModel{}
	diags := mlsinglemetricviewer.PopulateFromAPI(ctx, pm, nil, apiCfg)
	require.False(t, diags.HasError(), "%v", diags)

	cfg := pm.MlSingleMetricViewerConfig
	require.NotNil(t, cfg)
	assert.InDelta(t, float32(0), cfg.SelectedDetectorIndex.ValueFloat32(), 1e-6)
}

func TestPopulateFromAPI_selectedDetectorIndex_float32RoundTrip(t *testing.T) {
	ctx := context.Background()
	const idx float32 = 2.5
	apiCfg := makeAPIConfig(withSelectedDetectorIndex(idx))

	pm := &models.PanelModel{
		MlSingleMetricViewerConfig: &models.MlSingleMetricViewerConfigModel{
			JobIDs:                []types.String{types.StringValue("job-a")},
			SelectedDetectorIndex: types.Float32Value(idx),
		},
	}
	prior := &models.PanelModel{
		MlSingleMetricViewerConfig: &models.MlSingleMetricViewerConfigModel{
			JobIDs:                []types.String{types.StringValue("job-a")},
			SelectedDetectorIndex: types.Float32Value(idx),
		},
	}
	diags := mlsinglemetricviewer.PopulateFromAPI(ctx, pm, prior, apiCfg)
	require.False(t, diags.HasError(), "%v", diags)

	require.Equal(t, math.Float32bits(idx), math.Float32bits(pm.MlSingleMetricViewerConfig.SelectedDetectorIndex.ValueFloat32()))
}
