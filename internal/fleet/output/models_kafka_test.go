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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_fromAPIKafkaModel_preservesNullSasl(t *testing.T) {
	ctx := context.Background()

	kafkaObj := types.ObjectValueMust(getKafkaAttrTypes(ctx), map[string]attr.Value{
		"auth_type":         types.StringValue("user_pass"),
		"broker_timeout":    types.Float32Null(),
		"client_id":         types.StringNull(),
		"compression":       types.StringNull(),
		"compression_level": types.Int64Null(),
		"connection_type":   types.StringNull(),
		"topic":             types.StringValue("elastic-beats"),
		"partition":         types.StringNull(),
		"required_acks":     types.Int64Null(),
		"timeout":           types.Float32Null(),
		"version":           types.StringNull(),
		"username":          types.StringValue("kafka_user"),
		"password":          types.StringValue("kafka_password"),
		"key":               types.StringNull(),
		"headers":           types.ListNull(getHeadersAttrTypes(ctx)),
		"hash":              types.ObjectNull(getHashAttrTypes(ctx)),
		"random":            types.ObjectNull(getRandomAttrTypes(ctx)),
		"round_robin":       types.ObjectNull(getRoundRobinAttrTypes(ctx)),
		"sasl":              types.ObjectNull(getSaslAttrTypes(ctx)),
	})

	model := outputModel{
		Type:  types.StringValue("kafka"),
		Kafka: kafkaObj,
	}

	mechanism := kbapi.KibanaHTTPAPIsOutputKafkaSaslMechanismPLAIN
	diags := model.fromAPIKafkaModel(ctx, &kbapi.KibanaHTTPAPIsOutputKafka{
		Type:     kbapi.KibanaHTTPAPIsOutputKafkaTypeKafka,
		Name:     "Basic Kafka Output",
		Hosts:    []string{"kafka:9092"},
		AuthType: kbapi.KibanaHTTPAPIsOutputKafkaAuthTypeUserPass,
		Topic:    new("elastic-beats"),
		Sasl: &kbapi.KibanaHTTPAPIsOutputKafka_Sasl{
			Mechanism: &mechanism,
		},
	})
	require.False(t, diags.HasError())

	var result outputKafkaModel
	diags = model.Kafka.As(ctx, &result, basetypes.ObjectAsOptions{})
	require.False(t, diags.HasError())
	assert.True(t, result.Sasl.IsNull())
}

func Test_fromAPIKafkaModel_readsConfiguredSasl(t *testing.T) {
	ctx := context.Background()

	kafkaObj := types.ObjectValueMust(getKafkaAttrTypes(ctx), map[string]attr.Value{
		"auth_type":         types.StringValue("user_pass"),
		"broker_timeout":    types.Float32Null(),
		"client_id":         types.StringNull(),
		"compression":       types.StringNull(),
		"compression_level": types.Int64Null(),
		"connection_type":   types.StringNull(),
		"topic":             types.StringValue("elastic-beats"),
		"partition":         types.StringNull(),
		"required_acks":     types.Int64Null(),
		"timeout":           types.Float32Null(),
		"version":           types.StringNull(),
		"username":          types.StringValue("kafka_user"),
		"password":          types.StringValue("kafka_password"),
		"key":               types.StringNull(),
		"headers":           types.ListNull(getHeadersAttrTypes(ctx)),
		"hash":              types.ObjectNull(getHashAttrTypes(ctx)),
		"random":            types.ObjectNull(getRandomAttrTypes(ctx)),
		"round_robin":       types.ObjectNull(getRoundRobinAttrTypes(ctx)),
		"sasl": types.ObjectValueMust(getSaslAttrTypes(ctx), map[string]attr.Value{
			"mechanism": types.StringValue("SCRAM-SHA-256"),
		}),
	})

	model := outputModel{
		Type:  types.StringValue("kafka"),
		Kafka: kafkaObj,
	}

	mechanism := kbapi.KibanaHTTPAPIsOutputKafkaSaslMechanismSCRAMSHA256
	diags := model.fromAPIKafkaModel(ctx, &kbapi.KibanaHTTPAPIsOutputKafka{
		Type:     kbapi.KibanaHTTPAPIsOutputKafkaTypeKafka,
		Name:     "Kafka Output",
		Hosts:    []string{"kafka:9092"},
		AuthType: kbapi.KibanaHTTPAPIsOutputKafkaAuthTypeUserPass,
		Topic:    new("elastic-beats"),
		Sasl: &kbapi.KibanaHTTPAPIsOutputKafka_Sasl{
			Mechanism: &mechanism,
		},
	})
	require.False(t, diags.HasError())

	var result outputKafkaModel
	diags = model.Kafka.As(ctx, &result, basetypes.ObjectAsOptions{})
	require.False(t, diags.HasError())

	var sasl outputSaslModel
	diags = result.Sasl.As(ctx, &sasl, basetypes.ObjectAsOptions{})
	require.False(t, diags.HasError())
	assert.Equal(t, "SCRAM-SHA-256", sasl.Mechanism.ValueString())
}

