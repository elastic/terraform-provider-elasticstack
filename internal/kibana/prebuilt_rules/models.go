package prebuilt_rules

import (
	"context"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type prebuiltRuleModel struct {
	ID                    types.String `tfsdk:"id"`
	SpaceID               types.String `tfsdk:"space_id"`
	RulesInstalled        types.Int64  `tfsdk:"rules_installed"`
	RulesNotInstalled     types.Int64  `tfsdk:"rules_not_installed"`
	RulesNotUpdated       types.Int64  `tfsdk:"rules_not_updated"`
	TimelinesInstalled    types.Int64  `tfsdk:"timelines_installed"`
	TimelinesNotInstalled types.Int64  `tfsdk:"timelines_not_installed"`
	TimelinesNotUpdated   types.Int64  `tfsdk:"timelines_not_updated"`
}

func (model *prebuiltRuleModel) populateFromStatus(ctx context.Context, status *kbapi.ReadPrebuiltRulesAndTimelinesStatusResponse) {
	model.RulesInstalled = types.Int64Value(int64(status.JSON200.RulesInstalled))
	model.RulesNotInstalled = types.Int64Value(int64(status.JSON200.RulesNotInstalled))
	model.RulesNotUpdated = types.Int64Value(int64(status.JSON200.RulesNotUpdated))
	model.TimelinesInstalled = types.Int64Value(int64(status.JSON200.TimelinesInstalled))
	model.TimelinesNotInstalled = types.Int64Value(int64(status.JSON200.TimelinesNotInstalled))
	model.TimelinesNotUpdated = types.Int64Value(int64(status.JSON200.TimelinesNotUpdated))
}

func getPrebuiltRulesStatus(ctx context.Context, client *kibana_oapi.Client, spaceID string) (*kbapi.ReadPrebuiltRulesAndTimelinesStatusResponse, diag.Diagnostics) {
	resp, err := client.API.ReadPrebuiltRulesAndTimelinesStatusWithResponse(ctx, func(ctx context.Context, req *http.Request) error {
		if spaceID != "default" {
			req.Header.Set("kbn-space-id", spaceID)
		}
		return nil
	})

	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}

	if resp.StatusCode() != 200 {
		return nil, utils.FrameworkDiagFromError(fmt.Errorf("failed to get prebuilt rules status: %s", resp.Status()))
	}

	return resp, nil
}

func installPrebuiltRules(ctx context.Context, client *kibana_oapi.Client, spaceID string) diag.Diagnostics {
	resp, err := client.API.InstallPrebuiltRulesAndTimelinesWithResponse(ctx, func(ctx context.Context, req *http.Request) error {
		if spaceID != "default" {
			req.Header.Set("kbn-space-id", spaceID)
		}
		return nil
	})

	if err != nil {
		return utils.FrameworkDiagFromError(err)
	}

	if resp.StatusCode() != 200 {
		return utils.FrameworkDiagFromError(fmt.Errorf("failed to install prebuilt rules: %s - %s", resp.Status(), string(resp.Body)))
	}

	return nil
}

func needsRuleUpdate(ctx context.Context, client *kibana_oapi.Client, spaceID string) bool {
	status, diags := getPrebuiltRulesStatus(ctx, client, spaceID)
	if diags.HasError() {
		return true
	}
	return status.JSON200.RulesNotInstalled >= 1 || status.JSON200.RulesNotUpdated >= 1
}
