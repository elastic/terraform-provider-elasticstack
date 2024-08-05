package private_location

import (
	"context"
	"github.com/disaster37/go-kibana-rest/v8/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const resourceName = synthetics.MetadataPrefix + "private_location"

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &Resource{}
var _ resource.ResourceWithConfigure = &Resource{}
var _ resource.ResourceWithImportState = &Resource{}

type Resource struct {
	client *clients.ApiClient
}

type tfModelV0 struct {
	ID            types.String              `tfsdk:"id"`
	Label         types.String              `tfsdk:"label"`
	SpaceID       types.String              `tfsdk:"space_id"`
	AgentPolicyId types.String              `tfsdk:"agent_policy_id"`
	Tags          []types.String            `tfsdk:"tags"` //> string
	Geo           *synthetics.TFGeoConfigV0 `tfsdk:"geo"`
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

func privateLocationSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Synthetics private location config, see https://www.elastic.co/guide/en/kibana/current/create-private-location-api.html for more details",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Generated id for the private location. For monitor setup please use private location label.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"space_id": schema.StringAttribute{
				MarkdownDescription: "An identifier for the space. If space_id is not provided, the default space is used.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"label": schema.StringAttribute{
				Optional:            false,
				Required:            true,
				MarkdownDescription: "A label for the private location, used as unique identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"agent_policy_id": schema.StringAttribute{
				Optional:            false,
				Required:            true,
				MarkdownDescription: "The ID of the agent policy associated with the private location. To create a private location for synthetics monitor you need to create an agent policy in fleet and use its agentPolicyId",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "An array of tags to categorize the private location.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
					listplanmodifier.RequiresReplace(),
				},
			},
			"geo": synthetics.GeoConfigSchema(),
		},
	}
}

func (r *Resource) resourceReady(dg *diag.Diagnostics) bool {
	if r.client == nil {
		dg.AddError(
			"Unconfigured Client",
			"Expected configured client. Please report this issue to the provider developers.",
		)

		return false
	}
	return true
}

func (m *tfModelV0) toPrivateLocation() kbapi.PrivateLocation {
	var geoConfig *kbapi.SyntheticGeoConfig
	if m.Geo != nil {
		geoConfig = m.Geo.ToSyntheticGeoConfig()
	}

	var tags []string
	for _, tag := range m.Tags {
		tags = append(tags, tag.ValueString())
	}
	pLoc := kbapi.PrivateLocationConfig{
		Label:         m.Label.ValueString(),
		AgentPolicyId: m.AgentPolicyId.ValueString(),
		Tags:          tags,
		Geo:           geoConfig,
	}

	return kbapi.PrivateLocation{
		Id:                    m.ID.ValueString(),
		Namespace:             m.SpaceID.ValueString(),
		PrivateLocationConfig: pLoc,
	}
}

func toModelV0(pLoc kbapi.PrivateLocation) tfModelV0 {
	var tags []types.String
	for _, tag := range pLoc.Tags {
		tags = append(tags, types.StringValue(tag))
	}
	return tfModelV0{
		ID:            types.StringValue(pLoc.Id),
		Label:         types.StringValue(pLoc.Label),
		SpaceID:       types.StringValue(pLoc.Namespace),
		AgentPolicyId: types.StringValue(pLoc.AgentPolicyId),
		Tags:          tags,
		Geo:           synthetics.FromSyntheticGeoConfig(pLoc.Geo),
	}
}
