package output

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type outputModel struct {
	ID                   types.String `tfsdk:"id"`
	OutputID             types.String `tfsdk:"output_id"`
	Name                 types.String `tfsdk:"name"`
	Type                 types.String `tfsdk:"type"`
	Hosts                types.List   `tfsdk:"hosts"` //> string
	CaSha256             types.String `tfsdk:"ca_sha256"`
	CaTrustedFingerprint types.String `tfsdk:"ca_trusted_fingerprint"`
	DefaultIntegrations  types.Bool   `tfsdk:"default_integrations"`
	DefaultMonitoring    types.Bool   `tfsdk:"default_monitoring"`
	ConfigYaml           types.String `tfsdk:"config_yaml"`
	Ssl                  types.Object `tfsdk:"ssl"`   //> outputSslModel
	Kafka                types.Object `tfsdk:"kafka"` //> outputKafkaModel
}

type outputSslModel struct {
	CertificateAuthorities types.List   `tfsdk:"certificate_authorities"` //> string
	Certificate            types.String `tfsdk:"certificate"`
	Key                    types.String `tfsdk:"key"`
}

func (model *outputModel) populateFromAPI(ctx context.Context, union *kbapi.OutputUnion) (diags diag.Diagnostics) {
	if union == nil {
		return
	}

	doSsl := func(ssl *kbapi.OutputSsl) types.Object {
		if ssl != nil {
			p := path.Root("ssl")
			sslModel := outputSslModel{
				CertificateAuthorities: utils.SliceToListType_String(ctx, utils.Deref(ssl.CertificateAuthorities), p.AtName("certificate_authorities"), &diags),
				Certificate:            types.StringPointerValue(ssl.Certificate),
				Key:                    types.StringPointerValue(ssl.Key),
			}
			obj, nd := types.ObjectValueFrom(ctx, getSslAttrTypes(), sslModel)
			diags.Append(nd...)
			return obj
		} else {
			return types.ObjectNull(getSslAttrTypes())
		}
	}

	discriminator, err := union.Discriminator()
	if err != nil {
		diags.AddError(err.Error(), "")
		return
	}

	switch discriminator {
	case "elasticsearch":
		data, err := union.AsOutputElasticsearch()
		if err != nil {
			diags.AddError(err.Error(), "")
			return
		}

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
		model.Ssl = doSsl(data.Ssl)

	case "logstash":
		data, err := union.AsOutputLogstash()
		if err != nil {
			diags.AddError(err.Error(), "")
			return
		}

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
		model.Ssl = doSsl(data.Ssl)

	case "kafka":
		data, err := union.AsOutputKafka()
		if err != nil {
			diags.AddError(err.Error(), "")
			return
		}

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
		model.Ssl = doSsl(data.Ssl)

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

	default:
		diags.AddError(fmt.Sprintf("unhandled output type: %s", discriminator), "")
	}

	return
}

func (model outputModel) toAPICreateModel(ctx context.Context, client *clients.ApiClient) (kbapi.NewOutputUnion, diag.Diagnostics) {
	outputType := model.Type.ValueString()

	switch outputType {
	case "elasticsearch":
		return model.toAPICreateElasticsearchModel(ctx)
	case "logstash":
		return model.toAPICreateLogstashModel(ctx)
	case "kafka":
		if diags := assertKafkaSupport(ctx, client); diags.HasError() {
			return kbapi.NewOutputUnion{}, diags
		}

		return model.toAPICreateKafkaModel(ctx)
	default:
		return kbapi.NewOutputUnion{}, diag.Diagnostics{
			diag.NewErrorDiagnostic(fmt.Sprintf("unhandled output type: %s", outputType), ""),
		}
	}
}

func assertKafkaSupport(ctx context.Context, client *clients.ApiClient) diag.Diagnostics {
	var diags diag.Diagnostics

	// Check minimum version requirement for Kafka output type
	if supported, versionDiags := client.EnforceMinVersion(ctx, MinVersionOutputKafka); versionDiags.HasError() {
		diags.Append(utils.FrameworkDiagsFromSDK(versionDiags)...)
		return diags
	} else if !supported {
		diags.AddError("Unsupported version for Kafka output",
			fmt.Sprintf("Kafka output type requires server version %s or higher", MinVersionOutputKafka.String()))
		return diags
	}

	return nil
}

