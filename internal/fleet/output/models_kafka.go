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
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type outputKafkaModel struct {
	AuthType         types.String  `tfsdk:"auth_type"`
	BrokerTimeout    types.Float32 `tfsdk:"broker_timeout"`
	ClientID         types.String  `tfsdk:"client_id"`
	Compression      types.String  `tfsdk:"compression"`
	CompressionLevel types.Int64   `tfsdk:"compression_level"`
	ConnectionType   types.String  `tfsdk:"connection_type"`
	Topic            types.String  `tfsdk:"topic"`
	Partition        types.String  `tfsdk:"partition"`
	RequiredAcks     types.Int64   `tfsdk:"required_acks"`
	Timeout          types.Float32 `tfsdk:"timeout"`
	Version          types.String  `tfsdk:"version"`
	Username         types.String  `tfsdk:"username"`
	Password         types.String  `tfsdk:"password"`
	Key              types.String  `tfsdk:"key"`
	Headers          types.List    `tfsdk:"headers"`     // > outputHeadersModel
	Hash             types.Object  `tfsdk:"hash"`        // > outputHashModel
	Random           types.Object  `tfsdk:"random"`      // > outputRandomModel
	RoundRobin       types.Object  `tfsdk:"round_robin"` // > outputRoundRobinModel
	Sasl             types.Object  `tfsdk:"sasl"`        // > outputSaslModel
}

type outputHeadersModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type outputHashModel struct {
	Hash   types.String `tfsdk:"hash"`
	Random types.Bool   `tfsdk:"random"`
}

type outputRandomModel struct {
	GroupEvents types.Float64 `tfsdk:"group_events"`
}

type outputRoundRobinModel struct {
	GroupEvents types.Float64 `tfsdk:"group_events"`
}

type outputSaslModel struct {
	Mechanism types.String `tfsdk:"mechanism"`
}

func (m outputKafkaModel) toAPIHash(ctx context.Context) (*struct {
	Hash   *string `json:"hash,omitempty"`
	Random *bool   `json:"random,omitempty"`
}, diag.Diagnostics) {
	if !typeutils.IsKnown(m.Hash) {
		return nil, nil
	}

	var hashModel outputHashModel
	diags := m.Hash.As(ctx, &hashModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}

	return &struct {
		Hash   *string `json:"hash,omitempty"`
		Random *bool   `json:"random,omitempty"`
	}{
		Hash:   hashModel.Hash.ValueStringPointer(),
		Random: hashModel.Random.ValueBoolPointer(),
	}, diags
}

func (m outputKafkaModel) toAPIHeaders(ctx context.Context) (*[]struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}, diag.Diagnostics) {
	if !typeutils.IsKnown(m.Headers) {
		return nil, nil
	}

	var diags diag.Diagnostics
	headerModels := typeutils.ListTypeAs[outputHeadersModel](ctx, m.Headers, path.Root("kafka").AtName("headers"), &diags)
	if len(headerModels) == 0 {
		return nil, diags
	}

	headers := make([]struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}, len(headerModels))
	for i, h := range headerModels {
		headers[i] = struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}{
			Key:   h.Key.ValueString(),
			Value: h.Value.ValueString(),
		}
	}
	return &headers, diags
}

func (m outputKafkaModel) toAPIRandom(ctx context.Context) (*struct {
	GroupEvents *float32 `json:"group_events,omitempty"`
}, diag.Diagnostics) {
	if !typeutils.IsKnown(m.Random) {
		return nil, nil
	}

	var randomModel outputRandomModel
	diags := m.Random.As(ctx, &randomModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}

	return &struct {
		GroupEvents *float32 `json:"group_events,omitempty"`
	}{
		GroupEvents: func() *float32 {
			if !randomModel.GroupEvents.IsNull() {
				val := float32(randomModel.GroupEvents.ValueFloat64())
				return &val
			}
			return nil
		}(),
	}, diags
}

func (m outputKafkaModel) toAPIRoundRobin(ctx context.Context) (*struct {
	GroupEvents *float32 `json:"group_events,omitempty"`
}, diag.Diagnostics) {
	if !typeutils.IsKnown(m.RoundRobin) {
		return nil, nil
	}

	var roundRobinModel outputRoundRobinModel
	diags := m.RoundRobin.As(ctx, &roundRobinModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}
	return &struct {
		GroupEvents *float32 `json:"group_events,omitempty"`
	}{
		GroupEvents: func() *float32 {
			if !roundRobinModel.GroupEvents.IsNull() {
				val := float32(roundRobinModel.GroupEvents.ValueFloat64())
				return &val
			}
			return nil
		}(),
	}, nil
}

