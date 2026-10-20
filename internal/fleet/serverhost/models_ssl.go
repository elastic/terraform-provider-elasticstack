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
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/sslcore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type serverHostSslModel struct {
	sslcore.Values
	ClientAuth types.String `tfsdk:"client_auth"`
}

// sslAPIModel is the supported subset of the Kibana ssl request object. The
// generated create and update bodies declare distinct anonymous ssl types, so
// this model is copied into them with assignSSL.
type sslAPIModel struct {
	Certificate            *string   `json:"certificate,omitempty"`
	CertificateAuthorities *[]string `json:"certificate_authorities,omitempty"`
	ClientAuth             *string   `json:"client_auth,omitempty"`
	Key                    *string   `json:"key,omitempty"`
}

func objectValueToSSL(ctx context.Context, obj types.Object) (*sslAPIModel, diag.Diagnostics) {
	if !typeutils.IsKnown(obj) {
		return nil, nil
	}

	var diags diag.Diagnostics
	sslModel := typeutils.ObjectTypeAs[serverHostSslModel](ctx, obj, path.Root("ssl"), &diags)
	if diags.HasError() || sslModel == nil {
		return nil, diags
	}

	certificate, certificateAuthorities, key := sslcore.Extract(ctx, sslModel.Values, path.Root("ssl").AtName("certificate_authorities"), &diags)

	return &sslAPIModel{
		Certificate:            certificate,
		CertificateAuthorities: certificateAuthorities,
		ClientAuth:             sslModel.ClientAuth.ValueStringPointer(),
		Key:                    key,
	}, diags
}

// objectValueToSSLUpdate returns the ssl object for an update request. Kibana
// replaces the stored object on PUT and ignores an omitted ssl field, so
// removing ssl from configuration is sent as an empty object.
func objectValueToSSLUpdate(ctx context.Context, plan, prior types.Object) (*sslAPIModel, diag.Diagnostics) {
	ssl, diags := objectValueToSSL(ctx, plan)
	if diags.HasError() || ssl != nil {
		return ssl, diags
	}

	if typeutils.IsKnown(prior) {
		return &sslAPIModel{}, diags
	}

	return nil, diags
}

func assignSSL[T any](dst **T, src *sslAPIModel) error {
	if src == nil {
		return nil
	}

	raw, err := json.Marshal(src)
	if err != nil {
		return err
	}

	var ssl T
	if err := json.Unmarshal(raw, &ssl); err != nil {
		return err
	}
	*dst = &ssl
	return nil
}

func sslToObjectValue(
	ctx context.Context,
	certificate *string,
	certificateAuthorities *[]string,
	key *string,
	clientAuth *kbapi.ServerHostSslClientAuth,
) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	sslModel := serverHostSslModel{
		Values:     sslcore.Hydrate(ctx, certificate, certificateAuthorities, key, path.Root("ssl").AtName("certificate_authorities"), &diags),
		ClientAuth: typeutils.StringishPointerValue(clientAuth),
	}

	if sslModel.AllNil() && sslModel.ClientAuth.IsNull() {
		return types.ObjectNull(getSslAttrTypes(ctx)), diags
	}

	obj, d := types.ObjectValueFrom(ctx, getSslAttrTypes(ctx), sslModel)
	diags.Append(d...)
	return obj, diags
}

// deferredSSLFields lists the agent_* and es_* ssl fields stored on a host.
// They are not exposed in the schema and are erased by a PUT that replaces ssl.
func deferredSSLFields(host *kbapi.ServerHost) []string {
	if host == nil || host.Ssl == nil {
		return nil
	}

	ssl := host.Ssl
	var fields []string
	for _, f := range []struct {
		name  string
		isSet bool
	}{
		{"agent_certificate", typeutils.Deref(ssl.AgentCertificate) != ""},
		{"agent_certificate_authorities", len(typeutils.Deref(ssl.AgentCertificateAuthorities)) > 0},
		{"agent_key", typeutils.Deref(ssl.AgentKey) != ""},
		{"es_certificate", typeutils.Deref(ssl.EsCertificate) != ""},
		{"es_certificate_authorities", len(typeutils.Deref(ssl.EsCertificateAuthorities)) > 0},
		{"es_key", typeutils.Deref(ssl.EsKey) != ""},
	} {
		if f.isSet {
			fields = append(fields, f.name)
		}
	}
	return fields
}
