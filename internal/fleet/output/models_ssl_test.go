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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func Test_objectValueToSSL(t *testing.T) {
	type args struct {
		obj types.Object
	}
	tests := []struct {
		name    string
		args    args
		want    *outputSSLAPIModel
		wantErr bool
	}{
		{
			name: "returns nil when object is unknown",
			args: args{
				obj: types.ObjectUnknown(getSslAttrTypes()),
			},
		},
		{
			name: "returns an ssl object when populated without verification mode",
			args: args{
				obj: types.ObjectValueMust(
					getSslAttrTypes(),
					map[string]attr.Value{
						"certificate_authorities": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("ca")}),
						"certificate":             types.StringValue("cert"),
						"key":                     types.StringValue("key"),
						"verification_mode":       types.StringNull(),
					},
				),
			},
			want: &outputSSLAPIModel{
				Certificate:            new("cert"),
				CertificateAuthorities: &[]string{"ca"},
				Key:                    new("key"),
				VerificationMode:       nil,
			},
		},
		{
			name: "returns verification mode when populated",
			args: args{
				obj: types.ObjectValueMust(
					getSslAttrTypes(),
					map[string]attr.Value{
						"certificate_authorities": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("ca")}),
						"certificate":             types.StringValue("cert"),
						"key":                     types.StringValue("key"),
						"verification_mode":       types.StringValue("none"),
					},
				),
			},
			want: &outputSSLAPIModel{
				Certificate:            new("cert"),
				CertificateAuthorities: &[]string{"ca"},
				Key:                    new("key"),
				VerificationMode:       new(kbapi.KibanaHTTPAPIsOutputSslVerificationModeNone),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, diags := objectValueToSSL(context.Background(), tt.args.obj)
			if (diags.HasError()) != tt.wantErr {
				t.Errorf("objectValueToSSL() error = %v, wantErr %v", diags.HasError(), tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_objectValueToSSLUpdate(t *testing.T) {
	type args struct {
		obj types.Object
	}
	tests := []struct {
		name    string
		args    args
		want    *outputSSLAPIModel
		wantErr bool
	}{
		{
			name: "returns nil when object is unknown",
			args: args{
				obj: types.ObjectUnknown(getSslAttrTypes()),
			},
		},
		{
			name: "returns an ssl object when populated with verification mode",
			args: args{
				obj: types.ObjectValueMust(
					getSslAttrTypes(),
					map[string]attr.Value{
						"certificate_authorities": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("ca")}),
						"certificate":             types.StringValue("cert"),
						"key":                     types.StringValue("key"),
						"verification_mode":       types.StringValue("none"),
					},
				),
			},
			want: &outputSSLAPIModel{
				Certificate:            new("cert"),
				CertificateAuthorities: &[]string{"ca"},
				Key:                    new("key"),
				VerificationMode:       new(kbapi.KibanaHTTPAPIsOutputSslVerificationModeNone),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, diags := objectValueToSSLUpdate(context.Background(), tt.args.obj)
			if (diags.HasError()) != tt.wantErr {
				t.Errorf("objectValueToSSLUpdate() error = %v, wantErr %v", diags.HasError(), tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_sslToObjectValue(t *testing.T) {
	type args struct {
		certificate            *string
		certificateAuthorities *[]string
		key                    *string
		verificationMode       *kbapi.KibanaHTTPAPIsOutputSslVerificationMode
	}
	tests := []struct {
		name    string
		args    args
		want    types.Object
		wantErr bool
	}{
		{
			name: "returns nil when ssl is nil",
			args: args{
				certificate:            nil,
				certificateAuthorities: nil,
				key:                    nil,
				verificationMode:       nil,
			},
			want: types.ObjectNull(getSslAttrTypes()),
		},
		{
			name: "returns null object when ssl has all empty fields",
			args: args{
				certificate:            nil,
				certificateAuthorities: nil,
				key:                    nil,
				verificationMode:       nil,
			},
			want: types.ObjectNull(getSslAttrTypes()),
		},
		{
			name: "returns null object when ssl has empty string pointers and empty slice",
			args: args{
				certificate:            new(""),
				certificateAuthorities: &[]string{},
				key:                    new(""),
				verificationMode:       nil,
			},
			want: types.ObjectNull(getSslAttrTypes()),
		},
		{
			name: "returns an object when populated with nil verification mode",
			args: args{
				certificate:            new("cert"),
				certificateAuthorities: &[]string{"ca"},
				key:                    new("key"),
				verificationMode:       nil,
			},
			want: types.ObjectValueMust(
				getSslAttrTypes(),
				map[string]attr.Value{
					"certificate_authorities": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("ca")}),
					"certificate":             types.StringValue("cert"),
					"key":                     types.StringValue("key"),
					"verification_mode":       types.StringNull(),
				},
			),
		},
		{
			name: "returns an object when verification mode is populated",
			args: args{
				certificate:            new("cert"),
				certificateAuthorities: &[]string{"ca"},
				key:                    new("key"),
				verificationMode:       new(kbapi.KibanaHTTPAPIsOutputSslVerificationModeNone),
			},
			want: types.ObjectValueMust(
				getSslAttrTypes(),
				map[string]attr.Value{
					"certificate_authorities": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("ca")}),
					"certificate":             types.StringValue("cert"),
					"key":                     types.StringValue("key"),
					"verification_mode":       types.StringValue("none"),
				},
			),
		},
		{
			name: "returns an object when only verification mode is populated",
			args: args{
				certificate:            nil,
				certificateAuthorities: nil,
				key:                    nil,
				verificationMode:       new(kbapi.KibanaHTTPAPIsOutputSslVerificationModeNone),
			},
			want: types.ObjectValueMust(
				getSslAttrTypes(),
				map[string]attr.Value{
					"certificate_authorities": types.ListNull(types.StringType),
					"certificate":             types.StringNull(),
					"key":                     types.StringNull(),
					"verification_mode":       types.StringValue("none"),
				},
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, diags := sslToObjectValue(context.Background(), tt.args.certificate, tt.args.certificateAuthorities, tt.args.key, tt.args.verificationMode)
			if (diags.HasError()) != tt.wantErr {
				t.Errorf("sslToObjectValue() error = %v, wantErr %v", diags.HasError(), tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
