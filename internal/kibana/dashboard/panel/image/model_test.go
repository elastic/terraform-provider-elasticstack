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

package image

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func imagePanelModel(cfg *models.ImagePanelConfigModel, x, y int64) models.PanelModel {
	return models.PanelModel{
		Type:        types.StringValue(panelType),
		Grid:        models.PanelGridModel{X: types.Int64Value(x), Y: types.Int64Value(y)},
		ImageConfig: cfg,
	}
}

func Test_imagePanelToAPI_fileSrc(t *testing.T) {
	pm := imagePanelModel(&models.ImagePanelConfigModel{
		Src: models.ImagePanelSrcModel{
			File: &models.ImagePanelSrcFileModel{FileID: types.StringValue("file-abc")},
		},
		AltText:   types.StringValue("diagram"),
		ObjectFit: types.StringValue("cover"),
	}, 0, 0)
	item, diags := Handler{}.ToAPI(pm, nil)
	require.False(t, diags.HasError())

	img, err := item.AsKbnDashboardPanelTypeImage()
	require.NoError(t, err)
	src0, err := img.Config.ImageConfig.Src.AsKbnDashboardPanelTypeImageConfigImageConfigSrc0()
	require.NoError(t, err)
	assert.Equal(t, kbapi.File, src0.Type)
	assert.Equal(t, "file-abc", src0.FileId)
	require.NotNil(t, img.Config.ImageConfig.AltText)
	assert.Equal(t, "diagram", *img.Config.ImageConfig.AltText)
	require.NotNil(t, img.Config.ImageConfig.ObjectFit)
	assert.Equal(t, kbapi.KbnDashboardPanelTypeImageConfigImageConfigObjectFitCover, *img.Config.ImageConfig.ObjectFit)
}

func Test_imagePanelToAPI_urlSrc(t *testing.T) {
	pm := imagePanelModel(&models.ImagePanelConfigModel{
		Src: models.ImagePanelSrcModel{
			URL: &models.ImagePanelSrcURLModel{URL: types.StringValue("https://example.com/x.png")},
		},
	}, 1, 2)
	item, diags := Handler{}.ToAPI(pm, nil)
	require.False(t, diags.HasError())

	img, err := item.AsKbnDashboardPanelTypeImage()
	require.NoError(t, err)
	src1, err := img.Config.ImageConfig.Src.AsKbnDashboardPanelTypeImageConfigImageConfigSrc1()
	require.NoError(t, err)
	assert.Equal(t, kbapi.Url, src1.Type)
	assert.Equal(t, "https://example.com/x.png", src1.Url)
}

func Test_imagePanelToAPI_drilldowns(t *testing.T) {
	pm := imagePanelModel(&models.ImagePanelConfigModel{
		Src: models.ImagePanelSrcModel{
			File: &models.ImagePanelSrcFileModel{FileID: types.StringValue("f")},
		},
		Drilldowns: []models.ImagePanelDrilldownModel{
			{
				DashboardDrilldown: &models.ImagePanelDashboardDrilldownModel{
					DashboardID: types.StringValue("dash-1"),
					Label:       types.StringValue("Open dash"),
					Trigger:     types.StringValue("on_click_image"),
					UseFilters:  types.BoolValue(true),
				},
			},
			{
				URLDrilldown: &models.ImagePanelURLDrilldownModel{
					URL:          types.StringValue("https://kibana/{{kibana.host}}/"),
					Label:        types.StringValue("menu link"),
					Trigger:      types.StringValue("on_open_panel_menu"),
					EncodeURL:    types.BoolValue(false),
					OpenInNewTab: types.BoolValue(true),
				},
			},
		},
	}, 0, 0)
	item, diags := Handler{}.ToAPI(pm, nil)
	require.False(t, diags.HasError())

	img, err := item.AsKbnDashboardPanelTypeImage()
	require.NoError(t, err)
	require.NotNil(t, img.Config.Drilldowns)
	require.Len(t, *img.Config.Drilldowns, 2)

	dd0, err := (*img.Config.Drilldowns)[0].AsKbnDashboardPanelTypeImageConfigDrilldowns0()
	require.NoError(t, err)
	assert.Equal(t, "dash-1", dd0.DashboardId)
	require.NotNil(t, dd0.UseFilters)
	assert.True(t, *dd0.UseFilters)

	dd1, err := (*img.Config.Drilldowns)[1].AsKbnDashboardPanelTypeImageConfigDrilldowns1()
	require.NoError(t, err)
	assert.Equal(t, kbapi.KbnDashboardPanelTypeImageConfigDrilldowns1TriggerOnOpenPanelMenu, dd1.Trigger)
	require.NotNil(t, dd1.EncodeUrl)
	assert.False(t, *dd1.EncodeUrl)
}

