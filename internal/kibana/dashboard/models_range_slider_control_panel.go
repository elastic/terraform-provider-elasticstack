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

package dashboard

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type rangeSliderControlConfigModel struct {
	Title             types.String  `tfsdk:"title"`
	DataViewID        types.String  `tfsdk:"data_view_id"`
	FieldName         types.String  `tfsdk:"field_name"`
	UseGlobalFilters  types.Bool    `tfsdk:"use_global_filters"`
	IgnoreValidations types.Bool    `tfsdk:"ignore_validations"`
	Value             types.List    `tfsdk:"value"`
	Step              types.Float64 `tfsdk:"step"`
}

// populateRangeSliderControlFromAPI reads back a range slider control config from the API
// response and updates the panel model. Null-preservation semantics apply: if a field is
// null in the existing TF state, we do not overwrite it with a Kibana-returned value. If
// there is no existing config block, and Kibana returns an empty/absent config, we leave
// RangeSliderControlConfig as nil.
//
// tfPanel is the prior TF state/plan panel, or nil on import. When nil, the function
// populates all API-returned fields unconditionally (no prior intent to preserve).
func populateRangeSliderControlFromAPI(ctx context.Context, pm *panelModel, tfPanel *panelModel, apiConfig kbapi.KbnDashboardPanelRangeSliderControl_Config) {
	existing := pm.RangeSliderControlConfig

	// On import (tfPanel == nil) there is no prior intent. Populate from API unconditionally.
	if tfPanel == nil {
		pm.RangeSliderControlConfig = &rangeSliderControlConfigModel{
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
			existing.Step = types.Float64Value(float64(*apiConfig.Step))
		}
		return
	}

	// If the existing state has no config block, preserve nil intent.
	if existing == nil {
		return
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
		existing.Step = types.Float64Value(float64(*apiConfig.Step))
	}
}

// buildRangeSliderControlConfig writes the TF model fields into the API panel struct.
func buildRangeSliderControlConfig(ctx context.Context, pm panelModel, rsPanel *kbapi.KbnDashboardPanelRangeSliderControl) {
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
		var elems []string
		_ = cfg.Value.ElementsAs(ctx, &elems, false)
		rsPanel.Config.Value = &elems
	}
	if typeutils.IsKnown(cfg.Step) {
		v := float32(cfg.Step.ValueFloat64())
		rsPanel.Config.Step = &v
	}
}
