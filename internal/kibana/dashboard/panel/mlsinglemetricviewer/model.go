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

package mlsinglemetricviewer

import (
	"context"
	"fmt"
	"math"
	"math/big"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var mlSingleMetricViewerEntityObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		entityAttrStringValue:  types.StringType,
		entityAttrNumericValue: types.NumberType,
	},
}

// BuildConfig writes Terraform state from pm into panel's typed API config.
func BuildConfig(ctx context.Context, pm models.PanelModel, panel *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeMlSingleMetricViewer) diag.Diagnostics {
	cfg := pm.MlSingleMetricViewerConfig
	if cfg == nil {
		var diags diag.Diagnostics
		diags.AddError(
			"Missing ML single metric viewer panel configuration",
			"ML single metric viewer panels require `ml_single_metric_viewer_config`.",
		)
		return diags
	}

	apiCfg := kbapi.KibanaHTTPAPIsMlSingleMetricViewer{
		JobIds: typeutils.ValueStringSlice(cfg.JobIDs),
	}

	panelkit.BuildPresentationConfig(cfg.Title, cfg.Description, cfg.HideTitle, cfg.HideBorder,
		&apiCfg.Title, &apiCfg.Description, &apiCfg.HideTitle, &apiCfg.HideBorder)
	if typeutils.IsKnown(cfg.SelectedDetectorIndex) {
		v := cfg.SelectedDetectorIndex.ValueFloat32()
		apiCfg.SelectedDetectorIndex = &v
	}
	if typeutils.IsKnown(cfg.ForecastID) {
		apiCfg.ForecastId = cfg.ForecastID.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.FunctionDescription) {
		apiCfg.FunctionDescription = cfg.FunctionDescription.ValueStringPointer()
	}
	if cfg.TimeRange != nil {
		apiCfg.TimeRange = lenscommon.TimeRangeModelToAPI(cfg.TimeRange)
	}

	var diags diag.Diagnostics
	apiCfg.SelectedEntities = mlSingleMetricViewerSelectedEntitiesToAPI(ctx, cfg.SelectedEntities, &diags)
	if diags.HasError() {
		return diags
	}

	panel.Config = apiCfg
	return nil
}

// PopulateFromAPI maps Kibana ML single metric viewer config into Terraform panel state while preserving prior null intent.
func PopulateFromAPI(ctx context.Context, pm *models.PanelModel, prior *models.PanelModel, apiConfig kbapi.KibanaHTTPAPIsMlSingleMetricViewer) diag.Diagnostics {
	if prior == nil {
		cfg, diags := mlSingleMetricViewerConfigFromAPIImport(ctx, apiConfig)
		if diags.HasError() {
			return diags
		}
		pm.MlSingleMetricViewerConfig = cfg
		return nil
	}

	if pm.MlSingleMetricViewerConfig == nil && prior.MlSingleMetricViewerConfig != nil {
		cfg, diags := mlSingleMetricViewerConfigFromAPIImport(ctx, apiConfig)
		if diags.HasError() {
			return diags
		}
		pm.MlSingleMetricViewerConfig = cfg
	}

	existing := pm.MlSingleMetricViewerConfig
	if existing == nil {
		return nil
	}

	return mlSingleMetricViewerMergeFromAPI(ctx, existing, prior.MlSingleMetricViewerConfig, apiConfig)
}

func mlSingleMetricViewerConfigFromAPIImport(ctx context.Context, apiConfig kbapi.KibanaHTTPAPIsMlSingleMetricViewer) (*models.MlSingleMetricViewerConfigModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := &models.MlSingleMetricViewerConfigModel{
		JobIDs:                typeutils.StringSliceValue(apiConfig.JobIds),
		SelectedDetectorIndex: types.Float32PointerValue(apiConfig.SelectedDetectorIndex),
		ForecastID:            types.StringPointerValue(apiConfig.ForecastId),
		FunctionDescription:   types.StringPointerValue(apiConfig.FunctionDescription),
		SelectedEntities:      mlSingleMetricViewerSelectedEntitiesFromAPI(ctx, apiConfig.SelectedEntities, &diags),
		Title:                 types.StringPointerValue(apiConfig.Title),
		Description:           types.StringPointerValue(apiConfig.Description),
		HideTitle:             types.BoolPointerValue(apiConfig.HideTitle),
		HideBorder:            types.BoolPointerValue(apiConfig.HideBorder),
		TimeRange:             panelkit.TimeRangeFromAPI(apiConfig.TimeRange, nil),
	}
	return out, diags
}

func mlSingleMetricViewerMergeFromAPI(
	ctx context.Context,
	existing, prior *models.MlSingleMetricViewerConfigModel,
	apiConfig kbapi.KibanaHTTPAPIsMlSingleMetricViewer,
) diag.Diagnostics {
	var diags diag.Diagnostics

	existing.JobIDs = typeutils.StringSliceValue(apiConfig.JobIds)
	existing.SelectedDetectorIndex = panelkit.PreserveFloat32(existing.SelectedDetectorIndex, apiConfig.SelectedDetectorIndex)
	existing.ForecastID = panelkit.PreserveString(existing.ForecastID, apiConfig.ForecastId)
	existing.FunctionDescription = panelkit.PreserveString(existing.FunctionDescription, apiConfig.FunctionDescription)
	panelkit.ApplyPresentationFromAPI(&existing.Title, &existing.Description, &existing.HideTitle, &existing.HideBorder,
		apiConfig.Title, apiConfig.Description, apiConfig.HideTitle, apiConfig.HideBorder)

	var priorTR *models.TimeRangeModel
	if prior != nil {
		priorTR = prior.TimeRange
	}
	existing.TimeRange = panelkit.MergeTimeRange(existing.TimeRange, apiConfig.TimeRange, priorTR)

	if prior != nil && typeutils.IsKnown(prior.SelectedEntities) {
		existing.SelectedEntities = mlSingleMetricViewerSelectedEntitiesFromAPI(ctx, apiConfig.SelectedEntities, &diags)
	}

	if prior != nil {
		mlSingleMetricViewerPreserveNullIntentFromPrior(prior, existing)
	}

	return diags
}

