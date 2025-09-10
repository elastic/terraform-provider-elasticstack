package agent_policy

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                = &agentPolicyResource{}
	_ resource.ResourceWithConfigure   = &agentPolicyResource{}
	_ resource.ResourceWithImportState = &agentPolicyResource{}
)

var (
	MinVersionGlobalDataTags    = version.Must(version.NewVersion("8.15.0"))
	MinSupportsAgentlessVersion = version.Must(version.NewVersion("8.15.0"))
	MinVersionInactivityTimeout = version.Must(version.NewVersion("8.7.0"))
)

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &agentPolicyResource{}
}

type agentPolicyResource struct {
	client *clients.ApiClient
}

func (r *agentPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func (r *agentPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, "fleet_agent_policy")
}

func (r *agentPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("policy_id"), req, resp)
}

func (r *agentPolicyResource) buildFeatures(ctx context.Context) (features, diag.Diagnostics) {
	supportsGDT, diags := r.client.EnforceMinVersion(ctx, MinVersionGlobalDataTags)
	if diags.HasError() {
		return features{}, utils.FrameworkDiagsFromSDK(diags)
	}

	supportsSupportsAgentless, diags := r.client.EnforceMinVersion(ctx, MinSupportsAgentlessVersion)
	if diags.HasError() {
		return features{}, utils.FrameworkDiagsFromSDK(diags)
	}

	supportsInactivityTimeout, diags := r.client.EnforceMinVersion(ctx, MinVersionInactivityTimeout)
	if diags.HasError() {
		return features{}, utils.FrameworkDiagsFromSDK(diags)
	}

	return features{
		SupportsGlobalDataTags:    supportsGDT,
		SupportsSupportsAgentless: supportsSupportsAgentless,
		SupportsInactivityTimeout: supportsInactivityTimeout,
	}, nil
}
