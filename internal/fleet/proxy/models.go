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
	"fmt"
	"strconv"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
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

func (model *proxyModel) populateFromAPI(spaceID string, item kbapi.FleetProxyItem) diag.Diagnostics {
	var diags diag.Diagnostics

	model.ID = types.StringValue((&clients.CompositeID{ClusterID: spaceID, ResourceID: item.Id}).String())
	model.ProxyID = types.StringValue(item.Id)
	model.SpaceID = types.StringValue(spaceID)
	model.Name = types.StringValue(item.Name)
	model.URL = types.StringValue(item.Url)

	if item.Certificate != nil && *item.Certificate != "" {
		model.Certificate = types.StringValue(*item.Certificate)
	} else {
		model.Certificate = types.StringNull()
	}

	if item.CertificateKey != nil && *item.CertificateKey != "" {
		model.CertificateKey = types.StringValue(*item.CertificateKey)
	} else {
		model.CertificateKey = types.StringNull()
	}

	if item.CertificateAuthorities != nil && *item.CertificateAuthorities != "" {
		model.CertificateAuthorities = types.StringValue(*item.CertificateAuthorities)
	} else {
		model.CertificateAuthorities = types.StringNull()
	}

	if item.IsPreconfigured != nil {
		model.IsPreconfigured = types.BoolValue(*item.IsPreconfigured)
	} else {
		model.IsPreconfigured = types.BoolValue(false)
	}

	headersMap, headerDiags := proxyHeadersToModel(item.ProxyHeaders)
	diags.Append(headerDiags...)
	if !diags.HasError() {
		model.ProxyHeaders = headersMap
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

func (model proxyModel) toAPICreateModel() (kbapi.PostFleetProxiesJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	body := kbapi.PostFleetProxiesJSONRequestBody{
		Name: model.Name.ValueString(),
		Url:  model.URL.ValueString(),
	}

	if !model.ProxyID.IsNull() && !model.ProxyID.IsUnknown() {
		body.Id = model.ProxyID.ValueStringPointer()
	}

	if !model.Certificate.IsNull() && !model.Certificate.IsUnknown() {
		body.Certificate = model.Certificate.ValueStringPointer()
	}

	if !model.CertificateAuthorities.IsNull() && !model.CertificateAuthorities.IsUnknown() {
		body.CertificateAuthorities = model.CertificateAuthorities.ValueStringPointer()
	}

	if !model.CertificateKey.IsNull() && !model.CertificateKey.IsUnknown() {
		body.CertificateKey = model.CertificateKey.ValueStringPointer()
	}

	if !model.ProxyHeaders.IsNull() && !model.ProxyHeaders.IsUnknown() {
		headers, headerDiags := proxyHeadersFromModel(model.ProxyHeaders)
		diags.Append(headerDiags...)
		if diags.HasError() {
			return body, diags
		}
		body.ProxyHeaders = headers
	}

	return body, diags
}

func (model proxyModel) toAPIUpdateModel() (kbapi.PutFleetProxiesItemidJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	emptyHeaders := map[string]kbapi.FleetProxyHeaderValue{}
	body := kbapi.PutFleetProxiesItemidJSONRequestBody{
		Name:         model.Name.ValueStringPointer(),
		Url:          model.URL.ValueStringPointer(),
		ProxyHeaders: &emptyHeaders,
	}

	if !model.Certificate.IsNull() && !model.Certificate.IsUnknown() {
		body.Certificate = model.Certificate.ValueStringPointer()
	}

	if !model.CertificateAuthorities.IsNull() && !model.CertificateAuthorities.IsUnknown() {
		body.CertificateAuthorities = model.CertificateAuthorities.ValueStringPointer()
	}

	if !model.CertificateKey.IsNull() && !model.CertificateKey.IsUnknown() {
		body.CertificateKey = model.CertificateKey.ValueStringPointer()
	}

	if !model.ProxyHeaders.IsNull() && !model.ProxyHeaders.IsUnknown() && len(model.ProxyHeaders.Elements()) > 0 {
		headers, headerDiags := proxyHeadersFromModel(model.ProxyHeaders)
		diags.Append(headerDiags...)
		if diags.HasError() {
			return body, diags
		}
		body.ProxyHeaders = headers
	}

	return body, diags
}