func Test_populateImagePanelFromAPI_nullPreservation(t *testing.T) {
	useF := false
	tr := false
	tab := false
	encode := true
	objFit := kbapi.KbnDashboardPanelTypeImageConfigImageConfigObjectFitContain

	apiPanel := kbapi.KbnDashboardPanelTypeImage{}
	title := "t"
	apiPanel.Config.Title = &title
	apiPanel.Config.ImageConfig.ObjectFit = &objFit
	src := kbapi.KbnDashboardPanelTypeImageConfigImageConfigSrc0{Type: kbapi.File, FileId: "img-1"}
	require.NoError(t, apiPanel.Config.ImageConfig.Src.FromKbnDashboardPanelTypeImageConfigImageConfigSrc0(src))

	d0 := kbapi.KbnDashboardPanelTypeImageConfigDrilldowns0{
		DashboardId:  "d1",
		Label:        "l",
		Trigger:      kbapi.KbnDashboardPanelTypeImageConfigDrilldowns0TriggerOnClickImage,
		Type:         kbapi.KbnDashboardPanelTypeImageConfigDrilldowns0TypeDashboardDrilldown,
		UseFilters:   &useF,
		UseTimeRange: &tr,
		OpenInNewTab: &tab,
	}
	var dashItem kbapi.KbnDashboardPanelTypeImage_Config_Drilldowns_Item
	require.NoError(t, dashItem.FromKbnDashboardPanelTypeImageConfigDrilldowns0(d0))

	d1 := kbapi.KbnDashboardPanelTypeImageConfigDrilldowns1{
		Url:          "https://example.com",
		Label:        "u",
		Trigger:      kbapi.KbnDashboardPanelTypeImageConfigDrilldowns1TriggerOnClickImage,
		Type:         kbapi.KbnDashboardPanelTypeImageConfigDrilldowns1TypeUrlDrilldown,
		EncodeUrl:    &encode,
		OpenInNewTab: &tab,
	}
	var urlItem kbapi.KbnDashboardPanelTypeImage_Config_Drilldowns_Item
	require.NoError(t, urlItem.FromKbnDashboardPanelTypeImageConfigDrilldowns1(d1))

	items := []kbapi.KbnDashboardPanelTypeImage_Config_Drilldowns_Item{dashItem, urlItem}
	apiPanel.Config.Drilldowns = &items

	prior := models.PanelModel{
		ImageConfig: &models.ImagePanelConfigModel{
			Src: models.ImagePanelSrcModel{
				File: &models.ImagePanelSrcFileModel{FileID: types.StringValue("img-1")},
			},
			Title:     types.StringNull(),
			ObjectFit: types.StringNull(),
			Drilldowns: []models.ImagePanelDrilldownModel{
				{
					DashboardDrilldown: &models.ImagePanelDashboardDrilldownModel{
						DashboardID:  types.StringValue("d1"),
						Label:        types.StringValue("l"),
						Trigger:      types.StringValue("on_click_image"),
						UseFilters:   types.BoolNull(),
						UseTimeRange: types.BoolNull(),
						OpenInNewTab: types.BoolNull(),
					},
				},
				{
					URLDrilldown: &models.ImagePanelURLDrilldownModel{
						URL:          types.StringValue("https://example.com"),
						Label:        types.StringValue("u"),
						Trigger:      types.StringValue("on_click_image"),
						EncodeURL:    types.BoolNull(),
						OpenInNewTab: types.BoolNull(),
					},
				},
			},
		},
	}

	pm := prior
	PopulateFromAPI(&pm, &prior, apiPanel)

	cfg := pm.ImageConfig
	require.NotNil(t, cfg)
	assert.True(t, cfg.Title.IsNull(), "title omitted in TF should stay null even when API sets it")
	assert.True(t, cfg.ObjectFit.IsNull(), "object_fit null should stay null when API echoes contain default")

	require.Len(t, cfg.Drilldowns, 2)
	assert.True(t, cfg.Drilldowns[0].DashboardDrilldown.UseFilters.IsNull())
	assert.True(t, cfg.Drilldowns[0].DashboardDrilldown.UseTimeRange.IsNull())
	assert.True(t, cfg.Drilldowns[0].DashboardDrilldown.OpenInNewTab.IsNull())
	assert.True(t, cfg.Drilldowns[1].URLDrilldown.EncodeURL.IsNull())
	assert.True(t, cfg.Drilldowns[1].URLDrilldown.OpenInNewTab.IsNull())
}

