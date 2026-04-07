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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type outputSslModel struct {
	CertificateAuthorities types.List   `tfsdk:"certificate_authorities"` // > string
	Certificate            types.String `tfsdk:"certificate"`
	Key                    types.String `tfsdk:"key"`
}

type outputSSLAPIModel struct {
	Certificate            *string
	CertificateAuthorities *[]string
	Key                    *string
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

	return &outputSSLAPIModel{
		Certificate:            sslModel.Certificate.ValueStringPointer(),
		CertificateAuthorities: schemautil.SliceRef(typeutils.ListTypeToSliceString(ctx, sslModel.CertificateAuthorities, path.Root("certificate_authorities"), &diags)),
		Key:                    sslModel.Key.ValueStringPointer(),
	}, diags
}

func objectValueToSSLUpdate(ctx context.Context, obj types.Object) (*outputSSLAPIModel, diag.Diagnostics) {
	ssl, diags := objectValueToSSL(ctx, obj)
	if diags.HasError() || ssl == nil {
		return nil, diags
	}

	return ssl, diags
}

func (ssl *outputSSLAPIModel) toCreateElasticsearch() *struct {
	Certificate            *string                                                        `json:"certificate,omitempty"`
	CertificateAuthorities *[]string                                                      `json:"certificate_authorities,omitempty"`
	Key                    *string                                                        `json:"key,omitempty"`
	VerificationMode       *kbapi.KibanaHTTPAPIsNewOutputElasticsearchSslVerificationMode `json:"verification_mode,omitempty"`
} {
	if ssl == nil {
		return nil
	}

	return &struct {
		Certificate            *string                                                        `json:"certificate,omitempty"`
		CertificateAuthorities *[]string                                                      `json:"certificate_authorities,omitempty"`
		Key                    *string                                                        `json:"key,omitempty"`
		VerificationMode       *kbapi.KibanaHTTPAPIsNewOutputElasticsearchSslVerificationMode `json:"verification_mode,omitempty"`
	}{
		Certificate:            ssl.Certificate,
		CertificateAuthorities: ssl.CertificateAuthorities,
		Key:                    ssl.Key,
	}
}

func (ssl *outputSSLAPIModel) toCreateKafka() *struct {
	Certificate            *string                                                `json:"certificate,omitempty"`
	CertificateAuthorities *[]string                                              `json:"certificate_authorities,omitempty"`
	Key                    *string                                                `json:"key,omitempty"`
	VerificationMode       *kbapi.KibanaHTTPAPIsNewOutputKafkaSslVerificationMode `json:"verification_mode,omitempty"`
} {
	if ssl == nil {
		return nil
	}

	return &struct {
		Certificate            *string                                                `json:"certificate,omitempty"`
		CertificateAuthorities *[]string                                              `json:"certificate_authorities,omitempty"`
		Key                    *string                                                `json:"key,omitempty"`
		VerificationMode       *kbapi.KibanaHTTPAPIsNewOutputKafkaSslVerificationMode `json:"verification_mode,omitempty"`
	}{
		Certificate:            ssl.Certificate,
		CertificateAuthorities: ssl.CertificateAuthorities,
		Key:                    ssl.Key,
	}
}

func (ssl *outputSSLAPIModel) toCreateLogstash() *struct {
	Certificate            *string                                                   `json:"certificate,omitempty"`
	CertificateAuthorities *[]string                                                 `json:"certificate_authorities,omitempty"`
	Key                    *string                                                   `json:"key,omitempty"`
	VerificationMode       *kbapi.KibanaHTTPAPIsNewOutputLogstashSslVerificationMode `json:"verification_mode,omitempty"`
} {
	if ssl == nil {
		return nil
	}

	return &struct {
		Certificate            *string                                                   `json:"certificate,omitempty"`
		CertificateAuthorities *[]string                                                 `json:"certificate_authorities,omitempty"`
		Key                    *string                                                   `json:"key,omitempty"`
		VerificationMode       *kbapi.KibanaHTTPAPIsNewOutputLogstashSslVerificationMode `json:"verification_mode,omitempty"`
	}{
		Certificate:            ssl.Certificate,
		CertificateAuthorities: ssl.CertificateAuthorities,
		Key:                    ssl.Key,
	}
}

