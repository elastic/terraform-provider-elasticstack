package contracttest

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/iface"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
)

func appendReflectIssues(ctx context.Context, handler iface.Handler, fixture string, issues *[]string) {
	block := handler.PanelType() + "_config"
	if !panelkit.HasPanelConfigBlock(block) {
		return
	}

	item0, err := ParseDashboardPanel(fixture)
	if err != nil {
		*issues = append(*issues, fmt.Sprintf("[Reflect] parse: %v", err))
		return
	}

	var pm models.PanelModel
	if diags := handler.FromAPI(ctx, &pm, nil, item0); diags.HasError() {
		*issues = append(*issues, fmt.Sprintf("[Reflect] FromAPI: %s", summarizeDiags(diags)))
		return
	}

	if !panelkit.HasConfig(&pm, block) {
		*issues = append(*issues, fmt.Sprintf("[Reflect] expected HasConfig(%s) after FromAPI", block))
		return
	}
	panelkit.ClearConfig(&pm, block)
	if panelkit.HasConfig(&pm, block) {
		*issues = append(*issues, fmt.Sprintf("[Reflect] expected HasConfig(%s) false after ClearConfig", block))
		return
	}
}
