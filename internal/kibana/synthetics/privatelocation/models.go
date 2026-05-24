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
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Model struct {
	entitycore.KibanaConnectionField
	ID            types.String   `tfsdk:"id"`
	Label         types.String   `tfsdk:"label"`
	AgentPolicyID types.String   `tfsdk:"agent_policy_id"`
	SpaceID       types.String   `tfsdk:"space_id"`
	Tags          []types.String `tfsdk:"tags"`
	Geo           *tfGeoConfigV0 `tfsdk:"geo"`
}

var _ entitycore.KibanaResourceModel = Model{}
var _ entitycore.WithVersionRequirements = Model{}
var _ entitycore.KibanaUnscopedSpace = Model{}

// IsUnscopedSpace tells the [entitycore.KibanaResource] envelope to bypass its
// non-empty-space guard. The resource IS space-scoped — non-default spaces use
// the /s/<space_id>/ URL prefix — but the historical state contract represents
// the Kibana default space with an empty string (preserved across the
// envelope migration). The envelope's space validation enforces a non-empty
// configured space, which conflicts with that contract; opting into the
// "unscoped" hook is the documented escape hatch.
func (Model) IsUnscopedSpace() bool { return true }

func (m Model) GetID() types.String         { return m.ID }
func (m Model) GetResourceID() types.String { return m.ID }
func (m Model) GetSpaceID() types.String    { return m.SpaceID }

func (m Model) GetVersionRequirements() ([]entitycore.VersionRequirement, diag.Diagnostics) {
	effectiveSpace := versionGateSpaceID(m)
	if !requiresSpaceIDMinVersion(effectiveSpace) {
		return nil, nil
	}
	return []entitycore.VersionRequirement{
		{
			MinVersion:   *MinVersionSpaceID,
			ErrorMessage: fmt.Sprintf("Synthetics private locations in a non-default Kibana space require Elastic Stack %s or later.", MinVersionSpaceID),
		},
	}, nil
}

// versionGateSpaceID returns the Kibana space used for version-gating. When the
// resource id is a composite import id (<space_id>/<private_location_id>), the
// space segment is used if space_id is not yet in state.
func versionGateSpaceID(m Model) string {
	return effectiveSpaceID(m.SpaceID, compositeIDFromModel(m))
}

func compositeIDFromModel(m Model) *clients.CompositeID {
	compID, _ := clients.CompositeIDFromStr(m.GetID().ValueString())
	return compID
}

// privateLocationToCreateBody converts a Terraform model into a kbapi create request body.
// Tags are passed as *[]string and geo coordinates preserve float64 precision.
func (m Model) toCreateBody() kbapi.PostPrivateLocationJSONRequestBody {
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

// modelFromAPI maps a SyntheticsGetPrivateLocation API response onto a Model.
func modelFromAPI(loc kbapi.SyntheticsGetPrivateLocation, spaceID string) Model {
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

	tags := tagsFromAdditionalProperties(loc)

	return Model{
		ID:            types.StringValue(id),
		Label:         types.StringValue(label),
		AgentPolicyID: types.StringValue(agentPolicyID),
		SpaceID:       types.StringValue(spaceID),
		Tags:          synthetics.StringSliceValue(tags),
		Geo:           geoFromAPIResponse(loc.Geo),
	}
}
