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

package rangeslider

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// PopulateFromAPI reads back a range slider control config from the API
// response and updates the panel model. Null-preservation semantics apply: if a field is
// null in the existing TF state, we do not overwrite it with a Kibana-returned value. If
// there is no existing config block, and Kibana returns an empty/absent config, we leave
// RangeSliderControlConfig as nil.
//
// tfPanel is the prior TF state/plan panel, or nil on import. When nil, the function
// populates all API-returned fields unconditionally (no prior intent to preserve).
func PopulateFromAPI(ctx context.Context, pm *models.PanelModel, tfPanel *models.PanelModel, rs *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeRangeSliderControl) {
	if rs == nil {
		return
	}
	apiConfig := &rs.Config
	existing := pm.RangeSliderControlConfig

	// On import (tfPanel == nil) there is no prior intent. Populate from API unconditionally.
	if tfPanel == nil {
		pm.RangeSliderControlConfig = &models.RangeSliderControlConfigModel{
			DataViewID: types.StringValue(apiConfig.DataViewId),
			FieldName:  types.StringValue(apiConfig.FieldName),
		}
		existing = pm.RangeSliderControlConfig
		if apiConfig.Title != nil {
			existing.Title = types.StringValue(*apiConfig.Title)
		}
		if apiConfig.UseGlobalFilters != nil {
			existing.UseGlobalFilters = types.BoolValue(*apiConfig.UseGlobalFilters)
		}
		if apiConfig.IgnoreValidations != nil {
			existing.IgnoreValidations = types.BoolValue(*apiConfig.IgnoreValidations)
		}
		if apiConfig.Value != nil {
			v, _ := types.ListValueFrom(ctx, types.StringType, *apiConfig.Value)
			existing.Value = v
		}
		if apiConfig.Step != nil {
			existing.Step = types.Float32Value(*apiConfig.Step)
		}
		return
	}

	if existing == nil {
		if tfPanel == nil || tfPanel.RangeSliderControlConfig == nil {
			return
		}
		pm.RangeSliderControlConfig = &models.RangeSliderControlConfigModel{
			DataViewID: types.StringValue(apiConfig.DataViewId),
			FieldName:  types.StringValue(apiConfig.FieldName),
		}
		existing = pm.RangeSliderControlConfig
		if apiConfig.Title != nil {
			existing.Title = types.StringValue(*apiConfig.Title)
		}
		if apiConfig.UseGlobalFilters != nil {
			existing.UseGlobalFilters = types.BoolValue(*apiConfig.UseGlobalFilters)
		}
		if apiConfig.IgnoreValidations != nil {
			existing.IgnoreValidations = types.BoolValue(*apiConfig.IgnoreValidations)
		}
		if apiConfig.Value != nil {
			v, _ := types.ListValueFrom(ctx, types.StringType, *apiConfig.Value)
			existing.Value = v
		}
		if apiConfig.Step != nil {
			existing.Step = types.Float32Value(*apiConfig.Step)
		}
	}

	// Block exists in state — always update required fields, update optional only when non-null.
	existing.DataViewID = types.StringValue(apiConfig.DataViewId)
	existing.FieldName = types.StringValue(apiConfig.FieldName)

	if typeutils.IsKnown(existing.Title) && apiConfig.Title != nil {
		existing.Title = types.StringValue(*apiConfig.Title)
	}
	if typeutils.IsKnown(existing.UseGlobalFilters) && apiConfig.UseGlobalFilters != nil {
		existing.UseGlobalFilters = types.BoolValue(*apiConfig.UseGlobalFilters)
	}
	if typeutils.IsKnown(existing.IgnoreValidations) && apiConfig.IgnoreValidations != nil {
		existing.IgnoreValidations = types.BoolValue(*apiConfig.IgnoreValidations)
	}
	if typeutils.IsKnown(existing.Value) && apiConfig.Value != nil {
		v, _ := types.ListValueFrom(ctx, types.StringType, *apiConfig.Value)
		existing.Value = v
	}
	if typeutils.IsKnown(existing.Step) && apiConfig.Step != nil {
		existing.Step = types.Float32Value(*apiConfig.Step)
	}

	if tfPanel != nil && tfPanel.RangeSliderControlConfig != nil {
		rangeSliderPreserveNullIntentFromPrior(tfPanel.RangeSliderControlConfig, existing)
	}
}

func rangeSliderPreserveNullIntentFromPrior(prior, existing *models.RangeSliderControlConfigModel) {
	if prior == nil || existing == nil {
		return
	}
	if !typeutils.IsKnown(prior.Title) {
		existing.Title = types.StringNull()
	}
	if !typeutils.IsKnown(prior.UseGlobalFilters) {
		existing.UseGlobalFilters = types.BoolNull()
	}
	if !typeutils.IsKnown(prior.IgnoreValidations) {
		existing.IgnoreValidations = types.BoolNull()
	}
	if !typeutils.IsKnown(prior.Value) {
		existing.Value = types.ListNull(types.StringType)
	}
	if !typeutils.IsKnown(prior.Step) {
		existing.Step = types.Float32Null()
	}
}

// BuildConfig writes TF model fields into the API panel payload.
func BuildConfig(pm models.PanelModel, rsPanel *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeRangeSliderControl) {
	cfg := pm.RangeSliderControlConfig
	if cfg == nil {
		return
	}
	rsPanel.Config.DataViewId = cfg.DataViewID.ValueString()
	rsPanel.Config.FieldName = cfg.FieldName.ValueString()

	if typeutils.IsKnown(cfg.Title) {
		rsPanel.Config.Title = cfg.Title.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.UseGlobalFilters) {
		rsPanel.Config.UseGlobalFilters = cfg.UseGlobalFilters.ValueBoolPointer()
	}
	if typeutils.IsKnown(cfg.IgnoreValidations) {
		rsPanel.Config.IgnoreValidations = cfg.IgnoreValidations.ValueBoolPointer()
	}
	if typeutils.IsKnown(cfg.Value) {
		raw := cfg.Value.Elements()
		elems := make([]string, len(raw))
		for i, e := range raw {
			elems[i] = e.(types.String).ValueString()
		}
		rsPanel.Config.Value = &elems
	}
	if typeutils.IsKnown(cfg.Step) {
		v := cfg.Step.ValueFloat32()
		rsPanel.Config.Step = &v
	}
}
