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
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet"
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
	ProxyID          types.String `tfsdk:"proxy_id"`
	SpaceIDs         types.Set    `tfsdk:"space_ids"` // > string
}

func (m serverHostModel) GetID() types.String         { return m.ID }
func (m serverHostModel) GetResourceID() types.String { return m.HostID }
func (m serverHostModel) GetSpaceID() types.String {
	return fleet.SpaceIDFromSetOrDefault(m.SpaceIDs, "")
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
	m.ProxyID = typeutils.NonEmptyStringishPointerValue(data.ProxyId)

	// Note: SpaceIDs is not returned by the API for server hosts, so we preserve it from existing state.
	// It's only used to determine which API endpoint to call.
	spaceIDs, d := typeutils.SetFromAPIStringsPreserveKnownEmpty[string](ctx, nil, m.SpaceIDs)
	diags.Append(d...)
	m.SpaceIDs = spaceIDs

	return
}

func (m serverHostModel) toAPICreateModel(ctx context.Context) (body kbapi.PostFleetFleetServerHostsJSONRequestBody, diags diag.Diagnostics) {
	body = kbapi.PostFleetFleetServerHostsJSONRequestBody{
		HostUrls:  typeutils.ListTypeToSliceString(ctx, m.Hosts, path.Root("hosts"), &diags),
		Id:        typeutils.OptionalString(m.HostID),
		IsDefault: m.Default.ValueBoolPointer(),
		Name:      m.Name.ValueString(),
		ProxyId:   typeutils.OptionalString(m.ProxyID),
	}
	return
}

func (m serverHostModel) toAPIUpdateModel(ctx context.Context, prior serverHostModel) (body kbapi.PutFleetFleetServerHostsItemidJSONRequestBody, diags diag.Diagnostics) {
	body = kbapi.PutFleetFleetServerHostsItemidJSONRequestBody{
		HostUrls:  typeutils.SliceRef(typeutils.ListTypeToSliceString(ctx, m.Hosts, path.Root("hosts"), &diags)),
		IsDefault: m.Default.ValueBoolPointer(),
		Name:      m.Name.ValueStringPointer(),
		ProxyId:   proxyIDForUpdate(m.ProxyID, prior.ProxyID),
	}
	return
}

// proxyIDForUpdate returns a pointer suitable for the generated update body.
// A known non-empty plan value is sent as-is. When the plan unsets a
// previously set proxy_id, an empty string is sent rather than nil. The
// generated `json:"proxy_id,omitempty"` tag drops nil; on Kibana 9.5.5 omission
// also clears the proxy, but sending "" is defensive for stacks where an
// omitted field leaves the value unchanged. When proxy_id was already unset,
// nil is returned so the field stays omitted.
func proxyIDForUpdate(plan, prior types.String) *string {
	if typeutils.IsKnown(plan) && plan.ValueString() != "" {
		return plan.ValueStringPointer()
	}
	if typeutils.IsKnown(prior) && prior.ValueString() != "" {
		empty := ""
		return &empty
	}
	return nil
}
