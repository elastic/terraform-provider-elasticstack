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

package agentdownloadsource

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/sslcore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// agentDownloadSourceSSLAPIModel mirrors the anonymous `ssl` object on the
// Fleet agent download source request and response bodies, which unlike the
// Fleet output `ssl` object has no verification mode.
type agentDownloadSourceSSLAPIModel struct {
	Certificate            *string
	CertificateAuthorities *[]string
	Key                    *string
	// KeyWO is the write-only client key, read from configuration. It targets
	// the same Fleet secret as Key.
	KeyWO *string
}

// sslModel is the Terraform ssl object: the shared SSL values plus the
// write-only key sibling, which is only ever non-null in configuration.
type sslModel struct {
	sslcore.Values
	KeyWO types.String `tfsdk:"key_wo"`
}

func (m sslModel) allNil() bool {
	return m.Values.AllNil() && m.KeyWO.IsNull()
}

// kbapiSsl is the anonymous `ssl` struct type generated on
// PostFleetAgentDownloadSourcesJSONBody and
// PutFleetAgentDownloadSourcesSourceidJSONBody.
type kbapiSsl = struct {
	Certificate            *string   `json:"certificate,omitempty"`
	CertificateAuthorities *[]string `json:"certificate_authorities,omitempty"`
	Key                    *string   `json:"key,omitempty"`
}

func getSslAttrTypes(ctx context.Context) map[string]attr.Type {
	return getSchema(ctx).Attributes["ssl"].GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}

func sslToKibana(ssl *agentDownloadSourceSSLAPIModel) *kbapiSsl {
	if ssl == nil {
		return nil
	}

	// The client key is never sent on the plain wire field; it travels as
	// secrets.ssl.key (see model.secretsPayload).
	return &kbapiSsl{
		Certificate:            ssl.Certificate,
		CertificateAuthorities: ssl.CertificateAuthorities,
	}
}

func objectValueToSSL(ctx context.Context, obj types.Object) (*agentDownloadSourceSSLAPIModel, diag.Diagnostics) {
	if !typeutils.IsKnown(obj) {
		return nil, nil
	}

	var diags diag.Diagnostics
	model := typeutils.ObjectTypeAs[sslModel](ctx, obj, path.Root("ssl"), &diags)
	if diags.HasError() {
		return nil, diags
	}

	// Mirror the schema-level at-least-one validator: children that were
	// unknown at plan time may resolve to all-null at apply, and sending an
	// empty ssl object would persist a present-but-empty configuration the
	// read path normalizes back to null.
	if model.allNil() {
		diags.AddAttributeError(path.Root("ssl"), sslAtLeastOneSubattrSummary, sslAtLeastOneSubattrMessage)
		return nil, diags
	}

	certificate, certificateAuthorities, key := sslcore.Extract(ctx, model.Values, path.Root("ssl").AtName("certificate_authorities"), &diags)

	return &agentDownloadSourceSSLAPIModel{
		Certificate:            certificate,
		CertificateAuthorities: certificateAuthorities,
		Key:                    key,
		KeyWO:                  model.KeyWO.ValueStringPointer(),
	}, diags
}

// objectValueToSSLUpdate converts the planned ssl object for an update
// request. Removing a previously set ssl object must send an explicit empty
// object so the Fleet API clears the stored configuration; omitting ssl
// would leave it unchanged.
func objectValueToSSLUpdate(ctx context.Context, obj types.Object, priorObj types.Object) (*agentDownloadSourceSSLAPIModel, diag.Diagnostics) {
	ssl, diags := objectValueToSSL(ctx, obj)
	if diags.HasError() {
		return nil, diags
	}
	if ssl != nil {
		return ssl, diags
	}
	if !obj.IsNull() {
		return nil, diags
	}

	if !typeutils.IsKnown(priorObj) {
		return nil, diags
	}
	return &agentDownloadSourceSSLAPIModel{}, diags
}

// sslToObjectValue hydrates an ssl object from the Fleet response's ssl
// fields, normalizing a present-but-empty response object to null so a
// cleared ssl does not drift against a practitioner-omitted attribute.
func sslToObjectValue(
	ctx context.Context,
	certificate *string,
	certificateAuthorities *[]string,
	key *string,
) (types.Object, diag.Diagnostics) {
	attrTypes := getSslAttrTypes(ctx)
	var diags diag.Diagnostics
	values := sslcore.Hydrate(ctx, certificate, certificateAuthorities, key, path.Root("ssl").AtName("certificate_authorities"), &diags)
	if values.AllNil() {
		return types.ObjectNull(attrTypes), nil
	}

	obj, diagTemp := types.ObjectValueFrom(ctx, attrTypes, sslModel{Values: values, KeyWO: types.StringNull()})
	diags.Append(diagTemp...)
	return obj, diags
}

// withSSLWriteOnlyKey returns obj with its key_wo child replaced by keyWO.
// A null or unknown obj is returned unchanged.
func withSSLWriteOnlyKey(ctx context.Context, obj types.Object, keyWO types.String) (types.Object, diag.Diagnostics) {
	if !typeutils.IsKnown(obj) {
		return obj, nil
	}
	var diags diag.Diagnostics
	model := typeutils.ObjectTypeAs[sslModel](ctx, obj, path.Root("ssl"), &diags)
	if diags.HasError() {
		return obj, diags
	}
	model.KeyWO = keyWO
	out, objDiags := types.ObjectValueFrom(ctx, getSslAttrTypes(ctx), model)
	diags.Append(objDiags...)
	return out, diags
}