func (m outputKafkaModel) toAPISasl(ctx context.Context) (*struct {
	Mechanism *kbapi.KibanaHTTPAPIsNewOutputKafkaSaslMechanism `json:"mechanism,omitempty"`
}, diag.Diagnostics) {
	if !typeutils.IsKnown(m.Sasl) {
		return nil, nil
	}
	var saslModel outputSaslModel
	diags := m.Sasl.As(ctx, &saslModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}

	if saslModel.Mechanism.IsNull() {
		return nil, diags
	}

	mechanism := kbapi.KibanaHTTPAPIsNewOutputKafkaSaslMechanism(saslModel.Mechanism.ValueString())
	return &struct {
		Mechanism *kbapi.KibanaHTTPAPIsNewOutputKafkaSaslMechanism `json:"mechanism,omitempty"`
	}{
		Mechanism: &mechanism,
	}, diags
}

func (m outputKafkaModel) toUpdateAPISasl(ctx context.Context) (*struct {
	Mechanism *kbapi.KibanaHTTPAPIsUpdateOutputKafkaSaslMechanism `json:"mechanism,omitempty"`
}, diag.Diagnostics) {
	sasl, diags := m.toAPISasl(ctx)
	if diags.HasError() || sasl == nil {
		return nil, diags
	}

	mechanism := kbapi.KibanaHTTPAPIsUpdateOutputKafkaSaslMechanism(*sasl.Mechanism)
	return &struct {
		Mechanism *kbapi.KibanaHTTPAPIsUpdateOutputKafkaSaslMechanism "json:\"mechanism,omitempty\""
	}{
		Mechanism: &mechanism,
	}, diags
}

func (m outputKafkaModel) toAuthType() kbapi.KibanaHTTPAPIsNewOutputKafkaAuthType {
	if !typeutils.IsKnown(m.AuthType) {
		return kbapi.KibanaHTTPAPIsNewOutputKafkaAuthTypeNone
	}

	return kbapi.KibanaHTTPAPIsNewOutputKafkaAuthType(m.AuthType.ValueString())
}

func (m outputKafkaModel) toUpdateAuthType() *kbapi.KibanaHTTPAPIsUpdateOutputKafkaAuthType {
	if !typeutils.IsKnown(m.AuthType) {
		return nil
	}

	authType := kbapi.KibanaHTTPAPIsUpdateOutputKafkaAuthType(m.AuthType.ValueString())
	return &authType
}

func newCreateKafkaConnectionType(value string) (*kbapi.KibanaHTTPAPIsNewOutputKafkaConnectionType, error) {
	connectionType := kbapi.KibanaHTTPAPIsNewOutputKafkaConnectionType(value)
	if !connectionType.Valid() {
		return nil, fmt.Errorf("invalid Kafka connection_type %q", value)
	}

	return &connectionType, nil
}

func newUpdateKafkaConnectionType(value string) (*kbapi.KibanaHTTPAPIsUpdateOutputKafkaConnectionType, error) {
	connectionType := kbapi.KibanaHTTPAPIsUpdateOutputKafkaConnectionType(value)
	if !connectionType.Valid() {
		return nil, fmt.Errorf("invalid Kafka connection_type %q", value)
	}

	return &connectionType, nil
}

func kafkaStringValue(value types.String) *string {
	if !typeutils.IsKnown(value) {
		return nil
	}

	return value.ValueStringPointer()
}

// kafkaComputedFields holds the computed field values shared between create and update.
type kafkaComputedFields struct {
	brokerTimeout    *float32
	compression      *string
	compressionLevel *float32
	partition        *string
	requiredAcks     *int64
	timeout          *float32
	hash             *struct {
		Hash   *string `json:"hash,omitempty"`
		Random *bool   `json:"random,omitempty"`
	}
	headers *[]struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	random *struct {
		GroupEvents *float32 `json:"group_events,omitempty"`
	}
	roundRobin *struct {
		GroupEvents *float32 `json:"group_events,omitempty"`
	}
}

