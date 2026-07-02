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
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/iface"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type Handler struct{}

func (Handler) PanelType() string                  { return panelType }
func (Handler) SchemaAttribute() schema.Attribute  { return SchemaAttribute() }
func (Handler) ClassifyJSON(_ map[string]any) bool { return false }
func (Handler) PopulateJSONDefaults(config map[string]any) map[string]any {
	return config
}

func (Handler) PinnedHandler() iface.PinnedHandler { return newPinnedHandler() }

func (Handler) AlignStateFromPlan(ctx context.Context, plan, state *models.PanelModel) {
	_, _, _ = ctx, plan, state
}

func (Handler) FromAPI(ctx context.Context, pm, prior *models.PanelModel, item kbapi.DashboardPanelItem) diag.Diagnostics {
	return panelkit.SimpleFromAPI(ctx, pm, prior,
		item.AsKibanaHTTPAPIsKbnDashboardPanelTypeRangeSliderControl,
		func(p kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeRangeSliderControl) (kbapi.KibanaHTTPAPIsKbnDashboardPanelGrid, *string) {
			return p.Grid, p.Id
		},
		func(pm *models.PanelModel, prior *models.PanelModel, p kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeRangeSliderControl) diag.Diagnostics {
			return PopulateFromAPI(ctx, pm, prior, &p)
		},
	)
}

func (Handler) ToAPI(pm models.PanelModel, dashboard *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	var diags diag.Diagnostics
	_ = dashboard
	if pm.RangeSliderControlConfig == nil {
		diags.AddError(
			"Missing range slider control panel configuration",
			"Range slider control panels require `range_slider_control_config`.",
		)
		return kbapi.DashboardPanelItem{}, diags
	}
	grid := panelkit.GridToAPI(pm.Grid)
	id := panelkit.IDToAPI(pm.ID)
	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeRangeSliderControl{
		Grid: grid,
		Id:   id,
	}
	diags.Append(BuildConfig(pm, &panel)...)
	if diags.HasError() {
		return kbapi.DashboardPanelItem{}, diags
	}
	var panelItem kbapi.DashboardPanelItem
	if err := panelItem.FromKibanaHTTPAPIsKbnDashboardPanelTypeRangeSliderControl(panel); err != nil {
		diags.AddError("Failed to create range slider control panel", err.Error())
		return panelItem, diags
	}
	return panelItem, nil
}

// ValidatePanelConfig is a no-op: all range_slider_control_config validation is enforced by
// schema-level validators. `by_field.data_view_id` / `by_field.field_name` and
// `by_esql.esql_query` / `by_esql.values_source` are natively `Required: true` inside their
// respective nested blocks, and the by_field/by_esql union itself is enforced by
// ExactlyOneOfNestedAttrsValidator + objectvalidator.ConflictsWith in schema.go.
//
// Note: panelkit.ValidateDataViewFieldName (previously used here) assumed data_view_id/field_name
// lived directly under range_slider_control_config. Post-restructure they live two levels deep
// under range_slider_control_config.by_field, a shape ResolvePanelAttrsShape does not resolve, so
// calling it here would spuriously error "data_view_id is required" any time by_field is set.
func (Handler) ValidatePanelConfig(_ context.Context, _ map[string]attr.Value, _ path.Path) diag.Diagnostics {
	return nil
}
