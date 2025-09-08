package anomaly_detector

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
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
	create(ctx, req, resp, r.client)
}

func (r *anomalyDetectorJobResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	read(ctx, req, resp, r.client)
}

func (r *anomalyDetectorJobResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	update(ctx, req, resp, r.client)
}

func (r *anomalyDetectorJobResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteResource(ctx, req, resp, r.client)
}

func (r *anomalyDetectorJobResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Use the job ID directly as the import ID
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("job_id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
