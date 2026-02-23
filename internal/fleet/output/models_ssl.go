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

func objectValueToSSL(ctx context.Context, obj types.Object) (*kbapi.NewOutputSsl, diag.Diagnostics) {
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

	return &kbapi.NewOutputSsl{
		Certificate:            sslModel.Certificate.ValueStringPointer(),
		CertificateAuthorities: schemautil.SliceRef(typeutils.ListTypeToSliceString(ctx, sslModel.CertificateAuthorities, path.Root("certificate_authorities"), &diags)),
		Key:                    sslModel.Key.ValueStringPointer(),
	}, diags
}

func objectValueToSSLUpdate(ctx context.Context, obj types.Object) (*kbapi.UpdateOutputSsl, diag.Diagnostics) {
	ssl, diags := objectValueToSSL(ctx, obj)
	if diags.HasError() || ssl == nil {
		return nil, diags
	}

	return &kbapi.UpdateOutputSsl{
		Certificate:            ssl.Certificate,
		CertificateAuthorities: ssl.CertificateAuthorities,
		Key:                    ssl.Key,
	}, diags
}

func sslToObjectValue(ctx context.Context, ssl *kbapi.OutputSsl) (types.Object, diag.Diagnostics) {
	if ssl == nil {
		return types.ObjectNull(getSslAttrTypes()), nil
	}

	var diags diag.Diagnostics
	sslModel := outputSslModel{
		Certificate: typeutils.NonEmptyStringishPointerValue(ssl.Certificate),
		Key:         typeutils.NonEmptyStringishPointerValue(ssl.Key),
	}

	if cas := schemautil.Deref(ssl.CertificateAuthorities); len(cas) > 0 {
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
