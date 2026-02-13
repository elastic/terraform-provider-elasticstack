package ab_agent

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &AgentResource{}
	_ resource.ResourceWithConfigure   = &AgentResource{}
	_ resource.ResourceWithImportState = &AgentResource{}
)

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &AgentResource{}
}

type AgentResource struct {
	client *clients.ApiClient
}

func (r *AgentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func (r *AgentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, "kibana_ab_agent")
}

func (r *AgentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	stateModel := agentModel{
		ID:     types.StringValue(req.ID),
		Labels: types.ListNull(types.StringType),
		Tools:  types.ListNull(types.StringType),
	}

	diags := resp.State.Set(ctx, stateModel)
	resp.Diagnostics.Append(diags...)
}
