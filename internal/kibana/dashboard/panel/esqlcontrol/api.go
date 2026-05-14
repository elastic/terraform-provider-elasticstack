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

package esqlcontrol

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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const panelConfigAttrsKeyPrefix = panelType + "_config"

type Handler struct{}

func (Handler) PanelType() string                  { return panelType }
func (Handler) SchemaAttribute() schema.Attribute  { return SchemaAttribute() }
func (Handler) ClassifyJSON(_ map[string]any) bool { return false }
func (Handler) PopulateJSONDefaults(config map[string]any) map[string]any {
	return config
}

func (Handler) PinnedHandler() iface.PinnedHandler { return pinnedHandler{} }

func (Handler) AlignStateFromPlan(_ context.Context, plan, state *models.PanelModel) {
	AlignEsqlPanels(plan, state)
}

func (Handler) FromAPI(ctx context.Context, pm, prior *models.PanelModel, item kbapi.DashboardPanelItem) diag.Diagnostics {
	esqlPanel, err := item.AsKbnDashboardPanelTypeEsqlControl()
	if err != nil {
		var d diag.Diagnostics
		d.AddError("Dashboard panel decode", err.Error())
		return d
	}

	pm.Grid = panelkit.GridFromAPI(esqlPanel.Grid.X, esqlPanel.Grid.Y, esqlPanel.Grid.W, esqlPanel.Grid.H)
	pm.ID = panelkit.IDFromAPI(esqlPanel.Id)
	pm.ConfigJSON = panelConfigJSONNull()
	PopulateFromAPI(pm, prior, esqlPanel.Config)
	_ = ctx
	return nil
}

func (Handler) ToAPI(pm models.PanelModel, dashboard *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	_ = dashboard
	grid := panelkit.GridToAPI(pm.Grid)
	id := panelkit.IDToAPI(pm.ID)
	panel := kbapi.KbnDashboardPanelTypeEsqlControl{
		Grid: grid,
		Id:   id,
	}
	diags := BuildConfig(pm, &panel)
	var panelItem kbapi.DashboardPanelItem
	if diags.HasError() {
		return panelItem, diags
	}
	if err := panelItem.FromKbnDashboardPanelTypeEsqlControl(panel); err != nil {
		diags.AddError("Failed to create esql control panel", err.Error())
		return panelItem, diags
	}
	return panelItem, diags
}

func esqlAttrsShape(attrs map[string]attr.Value) (flat bool, nested types.Object, shaped bool) {
	if attrs == nil {
		return false, types.Object{}, false
	}
	if _, vn := attrs["variable_name"]; vn {
		if _, eq := attrs["esql_query"]; eq {
			return true, types.Object{}, true
		}
		return false, types.Object{}, false
	}
	raw, ok := attrs[panelConfigAttrsKeyPrefix]
	if !ok || raw == nil {
		return false, types.Object{}, false
	}
	obj, ok := raw.(types.Object)
	if !ok {
		return false, types.Object{}, false
	}
	return false, obj, true
}

// ValidatePanelConfig checks required shallow leaves for typed esql configs (covers contracttest flattened attrs).
func (Handler) ValidatePanelConfig(_ context.Context, attrs map[string]attr.Value, attrPath path.Path) diag.Diagnostics {
	var out diag.Diagnostics
	flat, nested, shaped := esqlAttrsShape(attrs)
	if !shaped {
		return out
	}

	attrRead := attrs
	cfgPath := attrPath
	if !flat {
		cfgPath = attrPath.AtName(panelConfigAttrsKeyPrefix)
		attrRead = nested.Attributes()
	}

	requiredString := func(name, label string) {
		v := attrRead[name]
		deferVal, missing := panelkit.StringAttrDeferOrMissing(v)
		if deferVal {
			return
		}
		if missing {
			out.AddAttributeError(cfgPath.AtName(name), "Invalid ES|QL control configuration", "`"+label+"` is required.")
		}
	}

	requiredString("variable_name", "variable_name")
	requiredString("variable_type", "variable_type")
	requiredString("esql_query", "esql_query")
	requiredString("control_type", "control_type")

	sel := attrRead["selected_options"]
	switch {
	case sel == nil:
		out.AddAttributeError(cfgPath.AtName("selected_options"), "Invalid ES|QL control configuration", "`selected_options` is required.")
	case sel.IsUnknown():
		// Defer until list elements are known (cross-resource graphs).
	default:
		if sel.IsNull() {
			out.AddAttributeError(cfgPath.AtName("selected_options"), "Invalid ES|QL control configuration", "`selected_options` is required.")
		}
	}
	return out
}