func (model outputModel) toAPIUpdateModel(ctx context.Context, client *clients.ApiClient) (union kbapi.UpdateOutputUnion, diags diag.Diagnostics) {
	outputType := model.Type.ValueString()

	switch outputType {
	case "elasticsearch":
		return model.toAPIUpdateElasticsearchModel(ctx)
	case "logstash":
		return model.toAPIUpdateLogstashModel(ctx)
	case "kafka":
		if diags := assertKafkaSupport(ctx, client); diags.HasError() {
			return kbapi.UpdateOutputUnion{}, diags
		}

		return model.toAPIUpdateKafkaModel(ctx)
	default:
		diags.AddError(fmt.Sprintf("unhandled output type: %s", outputType), "")
	}

	return
}

func (model outputModel) toAPICreateElasticsearchModel(ctx context.Context) (kbapi.NewOutputUnion, diag.Diagnostics) {
	ssl, diags := model.toAPISSL(ctx)
	if diags.HasError() {
		return kbapi.NewOutputUnion{}, diags
	}

	body := kbapi.NewOutputElasticsearch{
		Type:                 kbapi.NewOutputElasticsearchTypeElasticsearch,
		CaSha256:             model.CaSha256.ValueStringPointer(),
		CaTrustedFingerprint: model.CaTrustedFingerprint.ValueStringPointer(),
		ConfigYaml:           model.ConfigYaml.ValueStringPointer(),
		Hosts:                utils.ListTypeToSlice_String(ctx, model.Hosts, path.Root("hosts"), &diags),
		Id:                   model.OutputID.ValueStringPointer(),
		IsDefault:            model.DefaultIntegrations.ValueBoolPointer(),
		IsDefaultMonitoring:  model.DefaultMonitoring.ValueBoolPointer(),
		Name:                 model.Name.ValueString(),
		Ssl:                  ssl,
	}

	var union kbapi.NewOutputUnion
	err := union.FromNewOutputElasticsearch(body)
	if err != nil {
		diags.AddError(err.Error(), "")
		return kbapi.NewOutputUnion{}, diags
	}

	return union, diags
}

func (model outputModel) toAPICreateLogstashModel(ctx context.Context) (kbapi.NewOutputUnion, diag.Diagnostics) {
	ssl, diags := model.toAPISSL(ctx)
	if diags.HasError() {
		return kbapi.NewOutputUnion{}, diags
	}
	body := kbapi.NewOutputLogstash{
		Type:                 kbapi.NewOutputLogstashTypeLogstash,
		CaSha256:             model.CaSha256.ValueStringPointer(),
		CaTrustedFingerprint: model.CaTrustedFingerprint.ValueStringPointer(),
		ConfigYaml:           model.ConfigYaml.ValueStringPointer(),
		Hosts:                utils.ListTypeToSlice_String(ctx, model.Hosts, path.Root("hosts"), &diags),
		Id:                   model.OutputID.ValueStringPointer(),
		IsDefault:            model.DefaultIntegrations.ValueBoolPointer(),
		IsDefaultMonitoring:  model.DefaultMonitoring.ValueBoolPointer(),
		Name:                 model.Name.ValueString(),
		Ssl:                  ssl,
	}

	var union kbapi.NewOutputUnion
	err := union.FromNewOutputLogstash(body)
	if err != nil {
		diags.AddError(err.Error(), "")
		return kbapi.NewOutputUnion{}, diags
	}

	return union, diags
}

