// Package iface defines the Terraform dashboard panel Handler contract used by registry-driven routing.
package iface

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Handler binds a dashboard panel discriminator to schema, conversions, validation, and pinning.
type Handler interface {
	PanelType() string
	SchemaAttribute() schema.Attribute
	FromAPI(ctx context.Context, pm, prior *models.PanelModel, item kbapi.DashboardPanelItem) diag.Diagnostics
	ToAPI(pm models.PanelModel, dashboard *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics)
	ValidatePanelConfig(ctx context.Context, panelType string, attrs map[string]attr.Value, attrPath path.Path) diag.Diagnostics
	AlignStateFromPlan(ctx context.Context, plan, state *models.PanelModel)
	ClassifyJSON(config map[string]any) bool
	PopulateJSONDefaults(config map[string]any) map[string]any
	PinnedHandler() PinnedHandler
}

// PinnedHandler converts control panels pinned in the dashboard control bar (optional).
type PinnedHandler interface {
	FromAPI(ctx context.Context, prior *models.PinnedPanelModel, raw kbapi.DashboardPinnedPanels_Item) (models.PinnedPanelModel, diag.Diagnostics)
	ToAPI(ppm models.PinnedPanelModel) (kbapi.DashboardPinnedPanels_Item, diag.Diagnostics)
}
