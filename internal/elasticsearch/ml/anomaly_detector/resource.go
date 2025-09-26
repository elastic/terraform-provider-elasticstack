package anomaly_detector

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func NewAnomalyDetectorJobResource() resource.Resource {
	return &anomalyDetectorJobResource{}
}

type anomalyDetectorJobResource struct {
	client *clients.ApiClient
}

func (r *anomalyDetectorJobResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elasticsearch_ml_anomaly_detector"
}

func (r *anomalyDetectorJobResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func (r *anomalyDetectorJobResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.create(ctx, req, resp)
}

func (r *anomalyDetectorJobResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AnomalyDetectorJobTFModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, diags := r.read(ctx, &state)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *anomalyDetectorJobResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.update(ctx, req, resp)
}

func (r *anomalyDetectorJobResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	r.delete(ctx, req, resp)
}

// resourceReady checks if the client is ready for API calls
func (r *anomalyDetectorJobResource) resourceReady(diags *fwdiags.Diagnostics) bool {
	if r.client == nil {
		diags.AddError("Client not configured", "Provider client is not configured")
		return false
	}
	return true
}

func (r *anomalyDetectorJobResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Use the job ID directly as the import ID
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("job_id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
