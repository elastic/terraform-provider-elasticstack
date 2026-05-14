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

package optionslist

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/iface"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
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

func (Handler) AlignStateFromPlan(ctx context.Context, plan, state *models.PanelModel) {
	_, _, _ = ctx, plan, state
}

func (Handler) FromAPI(ctx context.Context, pm, prior *models.PanelModel, item kbapi.DashboardPanelItem) diag.Diagnostics {
	olPanel, err := item.AsKbnDashboardPanelTypeOptionsListControl()
	if err != nil {
		var d diag.Diagnostics
		d.AddError("Dashboard panel decode", err.Error())
		return d
	}

	pm.Grid = panelkit.GridFromAPI(olPanel.Grid.X, olPanel.Grid.Y, olPanel.Grid.W, olPanel.Grid.H)
	pm.ID = panelkit.IDFromAPI(olPanel.Id)
	if configBytes, err := json.Marshal(olPanel.Config); err == nil {
		pm.ConfigJSON = customtypes.NewJSONWithDefaultsValue(string(configBytes), jsonDefaultsFunc())
	}
	PopulateFromAPI(pm, prior, &olPanel)
	_ = ctx
	return nil
}

func (Handler) ToAPI(pm models.PanelModel, dashboard *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	_ = dashboard
	grid := panelkit.GridToAPI(pm.Grid)
	id := panelkit.IDToAPI(pm.ID)
	panel := kbapi.KbnDashboardPanelTypeOptionsListControl{
		Grid: grid,
		Id:   id,
	}
	BuildConfig(pm, &panel)
	var panelItem kbapi.DashboardPanelItem
	if err := panelItem.FromKbnDashboardPanelTypeOptionsListControl(panel); err != nil {
		var diags diag.Diagnostics
		diags.AddError("Failed to create options list control panel", err.Error())
		return panelItem, diags
	}
	return panelItem, nil
}

func optionsListAttrsShape(attrs map[string]attr.Value) (flat bool, obj types.Object, shaped bool) {
	if attrs == nil {
		return false, types.Object{}, false
	}
	if _, dv := attrs["data_view_id"]; dv {
		if _, fn := attrs["field_name"]; fn {
			return true, types.Object{}, true
		}
		return false, types.Object{}, false
	}
	raw, nested := attrs[panelConfigAttrsKeyPrefix]
	if !nested {
		return false, types.Object{}, false
	}
	obj, ok := raw.(types.Object)
	if !ok {
		return false, types.Object{}, false
	}
	return false, obj, true
}

// ValidatePanelConfig enforces presence of DataViewID and FieldName for options_list panels.
func (Handler) ValidatePanelConfig(_ context.Context, attrs map[string]attr.Value, attrPath path.Path) diag.Diagnostics {
	var out diag.Diagnostics
	flat, obj, shaped := optionsListAttrsShape(attrs)
	if !shaped {
		return out
	}

	cfgPath := attrPath
	var dataViewAttr, fieldNameAttr attr.Value
	switch {
	case flat:
		dataViewAttr, fieldNameAttr = attrs["data_view_id"], attrs["field_name"]
	default:
		at := obj.Attributes()
		cfgPath = attrPath.AtName(panelConfigAttrsKeyPrefix)
		dataViewAttr, fieldNameAttr = at["data_view_id"], at["field_name"]
	}

	writeErr := func(field, msg string) {
		out.AddAttributeError(cfgPath.AtName(field), "Invalid options list control configuration", msg)
	}
	if deferDV, missDV := panelkit.StringAttrDeferOrMissing(dataViewAttr); !deferDV && missDV {
		writeErr("data_view_id", "`data_view_id` is required.")
	}
	if deferFN, missFN := panelkit.StringAttrDeferOrMissing(fieldNameAttr); !deferFN && missFN {
		writeErr("field_name", "`field_name` is required.")
	}
	return out
}