func mlSingleMetricViewerSelectedEntitiesToAPI(
	ctx context.Context,
	m types.Map,
	diags *diag.Diagnostics,
) *map[string]kbapi.KibanaHTTPAPIsMlSingleMetricViewer_SelectedEntities_AdditionalProperties {
	if !typeutils.IsKnown(m) || m.IsNull() {
		return nil
	}
	raw := typeutils.MapTypeAs[models.MlSingleMetricViewerEntityModel](ctx, m, path.Empty(), diags)
	if diags.HasError() {
		return nil
	}
	if len(raw) == 0 {
		return nil
	}
	out := make(map[string]kbapi.KibanaHTTPAPIsMlSingleMetricViewer_SelectedEntities_AdditionalProperties, len(raw))
	for k, v := range raw {
		var prop kbapi.KibanaHTTPAPIsMlSingleMetricViewer_SelectedEntities_AdditionalProperties
		switch {
		case typeutils.IsKnown(v.StringValue):
			if err := prop.FromKibanaHTTPAPIsMlSingleMetricViewerSelectedEntities0(v.StringValue.ValueString()); err != nil {
				diags.AddError("Invalid ML single metric viewer configuration", err.Error())
				return nil
			}
		case typeutils.IsKnown(v.NumericValue):
			f64, acc := v.NumericValue.ValueBigFloat().Float64()
			if acc == big.Above || acc == big.Below {
				diags.AddError(
					"Invalid ML single metric viewer configuration",
					fmt.Sprintf("selected_entities[%q] numeric_value is out of float64 range.", k),
				)
				return nil
			}
			if math.IsNaN(f64) || f64 > math.MaxFloat32 || f64 < -math.MaxFloat32 {
				diags.AddError(
					"Invalid ML single metric viewer configuration",
					fmt.Sprintf("selected_entities[%q] numeric_value is out of float32 range.", k),
				)
				return nil
			}
			f := float32(f64)
			if err := prop.FromKibanaHTTPAPIsMlSingleMetricViewerSelectedEntities1(f); err != nil {
				diags.AddError("Invalid ML single metric viewer configuration", err.Error())
				return nil
			}
		default:
			diags.AddError(
				"Invalid ML single metric viewer configuration",
				fmt.Sprintf("selected_entities[%q] must set exactly one of `string_value` or `numeric_value`.", k),
			)
			return nil
		}
		out[k] = prop
	}
	if len(out) == 0 {
		return nil
	}
	return &out
}

func mlSingleMetricViewerSelectedEntitiesFromAPI(
	ctx context.Context,
	api *map[string]kbapi.KibanaHTTPAPIsMlSingleMetricViewer_SelectedEntities_AdditionalProperties,
	diags *diag.Diagnostics,
) types.Map {
	if api == nil || len(*api) == 0 {
		return types.MapNull(mlSingleMetricViewerEntityObjectType)
	}
	elems := make(map[string]models.MlSingleMetricViewerEntityModel, len(*api))
	for k, v := range *api {
		entity := models.MlSingleMetricViewerEntityModel{
			StringValue:  types.StringNull(),
			NumericValue: types.NumberNull(),
		}
		if s, err := v.AsKibanaHTTPAPIsMlSingleMetricViewerSelectedEntities0(); err == nil {
			entity.StringValue = types.StringValue(s)
		} else if n, err := v.AsKibanaHTTPAPIsMlSingleMetricViewerSelectedEntities1(); err == nil {
			entity.NumericValue = types.NumberValue(big.NewFloat(float64(n)))
		} else {
			diags.AddError(
				"Invalid ML single metric viewer panel configuration on read",
				fmt.Sprintf("selected_entities[%q] has an unsupported value type.", k),
			)
		}
		elems[k] = entity
	}
	return typeutils.MapValueFrom(ctx, elems, mlSingleMetricViewerEntityObjectType, path.Empty(), diags)
}

func mlSingleMetricViewerPreserveNullIntentFromPrior(prior, existing *models.MlSingleMetricViewerConfigModel) {
	if prior == nil || existing == nil {
		return
	}
	panelkit.NullPreserveFloat32FromPrior(prior.SelectedDetectorIndex, &existing.SelectedDetectorIndex)
	panelkit.NullPreserveStringFromPrior(prior.ForecastID, &existing.ForecastID)
	panelkit.NullPreserveStringFromPrior(prior.FunctionDescription, &existing.FunctionDescription)
	panelkit.NullPreserveMapFromPrior(prior.SelectedEntities, &existing.SelectedEntities)
	panelkit.NullPreserveBaseFromPrior(prior.Title, prior.Description, prior.HideTitle, prior.HideBorder,
		&existing.Title, &existing.Description, &existing.HideTitle, &existing.HideBorder)
	existing.TimeRange = panelkit.PreserveTimeRangeNullIntentFromPrior(prior.TimeRange, existing.TimeRange)
}
