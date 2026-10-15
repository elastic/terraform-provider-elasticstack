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

// Package sslcore holds the certificate/certificate_authorities/key
// conversion logic shared by Fleet resources that expose an `ssl` block
// (currently output and agentdownloadsource), so the null-normalizing and
// pointer-to-tfsdk-value mapping only has to be correct in one place.
// Resource-specific concerns -- extra sub-attributes such as output's
// verification_mode, update "clear" semantics, and schema-level validation
// -- stay in each resource package. agentdownloadsource's ssl object is
// exactly Values. output embeds Values and adds verification_mode.
package sslcore

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Values is the tfsdk object for an ssl block whose attributes are exactly
// certificate, certificate_authorities, and key. A resource that adds
// attributes embeds Values by value, with no tfsdk tag on the embed.
// terraform-plugin-framework promotes these fields onto the parent object,
// so the tags must match the schema attribute names on every embedder.
type Values struct {
	Certificate            types.String `tfsdk:"certificate"`
	CertificateAuthorities types.List   `tfsdk:"certificate_authorities"` // > string
	Key                    types.String `tfsdk:"key"`
}

// AllNil reports whether certificate, certificate_authorities, and key are
// all null.
func (v Values) AllNil() bool {
	return v.Certificate.IsNull() && v.CertificateAuthorities.IsNull() && v.Key.IsNull()
}

// Extract reads the certificate, certificate_authorities, and key
// sub-attributes out of an already-decoded ssl object model, converting them
// to the pointer representation the Fleet API models expect.
// certificateAuthoritiesPath attributes any list-conversion diagnostics.
func Extract(ctx context.Context, values Values, certificateAuthoritiesPath path.Path, diags *diag.Diagnostics) (certificate *string, certificateAuthorities *[]string, key *string) {
	certificate = values.Certificate.ValueStringPointer()
	key = values.Key.ValueStringPointer()
	certificateAuthorities = typeutils.SliceRef(typeutils.ListTypeToSliceString(ctx, values.CertificateAuthorities, certificateAuthoritiesPath, diags))
	return certificate, certificateAuthorities, key
}

// Hydrate builds the tfsdk-typed certificate, certificate_authorities, and
// key values from a Fleet API response's ssl fields, normalizing an empty
// certificate_authorities list to null the same way a practitioner-omitted
// attribute would decode. Callers use Values.AllNil to collapse a
// present-but-empty ssl object down to ObjectNull.
func Hydrate(ctx context.Context, certificate *string, certificateAuthorities *[]string, key *string, certificateAuthoritiesPath path.Path, diags *diag.Diagnostics) Values {
	var values Values
	values.Certificate = typeutils.NonEmptyStringishPointerValue(certificate)
	values.Key = typeutils.NonEmptyStringishPointerValue(key)

	if cas := typeutils.Deref(certificateAuthorities); len(cas) > 0 {
		values.CertificateAuthorities = typeutils.SliceToListTypeString(ctx, cas, certificateAuthoritiesPath, diags)
	} else {
		values.CertificateAuthorities = types.ListNull(types.StringType)
	}

	return values
}
