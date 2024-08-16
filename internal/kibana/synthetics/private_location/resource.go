package private_location

import (
	"context"
	"github.com/disaster37/go-kibana-rest/v8"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const resourceName = synthetics.MetadataPrefix + "private_location"

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &Resource{}
var _ resource.ResourceWithConfigure = &Resource{}
var _ resource.ResourceWithImportState = &Resource{}

type Resource struct {
	client *clients.ApiClient
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = privateLocationSchema()
}

func (r *Resource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	tflog.Info(ctx, "Import private location")
	resource.ImportStatePassthroughID(ctx, path.Root("label"), request, response)
}

func (r *Resource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(request.ProviderData)
	response.Diagnostics.Append(diags...)
	r.client = client
}

func (r *Resource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + resourceName
}

func (r *Resource) Update(ctx context.Context, _ resource.UpdateRequest, response *resource.UpdateResponse) {
	tflog.Warn(ctx, "Update isn't supported for elasticstack_"+resourceName)
	response.Diagnostics.AddError(
		"synthetics private location update not supported",
		"Synthetics private location could only be replaced. Please, note, that only unused locations could be deleted.",
	)
}

func (r *Resource) getKibanaClient(dg diag.Diagnostics) *kibana.Client {
	if !r.resourceReady(&dg) {
		return nil
	}

	kibanaClient, err := r.client.GetKibanaClient()
	if err != nil {
		dg.AddError("unable to get kibana client", err.Error())
		return nil
	}
	return kibanaClient
}