func (model outputModel) toAPICreateKafkaModel(ctx context.Context) (kbapi.NewOutputUnion, diag.Diagnostics) {
	ssl, diags := model.toAPISSL(ctx)
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

func (model outputModel) toAPIUpdateElasticsearchModel(ctx context.Context) (kbapi.UpdateOutputUnion, diag.Diagnostics) {
	ssl, diags := model.toUpdateAPISSL(ctx)
	if diags.HasError() {
		return kbapi.UpdateOutputUnion{}, diags
	}
	body := kbapi.UpdateOutputElasticsearch{
		Type:                 utils.Pointer(kbapi.Elasticsearch),
		CaSha256:             model.CaSha256.ValueStringPointer(),
		CaTrustedFingerprint: model.CaTrustedFingerprint.ValueStringPointer(),
		ConfigYaml:           model.ConfigYaml.ValueStringPointer(),
		Hosts:                utils.SliceRef(utils.ListTypeToSlice_String(ctx, model.Hosts, path.Root("hosts"), &diags)),
		IsDefault:            model.DefaultIntegrations.ValueBoolPointer(),
		IsDefaultMonitoring:  model.DefaultMonitoring.ValueBoolPointer(),
		Name:                 model.Name.ValueStringPointer(),
		Ssl:                  ssl,
	}

	var union kbapi.UpdateOutputUnion
	err := union.FromUpdateOutputElasticsearch(body)
	if err != nil {
		diags.AddError(err.Error(), "")
		return kbapi.UpdateOutputUnion{}, diags
	}

	return union, diags
}

func (model outputModel) toAPIUpdateLogstashModel(ctx context.Context) (kbapi.UpdateOutputUnion, diag.Diagnostics) {
	ssl, diags := model.toUpdateAPISSL(ctx)
	if diags.HasError() {
		return kbapi.UpdateOutputUnion{}, diags
	}
	body := kbapi.UpdateOutputLogstash{
		Type:                 utils.Pointer(kbapi.Logstash),
		CaSha256:             model.CaSha256.ValueStringPointer(),
		CaTrustedFingerprint: model.CaTrustedFingerprint.ValueStringPointer(),
		ConfigYaml:           model.ConfigYaml.ValueStringPointer(),
		Hosts:                utils.SliceRef(utils.ListTypeToSlice_String(ctx, model.Hosts, path.Root("hosts"), &diags)),
		IsDefault:            model.DefaultIntegrations.ValueBoolPointer(),
		IsDefaultMonitoring:  model.DefaultMonitoring.ValueBoolPointer(),
		Name:                 model.Name.ValueStringPointer(),
		Ssl:                  ssl,
	}

	var union kbapi.UpdateOutputUnion
	err := union.FromUpdateOutputLogstash(body)
	if err != nil {
		diags.AddError(err.Error(), "")
		return kbapi.UpdateOutputUnion{}, diags
	}

	return union, diags
}

func (model outputModel) toAPIUpdateKafkaModel(ctx context.Context) (kbapi.UpdateOutputUnion, diag.Diagnostics) {
	ssl, diags := model.toUpdateAPISSL(ctx)
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

func (model outputModel) toAPISSL(ctx context.Context) (*kbapi.NewOutputSsl, diag.Diagnostics) {
	if !utils.IsKnown(model.Ssl) {
		return nil, nil
	}
	var diags diag.Diagnostics
	sslModel := utils.ObjectTypeAs[outputSslModel](ctx, model.Ssl, path.Root("ssl"), &diags)
	if diags.HasError() {
		return nil, diags
	}

	if sslModel == nil {
		return nil, diags
	}

	return &kbapi.NewOutputSsl{
		Certificate:            sslModel.Certificate.ValueStringPointer(),
		CertificateAuthorities: utils.SliceRef(utils.ListTypeToSlice_String(ctx, sslModel.CertificateAuthorities, path.Root("certificate_authorities"), &diags)),
		Key:                    sslModel.Key.ValueStringPointer(),
	}, diags
}

func (model outputModel) toUpdateAPISSL(ctx context.Context) (*kbapi.UpdateOutputSsl, diag.Diagnostics) {
	ssl, diags := model.toAPISSL(ctx)
	if diags.HasError() || ssl == nil {
		return nil, diags
	}

	return &kbapi.UpdateOutputSsl{
		Certificate:            ssl.Certificate,
		CertificateAuthorities: ssl.CertificateAuthorities,
		Key:                    ssl.Key,
	}, diags
}
