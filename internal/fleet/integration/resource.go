package integration

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource               = &integrationResource{}
	_ resource.ResourceWithConfigure  = &integrationResource{}
	_ resource.ResourceWithModifyPlan = &integrationResource{}
)

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &integrationResource{}
}

type integrationResource struct {
	client *clients.ApiClient
}

func (r *integrationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func (r *integrationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, "fleet_integration")
}

func (r *integrationResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Only modify during create or update
	if req.Plan.Raw.IsNull() {
		return
	}

	var configVersion types.String
	diags := req.Config.GetAttribute(ctx, path.Root("version"), &configVersion)
	if diags.HasError() {
		return
	}

	// If version is unknown (not set in config), set it to the latest version
	if !utils.IsKnown(configVersion) {
		// Client must be configured for this to work
		if r.client == nil {
			// Client not yet configured, skip plan modification
			// This will be handled during actual creation/update
			return
		}

		fleetClient, err := r.client.GetFleetClient()
		if err != nil {
			// If we can't get fleet client during plan, skip modification
			// The error will be caught during actual resource operations
			return
		}

		var plan integrationModel
		diags := req.Plan.Get(ctx, &plan)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		packageName := plan.Name.ValueString()
		if packageName == "" {
			// Package name not set, can't determine latest version
			return
		}

		latestVersion, diags := fleet.GetLatestPackageVersion(ctx, fleetClient, packageName)
		// If we can't get the latest version during plan (e.g., network issues),
		// don't fail the plan - just skip the modification
		if diags.HasError() {
			return
		}

		// Update the version in the plan
		plan.Version = types.StringValue(latestVersion)

		// Set the modified plan
		diags = resp.Plan.Set(ctx, plan)
		resp.Diagnostics.Append(diags...)
	}
}
