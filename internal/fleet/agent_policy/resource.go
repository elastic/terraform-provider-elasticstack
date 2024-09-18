package agent_policy

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                = &agentPolicyResource{}
	_ resource.ResourceWithConfigure   = &agentPolicyResource{}
	_ resource.ResourceWithImportState = &agentPolicyResource{}
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
