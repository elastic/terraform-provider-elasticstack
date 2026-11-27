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

package serverhostds

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
	entitycore.KibanaConnectionField
	ID      types.String `tfsdk:"id"`
	HostID  types.String `tfsdk:"host_id"`
	Name    types.String `tfsdk:"name"`
	Default types.Bool   `tfsdk:"default"`
	SpaceID types.String `tfsdk:"space_id"`
	Hosts   types.List   `tfsdk:"hosts"` // > string
	ProxyID types.String `tfsdk:"proxy_id"`
	Ssl     types.Object `tfsdk:"ssl"`
}

type serverHostSslModel struct {
	Certificate            types.String `tfsdk:"certificate"`
	CertificateAuthorities types.List   `tfsdk:"certificate_authorities"` // > string
	ClientAuth             types.String `tfsdk:"client_auth"`
}

func (m serverHostModel) GetID() types.String      { return m.ID }
func (m serverHostModel) GetSpaceID() types.String { return m.SpaceID }

// GetResourceID returns a placeholder when host_id is not the selector: the
// data source envelope rejects an empty identity, and the read callback picks
// the lookup from the configuration, not from this value.
func (m serverHostModel) GetResourceID() types.String {
	if typeutils.IsKnown(m.HostID) {
		return m.HostID
	}
	return types.StringValue("_")
}

func (m *serverHostModel) populateFromAPI(ctx context.Context, data *kbapi.ServerHost) diag.Diagnostics {
	var diags diag.Diagnostics

	m.ID = types.StringValue(data.Id)
	m.HostID = types.StringValue(data.Id)
	m.Name = types.StringValue(data.Name)
	m.Default = types.BoolValue(typeutils.Deref(data.IsDefault))
	m.Hosts = typeutils.SliceToListTypeString(ctx, data.HostUrls, path.Root("hosts"), &diags)
	m.ProxyID = typeutils.NonEmptyStringishPointerValue(data.ProxyId)
	m.Ssl = sslFromAPI(ctx, data, &diags)

	return diags
}

func sslFromAPI(ctx context.Context, data *kbapi.ServerHost, diags *diag.Diagnostics) types.Object {
	if data.Ssl == nil {
		return types.ObjectNull(sslAttrTypes)
	}

	ssl := serverHostSslModel{
		Certificate:            typeutils.NonEmptyStringishPointerValue(data.Ssl.Certificate),
		CertificateAuthorities: types.ListNull(types.StringType),
		ClientAuth:             typeutils.NonEmptyStringishPointerValue(data.Ssl.ClientAuth),
	}
	if cas := typeutils.Deref(data.Ssl.CertificateAuthorities); len(cas) > 0 {
		ssl.CertificateAuthorities = typeutils.SliceToListTypeString(ctx, cas, path.Root("ssl").AtName(sslCertificateAuthoritiesAttr), diags)
	}

	if ssl.Certificate.IsNull() && ssl.CertificateAuthorities.IsNull() && ssl.ClientAuth.IsNull() {
		return types.ObjectNull(sslAttrTypes)
	}

	obj, d := types.ObjectValueFrom(ctx, sslAttrTypes, ssl)
	diags.Append(d...)
	return obj
}
