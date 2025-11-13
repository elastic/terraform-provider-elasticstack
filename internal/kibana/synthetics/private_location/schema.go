package private_location

import (
	_ "embed"
	"strings"

	"github.com/disaster37/go-kibana-rest/v8/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type tfModelV0 struct {
	ID            types.String   `tfsdk:"id"`
	Label         types.String   `tfsdk:"label"`
	AgentPolicyId types.String   `tfsdk:"agent_policy_id"`
	Tags          []types.String `tfsdk:"tags"` //> string
	Geo           *tfGeoConfigV0 `tfsdk:"geo"`
}

func privateLocationSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: syntheticsPrivateLocationDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Generated id for the private location. For monitor setup please use private location label.",
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
			"geo": geoConfigSchema(),
		},
	}
}

func (m *tfModelV0) toPrivateLocationConfig() kbapi.PrivateLocationConfig {
	var geoConfig *kbapi.SyntheticGeoConfig
	if m.Geo != nil {
		geoConfig = m.Geo.toSyntheticGeoConfig()
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

	return tfModelV0{
		ID:            types.StringValue(pLoc.Id),
		Label:         types.StringValue(pLoc.Label),
		AgentPolicyId: types.StringValue(pLoc.AgentPolicyId),
		Tags:          synthetics.StringSliceValue(pLoc.Tags),
		Geo:           fromSyntheticGeoConfig(pLoc.Geo),
	}
}

//go:embed resource-description.md
var syntheticsPrivateLocationDescription string

// Geographic configuration schema and types
func geoConfigSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional:    true,
		Description: "Geographic coordinates (WGS84) for the location",
		Attributes: map[string]schema.Attribute{
			"lat": schema.Float64Attribute{
				Optional:            false,
				Required:            true,
				MarkdownDescription: "The latitude of the location.",
			},
			"lon": schema.Float64Attribute{
				Optional:            false,
				Required:            true,
				MarkdownDescription: "The longitude of the location.",
			},
		},
	}
}

type tfGeoConfigV0 struct {
	Lat types.Float64 `tfsdk:"lat"`
	Lon types.Float64 `tfsdk:"lon"`
}

func (m *tfGeoConfigV0) toSyntheticGeoConfig() *kbapi.SyntheticGeoConfig {
	return &kbapi.SyntheticGeoConfig{
		Lat: m.Lat.ValueFloat64(),
		Lon: m.Lon.ValueFloat64(),
	}
}

func fromSyntheticGeoConfig(v *kbapi.SyntheticGeoConfig) *tfGeoConfigV0 {
	if v == nil {
		return nil
	}
	return &tfGeoConfigV0{
		Lat: types.Float64Value(v.Lat),
		Lon: types.Float64Value(v.Lon),
	}
}
