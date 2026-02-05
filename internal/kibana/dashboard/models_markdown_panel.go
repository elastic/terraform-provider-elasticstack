package dashboard

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type markdownConfigModel struct {
	Content     types.String `tfsdk:"content"`
	Description types.String `tfsdk:"description"`
	HideTitle   types.Bool   `tfsdk:"hide_title"`
	Title       types.String `tfsdk:"title"`
}

type markdownPanelConfigConverter struct{}

func (c markdownPanelConfigConverter) handlesAPIPanelConfig(panelType string, _ kbapi.DashboardPanelItem_Config) bool {
	return panelType == "DASHBOARD_MARKDOWN"
}

func (c markdownPanelConfigConverter) handlesTFPanelConfig(pm panelModel) bool {
	return pm.MarkdownConfig != nil
}

func (c markdownPanelConfigConverter) populateFromAPIPanel(_ context.Context, pm *panelModel, config kbapi.DashboardPanelItem_Config) diag.Diagnostics {
	config0, err := config.AsDashboardPanelItemConfig0()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	pm.MarkdownConfig = &markdownConfigModel{
		Content:     types.StringValue(config0.Content),
		Description: types.StringPointerValue(config0.Description),
		HideTitle:   types.BoolPointerValue(config0.HideTitle),
		Title:       types.StringPointerValue(config0.Title),
	}

	return nil
}

func (c markdownPanelConfigConverter) mapPanelToAPI(pm panelModel, apiConfig *kbapi.DashboardPanelItem_Config) diag.Diagnostics {
	config0 := kbapi.DashboardPanelItemConfig0{
		Content: pm.MarkdownConfig.Content.ValueString(),
	}
	if utils.IsKnown(pm.MarkdownConfig.Description) {
		config0.Description = utils.Pointer(pm.MarkdownConfig.Description.ValueString())
	}
	if utils.IsKnown(pm.MarkdownConfig.HideTitle) {
		config0.HideTitle = utils.Pointer(pm.MarkdownConfig.HideTitle.ValueBool())
	}
	if utils.IsKnown(pm.MarkdownConfig.Title) {
		config0.Title = utils.Pointer(pm.MarkdownConfig.Title.ValueString())
	}

	var diags diag.Diagnostics
	if err := apiConfig.FromDashboardPanelItemConfig0(config0); err != nil {
		diags.AddError("Failed to marshal panel config", err.Error())
	}

	return diags
}
