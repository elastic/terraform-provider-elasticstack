package datafeed

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func NewDatafeedResource() resource.Resource {
	return &datafeedResource{}
}

type datafeedResource struct {
	client *clients.ApiClient
}

func (r *datafeedResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elasticsearch_ml_datafeed"
}

func (r *datafeedResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func (r *datafeedResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.create(ctx, req, resp)
}

func (r *datafeedResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Datafeed
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, diags := r.read(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *datafeedResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.update(ctx, req, resp)
}

func (r *datafeedResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	r.delete(ctx, req, resp)
}

// resourceReady checks if the client is ready for API calls
func (r *datafeedResource) resourceReady(diags *fwdiags.Diagnostics) bool {
	if r.client == nil {
		diags.AddError("Client not configured", "Provider client is not configured")
		return false
	}
	return true
}

func (r *datafeedResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	compID, sdkDiags := clients.CompositeIdFromStr(req.ID)
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	datafeedID := compID.ResourceId

	// Set the datafeed_id attribute
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("datafeed_id"), datafeedID)...)
}
