package contracttest

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/iface"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

var jsonDefaultsNoOp = func(m map[string]any) map[string]any { return m }

type synthStatsHandler struct{}

func (synthStatsHandler) PanelType() string { return "synthetics_stats_overview" }

func (synthStatsHandler) SchemaAttribute() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"title": schema.StringAttribute{Optional: true},
		},
	}
}

func (synthStatsHandler) FromAPI(_ context.Context, pm, prior *models.PanelModel, item kbapi.DashboardPanelItem) diag.Diagnostics {
	apiPanel, err := item.AsKbnDashboardPanelTypeSyntheticsStatsOverview()
	if err != nil {
		var d diag.Diagnostics
		d.AddError("panel union", err.Error())
		return d
	}

	pm.Type = types.StringValue("synthetics_stats_overview")
	pm.Grid = panelkit.GridFromAPI(apiPanel.Grid.X, apiPanel.Grid.Y, apiPanel.Grid.W, apiPanel.Grid.H)
	pm.ID = panelkit.IDFromAPI(apiPanel.Id)
	pm.ConfigJSON = customtypes.NewJSONWithDefaultsNull(jsonDefaultsNoOp)

	cfg := apiPanel.Config
	if prior == nil {
		if cfg.Title == nil {
			return nil
		}
		pm.SyntheticsStatsOverviewConfig = &models.SyntheticsStatsOverviewConfigModel{
			Title: types.StringPointerValue(cfg.Title),
		}
		return nil
	}

	existing := pm.SyntheticsStatsOverviewConfig
	if existing == nil {
		return nil
	}
	if typeutils.IsKnown(existing.Title) {
		existing.Title = types.StringPointerValue(cfg.Title)
	}
	pm.SyntheticsStatsOverviewConfig = existing
	return nil
}

func (synthStatsHandler) ToAPI(pm models.PanelModel, dashboard *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	_ = dashboard
	grid := panelkit.GridToAPI(pm.Grid)
	id := panelkit.IDToAPI(pm.ID)
	sso := kbapi.KbnDashboardPanelTypeSyntheticsStatsOverview{
		Grid: grid,
		Id:   id,
		Type: kbapi.SyntheticsStatsOverview,
	}
	if cfg := pm.SyntheticsStatsOverviewConfig; cfg != nil && typeutils.IsKnown(cfg.Title) {
		t := cfg.Title.ValueString()
		sso.Config.Title = &t
	}

	var out kbapi.DashboardPanelItem
	if err := out.FromKbnDashboardPanelTypeSyntheticsStatsOverview(sso); err != nil {
		var d diag.Diagnostics
		d.AddError("ToAPI", err.Error())
		return kbapi.DashboardPanelItem{}, d
	}
	return out, nil
}

func (synthStatsHandler) ValidatePanelConfig(_ context.Context, panelType string, _ map[string]attr.Value, _ path.Path) diag.Diagnostics {
	if panelType != "synthetics_stats_overview" {
		return nil
	}
	return nil
}

func (synthStatsHandler) AlignStateFromPlan(_ context.Context, _, _ *models.PanelModel) {}

func (synthStatsHandler) ClassifyJSON(_ map[string]any) bool { return false }

func (synthStatsHandler) PopulateJSONDefaults(config map[string]any) map[string]any { return config }

func (synthStatsHandler) PinnedHandler() iface.PinnedHandler { return nil }

type brokenContractHandler struct{}

func (brokenContractHandler) PanelType() string { return "contracttest_broken_stub" }

func (brokenContractHandler) SchemaAttribute() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"required_field": schema.StringAttribute{Required: true},
		},
	}
}

func (brokenContractHandler) FromAPI(_ context.Context, pm, _ *models.PanelModel, item kbapi.DashboardPanelItem) diag.Diagnostics {
	disc, _ := item.Discriminator()
	pm.Type = types.StringValue(disc)
	return nil
}

func (brokenContractHandler) ToAPI(_ models.PanelModel, _ *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	return kbapi.DashboardPanelItem{}, nil
}

func (brokenContractHandler) ValidatePanelConfig(_ context.Context, _ string, _ map[string]attr.Value, _ path.Path) diag.Diagnostics {
	return nil
}

func (brokenContractHandler) AlignStateFromPlan(_ context.Context, _, _ *models.PanelModel) {}

func (brokenContractHandler) ClassifyJSON(_ map[string]any) bool { return false }

func (brokenContractHandler) PopulateJSONDefaults(config map[string]any) map[string]any {
	return config
}

func (brokenContractHandler) PinnedHandler() iface.PinnedHandler { return nil }

func TestContractHarness_syntheticsStatsOverviewSmoke(t *testing.T) {
	const fixture = `{
  "type": "synthetics_stats_overview",
  "grid": { "x": 0, "y": 0, "w": 24, "h": 8 },
  "id": "synth-contract",
  "config": { "title": "Hello synthetics" }
}`
	Run(t, synthStatsHandler{}, Config{FullAPIResponse: fixture})
	require.False(t, t.Failed(), "expected harness to pass for minimal synthetics handler")
}

func TestContractHarness_brokenHandlerReportsFailures(t *testing.T) {
	const fixture = `{
  "type": "contracttest_broken_stub",
  "grid": { "x": 0, "y": 0, "w": 6, "h": 4 },
  "config": { "requiredField": "ok" }
}`
	ctx := context.Background()
	msgs := runChecks(ctx, brokenContractHandler{}, Config{FullAPIResponse: fixture})
	require.NotEmpty(t, msgs)
}
