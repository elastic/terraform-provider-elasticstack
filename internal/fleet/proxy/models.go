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

package proxy

import (
	"context"
	"fmt"
	"strconv"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type proxyModel struct {
	ID                     types.String `tfsdk:"id"`
	KibanaConnection       types.List   `tfsdk:"kibana_connection"`
	ProxyID                types.String `tfsdk:"proxy_id"`
	SpaceID                types.String `tfsdk:"space_id"`
	Name                   types.String `tfsdk:"name"`
	URL                    types.String `tfsdk:"url"`
	Certificate            types.String `tfsdk:"certificate"`
	CertificateAuthorities types.String `tfsdk:"certificate_authorities"`
	CertificateKey         types.String `tfsdk:"certificate_key"`
	ProxyHeaders           types.Map    `tfsdk:"proxy_headers"`
	IsPreconfigured        types.Bool   `tfsdk:"is_preconfigured"`
}

var proxyMinVersion = version.Must(version.NewVersion("8.7.1"))

func (m proxyModel) GetID() types.String             { return m.ID }
func (m proxyModel) GetResourceID() types.String     { return m.ProxyID }
func (m proxyModel) GetSpaceID() types.String        { return m.SpaceID }
func (m proxyModel) GetKibanaConnection() types.List { return m.KibanaConnection }

func (m proxyModel) GetVersionRequirements(_ context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{
		{
			MinVersion:   *proxyMinVersion,
			ErrorMessage: fmt.Sprintf("Fleet proxies require Elastic Stack v%s or later.", proxyMinVersion),
		},
	}, nil
}

func (m *proxyModel) populateFromAPI(spaceID string, item kbapi.FleetProxyItem) diag.Diagnostics {
	var diags diag.Diagnostics

	m.ID = types.StringValue((&clients.CompositeID{ClusterID: spaceID, ResourceID: item.Id}).String())
	m.ProxyID = types.StringValue(item.Id)
	m.SpaceID = types.StringValue(spaceID)
	m.Name = types.StringValue(item.Name)
	m.URL = types.StringValue(item.Url)

	if item.Certificate != nil && *item.Certificate != "" {
		m.Certificate = types.StringValue(*item.Certificate)
	} else {
		m.Certificate = types.StringNull()
	}

	if item.CertificateKey != nil && *item.CertificateKey != "" {
		m.CertificateKey = types.StringValue(*item.CertificateKey)
	} else {
		m.CertificateKey = types.StringNull()
	}

	if item.CertificateAuthorities != nil && *item.CertificateAuthorities != "" {
		m.CertificateAuthorities = types.StringValue(*item.CertificateAuthorities)
	} else {
		m.CertificateAuthorities = types.StringNull()
	}

	if item.IsPreconfigured != nil {
		m.IsPreconfigured = types.BoolValue(*item.IsPreconfigured)
	} else {
		m.IsPreconfigured = types.BoolValue(false)
	}

	headersMap, headerDiags := proxyHeadersToModel(item.ProxyHeaders)
	diags.Append(headerDiags...)
	if !diags.HasError() {
		m.ProxyHeaders = headersMap
	}

	return diags
}

func proxyHeadersToModel(api *map[string]kbapi.FleetProxyHeaderValue) (types.Map, diag.Diagnostics) {
	var diags diag.Diagnostics

	if api == nil || len(*api) == 0 {
		return types.MapNull(types.StringType), diags
	}

	elems := make(map[string]attr.Value, len(*api))
	for k, v := range *api {
		s, ok := proxyHeaderValueToString(v)
		if !ok {
			diags.AddError(
				"Unsupported proxy header value",
				fmt.Sprintf("Proxy header %q has a value type the provider cannot represent in state.", k),
			)
			return types.MapNull(types.StringType), diags
		}
		elems[k] = types.StringValue(s)
	}

	headersMap, mapDiags := types.MapValue(types.StringType, elems)
	diags.Append(mapDiags...)
	return headersMap, diags
}

func proxyHeaderValueToString(v kbapi.FleetProxyHeaderValue) (string, bool) {
	if s, err := v.AsFleetProxyHeaderValueString(); err == nil {
		return s, true
	}
	if b, err := v.AsFleetProxyHeaderValueBoolean(); err == nil {
		return strconv.FormatBool(b), true
	}
	if n, err := v.AsFleetProxyHeaderValueNumber(); err == nil {
		return strconv.FormatFloat(float64(n), 'f', -1, 32), true
	}
	return "", false
}

func proxyHeadersFromModel(m types.Map) (*map[string]kbapi.FleetProxyHeaderValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	if m.IsNull() || m.IsUnknown() || len(m.Elements()) == 0 {
		return nil, diags
	}

	out := make(map[string]kbapi.FleetProxyHeaderValue, len(m.Elements()))
	for k, v := range m.Elements() {
		s := v.(types.String).ValueString()
		var hv kbapi.FleetProxyHeaderValue
		if err := hv.FromFleetProxyHeaderValueString(s); err != nil {
			diags.AddError("Failed to encode proxy header", fmt.Sprintf("Could not encode proxy header %q: %s", k, err))
			return nil, diags
		}
		out[k] = hv
	}

	return &out, diags
}

func (m proxyModel) toAPICreateModel() (kbapi.PostFleetProxiesJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	body := kbapi.PostFleetProxiesJSONRequestBody{
		Name: m.Name.ValueString(),
		Url:  m.URL.ValueString(),
	}

	if !m.ProxyID.IsNull() && !m.ProxyID.IsUnknown() {
		body.Id = m.ProxyID.ValueStringPointer()
	}

	if !m.Certificate.IsNull() && !m.Certificate.IsUnknown() {
		body.Certificate = m.Certificate.ValueStringPointer()
	}

	if !m.CertificateAuthorities.IsNull() && !m.CertificateAuthorities.IsUnknown() {
		body.CertificateAuthorities = m.CertificateAuthorities.ValueStringPointer()
	}

	if !m.CertificateKey.IsNull() && !m.CertificateKey.IsUnknown() {
		body.CertificateKey = m.CertificateKey.ValueStringPointer()
	}

	if !m.ProxyHeaders.IsNull() && !m.ProxyHeaders.IsUnknown() {
		headers, headerDiags := proxyHeadersFromModel(m.ProxyHeaders)
		diags.Append(headerDiags...)
		if diags.HasError() {
			return body, diags
		}
		body.ProxyHeaders = headers
	}

	return body, diags
}

func (m proxyModel) toAPIUpdateModel() (kbapi.PutFleetProxiesItemidJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	emptyHeaders := map[string]kbapi.FleetProxyHeaderValue{}
	body := kbapi.PutFleetProxiesItemidJSONRequestBody{
		Name:         m.Name.ValueStringPointer(),
		Url:          m.URL.ValueStringPointer(),
		ProxyHeaders: &emptyHeaders,
	}

	if !m.Certificate.IsNull() && !m.Certificate.IsUnknown() {
		body.Certificate = m.Certificate.ValueStringPointer()
	}

	if !m.CertificateAuthorities.IsNull() && !m.CertificateAuthorities.IsUnknown() {
		body.CertificateAuthorities = m.CertificateAuthorities.ValueStringPointer()
	}

	if !m.CertificateKey.IsNull() && !m.CertificateKey.IsUnknown() {
		body.CertificateKey = m.CertificateKey.ValueStringPointer()
	}

	if !m.ProxyHeaders.IsNull() && !m.ProxyHeaders.IsUnknown() && len(m.ProxyHeaders.Elements()) > 0 {
		headers, headerDiags := proxyHeadersFromModel(m.ProxyHeaders)
		diags.Append(headerDiags...)
		if diags.HasError() {
			return body, diags
		}
		body.ProxyHeaders = headers
	}

	return body, diags
}
