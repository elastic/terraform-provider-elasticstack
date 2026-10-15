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
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func Test_objectValueToSSL(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		obj     types.Object
		want    *agentDownloadSourceSSLAPIModel
		wantErr bool
	}{
		{
			name: "returns nil when object is null",
			obj:  types.ObjectNull(getSslAttrTypes(ctx)),
		},
		{
			name: "returns nil when object is unknown",
			obj:  types.ObjectUnknown(getSslAttrTypes(ctx)),
		},
		{
			name: "returns an ssl object when populated",
			obj: types.ObjectValueMust(
				getSslAttrTypes(ctx),
				map[string]attr.Value{
					"certificate_authorities": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("ca")}),
					"certificate":             types.StringValue("cert"),
					"key":                     types.StringValue("key"),
				},
			),
			want: &agentDownloadSourceSSLAPIModel{
				Certificate:            new("cert"),
				CertificateAuthorities: &[]string{"ca"},
				Key:                    new("key"),
			},
		},
		{
			name: "returns only a CA list when only certificate_authorities is set",
			obj: types.ObjectValueMust(
				getSslAttrTypes(ctx),
				map[string]attr.Value{
					"certificate_authorities": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("ca")}),
					"certificate":             types.StringNull(),
					"key":                     types.StringNull(),
				},
			),
			want: &agentDownloadSourceSSLAPIModel{
				CertificateAuthorities: &[]string{"ca"},
			},
		},
		{
			name: "returns only certificate and key when certificate_authorities is unset",
			obj: types.ObjectValueMust(
				getSslAttrTypes(ctx),
				map[string]attr.Value{
					"certificate_authorities": types.ListNull(types.StringType),
					"certificate":             types.StringValue("cert"),
					"key":                     types.StringValue("key"),
				},
			),
			want: &agentDownloadSourceSSLAPIModel{
				Certificate: new("cert"),
				Key:         new("key"),
			},
		},
		{
			name: "known all-null object is rejected",
			obj: types.ObjectValueMust(
				getSslAttrTypes(ctx),
				map[string]attr.Value{
					"certificate_authorities": types.ListNull(types.StringType),
					"certificate":             types.StringNull(),
					"key":                     types.StringNull(),
				},
			),
			wantErr: true,
		},
		{
			name: "unknown child defers rejection",
			obj: types.ObjectValueMust(
				getSslAttrTypes(ctx),
				map[string]attr.Value{
					"certificate_authorities": types.ListUnknown(types.StringType),
					"certificate":             types.StringNull(),
					"key":                     types.StringUnknown(),
				},
			),
			// Unknown children defer the all-null rejection; existing
			// semantics encode an unknown string as an empty-string pointer.
			want: &agentDownloadSourceSSLAPIModel{Key: new("")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, diags := objectValueToSSL(ctx, tt.obj)
			if (diags.HasError()) != tt.wantErr {
				t.Errorf("objectValueToSSL() error = %v, wantErr %v", diags, tt.wantErr)
				return
			}
			if tt.wantErr {
				assert.Equal(t, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(path.Root("ssl"), sslAtLeastOneSubattrSummary, sslAtLeastOneSubattrMessage)}, diags)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_objectValueToSSLUpdate(t *testing.T) {
	ctx := context.Background()
	attrTypes := getSslAttrTypes(ctx)
	sslObject := types.ObjectValueMust(
		attrTypes,
		map[string]attr.Value{
			"certificate_authorities": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("ca")}),
			"certificate":             types.StringValue("cert"),
			"key":                     types.StringValue("key"),
		},
	)

	tests := []struct {
		name     string
		obj      types.Object
		priorObj types.Object
		want     *agentDownloadSourceSSLAPIModel
		wantErr  bool
	}{
		{
			name:     "returns nil when object and prior are null",
			obj:      types.ObjectNull(attrTypes),
			priorObj: types.ObjectNull(attrTypes),
		},
		{
			name:     "returns nil when object is unknown",
			obj:      types.ObjectUnknown(attrTypes),
			priorObj: types.ObjectNull(attrTypes),
		},
		{
			name:     "returns an ssl object when populated",
			obj:      sslObject,
			priorObj: types.ObjectNull(attrTypes),
			want: &agentDownloadSourceSSLAPIModel{
				Certificate:            new("cert"),
				CertificateAuthorities: &[]string{"ca"},
				Key:                    new("key"),
			},
		},
		{
			name:     "returns an explicit empty object when removing a previously set object",
			obj:      types.ObjectNull(attrTypes),
			priorObj: sslObject,
			want:     &agentDownloadSourceSSLAPIModel{},
		},
		{
			name: "known all-null object is rejected",
			obj: types.ObjectValueMust(attrTypes, map[string]attr.Value{
				"certificate_authorities": types.ListNull(types.StringType),
				"certificate":             types.StringNull(),
				"key":                     types.StringNull(),
			}),
			priorObj: types.ObjectNull(attrTypes),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, diags := objectValueToSSLUpdate(ctx, tt.obj, tt.priorObj)
			if (diags.HasError()) != tt.wantErr {
				t.Errorf("objectValueToSSLUpdate() error = %v, wantErr %v", diags, tt.wantErr)
				return
			}
			if tt.wantErr {
				assert.Equal(t, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(path.Root("ssl"), sslAtLeastOneSubattrSummary, sslAtLeastOneSubattrMessage)}, diags)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_sslToObjectValue(t *testing.T) {
	ctx := context.Background()
	attrTypes := getSslAttrTypes(ctx)

	tests := []struct {
		name                 string
		certificate          *string
		certificateAuthority *[]string
		key                  *string
		want                 types.Object
	}{
		{
			name: "returns null object when ssl is absent",
			want: types.ObjectNull(attrTypes),
		},
		{
			name:                 "returns null object when ssl is present but empty",
			certificate:          new(""),
			certificateAuthority: &[]string{},
			key:                  new(""),
			want:                 types.ObjectNull(attrTypes),
		},
		{
			name:                 "returns a populated object",
			certificate:          new("cert"),
			certificateAuthority: &[]string{"ca"},
			key:                  new("key"),
			want: types.ObjectValueMust(
				attrTypes,
				map[string]attr.Value{
					"certificate_authorities": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("ca")}),
					"certificate":             types.StringValue("cert"),
					"key":                     types.StringValue("key"),
				},
			),
		},
		{
			name:                 "returns an object with only certificate_authorities",
			certificateAuthority: &[]string{"ca"},
			want: types.ObjectValueMust(
				attrTypes,
				map[string]attr.Value{
					"certificate_authorities": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("ca")}),
					"certificate":             types.StringNull(),
					"key":                     types.StringNull(),
				},
			),
		},
		{
			name:        "returns an object with certificate and key and a null CA list",
			certificate: new("cert"),
			key:         new("key"),
			want: types.ObjectValueMust(
				attrTypes,
				map[string]attr.Value{
					"certificate_authorities": types.ListNull(types.StringType),
					"certificate":             types.StringValue("cert"),
					"key":                     types.StringValue("key"),
				},
			),
		},
		{
			name:                 "normalizes an empty CA list to a null list",
			certificate:          new("cert"),
			certificateAuthority: &[]string{},
			want: types.ObjectValueMust(
				attrTypes,
				map[string]attr.Value{
					"certificate_authorities": types.ListNull(types.StringType),
					"certificate":             types.StringValue("cert"),
					"key":                     types.StringNull(),
				},
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, diags := sslToObjectValue(ctx, tt.certificate, tt.certificateAuthority, tt.key)
			if diags.HasError() {
				t.Errorf("sslToObjectValue() error = %v", diags)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_sslRoundTrip(t *testing.T) {
	ctx := context.Background()
	configObject := types.ObjectValueMust(
		getSslAttrTypes(ctx),
		map[string]attr.Value{
			"certificate_authorities": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("ca-1"), types.StringValue("ca-2")}),
			"certificate":             types.StringValue("cert"),
			"key":                     types.StringValue("key"),
		},
	)

	apiSSL, diags := objectValueToSSL(ctx, configObject)
	if diags.HasError() {
		t.Fatalf("objectValueToSSL() error = %v", diags)
	}

	requestBody := sslToKibana(apiSSL)
	hydrated, diags := sslToObjectValue(ctx, requestBody.Certificate, requestBody.CertificateAuthorities, requestBody.Key)
	if diags.HasError() {
		t.Fatalf("sslToObjectValue() error = %v", diags)
	}

	assert.Equal(t, configObject, hydrated)
}
