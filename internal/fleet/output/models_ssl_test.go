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
			name: "returns an ssl object when populated",
			args: args{
				obj: types.ObjectValueMust(
					getSslAttrTypes(),
					map[string]attr.Value{
						"certificate_authorities": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("ca")}),
						"certificate":             types.StringValue("cert"),
						"key":                     types.StringValue("key"),
					},
				),
			},
			want: &outputSSLAPIModel{
				Certificate:            new("cert"),
				CertificateAuthorities: &[]string{"ca"},
				Key:                    new("key"),
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
			name: "returns an ssl object when populated",
			args: args{
				obj: types.ObjectValueMust(
					getSslAttrTypes(),
					map[string]attr.Value{
						"certificate_authorities": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("ca")}),
						"certificate":             types.StringValue("cert"),
						"key":                     types.StringValue("key"),
					},
				),
			},
			want: &outputSSLAPIModel{
				Certificate:            new("cert"),
				CertificateAuthorities: &[]string{"ca"},
				Key:                    new("key"),
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
			},
			want: types.ObjectNull(getSslAttrTypes()),
		},
		{
			name: "returns null object when ssl has all empty fields",
			args: args{
				certificate:            nil,
				certificateAuthorities: nil,
				key:                    nil,
			},
			want: types.ObjectNull(getSslAttrTypes()),
		},
		{
			name: "returns null object when ssl has empty string pointers and empty slice",
			args: args{
				certificate:            new(""),
				certificateAuthorities: &[]string{},
				key:                    new(""),
			},
			want: types.ObjectNull(getSslAttrTypes()),
		},
		{
			name: "returns an object when populated",
			args: args{
				certificate:            new("cert"),
				certificateAuthorities: &[]string{"ca"},
				key:                    new("key"),
			},
			want: types.ObjectValueMust(
				getSslAttrTypes(),
				map[string]attr.Value{
					"certificate_authorities": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("ca")}),
					"certificate":             types.StringValue("cert"),
					"key":                     types.StringValue("key"),
				},
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, diags := sslToObjectValue(context.Background(), tt.args.certificate, tt.args.certificateAuthorities, tt.args.key)
			if (diags.HasError()) != tt.wantErr {
				t.Errorf("sslToObjectValue() error = %v, wantErr %v", diags.HasError(), tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_normalizeSSLFromPlan(t *testing.T) {
	mappedSSL := types.ObjectValueMust(
		getSslAttrTypes(),
		map[string]attr.Value{
			"certificate_authorities": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("ca")}),
			"certificate":             types.StringValue("cert"),
			"key":                     types.StringValue("key"),
		},
	)

	tests := []struct {
		name       string
		plannedSSL types.Object
		mappedSSL  types.Object
		want       types.Object
	}{
		{
			name:       "returns null when planned ssl is null",
			plannedSSL: types.ObjectNull(getSslAttrTypes()),
			mappedSSL:  mappedSSL,
			want:       types.ObjectNull(getSslAttrTypes()),
		},
		{
			name:       "keeps mapped value when planned ssl is unknown",
			plannedSSL: types.ObjectUnknown(getSslAttrTypes()),
			mappedSSL:  mappedSSL,
			want:       mappedSSL,
		},
		{
			name: "keeps mapped value when planned ssl configured",
			plannedSSL: types.ObjectValueMust(
				getSslAttrTypes(),
				map[string]attr.Value{
					"certificate_authorities": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("ca")}),
					"certificate":             types.StringValue("configured-cert"),
					"key":                     types.StringValue("configured-key"),
				},
			),
			mappedSSL: mappedSSL,
			want:      mappedSSL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeSSLFromPlan(tt.plannedSSL, tt.mappedSSL)
			assert.Equal(t, tt.want, got)
		})
	}
}