func computeKafkaFields(ctx context.Context, kafkaModel outputKafkaModel, diags *diag.Diagnostics) kafkaComputedFields {
	hash, hashDiags := kafkaModel.toAPIHash(ctx)
	diags.Append(hashDiags...)

	headers, headersDiags := kafkaModel.toAPIHeaders(ctx)
	diags.Append(headersDiags...)

	random, randomDiags := kafkaModel.toAPIRandom(ctx)
	diags.Append(randomDiags...)

	roundRobin, rrDiags := kafkaModel.toAPIRoundRobin(ctx)
	diags.Append(rrDiags...)

	var brokerTimeout *float32
	if typeutils.IsKnown(kafkaModel.BrokerTimeout) {
		val := kafkaModel.BrokerTimeout.ValueFloat32()
		brokerTimeout = &val
	}

	var compression *string
	if typeutils.IsKnown(kafkaModel.Compression) {
		val := kafkaModel.Compression.ValueString()
		compression = &val
	}

	var compressionLevel *float32
	if typeutils.IsKnown(kafkaModel.CompressionLevel) && kafkaModel.Compression.ValueString() == kafkaCompressionGzip {
		val := float32(kafkaModel.CompressionLevel.ValueInt64())
		compressionLevel = &val
	}

	var partition *string
	if typeutils.IsKnown(kafkaModel.Partition) {
		val := kafkaModel.Partition.ValueString()
		partition = &val
	}

	var requiredAcks *int64
	if typeutils.IsKnown(kafkaModel.RequiredAcks) {
		val := kafkaModel.RequiredAcks.ValueInt64()
		requiredAcks = &val
	}

	var timeout *float32
	if typeutils.IsKnown(kafkaModel.Timeout) {
		val := kafkaModel.Timeout.ValueFloat32()
		timeout = &val
	}

	return kafkaComputedFields{
		brokerTimeout:    brokerTimeout,
		compression:      compression,
		compressionLevel: compressionLevel,
		partition:        partition,
		requiredAcks:     requiredAcks,
		timeout:          timeout,
		hash:             hash,
		headers:          headers,
		random:           random,
		roundRobin:       roundRobin,
	}
}

func readOutputKafkaConnectionType(value *kbapi.KibanaHTTPAPIsOutputKafkaConnectionType) *string {
	if value == nil {
		return nil
	}

	connectionType := string(*value)
	return &connectionType
}

func readOutputKafkaCompressionLevel(value *float32) *int64 {
	if value == nil {
		return nil
	}

	converted := int64(*value)
	return &converted
}