func Test_populateImagePanelFromAPI_import_drilldownDefaultsAndObjectFitNull(t *testing.T) {
	useF := false
	tr := false
	tab := false
	encode := true
	objFit := kbapi.KbnDashboardPanelTypeImageConfigImageConfigObjectFitContain

	apiPanel := kbapi.KbnDashboardPanelTypeImage{}
	apiPanel.Config.ImageConfig.ObjectFit = &objFit
	src := kbapi.KbnDashboardPanelTypeImageConfigImageConfigSrc0{Type: kbapi.File, FileId: "img-1"}
	require.NoError(t, apiPanel.Config.ImageConfig.Src.FromKbnDashboardPanelTypeImageConfigImageConfigSrc0(src))

	d0 := kbapi.KbnDashboardPanelTypeImageConfigDrilldowns0{
		DashboardId:  "d1",
		Label:        "l",
		Trigger:      kbapi.KbnDashboardPanelTypeImageConfigDrilldowns0TriggerOnClickImage,
		Type:         kbapi.KbnDashboardPanelTypeImageConfigDrilldowns0TypeDashboardDrilldown,
		UseFilters:   &useF,
		UseTimeRange: &tr,
		OpenInNewTab: &tab,
	}
	var dashItem kbapi.KbnDashboardPanelTypeImage_Config_Drilldowns_Item
	require.NoError(t, dashItem.FromKbnDashboardPanelTypeImageConfigDrilldowns0(d0))

	d1 := kbapi.KbnDashboardPanelTypeImageConfigDrilldowns1{
		Url:          "https://example.com",
		Label:        "u",
		Trigger:      kbapi.KbnDashboardPanelTypeImageConfigDrilldowns1TriggerOnClickImage,
		Type:         kbapi.KbnDashboardPanelTypeImageConfigDrilldowns1TypeUrlDrilldown,
		EncodeUrl:    &encode,
		OpenInNewTab: &tab,
	}
	var urlItem kbapi.KbnDashboardPanelTypeImage_Config_Drilldowns_Item
	require.NoError(t, urlItem.FromKbnDashboardPanelTypeImageConfigDrilldowns1(d1))

	apiPanel.Config.Drilldowns = &[]kbapi.KbnDashboardPanelTypeImage_Config_Drilldowns_Item{dashItem, urlItem}

	pm := models.PanelModel{}
	PopulateFromAPI(&pm, nil, apiPanel)

	cfg := pm.ImageConfig
	require.NotNil(t, cfg)
	assert.True(t, cfg.ObjectFit.IsNull(), "import: object_fit contain default should be null in state")

	require.Len(t, cfg.Drilldowns, 2)
	dd := cfg.Drilldowns[0].DashboardDrilldown
	require.NotNil(t, dd)
	assert.True(t, dd.UseFilters.IsNull())
	assert.True(t, dd.UseTimeRange.IsNull())
	assert.True(t, dd.OpenInNewTab.IsNull())

	ud := cfg.Drilldowns[1].URLDrilldown
	require.NotNil(t, ud)
	assert.True(t, ud.EncodeURL.IsNull())
	assert.True(t, ud.OpenInNewTab.IsNull())
}

