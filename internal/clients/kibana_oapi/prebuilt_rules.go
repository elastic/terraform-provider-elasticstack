package kibana_oapi

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// GetPrebuiltRulesStatus retrieves the status of prebuilt rules and timelines for a given space.
func GetPrebuiltRulesStatus(ctx context.Context, client *Client, spaceID string) (*kbapi.ReadPrebuiltRulesAndTimelinesStatusResponse, diag.Diagnostics) {
	resp, err := client.API.ReadPrebuiltRulesAndTimelinesStatusWithResponse(ctx, SpaceAwarePathRequestEditor(spaceID))

	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	if resp.StatusCode() != 200 {
		return nil, diagutil.FrameworkDiagFromError(fmt.Errorf("failed to get prebuilt rules status: %s", resp.Status()))
	}

	return resp, nil
}

// InstallPrebuiltRules installs or updates prebuilt rules and timelines for a given space.
func InstallPrebuiltRules(ctx context.Context, client *Client, spaceID string) diag.Diagnostics {
	resp, err := client.API.InstallPrebuiltRulesAndTimelinesWithResponse(ctx, SpaceAwarePathRequestEditor(spaceID))

	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	if resp.StatusCode() != 200 {
		return diagutil.CheckHttpErrorFromFW(resp.HTTPResponse, "failed to install prebuilt rules")
	}

	return nil
}
