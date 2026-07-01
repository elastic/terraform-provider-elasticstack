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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Handler implements iface.Handler for the ml_single_metric_viewer dashboard panel discriminator.
type Handler struct {
	panelkit.NoopHandlerBase
}

func (Handler) PanelType() string                 { return panelType }
func (Handler) SchemaAttribute() schema.Attribute { return SchemaAttribute() }

func (Handler) FromAPI(ctx context.Context, pm, prior *models.PanelModel, item kbapi.DashboardPanelItem) diag.Diagnostics {
	return panelkit.SimpleFromAPI(ctx, pm, prior,
		item.AsKibanaHTTPAPIsKbnDashboardPanelTypeMlSingleMetricViewer,
		func(p kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeMlSingleMetricViewer) (kbapi.KibanaHTTPAPIsKbnDashboardPanelGrid, *string) {
			return p.Grid, p.Id
		},
		func(pm, prior *models.PanelModel, p kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeMlSingleMetricViewer) diag.Diagnostics {
			return PopulateFromAPI(ctx, pm, prior, p.Config)
		},
	)
}

func (Handler) ToAPI(pm models.PanelModel, dashboard *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	_ = dashboard
	return panelkit.SimpleToAPI(pm,
		func(grid kbapi.KibanaHTTPAPIsKbnDashboardPanelGrid, id *string) (kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeMlSingleMetricViewer, diag.Diagnostics) {
			if diags := panelkit.RejectConfigJSON(pm, panelType); diags.HasError() {
				return kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeMlSingleMetricViewer{}, diags
			}
			panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeMlSingleMetricViewer{Grid: grid, Id: id, Type: kbapi.MlSingleMetricViewer}
			return panel, BuildConfig(context.Background(), pm, &panel)
		},
		func(item *kbapi.DashboardPanelItem, panel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeMlSingleMetricViewer) error {
			return item.FromKibanaHTTPAPIsKbnDashboardPanelTypeMlSingleMetricViewer(panel)
		},
		"Failed to create ML single metric viewer panel",
	)
}

func (Handler) ValidatePanelConfig(_ context.Context, attrs map[string]attr.Value, attrPath path.Path) diag.Diagnostics {
	var out diag.Diagnostics
	flat, obj, cfgPath, skip, diags := panelkit.ResolveConfigBlock(attrs, attrPath, panelConfigBlock,
		"Missing ML single metric viewer panel configuration",
		"ML single metric viewer panels require `ml_single_metric_viewer_config`.",
		"job_ids")
	out.Append(diags...)
	if skip {
		return out
	}

	var jobIDsVal attr.Value
	if flat {
		jobIDsVal = attrs["job_ids"]
	} else {
		jobIDsVal = obj.Attributes()["job_ids"]
	}
	switch {
	case jobIDsVal == nil || jobIDsVal.IsUnknown():
	case jobIDsVal.IsNull():
		out.AddAttributeError(cfgPath.AtName("job_ids"), "Invalid ML single metric viewer configuration", "`job_ids` is required.")
	default:
		if list, ok := jobIDsVal.(types.List); ok {
			switch {
			case list.IsNull() || list.IsUnknown():
			case len(list.Elements()) == 0:
				out.AddAttributeError(cfgPath.AtName("job_ids"), "Invalid ML single metric viewer configuration", "`job_ids` must contain exactly one entry.")
			case len(list.Elements()) > 1:
				out.AddAttributeError(cfgPath.AtName("job_ids"), "Invalid ML single metric viewer configuration", "`job_ids` must contain exactly one entry.")
			}
		}
	}

	return out
}
