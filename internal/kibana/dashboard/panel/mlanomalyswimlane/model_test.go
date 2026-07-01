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

package mlanomalyswimlane_test

import (
	"math"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/mlanomalyswimlane"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeOverallAPIUnion(opts ...func(*kbapi.KibanaHTTPAPIsMlAnomalySwimlane0)) kbapi.KibanaHTTPAPIsMlAnomalySwimlane {
	cfg := kbapi.KibanaHTTPAPIsMlAnomalySwimlane0{
		SwimlaneType: kbapi.Overall,
		JobIds:       []string{"job-a"},
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	var union kbapi.KibanaHTTPAPIsMlAnomalySwimlane
	if err := union.FromKibanaHTTPAPIsMlAnomalySwimlane0(cfg); err != nil {
		panic(err)
	}
	return union
}

func makeViewByAPIUnion(opts ...func(*kbapi.KibanaHTTPAPIsMlAnomalySwimlane1)) kbapi.KibanaHTTPAPIsMlAnomalySwimlane {
	cfg := kbapi.KibanaHTTPAPIsMlAnomalySwimlane1{
		SwimlaneType: kbapi.ViewBy,
		JobIds:       []string{"job-a"},
		ViewBy:       "host.name",
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	var union kbapi.KibanaHTTPAPIsMlAnomalySwimlane
	if err := union.FromKibanaHTTPAPIsMlAnomalySwimlane1(cfg); err != nil {
		panic(err)
	}
	return union
}

func withOverallTitle(title string) func(*kbapi.KibanaHTTPAPIsMlAnomalySwimlane0) {
	return func(c *kbapi.KibanaHTTPAPIsMlAnomalySwimlane0) { c.Title = &title }
}

func withOverallDescription(desc string) func(*kbapi.KibanaHTTPAPIsMlAnomalySwimlane0) {
	return func(c *kbapi.KibanaHTTPAPIsMlAnomalySwimlane0) { c.Description = &desc }
}

func withOverallHideTitle(v bool) func(*kbapi.KibanaHTTPAPIsMlAnomalySwimlane0) {
	return func(c *kbapi.KibanaHTTPAPIsMlAnomalySwimlane0) { c.HideTitle = &v }
}

func withOverallHideBorder(v bool) func(*kbapi.KibanaHTTPAPIsMlAnomalySwimlane0) {
	return func(c *kbapi.KibanaHTTPAPIsMlAnomalySwimlane0) { c.HideBorder = &v }
}

func withOverallPerPage(v float32) func(*kbapi.KibanaHTTPAPIsMlAnomalySwimlane0) {
	return func(c *kbapi.KibanaHTTPAPIsMlAnomalySwimlane0) { c.PerPage = &v }
}

func withOverallTimeRange(from, to string, mode *kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchemaMode) func(*kbapi.KibanaHTTPAPIsMlAnomalySwimlane0) {
	return func(c *kbapi.KibanaHTTPAPIsMlAnomalySwimlane0) {
		c.TimeRange = &kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema{
			From: from,
			To:   to,
			Mode: mode,
		}
	}
}

func withViewByPerPage(v float32) func(*kbapi.KibanaHTTPAPIsMlAnomalySwimlane1) {
	return func(c *kbapi.KibanaHTTPAPIsMlAnomalySwimlane1) { c.PerPage = &v }
}

func withViewByTitle(title string) func(*kbapi.KibanaHTTPAPIsMlAnomalySwimlane1) {
	return func(c *kbapi.KibanaHTTPAPIsMlAnomalySwimlane1) { c.Title = &title }
}

func withViewByDescription(desc string) func(*kbapi.KibanaHTTPAPIsMlAnomalySwimlane1) {
	return func(c *kbapi.KibanaHTTPAPIsMlAnomalySwimlane1) { c.Description = &desc }
}

func withViewByHideTitle(v bool) func(*kbapi.KibanaHTTPAPIsMlAnomalySwimlane1) {
	return func(c *kbapi.KibanaHTTPAPIsMlAnomalySwimlane1) { c.HideTitle = &v }
}

func withViewByHideBorder(v bool) func(*kbapi.KibanaHTTPAPIsMlAnomalySwimlane1) {
	return func(c *kbapi.KibanaHTTPAPIsMlAnomalySwimlane1) { c.HideBorder = &v }
}

func withViewByTimeRange(from, to string, mode *kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchemaMode) func(*kbapi.KibanaHTTPAPIsMlAnomalySwimlane1) {
	return func(c *kbapi.KibanaHTTPAPIsMlAnomalySwimlane1) {
		c.TimeRange = &kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema{
			From: from,
			To:   to,
			Mode: mode,
		}
	}
}

func TestBuildConfig_overall_minimal(t *testing.T) {
	pm := models.PanelModel{
		MlAnomalySwimlaneConfig: &models.MlAnomalySwimlaneConfigModel{
			SwimlaneType: types.StringValue("overall"),
			JobIDs:       []types.String{types.StringValue("job-a")},
			ViewBy:       types.StringNull(),
		},
	}
	var panel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeMlAnomalySwimlane
	diags := mlanomalyswimlane.BuildConfig(pm, &panel)
	require.False(t, diags.HasError(), "%v", diags)

	cfg0, err := panel.Config.AsKibanaHTTPAPIsMlAnomalySwimlane0()
	require.NoError(t, err)
	assert.Equal(t, kbapi.Overall, cfg0.SwimlaneType)
	assert.Equal(t, []string{"job-a"}, cfg0.JobIds)
	assert.Nil(t, cfg0.Title)
	assert.Nil(t, cfg0.PerPage)
}

func TestBuildConfig_viewBy_minimal(t *testing.T) {
	pm := models.PanelModel{
		MlAnomalySwimlaneConfig: &models.MlAnomalySwimlaneConfigModel{
			SwimlaneType: types.StringValue("viewBy"),
			JobIDs:       []types.String{types.StringValue("job-a"), types.StringValue("job-b")},
			ViewBy:       types.StringValue("host.name"),
		},
	}
	var panel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeMlAnomalySwimlane
	diags := mlanomalyswimlane.BuildConfig(pm, &panel)
	require.False(t, diags.HasError(), "%v", diags)

	cfg1, err := panel.Config.AsKibanaHTTPAPIsMlAnomalySwimlane1()
	require.NoError(t, err)
	assert.Equal(t, kbapi.ViewBy, cfg1.SwimlaneType)
	assert.Equal(t, []string{"job-a", "job-b"}, cfg1.JobIds)
	assert.Equal(t, "host.name", cfg1.ViewBy)
}

func TestBuildConfig_withOptionalFields(t *testing.T) {
	pm := models.PanelModel{
		MlAnomalySwimlaneConfig: &models.MlAnomalySwimlaneConfigModel{
			SwimlaneType: types.StringValue("viewBy"),
			JobIDs:       []types.String{types.StringValue("job-a")},
			ViewBy:       types.StringValue("host.name"),
			PerPage:      types.Float32Value(10),
			Title:        types.StringValue("Swim Lane"),
			Description:  types.StringValue("desc"),
			HideTitle:    types.BoolValue(true),
			HideBorder:   types.BoolValue(false),
			TimeRange: &models.TimeRangeModel{
				From: types.StringValue("now-7d"),
				To:   types.StringValue("now"),
				Mode: types.StringValue("relative"),
			},
		},
	}
	var panel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeMlAnomalySwimlane
	diags := mlanomalyswimlane.BuildConfig(pm, &panel)
	require.False(t, diags.HasError(), "%v", diags)

	cfg1, err := panel.Config.AsKibanaHTTPAPIsMlAnomalySwimlane1()
	require.NoError(t, err)
	require.NotNil(t, cfg1.Title)
	assert.Equal(t, "Swim Lane", *cfg1.Title)
	require.NotNil(t, cfg1.Description)
	assert.Equal(t, "desc", *cfg1.Description)
	require.NotNil(t, cfg1.HideTitle)
	assert.True(t, *cfg1.HideTitle)
	require.NotNil(t, cfg1.HideBorder)
	assert.False(t, *cfg1.HideBorder)
	require.NotNil(t, cfg1.PerPage)
	require.Equal(t, math.Float32bits(float32(10)), math.Float32bits(*cfg1.PerPage))
	require.NotNil(t, cfg1.TimeRange)
	assert.Equal(t, "now-7d", cfg1.TimeRange.From)
	assert.Equal(t, "now", cfg1.TimeRange.To)
	require.NotNil(t, cfg1.TimeRange.Mode)
	assert.Equal(t, kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchemaModeRelative, *cfg1.TimeRange.Mode)
}

func TestBuildConfig_missingConfigBlock(t *testing.T) {
	var panel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeMlAnomalySwimlane
	diags := mlanomalyswimlane.BuildConfig(models.PanelModel{}, &panel)
	require.True(t, diags.HasError())
	assert.Equal(t, "Missing ML anomaly swim lane panel configuration", diags[0].Summary())
}

func TestBuildConfig_perPage_float32RoundTrip(t *testing.T) {
	const perPage float32 = 12.5
	pm := models.PanelModel{
		MlAnomalySwimlaneConfig: &models.MlAnomalySwimlaneConfigModel{
			SwimlaneType: types.StringValue("viewBy"),
			JobIDs:       []types.String{types.StringValue("job-a")},
			ViewBy:       types.StringValue("host.name"),
			PerPage:      types.Float32Value(perPage),
		},
	}
	var panel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeMlAnomalySwimlane
	diags := mlanomalyswimlane.BuildConfig(pm, &panel)
	require.False(t, diags.HasError(), "%v", diags)

	cfg1, err := panel.Config.AsKibanaHTTPAPIsMlAnomalySwimlane1()
	require.NoError(t, err)
	require.NotNil(t, cfg1.PerPage)
	require.Equal(t, math.Float32bits(perPage), math.Float32bits(*cfg1.PerPage))
}

func TestPopulateFromAPI_import_overall(t *testing.T) {
	mode := kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchemaModeRelative
	apiCfg := makeOverallAPIUnion(
		withOverallTitle("title"),
		withOverallDescription("desc"),
		withOverallHideTitle(true),
		withOverallHideBorder(false),
		withOverallPerPage(10),
		withOverallTimeRange("now-7d", "now", &mode),
	)
	pm := &models.PanelModel{}
	diags := mlanomalyswimlane.PopulateFromAPI(pm, nil, apiCfg)
	require.False(t, diags.HasError(), "%v", diags)

	cfg := pm.MlAnomalySwimlaneConfig
	require.NotNil(t, cfg)
	assert.Equal(t, "overall", cfg.SwimlaneType.ValueString())
	assert.Len(t, cfg.JobIDs, 1)
	assert.Equal(t, "job-a", cfg.JobIDs[0].ValueString())
	assert.True(t, cfg.ViewBy.IsNull())
	assert.Equal(t, "title", cfg.Title.ValueString())
	assert.Equal(t, float32(10), cfg.PerPage.ValueFloat32())
}

func TestPopulateFromAPI_import_viewBy(t *testing.T) {
	apiCfg := makeViewByAPIUnion(withViewByPerPage(15))
	pm := &models.PanelModel{}
	diags := mlanomalyswimlane.PopulateFromAPI(pm, nil, apiCfg)
	require.False(t, diags.HasError(), "%v", diags)

	cfg := pm.MlAnomalySwimlaneConfig
	require.NotNil(t, cfg)
	assert.Equal(t, "viewBy", cfg.SwimlaneType.ValueString())
	assert.Equal(t, "host.name", cfg.ViewBy.ValueString())
	assert.Equal(t, float32(15), cfg.PerPage.ValueFloat32())
}

func TestPopulateFromAPI_overall_nullPreservation(t *testing.T) {
	mode := kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchemaModeAbsolute
	apiCfg := makeOverallAPIUnion(
		withOverallTitle("API title"),
		withOverallDescription("API desc"),
		withOverallHideTitle(true),
		withOverallHideBorder(false),
		withOverallPerPage(20),
		withOverallTimeRange("now-30d", "now", &mode),
	)

	pm := &models.PanelModel{
		MlAnomalySwimlaneConfig: &models.MlAnomalySwimlaneConfigModel{
			SwimlaneType: types.StringValue("overall"),
			JobIDs:       []types.String{types.StringValue("job-a")},
			ViewBy:       types.StringNull(),
			PerPage:      types.Float32Null(),
			Title:        types.StringNull(),
			Description:  types.StringNull(),
			HideTitle:    types.BoolNull(),
			HideBorder:   types.BoolNull(),
			TimeRange:    nil,
		},
	}
	prior := &models.PanelModel{
		MlAnomalySwimlaneConfig: &models.MlAnomalySwimlaneConfigModel{
			SwimlaneType: types.StringValue("overall"),
			JobIDs:       []types.String{types.StringValue("job-a")},
			ViewBy:       types.StringNull(),
			PerPage:      types.Float32Null(),
			Title:        types.StringNull(),
			Description:  types.StringNull(),
			HideTitle:    types.BoolNull(),
			HideBorder:   types.BoolNull(),
			TimeRange:    nil,
		},
	}
	diags := mlanomalyswimlane.PopulateFromAPI(pm, prior, apiCfg)
	require.False(t, diags.HasError(), "%v", diags)

	cfg := pm.MlAnomalySwimlaneConfig
	require.NotNil(t, cfg)
	assert.True(t, cfg.ViewBy.IsNull())
	assert.True(t, cfg.PerPage.IsNull())
	assert.True(t, cfg.Title.IsNull())
	assert.True(t, cfg.Description.IsNull())
	assert.True(t, cfg.HideTitle.IsNull())
	assert.True(t, cfg.HideBorder.IsNull())
	assert.Nil(t, cfg.TimeRange)
}

func TestPopulateFromAPI_viewBy_nullPreservation(t *testing.T) {
	mode := kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchemaModeRelative
	apiCfg := makeViewByAPIUnion(
		withViewByPerPage(25),
		withViewByTitle("API title"),
		withViewByDescription("API desc"),
		withViewByHideTitle(true),
		withViewByHideBorder(false),
		withViewByTimeRange("now-30d", "now", &mode),
	)

	pm := &models.PanelModel{
		MlAnomalySwimlaneConfig: &models.MlAnomalySwimlaneConfigModel{
			SwimlaneType: types.StringValue("viewBy"),
			JobIDs:       []types.String{types.StringValue("job-a")},
			ViewBy:       types.StringValue("host.name"),
			PerPage:      types.Float32Null(),
			Title:        types.StringNull(),
			Description:  types.StringNull(),
			HideTitle:    types.BoolNull(),
			HideBorder:   types.BoolNull(),
			TimeRange:    nil,
		},
	}
	prior := &models.PanelModel{
		MlAnomalySwimlaneConfig: &models.MlAnomalySwimlaneConfigModel{
			SwimlaneType: types.StringValue("viewBy"),
			JobIDs:       []types.String{types.StringValue("job-a")},
			ViewBy:       types.StringValue("host.name"),
			PerPage:      types.Float32Null(),
			Title:        types.StringNull(),
			Description:  types.StringNull(),
			HideTitle:    types.BoolNull(),
			HideBorder:   types.BoolNull(),
			TimeRange:    nil,
		},
	}
	diags := mlanomalyswimlane.PopulateFromAPI(pm, prior, apiCfg)
	require.False(t, diags.HasError(), "%v", diags)

	cfg := pm.MlAnomalySwimlaneConfig
	require.NotNil(t, cfg)
	assert.Equal(t, "host.name", cfg.ViewBy.ValueString())
	assert.True(t, cfg.PerPage.IsNull())
	assert.True(t, cfg.Title.IsNull())
	assert.True(t, cfg.Description.IsNull())
	assert.True(t, cfg.HideTitle.IsNull())
	assert.True(t, cfg.HideBorder.IsNull())
	assert.Nil(t, cfg.TimeRange)
}

func TestPopulateFromAPI_timeRangeSubfields_nullPreservation(t *testing.T) {
	mode := kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchemaModeRelative
	apiCfg := makeViewByAPIUnion(func(c *kbapi.KibanaHTTPAPIsMlAnomalySwimlane1) {
		c.TimeRange = &kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema{
			From: "now-7d",
			To:   "now",
			Mode: &mode,
		}
	})

	pm := &models.PanelModel{
		MlAnomalySwimlaneConfig: &models.MlAnomalySwimlaneConfigModel{
			SwimlaneType: types.StringValue("viewBy"),
			JobIDs:       []types.String{types.StringValue("job-a")},
			ViewBy:       types.StringValue("host.name"),
			TimeRange: &models.TimeRangeModel{
				From: types.StringValue("now-7d"),
				To:   types.StringValue("now"),
				Mode: types.StringNull(),
			},
		},
	}
	prior := &models.PanelModel{
		MlAnomalySwimlaneConfig: &models.MlAnomalySwimlaneConfigModel{
			SwimlaneType: types.StringValue("viewBy"),
			JobIDs:       []types.String{types.StringValue("job-a")},
			ViewBy:       types.StringValue("host.name"),
			TimeRange: &models.TimeRangeModel{
				From: types.StringValue("now-7d"),
				To:   types.StringValue("now"),
				Mode: types.StringNull(),
			},
		},
	}
	diags := mlanomalyswimlane.PopulateFromAPI(pm, prior, apiCfg)
	require.False(t, diags.HasError(), "%v", diags)

	cfg := pm.MlAnomalySwimlaneConfig
	require.NotNil(t, cfg.TimeRange)
	assert.Equal(t, "now-7d", cfg.TimeRange.From.ValueString())
	assert.Equal(t, "now", cfg.TimeRange.To.ValueString())
	assert.True(t, cfg.TimeRange.Mode.IsNull())
}

func TestPopulateFromAPI_perPage_float32RoundTrip(t *testing.T) {
	const perPage float32 = 12.5
	apiCfg := makeViewByAPIUnion(withViewByPerPage(perPage))

	pm := &models.PanelModel{
		MlAnomalySwimlaneConfig: &models.MlAnomalySwimlaneConfigModel{
			SwimlaneType: types.StringValue("viewBy"),
			JobIDs:       []types.String{types.StringValue("job-a")},
			ViewBy:       types.StringValue("host.name"),
			PerPage:      types.Float32Value(perPage),
		},
	}
	prior := &models.PanelModel{
		MlAnomalySwimlaneConfig: &models.MlAnomalySwimlaneConfigModel{
			SwimlaneType: types.StringValue("viewBy"),
			JobIDs:       []types.String{types.StringValue("job-a")},
			ViewBy:       types.StringValue("host.name"),
			PerPage:      types.Float32Value(perPage),
		},
	}
	diags := mlanomalyswimlane.PopulateFromAPI(pm, prior, apiCfg)
	require.False(t, diags.HasError(), "%v", diags)

	require.Equal(t, math.Float32bits(perPage), math.Float32bits(pm.MlAnomalySwimlaneConfig.PerPage.ValueFloat32()))
}
