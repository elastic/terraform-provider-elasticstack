package ab_tool

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &ToolResource{}
	_ resource.ResourceWithConfigure   = &ToolResource{}
	_ resource.ResourceWithImportState = &ToolResource{}
)

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &ToolResource{}
}

type ToolResource struct {
	client *clients.ApiClient
}

func (r *ToolResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func (r *ToolResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, "kibana_ab_tool")
}

func (r *ToolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	stateModel := toolModel{
		ID:   types.StringValue(req.ID),
		Tags: types.ListNull(types.StringType),
	}

	diags := resp.State.Set(ctx, stateModel)
	resp.Diagnostics.Append(diags...)
}