func Test_populateImagePanelFromAPI_import_setsSrcBranches(t *testing.T) {
	apiPanel := kbapi.KbnDashboardPanelTypeImage{}
	src := kbapi.KbnDashboardPanelTypeImageConfigImageConfigSrc0{Type: kbapi.File, FileId: "fid"}
	require.NoError(t, apiPanel.Config.ImageConfig.Src.FromKbnDashboardPanelTypeImageConfigImageConfigSrc0(src))

	pm := models.PanelModel{}
	PopulateFromAPI(&pm, nil, apiPanel)
	require.NotNil(t, pm.ImageConfig)
	assert.NotNil(t, pm.ImageConfig.Src.File)
	assert.Nil(t, pm.ImageConfig.Src.URL)
}

func Test_imagePanelSrcValidator(t *testing.T) {
	ctx := context.Background()
	v := panelkit.ExactlyOneOfNestedAttrsValidator(panelkit.ExactlyOneOfNestedAttrsOpts{
		AttrNames:     []string{"file", "url"},
		Summary:       "Invalid image_config.src",
		MissingDetail: "Exactly one of `file` or `url` must be set inside `src`.",
		TooManyDetail: "Exactly one of `file` or `url` must be set inside `src`, not both.",
	})

	fileType := types.ObjectType{AttrTypes: map[string]attr.Type{"file_id": types.StringType}}
	urlType := types.ObjectType{AttrTypes: map[string]attr.Type{"url": types.StringType}}
	srcType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"file": fileType,
		"url":  urlType,
	}}

	fileObj := types.ObjectValueMust(fileType.AttrTypes, map[string]attr.Value{"file_id": types.StringValue("a")})
	urlObj := types.ObjectValueMust(urlType.AttrTypes, map[string]attr.Value{"url": types.StringValue("https://x")})

	t.Run("rejects both file and url", func(t *testing.T) {
		ov := types.ObjectValueMust(srcType.AttrTypes, map[string]attr.Value{
			"file": fileObj,
			"url":  urlObj,
		})
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("src")}, &resp)
		require.True(t, resp.Diagnostics.HasError())
	})

	t.Run("rejects neither", func(t *testing.T) {
		ov := types.ObjectValueMust(srcType.AttrTypes, map[string]attr.Value{
			"file": types.ObjectNull(fileType.AttrTypes),
			"url":  types.ObjectNull(urlType.AttrTypes),
		})
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("src")}, &resp)
		require.True(t, resp.Diagnostics.HasError())
	})

	t.Run("accepts file only", func(t *testing.T) {
		ov := types.ObjectValueMust(srcType.AttrTypes, map[string]attr.Value{
			"file": fileObj,
			"url":  types.ObjectNull(urlType.AttrTypes),
		})
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("src")}, &resp)
		require.False(t, resp.Diagnostics.HasError())
	})
}

func Test_imagePanel_objectFitValidator(t *testing.T) {
	ctx := context.Background()
	v := stringvalidator.OneOf("fill", "contain", "cover", "none")
	var resp validator.StringResponse
	v.ValidateString(ctx, validator.StringRequest{
		ConfigValue: types.StringValue("stretch"),
		Path:        path.Root("object_fit"),
	}, &resp)
	require.True(t, resp.Diagnostics.HasError())
}

func Test_imagePanel_dashboardDrilldownTriggerValidator(t *testing.T) {
	ctx := context.Background()
	v := stringvalidator.OneOf("on_click_image")
	var resp validator.StringResponse
	v.ValidateString(ctx, validator.StringRequest{
		ConfigValue: types.StringValue("on_open_panel_menu"),
		Path:        path.Root("trigger"),
	}, &resp)
	require.True(t, resp.Diagnostics.HasError())
}