func (model outputModel) toAPICreateKafkaModel(ctx context.Context) (kbapi.NewOutputUnion, diag.Diagnostics) {
	ssl, diags := objectValueToSSL(ctx, model.Ssl)
	if diags.HasError() {
		return kbapi.NewOutputUnion{}, diags
	}

	var kafkaModel outputKafkaModel
	if !model.Kafka.IsNull() {
		kafkaObj := typeutils.ObjectTypeAs[outputKafkaModel](ctx, model.Kafka, path.Root("kafka"), &diags)
		kafkaModel = *kafkaObj
	}

	fields := computeKafkaFields(ctx, kafkaModel, &diags)

	sasl, saslDiags := kafkaModel.toAPISasl(ctx)
	diags.Append(saslDiags...)

	var err error
	var connectionType *kbapi.KibanaHTTPAPIsNewOutputKafkaConnectionType
	if connectionTypeValue := kafkaStringValue(kafkaModel.ConnectionType); connectionTypeValue != nil {
		connectionType, err = newCreateKafkaConnectionType(*connectionTypeValue)
		if err != nil {
			diags.AddError(err.Error(), "")
		}
	}

	var compression *kbapi.KibanaHTTPAPIsNewOutputKafkaCompression
	if fields.compression != nil {
		comp := kbapi.KibanaHTTPAPIsNewOutputKafkaCompression(*fields.compression)
		compression = &comp
	}

	var partition *kbapi.KibanaHTTPAPIsNewOutputKafkaPartition
	if fields.partition != nil {
		part := kbapi.KibanaHTTPAPIsNewOutputKafkaPartition(*fields.partition)
		partition = &part
	}

	var requiredAcks *kbapi.KibanaHTTPAPIsNewOutputKafkaRequiredAcks
	if fields.requiredAcks != nil {
		val := kbapi.KibanaHTTPAPIsNewOutputKafkaRequiredAcks(*fields.requiredAcks)
		requiredAcks = &val
	}

	body := kbapi.KibanaHTTPAPIsNewOutputKafka{
		Type:                 kbapi.KibanaHTTPAPIsNewOutputKafkaTypeKafka,
		CaSha256:             model.CaSha256.ValueStringPointer(),
		CaTrustedFingerprint: model.CaTrustedFingerprint.ValueStringPointer(),
		ConfigYaml:           model.ConfigYaml.ValueStringPointer(),
		Hosts:                typeutils.ListTypeToSliceString(ctx, model.Hosts, path.Root("hosts"), &diags),
		Id:                   typeutils.OptionalString(model.OutputID),
		IsDefault:            model.DefaultIntegrations.ValueBoolPointer(),
		IsDefaultMonitoring:  model.DefaultMonitoring.ValueBoolPointer(),
		Name:                 model.Name.ValueString(),
		Ssl:                  ssl.toAPI(),
		AuthType:             kafkaModel.toAuthType(),
		BrokerTimeout:        fields.brokerTimeout,
		ClientId:             kafkaModel.ClientID.ValueStringPointer(),
		Compression:          compression,
		CompressionLevel:     fields.compressionLevel,
		ConnectionType:       connectionType,
		Topic:                kafkaModel.Topic.ValueStringPointer(),
		Partition:            partition,
		RequiredAcks:         requiredAcks,
		Timeout:              fields.timeout,
		Version:              kafkaModel.Version.ValueStringPointer(),
		Username:             kafkaStringValue(kafkaModel.Username),
		Password:             kafkaStringValue(kafkaModel.Password),
		Key:                  kafkaModel.Key.ValueStringPointer(),
		Headers:              fields.headers,
		Hash:                 fields.hash,
		Random:               fields.random,
		RoundRobin:           fields.roundRobin,
		Sasl:                 sasl,
	}

	var union kbapi.NewOutputUnion
	err = union.FromKibanaHTTPAPIsNewOutputKafka(body)
	if err != nil {
		diags.AddError(err.Error(), "")
		return kbapi.NewOutputUnion{}, diags
	}

	return union, diags
}

func (model outputModel) toAPIUpdateKafkaModel(ctx context.Context) (kbapi.UpdateOutputUnion, diag.Diagnostics) {
	ssl, diags := objectValueToSSLUpdate(ctx, model.Ssl)
	if diags.HasError() {
		return kbapi.UpdateOutputUnion{}, diags
	}

	var kafkaModel outputKafkaModel
	if !model.Kafka.IsNull() {
		kafkaObj := typeutils.ObjectTypeAs[outputKafkaModel](ctx, model.Kafka, path.Root("kafka"), &diags)
		kafkaModel = *kafkaObj
	}

	fields := computeKafkaFields(ctx, kafkaModel, &diags)

	sasl, saslDiags := kafkaModel.toUpdateAPISasl(ctx)
	diags.Append(saslDiags...)

	var err error
	var connectionType *kbapi.KibanaHTTPAPIsUpdateOutputKafkaConnectionType
	if connectionTypeValue := kafkaStringValue(kafkaModel.ConnectionType); connectionTypeValue != nil {
		connectionType, err = newUpdateKafkaConnectionType(*connectionTypeValue)
		if err != nil {
			diags.AddError(err.Error(), "")
		}
	}

	var compression *kbapi.KibanaHTTPAPIsUpdateOutputKafkaCompression
	if fields.compression != nil {
		comp := kbapi.KibanaHTTPAPIsUpdateOutputKafkaCompression(*fields.compression)
		compression = &comp
	}

	var partition *kbapi.KibanaHTTPAPIsUpdateOutputKafkaPartition
	if fields.partition != nil {
		part := kbapi.KibanaHTTPAPIsUpdateOutputKafkaPartition(*fields.partition)
		partition = &part
	}

	var requiredAcks *kbapi.KibanaHTTPAPIsUpdateOutputKafkaRequiredAcks
	if fields.requiredAcks != nil {
		val := kbapi.KibanaHTTPAPIsUpdateOutputKafkaRequiredAcks(*fields.requiredAcks)
		requiredAcks = &val
	}

	outputType := kbapi.Kafka
	body := kbapi.KibanaHTTPAPIsUpdateOutputKafka{
		Type:                 &outputType,
		CaSha256:             model.CaSha256.ValueStringPointer(),
		CaTrustedFingerprint: model.CaTrustedFingerprint.ValueStringPointer(),
		ConfigYaml:           model.ConfigYaml.ValueStringPointer(),
		Hosts:                typeutils.SliceRef(typeutils.ListTypeToSliceString(ctx, model.Hosts, path.Root("hosts"), &diags)),
		IsDefault:            model.DefaultIntegrations.ValueBoolPointer(),
		IsDefaultMonitoring:  model.DefaultMonitoring.ValueBoolPointer(),
		Name:                 model.Name.ValueString(),
		Ssl:                  ssl.toAPI(),
		AuthType:             kafkaModel.toUpdateAuthType(),
		BrokerTimeout:        fields.brokerTimeout,
		ClientId:             kafkaModel.ClientID.ValueStringPointer(),
		Compression:          compression,
		CompressionLevel:     fields.compressionLevel,
		ConnectionType:       connectionType,
		Topic:                kafkaModel.Topic.ValueStringPointer(),
		Partition:            partition,
		RequiredAcks:         requiredAcks,
		Timeout:              fields.timeout,
		Version:              kafkaModel.Version.ValueStringPointer(),
		Username:             kafkaStringValue(kafkaModel.Username),
		Password:             kafkaStringValue(kafkaModel.Password),
		Key:                  kafkaModel.Key.ValueStringPointer(),
		Headers:              fields.headers,
		Hash:                 fields.hash,
		Random:               fields.random,
		RoundRobin:           fields.roundRobin,
		Sasl:                 sasl,
	}

	var union kbapi.UpdateOutputUnion
	err = union.FromKibanaHTTPAPIsUpdateOutputKafka(body)
	if err != nil {
		diags.AddError(err.Error(), "")
		return kbapi.UpdateOutputUnion{}, diags
	}

	return union, diags
}

