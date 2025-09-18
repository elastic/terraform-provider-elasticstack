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

func Test_outputKafkaModel_toAPIHash(t *testing.T) {
	type fields struct {
		Hash types.Object
	}
	tests := []struct {
		name   string
		fields fields
		want   *struct {
			Hash   *string `json:"hash,omitempty"`
			Random *bool   `json:"random,omitempty"`
		}
		wantErr bool
	}{
		{
			name: "returns nil when hash is unknown",
			fields: fields{
				Hash: types.ObjectUnknown(getHashAttrTypes()),
			},
		},
		{
			name: "returns a hash object when all fields are set",
			fields: fields{
				Hash: types.ObjectValueMust(
					getHashAttrTypes(),
					map[string]attr.Value{
						"hash":   types.StringValue("field"),
						"random": types.BoolValue(true),
					},
				),
			},
			want: &struct {
				Hash   *string `json:"hash,omitempty"`
				Random *bool   `json:"random,omitempty"`
			}{
				Hash:   utils.Pointer("field"),
				Random: utils.Pointer(true),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := outputKafkaModel{
				Hash: tt.fields.Hash,
			}
			got, diags := m.toAPIHash(context.Background())
			if (diags.HasError()) != tt.wantErr {
				t.Errorf("outputKafkaModel.toAPIHash() error = %v, wantErr %v", diags.HasError(), tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_outputKafkaModel_toAPIHeaders(t *testing.T) {
	type fields struct {
		Headers types.List
	}
	tests := []struct {
		name   string
		fields fields
		want   *[]struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}
		wantErr bool
	}{
		{
			name: "returns nil when headers are unknown",
			fields: fields{
				Headers: types.ListUnknown(getHeadersAttrTypes()),
			},
		},
		{
			name: "returns headers when populated",
			fields: fields{
				Headers: types.ListValueMust(
					getHeadersAttrTypes(),
					[]attr.Value{
						types.ObjectValueMust(getHeadersAttrTypes().(types.ObjectType).AttrTypes, map[string]attr.Value{
							"key":   types.StringValue("key-1"),
							"value": types.StringValue("value-1"),
						}),
						types.ObjectValueMust(getHeadersAttrTypes().(types.ObjectType).AttrTypes, map[string]attr.Value{
							"key":   types.StringValue("key-2"),
							"value": types.StringValue("value-2"),
						}),
					},
				),
			},
			want: &[]struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			}{
				{Key: "key-1", Value: "value-1"},
				{Key: "key-2", Value: "value-2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := outputKafkaModel{
				Headers: tt.fields.Headers,
			}
			got, diags := m.toAPIHeaders(context.Background())
			if (diags.HasError()) != tt.wantErr {
				t.Errorf("outputKafkaModel.toAPIHeaders() error = %v, wantErr %v", diags.HasError(), tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_outputKafkaModel_toAPIRandom(t *testing.T) {
	type fields struct {
		Random types.Object
	}
	tests := []struct {
		name   string
		fields fields
		want   *struct {
			GroupEvents *float32 `json:"group_events,omitempty"`
		}
		wantErr bool
	}{
		{
			name: "returns nil when random is unknown",
			fields: fields{
				Random: types.ObjectUnknown(getRandomAttrTypes()),
			},
		},
		{
			name: "returns a random object when populated",
			fields: fields{
				Random: types.ObjectValueMust(
					getRandomAttrTypes(),
					map[string]attr.Value{
						"group_events": types.Float64Value(1),
					},
				),
			},
			want: &struct {
				GroupEvents *float32 `json:"group_events,omitempty"`
			}{
				GroupEvents: utils.Pointer(float32(1)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := outputKafkaModel{
				Random: tt.fields.Random,
			}
			got, diags := m.toAPIRandom(context.Background())
			if (diags.HasError()) != tt.wantErr {
				t.Errorf("outputKafkaModel.toAPIRandom() error = %v, wantErr %v", diags.HasError(), tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_outputKafkaModel_toAPIRoundRobin(t *testing.T) {
	type fields struct {
		RoundRobin types.Object
	}
	tests := []struct {
		name   string
		fields fields
		want   *struct {
			GroupEvents *float32 `json:"group_events,omitempty"`
		}
		wantErr bool
	}{
		{
			name: "returns nil when round_robin is unknown",
			fields: fields{
				RoundRobin: types.ObjectUnknown(getRoundRobinAttrTypes()),
			},
		},
		{
			name: "returns a round_robin object when populated",
			fields: fields{
				RoundRobin: types.ObjectValueMust(
					getRoundRobinAttrTypes(),
					map[string]attr.Value{
						"group_events": types.Float64Value(1),
					},
				),
			},
			want: &struct {
				GroupEvents *float32 `json:"group_events,omitempty"`
			}{
				GroupEvents: utils.Pointer(float32(1)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := outputKafkaModel{
				RoundRobin: tt.fields.RoundRobin,
			}
			got, diags := m.toAPIRoundRobin(context.Background())
			if (diags.HasError()) != tt.wantErr {
				t.Errorf("outputKafkaModel.toAPIRoundRobin() error = %v, wantErr %v", diags.HasError(), tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_outputKafkaModel_toAPISasl(t *testing.T) {
	type fields struct {
		Sasl types.Object
	}
	tests := []struct {
		name   string
		fields fields
		want   *struct {
			Mechanism *kbapi.NewOutputKafkaSaslMechanism `json:"mechanism,omitempty"`
		}
		wantErr bool
	}{
		{
			name: "returns nil when sasl is unknown",
			fields: fields{
				Sasl: types.ObjectUnknown(getSaslAttrTypes()),
			},
		},
		{
			name: "returns a sasl object when populated",
			fields: fields{
				Sasl: types.ObjectValueMust(
					getSaslAttrTypes(),
					map[string]attr.Value{
						"mechanism": types.StringValue("plain"),
					},
				),
			},
			want: &struct {
				Mechanism *kbapi.NewOutputKafkaSaslMechanism `json:"mechanism,omitempty"`
			}{
				Mechanism: utils.Pointer(kbapi.NewOutputKafkaSaslMechanism("plain")),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := outputKafkaModel{
				Sasl: tt.fields.Sasl,
			}
			got, diags := m.toAPISasl(context.Background())
			if (diags.HasError()) != tt.wantErr {
				t.Errorf("outputKafkaModel.toAPISasl() error = %v, wantErr %v", diags.HasError(), tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_outputKafkaModel_toUpdateAPISasl(t *testing.T) {
	type fields struct {
		Sasl types.Object
	}
	tests := []struct {
		name   string
		fields fields
		want   *struct {
			Mechanism *kbapi.UpdateOutputKafkaSaslMechanism `json:"mechanism,omitempty"`
		}
		wantErr bool
	}{
		{
			name: "returns nil when sasl is unknown",
			fields: fields{
				Sasl: types.ObjectUnknown(getSaslAttrTypes()),
			},
		},
		{
			name: "returns a sasl object when populated",
			fields: fields{
				Sasl: types.ObjectValueMust(
					getSaslAttrTypes(),
					map[string]attr.Value{
						"mechanism": types.StringValue("plain"),
					},
				),
			},
			want: &struct {
				Mechanism *kbapi.UpdateOutputKafkaSaslMechanism `json:"mechanism,omitempty"`
			}{
				Mechanism: utils.Pointer(kbapi.UpdateOutputKafkaSaslMechanism("plain")),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := outputKafkaModel{
				Sasl: tt.fields.Sasl,
			}
			got, diags := m.toUpdateAPISasl(context.Background())
			if (diags.HasError()) != tt.wantErr {
				t.Errorf("outputKafkaModel.toUpdateAPISasl() error = %v, wantErr %v", diags.HasError(), tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_outputKafkaModel_toAuthType(t *testing.T) {
	type fields struct {
		AuthType types.String
	}
	tests := []struct {
		name   string
		fields fields
		want   kbapi.NewOutputKafkaAuthType
	}{
		{
			name: "returns none when auth_type is unknown",
			fields: fields{
				AuthType: types.StringUnknown(),
			},
			want: kbapi.NewOutputKafkaAuthTypeNone,
		},
		{
			name: "returns an auth_type object when populated",
			fields: fields{
				AuthType: types.StringValue("user"),
			},
			want: kbapi.NewOutputKafkaAuthType("user"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := outputKafkaModel{
				AuthType: tt.fields.AuthType,
			}
			assert.Equal(t, tt.want, m.toAuthType())
		})
	}
}

func Test_outputKafkaModel_toUpdateAuthType(t *testing.T) {
	type fields struct {
		AuthType types.String
	}
	tests := []struct {
		name   string
		fields fields
		want   *kbapi.UpdateOutputKafkaAuthType
	}{
		{
			name: "returns nil when auth_type is unknown",
			fields: fields{
				AuthType: types.StringUnknown(),
			},
		},
		{
			name: "returns an auth_type object when populated",
			fields: fields{
				AuthType: types.StringValue("user"),
			},
			want: utils.Pointer(kbapi.UpdateOutputKafkaAuthType("user")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := outputKafkaModel{
				AuthType: tt.fields.AuthType,
			}
			assert.Equal(t, tt.want, m.toUpdateAuthType())
		})
	}
}
