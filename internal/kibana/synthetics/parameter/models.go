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

package parameter

import (
	"slices"

	kboapi "github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Model struct {
	entitycore.ResourceTimeoutsField
	entitycore.KibanaConnectionField
	ID                types.String   `tfsdk:"id"`
	SpaceID           types.String   `tfsdk:"space_id"`
	Key               types.String   `tfsdk:"key"`
	Value             types.String   `tfsdk:"value"`
	Description       types.String   `tfsdk:"description"`
	Tags              []types.String `tfsdk:"tags"`
	ShareAcrossSpaces types.Bool     `tfsdk:"share_across_spaces"`
}

var _ entitycore.KibanaResourceModel = Model{}

func (m Model) GetID() types.String { return m.ID }

func (m Model) GetResourceID() types.String {
	if m.ID.IsNull() || m.ID.IsUnknown() {
		return types.StringNull()
	}
	compID, _ := synthetics.TryReadCompositeID(m.ID.ValueString())
	if compID != nil {
		return types.StringValue(compID.ResourceID)
	}
	return m.ID
}

func (m Model) GetSpaceID() types.String { return m.SpaceID }

func (m Model) toParameterRequest(forUpdate bool) kboapi.SyntheticsParameterRequest {
	// share_across_spaces is not allowed to be set when updating an existing
	// global parameter.
	var shareAcrossSpaces *bool
	if !forUpdate {
		shareAcrossSpaces = m.ShareAcrossSpaces.ValueBoolPointer()
	}

	return kboapi.SyntheticsParameterRequest{
		Key:         m.Key.ValueString(),
		Value:       m.Value.ValueString(),
		Description: new(m.Description.ValueString()),
		// We need this to marshal as an empty JSON array, not null.
		Tags:              new(typeutils.NonNilSlice(typeutils.ValueStringSlice(m.Tags))),
		ShareAcrossSpaces: shareAcrossSpaces,
	}
}

func modelFromOAPI(param kboapi.SyntheticsGetParameterResponse, spaceID string) Model {
	// Namespaces is omitempty in the Kibana API and is only populated for users
	// with read-only permissions; treat a missing list as not shared across spaces.
	allSpaces := param.Namespaces != nil && slices.Equal(*param.Namespaces, []string{"*"})

	var id types.String
	if param.Id != nil {
		id = types.StringValue((&clients.CompositeID{ClusterID: spaceID, ResourceID: *param.Id}).String())
	}

	return Model{
		ID:          id,
		SpaceID:     types.StringValue(spaceID),
		Key:         types.StringPointerValue(param.Key),
		Value:       types.StringPointerValue(param.Value),
		Description: types.StringPointerValue(param.Description),
		// Terraform, like json.Marshal, treats empty slices as null. We need an
		// actual backing array of size 0.
		Tags:              typeutils.NonNilSlice(typeutils.StringSliceValue(typeutils.Deref(param.Tags))),
		ShareAcrossSpaces: types.BoolValue(allSpaces),
	}
}