func (model *outputModel) fromAPIKafkaModel(ctx context.Context, data *kbapi.KibanaHTTPAPIsOutputKafka) (diags diag.Diagnostics) {
	diags = model.fromAPICommonFields(ctx, commonOutputReadData{
		id:                   data.Id,
		name:                 data.Name,
		outputType:           string(data.Type),
		hosts:                data.Hosts,
		caSha256:             data.CaSha256,
		caTrustedFingerprint: data.CaTrustedFingerprint,
		isDefault:            data.IsDefault,
		isDefaultMonitoring:  data.IsDefaultMonitoring,
		configYaml:           data.ConfigYaml,
		ssl:                  data.Ssl,
	})

	// Capture the configured password and sasl before re-initializing kafkaModel
	// so that we can preserve them when Fleet omits/redacts or adds server-side
	// defaults that the user did not configure.
	configuredPassword := types.StringNull()
	saslExplicitlyNull := false
	if typeutils.IsKnown(model.Kafka) {
		var existing outputKafkaModel
		existingDiags := model.Kafka.As(ctx, &existing, basetypes.ObjectAsOptions{})
		diags.Append(existingDiags...)
		if !existingDiags.HasError() {
			configuredPassword = existing.Password
			if !existing.Sasl.IsUnknown() {
				saslExplicitlyNull = existing.Sasl.IsNull()
			}
		}
	}

	// Kafka-specific fields - initialize kafka nested object
	kafkaModel := outputKafkaModel{}
	kafkaModel.AuthType = types.StringValue(string(data.AuthType))
	kafkaModel.BrokerTimeout = types.Float32PointerValue(data.BrokerTimeout)
	kafkaModel.ClientID = types.StringPointerValue(data.ClientId)
	kafkaModel.Compression = types.StringPointerValue((*string)(data.Compression))
	// Handle CompressionLevel
	if compressionLevel := readOutputKafkaCompressionLevel(data.CompressionLevel); compressionLevel != nil {
		kafkaModel.CompressionLevel = types.Int64Value(*compressionLevel)
	} else {
		kafkaModel.CompressionLevel = types.Int64Null()
	}
	// Handle ConnectionType
	kafkaModel.ConnectionType = types.StringPointerValue(readOutputKafkaConnectionType(data.ConnectionType))
	kafkaModel.Topic = types.StringPointerValue(data.Topic)
	kafkaModel.Partition = types.StringPointerValue((*string)(data.Partition))
	if data.RequiredAcks != nil {
		kafkaModel.RequiredAcks = types.Int64Value(int64(*data.RequiredAcks))
	} else {
		kafkaModel.RequiredAcks = types.Int64Null()
	}

	kafkaModel.Timeout = types.Float32PointerValue(data.Timeout)
	kafkaModel.Version = types.StringPointerValue(data.Version)
	kafkaModel.Username = types.StringPointerValue(data.Username)
	switch {
	case data.Password != nil:
		kafkaModel.Password = types.StringPointerValue(data.Password)
	case typeutils.IsKnown(configuredPassword):
		// Fleet redacts kafka.password in API responses (the value is stored
		// in the secret store and only a reference comes back). Preserve the
		// configured value so Terraform does not see an inconsistent change.
		kafkaModel.Password = configuredPassword
	default:
		kafkaModel.Password = types.StringNull()
	}
	kafkaModel.Key = types.StringPointerValue(data.Key)

	// Handle headers
	if data.Headers != nil {
		headerModels := make([]outputHeadersModel, len(*data.Headers))
		for i, header := range *data.Headers {
			headerModels[i] = outputHeadersModel{
				Key:   types.StringValue(header.Key),
				Value: types.StringValue(header.Value),
			}
		}
		list, nd := types.ListValueFrom(ctx, getHeadersAttrTypes(ctx), headerModels)
		diags.Append(nd...)
		kafkaModel.Headers = list
	} else {
		kafkaModel.Headers = types.ListNull(getHeadersAttrTypes(ctx))
	}

	// Handle hash
	if data.Hash != nil {
		hashModel := outputHashModel{
			Hash:   types.StringPointerValue(data.Hash.Hash),
			Random: types.BoolPointerValue(data.Hash.Random),
		}
		obj, nd := types.ObjectValueFrom(ctx, getHashAttrTypes(ctx), hashModel)
		diags.Append(nd...)
		kafkaModel.Hash = obj
	} else {
		kafkaModel.Hash = types.ObjectNull(getHashAttrTypes(ctx))
	}

	// Handle random
	if data.Random != nil {
		randomModel := outputRandomModel{
			GroupEvents: func() types.Float64 {
				if data.Random.GroupEvents != nil {
					return types.Float64Value(float64(*data.Random.GroupEvents))
				}
				return types.Float64Null()
			}(),
		}
		obj, nd := types.ObjectValueFrom(ctx, getRandomAttrTypes(ctx), randomModel)
		diags.Append(nd...)
		kafkaModel.Random = obj
	} else {
		kafkaModel.Random = types.ObjectNull(getRandomAttrTypes(ctx))
	}

	// Handle round_robin
	if data.RoundRobin != nil {
		roundRobinModel := outputRoundRobinModel{
			GroupEvents: func() types.Float64 {
				if data.RoundRobin.GroupEvents != nil {
					return types.Float64Value(float64(*data.RoundRobin.GroupEvents))
				}
				return types.Float64Null()
			}(),
		}
		obj, nd := types.ObjectValueFrom(ctx, getRoundRobinAttrTypes(ctx), roundRobinModel)
		diags.Append(nd...)
		kafkaModel.RoundRobin = obj
	} else {
		kafkaModel.RoundRobin = types.ObjectNull(getRoundRobinAttrTypes(ctx))
	}

	// Handle sasl
	switch {
	case saslExplicitlyNull:
		// Fleet may return a default sasl block (e.g. mechanism=PLAIN for
		// user_pass auth) even when the user did not configure sasl. Preserve
		// the configured null so Terraform does not see an inconsistent change.
		kafkaModel.Sasl = types.ObjectNull(getSaslAttrTypes(ctx))
	case data.Sasl != nil:
		saslModel := outputSaslModel{
			Mechanism: func() types.String {
				if data.Sasl.Mechanism != nil {
					return types.StringValue(string(*data.Sasl.Mechanism))
				}
				return types.StringNull()
			}(),
		}
		obj, nd := types.ObjectValueFrom(ctx, getSaslAttrTypes(ctx), saslModel)
		diags.Append(nd...)
		kafkaModel.Sasl = obj
	default:
		kafkaModel.Sasl = types.ObjectNull(getSaslAttrTypes(ctx))
	}

	// Set the kafka nested object on the main model
	kafkaObj, nd := types.ObjectValueFrom(ctx, getKafkaAttrTypes(ctx), kafkaModel)
	diags.Append(nd...)
	model.Kafka = kafkaObj

	clearRemoteElasticsearchOnlyFields(model)

	return
}