func Test_kafkaCompressionLevel(t *testing.T) {
	tests := []struct {
		name             string
		compression      types.String
		compressionLevel types.Int64
		want             *float32
	}{
		{
			name:        "returns nil when compression is not gzip",
			compression: types.StringValue("snappy"),
		},
		{
			name:        "returns nil when compression is unknown",
			compression: types.StringUnknown(),
		},
		{
			name:             "returns explicit level for gzip",
			compression:      types.StringValue("gzip"),
			compressionLevel: types.Int64Value(6),
			want:             new(float32(6)),
		},
		{
			name:        "defaults to 4 for gzip without explicit level",
			compression: types.StringValue("gzip"),
			want:        new(float32(4)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, kafkaCompressionLevel(tt.compression, tt.compressionLevel))
		})
	}
}

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
				Hash: types.ObjectUnknown(getHashAttrTypes(context.Background())),
			},
		},
		{
			name: "returns a hash object when all fields are set",
			fields: fields{
				Hash: types.ObjectValueMust(
					getHashAttrTypes(context.Background()),
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
				Hash:   new("field"),
				Random: new(true),
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
				Headers: types.ListUnknown(getHeadersAttrTypes(context.Background())),
			},
		},
		{
			name: "returns headers when populated",
			fields: fields{
				Headers: types.ListValueMust(
					getHeadersAttrTypes(context.Background()),
					[]attr.Value{
						types.ObjectValueMust(getHeadersAttrTypes(context.Background()).(types.ObjectType).AttrTypes, map[string]attr.Value{
							"key":   types.StringValue("key-1"),
							"value": types.StringValue("value-1"),
						}),
						types.ObjectValueMust(getHeadersAttrTypes(context.Background()).(types.ObjectType).AttrTypes, map[string]attr.Value{
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
				Random: types.ObjectUnknown(getRandomAttrTypes(context.Background())),
			},
		},
		{
			name: "returns a random object when populated",
			fields: fields{
				Random: types.ObjectValueMust(
					getRandomAttrTypes(context.Background()),
					map[string]attr.Value{
						"group_events": types.Float64Value(1),
					},
				),
			},
			want: &struct {
				GroupEvents *float32 `json:"group_events,omitempty"`
			}{
				GroupEvents: new(float32(1)),
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
				RoundRobin: types.ObjectUnknown(getRoundRobinAttrTypes(context.Background())),
			},
		},
		{
			name: "returns a round_robin object when populated",
			fields: fields{
				RoundRobin: types.ObjectValueMust(
					getRoundRobinAttrTypes(context.Background()),
					map[string]attr.Value{
						"group_events": types.Float64Value(1),
					},
				),
			},
			want: &struct {
				GroupEvents *float32 `json:"group_events,omitempty"`
			}{
				GroupEvents: new(float32(1)),
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
			Mechanism *kbapi.KibanaHTTPAPIsNewOutputKafkaSaslMechanism `json:"mechanism,omitempty"`
		}
		wantErr bool
	}{
		{
			name: "returns nil when sasl is unknown",
			fields: fields{
				Sasl: types.ObjectUnknown(getSaslAttrTypes(context.Background())),
			},
		},
		{
			name: "returns a sasl object when populated",
			fields: fields{
				Sasl: types.ObjectValueMust(
					getSaslAttrTypes(context.Background()),
					map[string]attr.Value{
						"mechanism": types.StringValue("plain"),
					},
				),
			},
			want: &struct {
				Mechanism *kbapi.KibanaHTTPAPIsNewOutputKafkaSaslMechanism `json:"mechanism,omitempty"`
			}{
				Mechanism: new(kbapi.KibanaHTTPAPIsNewOutputKafkaSaslMechanism("plain")),
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
			Mechanism *kbapi.KibanaHTTPAPIsUpdateOutputKafkaSaslMechanism `json:"mechanism,omitempty"`
		}
		wantErr bool
	}{
		{
			name: "returns nil when sasl is unknown",
			fields: fields{
				Sasl: types.ObjectUnknown(getSaslAttrTypes(context.Background())),
			},
		},
		{
			name: "returns a sasl object when populated",
			fields: fields{
				Sasl: types.ObjectValueMust(
					getSaslAttrTypes(context.Background()),
					map[string]attr.Value{
						"mechanism": types.StringValue("plain"),
					},
				),
			},
			want: &struct {
				Mechanism *kbapi.KibanaHTTPAPIsUpdateOutputKafkaSaslMechanism `json:"mechanism,omitempty"`
			}{
				Mechanism: new(kbapi.KibanaHTTPAPIsUpdateOutputKafkaSaslMechanism("plain")),
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
		want   kbapi.KibanaHTTPAPIsNewOutputKafkaAuthType
	}{
		{
			name: "returns none when auth_type is unknown",
			fields: fields{
				AuthType: types.StringUnknown(),
			},
			want: kbapi.KibanaHTTPAPIsNewOutputKafkaAuthTypeNone,
		},
		{
			name: "returns an auth_type object when populated",
			fields: fields{
				AuthType: types.StringValue("user"),
			},
			want: kbapi.KibanaHTTPAPIsNewOutputKafkaAuthType("user"),
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
		want   *kbapi.KibanaHTTPAPIsUpdateOutputKafkaAuthType
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
			want: func() *kbapi.KibanaHTTPAPIsUpdateOutputKafkaAuthType {
				value := kbapi.KibanaHTTPAPIsUpdateOutputKafkaAuthType("user")
				return &value
			}(),
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
