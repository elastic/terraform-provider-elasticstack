// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package privatelocation

import (
	_ "embed"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics"
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type tfModelV0 struct {
	ID               types.String   `tfsdk:"id"`
	KibanaConnection types.List     `tfsdk:"kibana_connection"`
	Label            types.String   `tfsdk:"label"`
	AgentPolicyID    types.String   `tfsdk:"agent_policy_id"`
	SpaceID          types.String   `tfsdk:"space_id"`
	Tags             []types.String `tfsdk:"tags"` // > string
	Geo              *tfGeoConfigV0 `tfsdk:"geo"`
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
				MarkdownDescription: agentPolicyIDDescription,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"space_id": schema.StringAttribute{
				MarkdownDescription: spaceIDDescription,
				Optional:            true,
				Computed:            true,
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

		Blocks: map[string]schema.Block{
			"kibana_connection": providerschema.GetKbFWConnectionBlock(),
		}}
}

// privateLocationToCreateBody converts a Terraform model into a kbapi create request body.
// Tags are passed as *[]string and geo coordinates preserve float64 precision.
func privateLocationToCreateBody(m tfModelV0) kbapi.PostPrivateLocationJSONRequestBody {
	body := kbapi.PostPrivateLocationJSONRequestBody{
		Label:         m.Label.ValueString(),
		AgentPolicyId: m.AgentPolicyID.ValueString(),
	}

	tags := synthetics.ValueStringSlice(m.Tags)
	if len(tags) > 0 {
		body.Tags = &tags
	}

	if m.Geo != nil {
		body.Geo = &struct {
			Lat float64 `json:"lat"`
			Lon float64 `json:"lon"`
		}{
			Lat: m.Geo.Lat.ValueFloat64(),
			Lon: m.Geo.Lon.ValueFloat64(),
		}
	}

	return body
}

// privateLocationFromAPI maps a SyntheticsGetPrivateLocation API response onto a tfModelV0.
// Tags are read from AdditionalProperties["tags"] when the field is not a first-class struct field.
func privateLocationFromAPI(loc kbapi.SyntheticsGetPrivateLocation, spaceID string, kibanaConnection types.List) tfModelV0 {
	id := ""
	if loc.Id != nil {
		id = *loc.Id
	}

	label := ""
	if loc.Label != nil {
		label = *loc.Label
	}

	agentPolicyID := ""
	if loc.AgentPolicyId != nil {
		agentPolicyID = *loc.AgentPolicyId
	}

	// Tags are not a first-class field on SyntheticsGetPrivateLocation; they may land
	// in AdditionalProperties when the generated model omits them.
	tags := tagsFromAdditionalProperties(loc)

	return tfModelV0{
		ID:               types.StringValue(id),
		Label:            types.StringValue(label),
		AgentPolicyID:    types.StringValue(agentPolicyID),
		SpaceID:          types.StringValue(spaceID),
		Tags:             synthetics.StringSliceValue(tags),
		Geo:              geoFromAPIResponse(loc.Geo),
		KibanaConnection: kibanaConnection,
	}
}

// tagsFromAdditionalProperties extracts the tags slice from the API response.
// The generated OpenAPI struct does not have a first-class Tags field, so tags are
// stored in AdditionalProperties["tags"] as []interface{}.
func tagsFromAdditionalProperties(loc kbapi.SyntheticsGetPrivateLocation) []string {
	val, found := loc.Get("tags")
	if !found || val == nil {
		return nil
	}
	rawSlice, ok := val.([]any)
	if !ok {
		return nil
	}
	tags := make([]string, 0, len(rawSlice))
	for _, v := range rawSlice {
		if s, ok := v.(string); ok {
			tags = append(tags, s)
		}
	}
	return tags
}

// geoFromAPIResponse converts the nested geo struct from the API response to a tfGeoConfigV0.
func geoFromAPIResponse(geo *struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}) *tfGeoConfigV0 {
	if geo == nil {
		return nil
	}
	return &tfGeoConfigV0{
		Lat: NewFloat32PrecisionValue(float64(geo.Lat)),
		Lon: NewFloat32PrecisionValue(float64(geo.Lon)),
	}
}

// effectiveSpaceID returns the Kibana space for API calls. When the resource id
// is a composite import id (<space_id>/<private_location_id>), the space
// segment is used if space_id is not yet in state (for example right after import).
func effectiveSpaceID(spaceID types.String, compositeID *clients.CompositeID) string {
	s := spaceID.ValueString()
	if compositeID != nil && (spaceID.IsNull() || spaceID.IsUnknown() || s == "") {
		return compositeID.ClusterID
	}
	return s
}

//go:embed resource-description.md
var syntheticsPrivateLocationDescription string

//go:embed agent_policy_id-description.md
var agentPolicyIDDescription string

//go:embed descriptions/space_id.md
var spaceIDDescription string

// Geographic configuration schema and types
func geoConfigSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional:    true,
		Description: "Geographic coordinates (WGS84) for the location",
		Attributes: map[string]schema.Attribute{
			"lat": schema.Float64Attribute{
				Optional:            false,
				Required:            true,
				CustomType:          Float32PrecisionType{},
				MarkdownDescription: "The latitude of the location.",
			},
			"lon": schema.Float64Attribute{
				Optional:            false,
				Required:            true,
				CustomType:          Float32PrecisionType{},
				MarkdownDescription: "The longitude of the location.",
			},
		},
	}
}

type tfGeoConfigV0 struct {
	Lat Float32PrecisionValue `tfsdk:"lat"`
	Lon Float32PrecisionValue `tfsdk:"lon"`
}
