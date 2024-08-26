package private_location

import (
	"github.com/disaster37/go-kibana-rest/v8/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
)

type tfModelV0 struct {
	ID            types.String              `tfsdk:"id"`
	Label         types.String              `tfsdk:"label"`
	SpaceID       types.String              `tfsdk:"space_id"`
	AgentPolicyId types.String              `tfsdk:"agent_policy_id"`
	Tags          []types.String            `tfsdk:"tags"` //> string
	Geo           *synthetics.TFGeoConfigV0 `tfsdk:"geo"`
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

func (m *tfModelV0) toPrivateLocationConfig() kbapi.PrivateLocationConfig {
	var geoConfig *kbapi.SyntheticGeoConfig
	if m.Geo != nil {
		geoConfig = m.Geo.ToSyntheticGeoConfig()
	}

	return kbapi.PrivateLocationConfig{
		Label:         m.Label.ValueString(),
		AgentPolicyId: m.AgentPolicyId.ValueString(),
		Tags:          synthetics.ValueStringSlice(m.Tags),
		Geo:           geoConfig,
	}
}

func tryReadCompositeId(id string) (*clients.CompositeId, diag.Diagnostics) {
	if strings.Contains(id, "/") {
		compositeId, diagnostics := synthetics.GetCompositeId(id)
		return compositeId, diagnostics
	}
	return nil, diag.Diagnostics{}
}

func toModelV0(pLoc kbapi.PrivateLocation) tfModelV0 {

	resourceID := clients.CompositeId{
		ClusterId:  pLoc.Namespace,
		ResourceId: pLoc.Id,
	}

	return tfModelV0{
		ID:            types.StringValue(resourceID.String()),
		Label:         types.StringValue(pLoc.Label),
		SpaceID:       types.StringValue(pLoc.Namespace),
		AgentPolicyId: types.StringValue(pLoc.AgentPolicyId),
		Tags:          synthetics.StringSliceValue(pLoc.Tags),
		Geo:           synthetics.FromSyntheticGeoConfig(pLoc.Geo),
	}
}