func (ssl *outputSSLAPIModel) toUpdateElasticsearch() *struct {
	Certificate            *string                                             `json:"certificate,omitempty"`
	CertificateAuthorities *[]string                                           `json:"certificate_authorities,omitempty"`
	Key                    *string                                             `json:"key,omitempty"`
	VerificationMode       *kbapi.UpdateOutputElasticsearchSslVerificationMode `json:"verification_mode,omitempty"`
} {
	if ssl == nil {
		return nil
	}

	return &struct {
		Certificate            *string                                             `json:"certificate,omitempty"`
		CertificateAuthorities *[]string                                           `json:"certificate_authorities,omitempty"`
		Key                    *string                                             `json:"key,omitempty"`
		VerificationMode       *kbapi.UpdateOutputElasticsearchSslVerificationMode `json:"verification_mode,omitempty"`
	}{
		Certificate:            ssl.Certificate,
		CertificateAuthorities: ssl.CertificateAuthorities,
		Key:                    ssl.Key,
	}
}

func (ssl *outputSSLAPIModel) toUpdateKafka() *struct {
	Certificate            *string                                     `json:"certificate,omitempty"`
	CertificateAuthorities *[]string                                   `json:"certificate_authorities,omitempty"`
	Key                    *string                                     `json:"key,omitempty"`
	VerificationMode       *kbapi.UpdateOutputKafkaSslVerificationMode `json:"verification_mode,omitempty"`
} {
	if ssl == nil {
		return nil
	}

	return &struct {
		Certificate            *string                                     `json:"certificate,omitempty"`
		CertificateAuthorities *[]string                                   `json:"certificate_authorities,omitempty"`
		Key                    *string                                     `json:"key,omitempty"`
		VerificationMode       *kbapi.UpdateOutputKafkaSslVerificationMode `json:"verification_mode,omitempty"`
	}{
		Certificate:            ssl.Certificate,
		CertificateAuthorities: ssl.CertificateAuthorities,
		Key:                    ssl.Key,
	}
}

func (ssl *outputSSLAPIModel) toUpdateLogstash() *struct {
	Certificate            *string                                        `json:"certificate,omitempty"`
	CertificateAuthorities *[]string                                      `json:"certificate_authorities,omitempty"`
	Key                    *string                                        `json:"key,omitempty"`
	VerificationMode       *kbapi.UpdateOutputLogstashSslVerificationMode `json:"verification_mode,omitempty"`
} {
	if ssl == nil {
		return nil
	}

	return &struct {
		Certificate            *string                                        `json:"certificate,omitempty"`
		CertificateAuthorities *[]string                                      `json:"certificate_authorities,omitempty"`
		Key                    *string                                        `json:"key,omitempty"`
		VerificationMode       *kbapi.UpdateOutputLogstashSslVerificationMode `json:"verification_mode,omitempty"`
	}{
		Certificate:            ssl.Certificate,
		CertificateAuthorities: ssl.CertificateAuthorities,
		Key:                    ssl.Key,
	}
}

func sslToObjectValue(ctx context.Context, certificate *string, certificateAuthorities *[]string, key *string) (types.Object, diag.Diagnostics) {
	if certificate == nil && certificateAuthorities == nil && key == nil {
		return types.ObjectNull(getSslAttrTypes()), nil
	}

	var diags diag.Diagnostics
	sslModel := outputSslModel{
		Certificate: typeutils.NonEmptyStringishPointerValue(certificate),
		Key:         typeutils.NonEmptyStringishPointerValue(key),
	}

	if cas := schemautil.Deref(certificateAuthorities); len(cas) > 0 {
		sslModel.CertificateAuthorities = typeutils.SliceToListTypeString(ctx, cas, path.Root("ssl").AtName("certificate_authorities"), &diags)
	} else {
		sslModel.CertificateAuthorities = types.ListNull(types.StringType)
	}

	if sslModel.CertificateAuthorities.IsNull() && sslModel.Certificate.IsNull() && sslModel.Key.IsNull() {
		return types.ObjectNull(getSslAttrTypes()), nil
	}

	obj, diagTemp := types.ObjectValueFrom(ctx, getSslAttrTypes(), sslModel)
	diags.Append(diagTemp...)
	return obj, diags
}

// normalizeSSLFromPlan keeps ssl explicitly null when users omit it in configuration.
// This avoids post-apply inconsistencies for nested sensitive values when APIs return
// partial SSL objects while plan did not configure SSL.
func normalizeSSLFromPlan(plannedSSL types.Object, mappedSSL types.Object) types.Object {
	if plannedSSL.IsNull() {
		return types.ObjectNull(getSslAttrTypes())
	}
	return mappedSSL
}
