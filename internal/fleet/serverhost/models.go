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

package serverhost

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type serverHostModel struct {
	entitycore.ResourceTimeoutsField
	ID               types.String `tfsdk:"id"`
	KibanaConnection types.List   `tfsdk:"kibana_connection"`
	HostID           types.String `tfsdk:"host_id"`
	Name             types.String `tfsdk:"name"`
	Hosts            types.List   `tfsdk:"hosts"`
	Default          types.Bool   `tfsdk:"default"`
	SpaceIDs         types.Set    `tfsdk:"space_ids"` // > string
}

func (m serverHostModel) GetID() types.String         { return m.ID }
func (m serverHostModel) GetResourceID() types.String { return m.HostID }
func (m serverHostModel) GetSpaceID() types.String {
	if m.SpaceIDs.IsNull() || m.SpaceIDs.IsUnknown() {
		return types.StringValue("")
	}
	for _, elem := range m.SpaceIDs.Elements() {
		s, ok := elem.(types.String)
		if !ok || s.IsNull() || s.IsUnknown() {
			continue
		}
		if v := s.ValueString(); v != "" {
			return s
		}
	}
	return types.StringValue("")
}
func (m serverHostModel) GetKibanaConnection() types.List { return m.KibanaConnection }

// IsUnscopedSpace implements entitycore.KibanaUnscopedSpace.
func (m serverHostModel) IsUnscopedSpace() bool { return true }

func (m *serverHostModel) populateFromAPI(ctx context.Context, data *kbapi.ServerHost) (diags diag.Diagnostics) {
	if data == nil {
		return nil
	}

	m.ID = types.StringValue(data.Id)
	m.HostID = types.StringValue(data.Id)
	m.Name = types.StringValue(data.Name)
	m.Hosts = typeutils.SliceToListTypeString(ctx, data.HostUrls, path.Root("hosts"), &diags)
	m.Default = types.BoolPointerValue(data.IsDefault)

	// Note: SpaceIDs is not returned by the API for server hosts, so we preserve it from existing state.
	// It's only used to determine which API endpoint to call.
	// If space_ids is unknown (not provided by user), set to null to satisfy Terraform's requirement.
	if m.SpaceIDs.IsUnknown() {
		m.SpaceIDs = types.SetNull(types.StringType)
	}

	return
}

func (m serverHostModel) toAPICreateModel(ctx context.Context) (body kbapi.PostFleetFleetServerHostsJSONRequestBody, diags diag.Diagnostics) {
	body = kbapi.PostFleetFleetServerHostsJSONRequestBody{
		HostUrls:  typeutils.ListTypeToSliceString(ctx, m.Hosts, path.Root("hosts"), &diags),
		Id:        typeutils.OptionalString(m.HostID),
		IsDefault: m.Default.ValueBoolPointer(),
		Name:      m.Name.ValueString(),
	}
	return
}

func (m serverHostModel) toAPIUpdateModel(ctx context.Context) (body kbapi.PutFleetFleetServerHostsItemidJSONRequestBody, diags diag.Diagnostics) {
	body = kbapi.PutFleetFleetServerHostsItemidJSONRequestBody{
		HostUrls:  typeutils.SliceRef(typeutils.ListTypeToSliceString(ctx, m.Hosts, path.Root("hosts"), &diags)),
		IsDefault: m.Default.ValueBoolPointer(),
		Name:      m.Name.ValueStringPointer(),
	}
	return
}
