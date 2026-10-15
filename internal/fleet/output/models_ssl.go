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

package output

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/sslcore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type outputSslModel struct {
	CertificateAuthorities types.List   `tfsdk:"certificate_authorities"` // > string
	Certificate            types.String `tfsdk:"certificate"`
	Key                    types.String `tfsdk:"key"`
	VerificationMode       types.String `tfsdk:"verification_mode"`
}

type outputSSLAPIModel struct {
	Certificate            *string
	CertificateAuthorities *[]string
	Key                    *string
	VerificationMode       *kbapi.KibanaHTTPAPIsOutputSslVerificationMode
}

func objectValueToSSL(ctx context.Context, obj types.Object) (*outputSSLAPIModel, diag.Diagnostics) {
	if !typeutils.IsKnown(obj) {
		return nil, nil
	}

	var diags diag.Diagnostics
	sslModel := typeutils.ObjectTypeAs[outputSslModel](ctx, obj, path.Root("ssl"), &diags)
	if diags.HasError() {
		return nil, diags
	}

	if sslModel == nil {
		return nil, diags
	}

	var verificationMode *kbapi.KibanaHTTPAPIsOutputSslVerificationMode
	if typeutils.IsKnown(sslModel.VerificationMode) {
		mode := kbapi.KibanaHTTPAPIsOutputSslVerificationMode(sslModel.VerificationMode.ValueString())
		verificationMode = &mode
	}

	certificate, certificateAuthorities, key := sslcore.Extract(ctx, sslcore.Values{
		Certificate:            sslModel.Certificate,
		CertificateAuthorities: sslModel.CertificateAuthorities,
		Key:                    sslModel.Key,
	}, path.Root("certificate_authorities"), &diags)

	return &outputSSLAPIModel{
		Certificate:            certificate,
		CertificateAuthorities: certificateAuthorities,
		Key:                    key,
		VerificationMode:       verificationMode,
	}, diags
}

// objectValueToSSLUpdate converts the planned ssl object for an update
// request. Unlike agentdownloadsource's equivalent, it does not send an
// explicit empty ssl object when a previously set ssl block is removed from
// config; that clearing-semantics gap is a pre-existing, documented
// difference between the two resources rather than an oversight introduced
// by this refactor.
func objectValueToSSLUpdate(ctx context.Context, obj types.Object) (*outputSSLAPIModel, diag.Diagnostics) {
	ssl, diags := objectValueToSSL(ctx, obj)
	if diags.HasError() || ssl == nil {
		return nil, diags
	}

	return ssl, diags
}

func (ssl *outputSSLAPIModel) toAPI() *kbapi.KibanaHTTPAPIsOutputSsl {
	if ssl == nil {
		return nil
	}

	return &kbapi.KibanaHTTPAPIsOutputSsl{
		Certificate:            ssl.Certificate,
		CertificateAuthorities: ssl.CertificateAuthorities,
		Key:                    ssl.Key,
		VerificationMode:       ssl.VerificationMode,
	}
}

func sslToObjectValue(
	ctx context.Context,
	certificate *string,
	certificateAuthorities *[]string,
	key *string,
	verificationMode *kbapi.KibanaHTTPAPIsOutputSslVerificationMode,
) (types.Object, diag.Diagnostics) {
	if sslcore.AllNil(certificate, certificateAuthorities, key) && verificationMode == nil {
		return types.ObjectNull(getSslAttrTypes(ctx)), nil
	}

	var diags diag.Diagnostics
	values, allNull := sslcore.Hydrate(ctx, certificate, certificateAuthorities, key, path.Root("ssl").AtName("certificate_authorities"), &diags)

	sslModel := outputSslModel{
		Certificate:            values.Certificate,
		CertificateAuthorities: values.CertificateAuthorities,
		Key:                    values.Key,
		VerificationMode:       typeutils.StringishPointerValue(verificationMode),
	}

	if allNull && sslModel.VerificationMode.IsNull() {
		return types.ObjectNull(getSslAttrTypes(ctx)), nil
	}

	obj, diagTemp := types.ObjectValueFrom(ctx, getSslAttrTypes(ctx), sslModel)
	diags.Append(diagTemp...)
	return obj, diags
}
