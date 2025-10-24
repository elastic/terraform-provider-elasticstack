package output

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type outputKafkaModel struct {
	AuthType         types.String  `tfsdk:"auth_type"`
	BrokerTimeout    types.Float32 `tfsdk:"broker_timeout"`
	ClientId         types.String  `tfsdk:"client_id"`
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
	Headers          types.List    `tfsdk:"headers"`     //> outputHeadersModel
	Hash             types.Object  `tfsdk:"hash"`        //> outputHashModel
	Random           types.Object  `tfsdk:"random"`      //> outputRandomModel
	RoundRobin       types.Object  `tfsdk:"round_robin"` //> outputRoundRobinModel
	Sasl             types.Object  `tfsdk:"sasl"`        //> outputSaslModel
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
	if !utils.IsKnown(m.Hash) {
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
	if !utils.IsKnown(m.Headers) {
		return nil, nil
	}

	var diags diag.Diagnostics
	headerModels := utils.ListTypeAs[outputHeadersModel](ctx, m.Headers, path.Root("kafka").AtName("headers"), &diags)
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
	if !utils.IsKnown(m.Random) {
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
	if !utils.IsKnown(m.RoundRobin) {
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
	Mechanism *kbapi.NewOutputKafkaSaslMechanism `json:"mechanism,omitempty"`
}, diag.Diagnostics) {
	if !utils.IsKnown(m.Sasl) {
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

	mechanism := kbapi.NewOutputKafkaSaslMechanism(saslModel.Mechanism.ValueString())
	return &struct {
		Mechanism *kbapi.NewOutputKafkaSaslMechanism `json:"mechanism,omitempty"`
	}{
		Mechanism: &mechanism,
	}, diags
}

func (m outputKafkaModel) toUpdateAPISasl(ctx context.Context) (*struct {
	Mechanism *kbapi.UpdateOutputKafkaSaslMechanism `json:"mechanism,omitempty"`
}, diag.Diagnostics) {
	sasl, diags := m.toAPISasl(ctx)
	if diags.HasError() || sasl == nil {
		return nil, diags
	}

	mechanism := kbapi.UpdateOutputKafkaSaslMechanism(*sasl.Mechanism)
	return &struct {
		Mechanism *kbapi.UpdateOutputKafkaSaslMechanism "json:\"mechanism,omitempty\""
	}{
		Mechanism: &mechanism,
	}, diags
}

func (m outputKafkaModel) toAuthType() kbapi.NewOutputKafkaAuthType {
	if !utils.IsKnown(m.AuthType) {
		return kbapi.NewOutputKafkaAuthTypeNone
	}

	return kbapi.NewOutputKafkaAuthType(m.AuthType.ValueString())
}

func (m outputKafkaModel) toUpdateAuthType() *kbapi.UpdateOutputKafkaAuthType {
	if !utils.IsKnown(m.AuthType) {
		return nil
	}

	return utils.Pointer(kbapi.UpdateOutputKafkaAuthType(m.AuthType.ValueString()))
}

func (model outputModel) toAPICreateKafkaModel(ctx context.Context) (kbapi.NewOutputUnion, diag.Diagnostics) {
	ssl, diags := objectValueToSSL(ctx, model.Ssl)
	if diags.HasError() {
		return kbapi.NewOutputUnion{}, diags
	}

	// Extract kafka model from nested structure
	var kafkaModel outputKafkaModel
	if !model.Kafka.IsNull() {
		kafkaObj := utils.ObjectTypeAs[outputKafkaModel](ctx, model.Kafka, path.Root("kafka"), &diags)
		kafkaModel = *kafkaObj
	}

	hash, hashDiags := kafkaModel.toAPIHash(ctx)
	diags.Append(hashDiags...)

	headers, headersDiags := kafkaModel.toAPIHeaders(ctx)
	diags.Append(headersDiags...)

	random, randomDiags := kafkaModel.toAPIRandom(ctx)
	diags.Append(randomDiags...)

	roundRobin, rrDiags := kafkaModel.toAPIRoundRobin(ctx)
	diags.Append(rrDiags...)

	sasl, saslDiags := kafkaModel.toAPISasl(ctx)
	diags.Append(saslDiags...)

	body := kbapi.NewOutputKafka{
		Type:                 kbapi.NewOutputKafkaTypeKafka,
		CaSha256:             model.CaSha256.ValueStringPointer(),
		CaTrustedFingerprint: model.CaTrustedFingerprint.ValueStringPointer(),
		ConfigYaml:           model.ConfigYaml.ValueStringPointer(),
		Hosts:                utils.ListTypeToSlice_String(ctx, model.Hosts, path.Root("hosts"), &diags),
		Id:                   model.OutputID.ValueStringPointer(),
		IsDefault:            model.DefaultIntegrations.ValueBoolPointer(),
		IsDefaultMonitoring:  model.DefaultMonitoring.ValueBoolPointer(),
		Name:                 model.Name.ValueString(),
		Ssl:                  ssl,
		// Kafka-specific fields
		AuthType: kafkaModel.toAuthType(),
		BrokerTimeout: func() *float32 {
			if !utils.IsKnown(kafkaModel.BrokerTimeout) {
				return nil
			}
			val := kafkaModel.BrokerTimeout.ValueFloat32()
			return &val
		}(),
		ClientId: kafkaModel.ClientId.ValueStringPointer(),
		Compression: func() *kbapi.NewOutputKafkaCompression {
			if !utils.IsKnown(kafkaModel.Compression) {
				return nil
			}
			comp := kbapi.NewOutputKafkaCompression(kafkaModel.Compression.ValueString())
			return &comp
		}(),
		CompressionLevel: func() *int {
			if !utils.IsKnown(kafkaModel.CompressionLevel) || kafkaModel.Compression.ValueString() != "gzip" {
				return nil
			}

			val := int(kafkaModel.CompressionLevel.ValueInt64())
			return &val
		}(),
		ConnectionType: kafkaModel.ConnectionType.ValueStringPointer(),
		Topic:          kafkaModel.Topic.ValueStringPointer(),
		Partition: func() *kbapi.NewOutputKafkaPartition {
			if !utils.IsKnown(kafkaModel.Partition) {
				return nil
			}
			part := kbapi.NewOutputKafkaPartition(kafkaModel.Partition.ValueString())
			return &part
		}(),
		RequiredAcks: func() *kbapi.NewOutputKafkaRequiredAcks {
			if !utils.IsKnown(kafkaModel.RequiredAcks) {
				return nil
			}
			val := kbapi.NewOutputKafkaRequiredAcks(kafkaModel.RequiredAcks.ValueInt64())
			return &val
		}(),
		Timeout: func() *float32 {
			if !utils.IsKnown(kafkaModel.Timeout) {
				return nil
			}

			val := kafkaModel.Timeout.ValueFloat32()
			return &val
		}(),
		Version:    kafkaModel.Version.ValueStringPointer(),
		Username:   kafkaModel.Username.ValueStringPointer(),
		Password:   kafkaModel.Password.ValueStringPointer(),
		Key:        kafkaModel.Key.ValueStringPointer(),
		Headers:    headers,
		Hash:       hash,
		Random:     random,
		RoundRobin: roundRobin,
		Sasl:       sasl,
	}

	var union kbapi.NewOutputUnion
	err := union.FromNewOutputKafka(body)
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

	// Extract kafka model from nested structure
	var kafkaModel outputKafkaModel
	if !model.Kafka.IsNull() {
		kafkaObj := utils.ObjectTypeAs[outputKafkaModel](ctx, model.Kafka, path.Root("kafka"), &diags)
		kafkaModel = *kafkaObj
	}

	hash, hashDiags := kafkaModel.toAPIHash(ctx)
	diags.Append(hashDiags...)

	headers, headersDiags := kafkaModel.toAPIHeaders(ctx)
	diags.Append(headersDiags...)

	random, randomDiags := kafkaModel.toAPIRandom(ctx)
	diags.Append(randomDiags...)

	roundRobin, rrDiags := kafkaModel.toAPIRoundRobin(ctx)
	diags.Append(rrDiags...)

	sasl, saslDiags := kafkaModel.toUpdateAPISasl(ctx)
	diags.Append(saslDiags...)

	body := kbapi.UpdateOutputKafka{
		Type:                 utils.Pointer(kbapi.Kafka),
		CaSha256:             model.CaSha256.ValueStringPointer(),
		CaTrustedFingerprint: model.CaTrustedFingerprint.ValueStringPointer(),
		ConfigYaml:           model.ConfigYaml.ValueStringPointer(),
		Hosts:                utils.SliceRef(utils.ListTypeToSlice_String(ctx, model.Hosts, path.Root("hosts"), &diags)),
		IsDefault:            model.DefaultIntegrations.ValueBoolPointer(),
		IsDefaultMonitoring:  model.DefaultMonitoring.ValueBoolPointer(),
		Name:                 model.Name.ValueString(),
		Ssl:                  ssl,
		// Kafka-specific fields
		AuthType: kafkaModel.toUpdateAuthType(),
		BrokerTimeout: func() *float32 {
			if !utils.IsKnown(kafkaModel.BrokerTimeout) {
				return nil
			}
			val := kafkaModel.BrokerTimeout.ValueFloat32()
			return &val
		}(),
		ClientId: kafkaModel.ClientId.ValueStringPointer(),
		Compression: func() *kbapi.UpdateOutputKafkaCompression {
			if !utils.IsKnown(kafkaModel.Compression) {
				return nil
			}
			comp := kbapi.UpdateOutputKafkaCompression(kafkaModel.Compression.ValueString())
			return &comp
		}(),
		CompressionLevel: func() *int {
			if !utils.IsKnown(kafkaModel.CompressionLevel) || kafkaModel.Compression.ValueString() != "gzip" {
				return nil
			}
			val := int(kafkaModel.CompressionLevel.ValueInt64())
			return &val
		}(),
		ConnectionType: kafkaModel.ConnectionType.ValueStringPointer(),
		Topic:          kafkaModel.Topic.ValueStringPointer(),
		Partition: func() *kbapi.UpdateOutputKafkaPartition {
			if !utils.IsKnown(kafkaModel.Partition) {
				return nil
			}
			part := kbapi.UpdateOutputKafkaPartition(kafkaModel.Partition.ValueString())
			return &part
		}(),
		RequiredAcks: func() *kbapi.UpdateOutputKafkaRequiredAcks {
			if !utils.IsKnown(kafkaModel.RequiredAcks) {
				return nil
			}
			val := kbapi.UpdateOutputKafkaRequiredAcks(kafkaModel.RequiredAcks.ValueInt64())
			return &val
		}(),
		Timeout: func() *float32 {
			if !utils.IsKnown(kafkaModel.Timeout) {
				return nil
			}
			val := kafkaModel.Timeout.ValueFloat32()
			return &val
		}(),
		Version:    kafkaModel.Version.ValueStringPointer(),
		Username:   kafkaModel.Username.ValueStringPointer(),
		Password:   kafkaModel.Password.ValueStringPointer(),
		Key:        kafkaModel.Key.ValueStringPointer(),
		Headers:    headers,
		Hash:       hash,
		Random:     random,
		RoundRobin: roundRobin,
		Sasl:       sasl,
	}

	var union kbapi.UpdateOutputUnion
	err := union.FromUpdateOutputKafka(body)
	if err != nil {
		diags.AddError(err.Error(), "")
		return kbapi.UpdateOutputUnion{}, diags
	}

	return union, diags
}

func (model *outputModel) fromAPIKafkaModel(ctx context.Context, data *kbapi.OutputKafka) (diags diag.Diagnostics) {
	model.ID = types.StringPointerValue(data.Id)
	model.OutputID = types.StringPointerValue(data.Id)
	model.Name = types.StringValue(data.Name)
	model.Type = types.StringValue(string(data.Type))
	model.Hosts = utils.SliceToListType_String(ctx, data.Hosts, path.Root("hosts"), &diags)
	model.CaSha256 = types.StringPointerValue(data.CaSha256)
	model.CaTrustedFingerprint = types.StringPointerValue(data.CaTrustedFingerprint)
	model.DefaultIntegrations = types.BoolPointerValue(data.IsDefault)
	model.DefaultMonitoring = types.BoolPointerValue(data.IsDefaultMonitoring)
	model.ConfigYaml = types.StringPointerValue(data.ConfigYaml)
	model.Ssl, diags = sslToObjectValue(ctx, data.Ssl)

	// Kafka-specific fields - initialize kafka nested object
	kafkaModel := outputKafkaModel{}
	kafkaModel.AuthType = types.StringValue(string(data.AuthType))
	kafkaModel.BrokerTimeout = types.Float32PointerValue(data.BrokerTimeout)
	kafkaModel.ClientId = types.StringPointerValue(data.ClientId)
	kafkaModel.Compression = types.StringPointerValue((*string)(data.Compression))
	// Handle CompressionLevel
	if data.CompressionLevel != nil {
		kafkaModel.CompressionLevel = types.Int64Value(int64(*data.CompressionLevel))
	} else {
		kafkaModel.CompressionLevel = types.Int64Null()
	}
	// Handle ConnectionType
	kafkaModel.ConnectionType = types.StringPointerValue(data.ConnectionType)
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
	kafkaModel.Password = types.StringPointerValue(data.Password)
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
		list, nd := types.ListValueFrom(ctx, getHeadersAttrTypes(), headerModels)
		diags.Append(nd...)
		kafkaModel.Headers = list
	} else {
		kafkaModel.Headers = types.ListNull(getHeadersAttrTypes())
	}

	// Handle hash
	if data.Hash != nil {
		hashModel := outputHashModel{
			Hash:   types.StringPointerValue(data.Hash.Hash),
			Random: types.BoolPointerValue(data.Hash.Random),
		}
		obj, nd := types.ObjectValueFrom(ctx, getHashAttrTypes(), hashModel)
		diags.Append(nd...)
		kafkaModel.Hash = obj
	} else {
		kafkaModel.Hash = types.ObjectNull(getHashAttrTypes())
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
		obj, nd := types.ObjectValueFrom(ctx, getRandomAttrTypes(), randomModel)
		diags.Append(nd...)
		kafkaModel.Random = obj
	} else {
		kafkaModel.Random = types.ObjectNull(getRandomAttrTypes())
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
		obj, nd := types.ObjectValueFrom(ctx, getRoundRobinAttrTypes(), roundRobinModel)
		diags.Append(nd...)
		kafkaModel.RoundRobin = obj
	} else {
		kafkaModel.RoundRobin = types.ObjectNull(getRoundRobinAttrTypes())
	}

	// Handle sasl
	if data.Sasl != nil {
		saslModel := outputSaslModel{
			Mechanism: func() types.String {
				if data.Sasl.Mechanism != nil {
					return types.StringValue(string(*data.Sasl.Mechanism))
				}
				return types.StringNull()
			}(),
		}
		obj, nd := types.ObjectValueFrom(ctx, getSaslAttrTypes(), saslModel)
		diags.Append(nd...)
		kafkaModel.Sasl = obj
	} else {
		kafkaModel.Sasl = types.ObjectNull(getSaslAttrTypes())
	}

	// Set the kafka nested object on the main model
	kafkaObj, nd := types.ObjectValueFrom(ctx, getKafkaAttrTypes(), kafkaModel)
	diags.Append(nd...)
	model.Kafka = kafkaObj

	// Note: SpaceIds is not returned by the API for outputs, so we preserve it from existing state
	// It's only used to determine which API endpoint to call
	// If space_ids is unknown (not provided by user), set to null to satisfy Terraform's requirement
	if model.SpaceIds.IsNull() || model.SpaceIds.IsUnknown() {
		model.SpaceIds = types.ListNull(types.StringType)
	}

	return
}
