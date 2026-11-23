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
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var serverHostSslMinVersion = version.Must(version.NewVersion("9.1.0"))

type serverHostModel struct {
	entitycore.ResourceTimeoutsField
	entitycore.KibanaConnectionField
	ID       types.String `tfsdk:"id"`
	HostID   types.String `tfsdk:"host_id"`
	Name     types.String `tfsdk:"name"`
	Hosts    types.List   `tfsdk:"hosts"`
	Default  types.Bool   `tfsdk:"default"`
	ProxyID  types.String `tfsdk:"proxy_id"`
	SpaceIDs types.Set    `tfsdk:"space_ids"` // > string
	Ssl      types.Object `tfsdk:"ssl"`
}

func (m serverHostModel) GetID() types.String         { return m.ID }
func (m serverHostModel) GetResourceID() types.String { return m.HostID }
func (m serverHostModel) GetSpaceID() types.String {
	return fleet.SpaceIDFromSetOrDefault(m.SpaceIDs, "")
}

// IsUnscopedSpace implements entitycore.KibanaUnscopedSpace.
func (m serverHostModel) IsUnscopedSpace() bool { return true }

func (m serverHostModel) GetVersionRequirements(_ context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	var reqs []entitycore.VersionRequirement
	if typeutils.IsKnown(m.Ssl) {
		reqs = append(reqs, entitycore.NewAttributeVersionRequirement(
			path.Root("ssl"),
			*serverHostSslMinVersion,
			fmt.Sprintf("ssl is only supported in Elastic Stack %s and above", serverHostSslMinVersion),
		))
	}
	return reqs, nil
}

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

	if data.Ssl == nil {
		m.Ssl = types.ObjectNull(sslAttrTypes)
	} else {
		var d diag.Diagnostics
		m.Ssl, d = sslToObjectValue(ctx, data.Ssl.Certificate, data.Ssl.CertificateAuthorities, data.Ssl.Key, data.Ssl.ClientAuth)
		diags.Append(d...)
	}

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

	ssl, d := objectValueToSSL(ctx, m.Ssl)
	diags.Append(d...)
	if err := assignSSL(&body.Ssl, ssl); err != nil {
		diags.AddError("Failed to build ssl request", err.Error())
	}
	return
}

func (m serverHostModel) toAPIUpdateModel(ctx context.Context, prior serverHostModel) (body kbapi.PutFleetFleetServerHostsItemidJSONRequestBody, diags diag.Diagnostics) {
	body = kbapi.PutFleetFleetServerHostsItemidJSONRequestBody{
		HostUrls:  typeutils.SliceRef(typeutils.ListTypeToSliceString(ctx, m.Hosts, path.Root("hosts"), &diags)),
		IsDefault: m.Default.ValueBoolPointer(),
		Name:      m.Name.ValueStringPointer(),
		ProxyId:   fleet.ProxyIDForUpdate(m.ProxyID, prior.ProxyID),
	}

	ssl, d := objectValueToSSLUpdate(ctx, m.Ssl, prior.Ssl)
	diags.Append(d...)
	if err := assignSSL(&body.Ssl, ssl); err != nil {
		diags.AddError("Failed to build ssl request", err.Error())
	}
	return
}
