package synthetics

import (
	"context"
	"github.com/disaster37/go-kibana-rest/v8"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

const resourceName = MetadataPrefix + "monitor"

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &Resource{}
var _ resource.ResourceWithConfigure = &Resource{}
var _ resource.ResourceWithImportState = &Resource{}
var _ resource.ResourceWithConfigValidators = &Resource{}
var _ ESApiClient = &Resource{}

type ESApiClient interface {
	GetClient() *clients.ApiClient
}

func GetKibanaClient(c ESApiClient, dg diag.Diagnostics) *kibana.Client {

	client := c.GetClient()
	if client == nil {
		dg.AddError(
			"Unconfigured Client",
			"Expected configured client. Please report this issue to the provider developers.",
		)
		return nil
	}

	kibanaClient, err := client.GetKibanaClient()
	if err != nil {
		dg.AddError("unable to get kibana client", err.Error())
		return nil
	}
	return kibanaClient
}

type Resource struct {
	client *clients.ApiClient
	ESApiClient
}

func (r *Resource) GetClient() *clients.ApiClient {
	return r.client
}
func (r *Resource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("http"),
			path.MatchRoot("tcp"),
			// other monitor config types: icmp, browser
		),
		resourcevalidator.AtLeastOneOf(
			path.MatchRoot("locations"),
			path.MatchRoot("private_locations"),
		),
	}
}

func (r *Resource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *Resource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(request.ProviderData)
	response.Diagnostics.Append(diags...)
	r.client = client
}

func (r *Resource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + resourceName
}

func (r *Resource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = monitorConfigSchema()

}

func (r *Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	//TODO implement me
	panic("implement me")
}

func (r *Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	//TODO implement me
	panic("implement me")
}

func (r *Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	//TODO implement me
	panic("implement me")
}

func (r *Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	//TODO implement me
	panic("implement me")
}
