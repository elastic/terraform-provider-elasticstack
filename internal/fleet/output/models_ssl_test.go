package output

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
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
		want    *kbapi.NewOutputSsl
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
			want: &kbapi.NewOutputSsl{
				Certificate:            utils.Pointer("cert"),
				CertificateAuthorities: &[]string{"ca"},
				Key:                    utils.Pointer("key"),
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
		want    *kbapi.UpdateOutputSsl
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
			want: &kbapi.UpdateOutputSsl{
				Certificate:            utils.Pointer("cert"),
				CertificateAuthorities: &[]string{"ca"},
				Key:                    utils.Pointer("key"),
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
		ssl *kbapi.OutputSsl
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
				ssl: nil,
			},
			want: types.ObjectNull(getSslAttrTypes()),
		},
		{
			name: "returns an object when populated",
			args: args{
				ssl: &kbapi.OutputSsl{
					Certificate:            utils.Pointer("cert"),
					CertificateAuthorities: &[]string{"ca"},
					Key:                    utils.Pointer("key"),
				},
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
			got, diags := sslToObjectValue(context.Background(), tt.args.ssl)
			if (diags.HasError()) != tt.wantErr {
				t.Errorf("sslToObjectValue() error = %v, wantErr %v", diags.HasError(), tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
