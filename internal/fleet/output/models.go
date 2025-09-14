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

type outputKafkaModel struct {
	AuthType         types.String  `tfsdk:"auth_type"`
	BrokerTimeout    types.Float64 `tfsdk:"broker_timeout"`
	ClientId         types.String  `tfsdk:"client_id"`
	Compression      types.String  `tfsdk:"compression"`
	CompressionLevel types.Float64 `tfsdk:"compression_level"`
	ConnectionType   types.String  `tfsdk:"connection_type"`
	Topic            types.String  `tfsdk:"topic"`
	Partition        types.String  `tfsdk:"partition"`
	RequiredAcks     types.Int64   `tfsdk:"required_acks"`
	Timeout          types.Float64 `tfsdk:"timeout"`
	Version          types.String  `tfsdk:"version"`
	Username         types.String  `tfsdk:"username"`
	Password         types.String  `tfsdk:"password"`
	Key              types.String  `tfsdk:"key"`
	Headers          types.List    `tfsdk:"headers"`     //> outputHeadersModel
	Hash             types.List    `tfsdk:"hash"`        //> outputHashModel
	Random           types.List    `tfsdk:"random"`      //> outputRandomModel
	RoundRobin       types.List    `tfsdk:"round_robin"` //> outputRoundRobinModel
	Sasl             types.List    `tfsdk:"sasl"`        //> outputSaslModel
}

type outputSslModel struct {
	CertificateAuthorities types.List   `tfsdk:"certificate_authorities"` //> string
	Certificate            types.String `tfsdk:"certificate"`
	Key                    types.String `tfsdk:"key"`
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
		if data.BrokerTimeout != nil {
			kafkaModel.BrokerTimeout = types.Float64Value(float64(*data.BrokerTimeout))
		} else {
			kafkaModel.BrokerTimeout = types.Float64Null()
		}
		kafkaModel.ClientId = types.StringPointerValue(data.ClientId)
		if data.Compression != nil {
			kafkaModel.Compression = types.StringValue(string(*data.Compression))
		} else {
			kafkaModel.Compression = types.StringNull()
		}
		// Handle CompressionLevel
		if data.CompressionLevel != nil {
			kafkaModel.CompressionLevel = types.Float64Value(float64(*data.CompressionLevel))
		} else {
			kafkaModel.CompressionLevel = types.Float64Null()
		}
		// Handle ConnectionType
		if data.ConnectionType != nil {
			kafkaModel.ConnectionType = types.StringValue(*data.ConnectionType)
		} else {
			kafkaModel.ConnectionType = types.StringNull()
		}
		kafkaModel.Topic = types.StringPointerValue(data.Topic)
		if data.Partition != nil {
			kafkaModel.Partition = types.StringValue(string(*data.Partition))
		} else {
			kafkaModel.Partition = types.StringNull()
		}
		if data.RequiredAcks != nil {
			kafkaModel.RequiredAcks = types.Int64Value(int64(*data.RequiredAcks))
		} else {
			kafkaModel.RequiredAcks = types.Int64Null()
		}
		if data.Timeout != nil {
			kafkaModel.Timeout = types.Float64Value(float64(*data.Timeout))
		} else {
			kafkaModel.Timeout = types.Float64Null()
		}
		kafkaModel.Version = types.StringPointerValue(data.Version)
		if data.Username != nil {
			kafkaModel.Username = types.StringValue(*data.Username)
		} else {
			kafkaModel.Username = types.StringNull()
		}
		if data.Password != nil {
			kafkaModel.Password = types.StringValue(*data.Password)
		} else {
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
			list, nd := types.ListValueFrom(ctx, getHeadersAttrTypes(), headerModels)
			diags.Append(nd...)
			kafkaModel.Headers = list
		} else {
			kafkaModel.Headers = types.ListNull(getHeadersAttrTypes())
		}

		// Handle hash
		if data.Hash != nil {
			hashModels := []outputHashModel{{
				Hash:   types.StringPointerValue(data.Hash.Hash),
				Random: types.BoolPointerValue(data.Hash.Random),
			}}
			list, nd := types.ListValueFrom(ctx, getHashAttrTypes(), hashModels)
			diags.Append(nd...)
			kafkaModel.Hash = list
		} else {
			kafkaModel.Hash = types.ListNull(getHashAttrTypes())
		}

		// Handle random
		if data.Random != nil {
			randomModels := []outputRandomModel{{
				GroupEvents: func() types.Float64 {
					if data.Random.GroupEvents != nil {
						return types.Float64Value(float64(*data.Random.GroupEvents))
					}
					return types.Float64Null()
				}(),
			}}
			list, nd := types.ListValueFrom(ctx, getRandomAttrTypes(), randomModels)
			diags.Append(nd...)
			kafkaModel.Random = list
		} else {
			kafkaModel.Random = types.ListNull(getRandomAttrTypes())
		}

		// Handle round_robin
		if data.RoundRobin != nil {
			roundRobinModels := []outputRoundRobinModel{{
				GroupEvents: func() types.Float64 {
					if data.RoundRobin.GroupEvents != nil {
						return types.Float64Value(float64(*data.RoundRobin.GroupEvents))
					}
					return types.Float64Null()
				}(),
			}}
			list, nd := types.ListValueFrom(ctx, getRoundRobinAttrTypes(), roundRobinModels)
			diags.Append(nd...)
			kafkaModel.RoundRobin = list
		} else {
			kafkaModel.RoundRobin = types.ListNull(getRoundRobinAttrTypes())
		}

		// Handle sasl
		if data.Sasl != nil {
			saslModels := []outputSaslModel{{
				Mechanism: func() types.String {
					if data.Sasl.Mechanism != nil {
						return types.StringValue(string(*data.Sasl.Mechanism))
					}
					return types.StringNull()
				}(),
			}}
			list, nd := types.ListValueFrom(ctx, getSaslAttrTypes(), saslModels)
			diags.Append(nd...)
			kafkaModel.Sasl = list
		} else {
			kafkaModel.Sasl = types.ListNull(getSaslAttrTypes())
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

func (model outputModel) toAPICreateModel(ctx context.Context, client *clients.ApiClient) (union kbapi.NewOutputUnion, diags diag.Diagnostics) {
	doSsl := func() *kbapi.NewOutputSsl {
		if utils.IsKnown(model.Ssl) {
			sslModel := utils.ObjectTypeAs[outputSslModel](ctx, model.Ssl, path.Root("ssl"), &diags)
			if sslModel != nil {
				return &kbapi.NewOutputSsl{
					Certificate:            sslModel.Certificate.ValueStringPointer(),
					CertificateAuthorities: utils.SliceRef(utils.ListTypeToSlice_String(ctx, sslModel.CertificateAuthorities, path.Root("certificate_authorities"), &diags)),
					Key:                    sslModel.Key.ValueStringPointer(),
				}
			}
		}
		return nil
	}

	outputType := model.Type.ValueString()
	switch outputType {
	case "elasticsearch":
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
			Ssl:                  doSsl(),
		}

		err := union.FromNewOutputElasticsearch(body)
		if err != nil {
			diags.AddError(err.Error(), "")
			return
		}

	case "logstash":
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
			Ssl:                  doSsl(),
		}

		err := union.FromNewOutputLogstash(body)
		if err != nil {
			diags.AddError(err.Error(), "")
			return
		}

	case "kafka":
		// Check minimum version requirement for Kafka output type
		if supported, versionDiags := client.EnforceMinVersion(ctx, MinVersionOutputKafka); versionDiags.HasError() {
			diags.Append(utils.FrameworkDiagsFromSDK(versionDiags)...)
			return
		} else if !supported {
			diags.AddError("Unsupported version for Kafka output",
				fmt.Sprintf("Kafka output type requires server version %s or higher", MinVersionOutputKafka.String()))
			return
		}

		// Extract kafka model from nested structure
		var kafkaModel outputKafkaModel
		if !model.Kafka.IsNull() {
			kafkaObj := utils.ObjectTypeAs[outputKafkaModel](ctx, model.Kafka, path.Root("kafka"), &diags)
			kafkaModel = *kafkaObj
		}

		// Helper functions for Kafka-specific complex types
		doHeaders := func() *[]struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} {
			if utils.IsKnown(kafkaModel.Headers) {
				headerModels := utils.ListTypeAs[outputHeadersModel](ctx, kafkaModel.Headers, path.Root("kafka").AtName("headers"), &diags)
				if len(headerModels) > 0 {
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
					return &headers
				}
			}
			return nil
		}

		doHash := func() *struct {
			Hash   *string `json:"hash,omitempty"`
			Random *bool   `json:"random,omitempty"`
		} {
			if utils.IsKnown(kafkaModel.Hash) {
				hashModels := utils.ListTypeAs[outputHashModel](ctx, kafkaModel.Hash, path.Root("kafka").AtName("hash"), &diags)
				if len(hashModels) > 0 {
					return &struct {
						Hash   *string `json:"hash,omitempty"`
						Random *bool   `json:"random,omitempty"`
					}{
						Hash:   hashModels[0].Hash.ValueStringPointer(),
						Random: hashModels[0].Random.ValueBoolPointer(),
					}
				}
			}
			return nil
		}

		doRandom := func() *struct {
			GroupEvents *float32 `json:"group_events,omitempty"`
		} {
			if utils.IsKnown(kafkaModel.Random) {
				randomModels := utils.ListTypeAs[outputRandomModel](ctx, kafkaModel.Random, path.Root("kafka").AtName("random"), &diags)
				if len(randomModels) > 0 {
					return &struct {
						GroupEvents *float32 `json:"group_events,omitempty"`
					}{
						GroupEvents: func() *float32 {
							if !randomModels[0].GroupEvents.IsNull() {
								val := float32(randomModels[0].GroupEvents.ValueFloat64())
								return &val
							}
							return nil
						}(),
					}
				}
			}
			return nil
		}

		doRoundRobin := func() *struct {
			GroupEvents *float32 `json:"group_events,omitempty"`
		} {
			if utils.IsKnown(kafkaModel.RoundRobin) {
				roundRobinModels := utils.ListTypeAs[outputRoundRobinModel](ctx, kafkaModel.RoundRobin, path.Root("kafka").AtName("round_robin"), &diags)
				if len(roundRobinModels) > 0 {
					return &struct {
						GroupEvents *float32 `json:"group_events,omitempty"`
					}{
						GroupEvents: func() *float32 {
							if !roundRobinModels[0].GroupEvents.IsNull() {
								val := float32(roundRobinModels[0].GroupEvents.ValueFloat64())
								return &val
							}
							return nil
						}(),
					}
				}
			}
			return nil
		}

		doSasl := func() *struct {
			Mechanism *kbapi.NewOutputKafkaSaslMechanism `json:"mechanism,omitempty"`
		} {
			if utils.IsKnown(kafkaModel.Sasl) {
				saslModels := utils.ListTypeAs[outputSaslModel](ctx, kafkaModel.Sasl, path.Root("kafka").AtName("sasl"), &diags)
				if len(saslModels) > 0 && !saslModels[0].Mechanism.IsNull() {
					mechanism := kbapi.NewOutputKafkaSaslMechanism(saslModels[0].Mechanism.ValueString())
					return &struct {
						Mechanism *kbapi.NewOutputKafkaSaslMechanism `json:"mechanism,omitempty"`
					}{
						Mechanism: &mechanism,
					}
				}
			}
			return nil
		}

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
			Ssl:                  doSsl(),
			// Kafka-specific fields
			AuthType: func() kbapi.NewOutputKafkaAuthType {
				if !kafkaModel.AuthType.IsNull() {
					return kbapi.NewOutputKafkaAuthType(kafkaModel.AuthType.ValueString())
				}
				return kbapi.NewOutputKafkaAuthTypeNone
			}(),
			BrokerTimeout: func() *float32 {
				if !kafkaModel.BrokerTimeout.IsNull() {
					val := float32(kafkaModel.BrokerTimeout.ValueFloat64())
					return &val
				}
				return nil
			}(),
			ClientId: kafkaModel.ClientId.ValueStringPointer(),
			Compression: func() *kbapi.NewOutputKafkaCompression {
				if !kafkaModel.Compression.IsNull() {
					comp := kbapi.NewOutputKafkaCompression(kafkaModel.Compression.ValueString())
					return &comp
				}
				return nil
			}(),
			CompressionLevel: func() *float32 {
				if !kafkaModel.CompressionLevel.IsNull() && !kafkaModel.Compression.IsNull() && kafkaModel.Compression.ValueString() == "gzip" {
					val := float32(kafkaModel.CompressionLevel.ValueFloat64())
					return &val
				}
				return nil
			}(),
			ConnectionType: kafkaModel.ConnectionType.ValueStringPointer(),
			Topic:          kafkaModel.Topic.ValueStringPointer(),
			Partition: func() *kbapi.NewOutputKafkaPartition {
				if !kafkaModel.Partition.IsNull() {
					part := kbapi.NewOutputKafkaPartition(kafkaModel.Partition.ValueString())
					return &part
				}
				return nil
			}(),
			RequiredAcks: func() *kbapi.NewOutputKafkaRequiredAcks {
				if !kafkaModel.RequiredAcks.IsNull() {
					acks := kbapi.NewOutputKafkaRequiredAcks(kafkaModel.RequiredAcks.ValueInt64())
					return &acks
				}
				return nil
			}(),
			Timeout: func() *float32 {
				if !kafkaModel.Timeout.IsNull() {
					val := float32(kafkaModel.Timeout.ValueFloat64())
					return &val
				}
				return nil
			}(),
			Version:    kafkaModel.Version.ValueStringPointer(),
			Username:   kafkaModel.Username.ValueStringPointer(),
			Password:   kafkaModel.Password.ValueStringPointer(),
			Key:        kafkaModel.Key.ValueStringPointer(),
			Headers:    doHeaders(),
			Hash:       doHash(),
			Random:     doRandom(),
			RoundRobin: doRoundRobin(),
			Sasl:       doSasl(),
		}

		err := union.FromNewOutputKafka(body)
		if err != nil {
			diags.AddError(err.Error(), "")
			return
		}

	default:
		diags.AddError(fmt.Sprintf("unhandled output type: %s", outputType), "")
	}

	return
}

func (model outputModel) toAPIUpdateModel(ctx context.Context, client *clients.ApiClient) (union kbapi.UpdateOutputUnion, diags diag.Diagnostics) {
	doSsl := func() *kbapi.UpdateOutputSsl {
		if utils.IsKnown(model.Ssl) {
			sslModel := utils.ObjectTypeAs[outputSslModel](ctx, model.Ssl, path.Root("ssl"), &diags)
			if sslModel != nil {
				return &kbapi.UpdateOutputSsl{
					Certificate:            sslModel.Certificate.ValueStringPointer(),
					CertificateAuthorities: utils.SliceRef(utils.ListTypeToSlice_String(ctx, sslModel.CertificateAuthorities, path.Root("certificate_authorities"), &diags)),
					Key:                    sslModel.Key.ValueStringPointer(),
				}
			}
		}
		return nil
	}

	outputType := model.Type.ValueString()
	switch outputType {
	case "elasticsearch":
		body := kbapi.UpdateOutputElasticsearch{
			Type:                 utils.Pointer(kbapi.Elasticsearch),
			CaSha256:             model.CaSha256.ValueStringPointer(),
			CaTrustedFingerprint: model.CaTrustedFingerprint.ValueStringPointer(),
			ConfigYaml:           model.ConfigYaml.ValueStringPointer(),
			Hosts:                utils.SliceRef(utils.ListTypeToSlice_String(ctx, model.Hosts, path.Root("hosts"), &diags)),
			IsDefault:            model.DefaultIntegrations.ValueBoolPointer(),
			IsDefaultMonitoring:  model.DefaultMonitoring.ValueBoolPointer(),
			Name:                 model.Name.ValueStringPointer(),
			Ssl:                  doSsl(),
		}

		err := union.FromUpdateOutputElasticsearch(body)
		if err != nil {
			diags.AddError(err.Error(), "")
			return
		}

	case "logstash":
		body := kbapi.UpdateOutputLogstash{
			Type:                 utils.Pointer(kbapi.Logstash),
			CaSha256:             model.CaSha256.ValueStringPointer(),
			CaTrustedFingerprint: model.CaTrustedFingerprint.ValueStringPointer(),
			ConfigYaml:           model.ConfigYaml.ValueStringPointer(),
			Hosts:                utils.SliceRef(utils.ListTypeToSlice_String(ctx, model.Hosts, path.Root("hosts"), &diags)),
			IsDefault:            model.DefaultIntegrations.ValueBoolPointer(),
			IsDefaultMonitoring:  model.DefaultMonitoring.ValueBoolPointer(),
			Name:                 model.Name.ValueStringPointer(),
			Ssl:                  doSsl(),
		}

		err := union.FromUpdateOutputLogstash(body)
		if err != nil {
			diags.AddError(err.Error(), "")
			return
		}

	case "kafka":
		// Check minimum version requirement for Kafka output type
		if supported, versionDiags := client.EnforceMinVersion(ctx, MinVersionOutputKafka); versionDiags.HasError() {
			diags.Append(utils.FrameworkDiagsFromSDK(versionDiags)...)
			return
		} else if !supported {
			diags.AddError("Unsupported version for Kafka output",
				fmt.Sprintf("Kafka output type requires server version %s or higher", MinVersionOutputKafka.String()))
			return
		}

		// Extract kafka model from nested structure
		var kafkaModel outputKafkaModel
		if !model.Kafka.IsNull() {
			kafkaObj := utils.ObjectTypeAs[outputKafkaModel](ctx, model.Kafka, path.Root("kafka"), &diags)
			kafkaModel = *kafkaObj
		}

		// Helper functions for Kafka-specific complex types (Update version)
		doHeaders := func() *[]struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} {
			if utils.IsKnown(kafkaModel.Headers) {
				headerModels := utils.ListTypeAs[outputHeadersModel](ctx, kafkaModel.Headers, path.Root("kafka").AtName("headers"), &diags)
				if len(headerModels) > 0 {
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
					return &headers
				}
			}
			return nil
		}

		doHash := func() *struct {
			Hash   *string `json:"hash,omitempty"`
			Random *bool   `json:"random,omitempty"`
		} {
			if utils.IsKnown(kafkaModel.Hash) {
				hashModels := utils.ListTypeAs[outputHashModel](ctx, kafkaModel.Hash, path.Root("kafka").AtName("hash"), &diags)
				if len(hashModels) > 0 {
					return &struct {
						Hash   *string `json:"hash,omitempty"`
						Random *bool   `json:"random,omitempty"`
					}{
						Hash:   hashModels[0].Hash.ValueStringPointer(),
						Random: hashModels[0].Random.ValueBoolPointer(),
					}
				}
			}
			return nil
		}

		doRandom := func() *struct {
			GroupEvents *float32 `json:"group_events,omitempty"`
		} {
			if utils.IsKnown(kafkaModel.Random) {
				randomModels := utils.ListTypeAs[outputRandomModel](ctx, kafkaModel.Random, path.Root("kafka").AtName("random"), &diags)
				if len(randomModels) > 0 {
					return &struct {
						GroupEvents *float32 `json:"group_events,omitempty"`
					}{
						GroupEvents: func() *float32 {
							if !randomModels[0].GroupEvents.IsNull() {
								val := float32(randomModels[0].GroupEvents.ValueFloat64())
								return &val
							}
							return nil
						}(),
					}
				}
			}
			return nil
		}

		doRoundRobin := func() *struct {
			GroupEvents *float32 `json:"group_events,omitempty"`
		} {
			if utils.IsKnown(kafkaModel.RoundRobin) {
				roundRobinModels := utils.ListTypeAs[outputRoundRobinModel](ctx, kafkaModel.RoundRobin, path.Root("kafka").AtName("round_robin"), &diags)
				if len(roundRobinModels) > 0 {
					return &struct {
						GroupEvents *float32 `json:"group_events,omitempty"`
					}{
						GroupEvents: func() *float32 {
							if !roundRobinModels[0].GroupEvents.IsNull() {
								val := float32(roundRobinModels[0].GroupEvents.ValueFloat64())
								return &val
							}
							return nil
						}(),
					}
				}
			}
			return nil
		}

		doSasl := func() *struct {
			Mechanism *kbapi.UpdateOutputKafkaSaslMechanism `json:"mechanism,omitempty"`
		} {
			if utils.IsKnown(kafkaModel.Sasl) {
				saslModels := utils.ListTypeAs[outputSaslModel](ctx, kafkaModel.Sasl, path.Root("kafka").AtName("sasl"), &diags)
				if len(saslModels) > 0 && !saslModels[0].Mechanism.IsNull() {
					mechanism := kbapi.UpdateOutputKafkaSaslMechanism(saslModels[0].Mechanism.ValueString())
					return &struct {
						Mechanism *kbapi.UpdateOutputKafkaSaslMechanism `json:"mechanism,omitempty"`
					}{
						Mechanism: &mechanism,
					}
				}
			}
			return nil
		}

		body := kbapi.UpdateOutputKafka{
			Type:                 utils.Pointer(kbapi.Kafka),
			CaSha256:             model.CaSha256.ValueStringPointer(),
			CaTrustedFingerprint: model.CaTrustedFingerprint.ValueStringPointer(),
			ConfigYaml:           model.ConfigYaml.ValueStringPointer(),
			Hosts:                utils.SliceRef(utils.ListTypeToSlice_String(ctx, model.Hosts, path.Root("hosts"), &diags)),
			IsDefault:            model.DefaultIntegrations.ValueBoolPointer(),
			IsDefaultMonitoring:  model.DefaultMonitoring.ValueBoolPointer(),
			Name:                 model.Name.ValueString(),
			Ssl:                  doSsl(),
			// Kafka-specific fields
			AuthType: func() *kbapi.UpdateOutputKafkaAuthType {
				if !kafkaModel.AuthType.IsNull() {
					authType := kbapi.UpdateOutputKafkaAuthType(kafkaModel.AuthType.ValueString())
					return &authType
				}
				return nil
			}(),
			BrokerTimeout: func() *float32 {
				if !kafkaModel.BrokerTimeout.IsNull() {
					val := float32(kafkaModel.BrokerTimeout.ValueFloat64())
					return &val
				}
				return nil
			}(),
			ClientId: kafkaModel.ClientId.ValueStringPointer(),
			Compression: func() *kbapi.UpdateOutputKafkaCompression {
				if !kafkaModel.Compression.IsNull() {
					comp := kbapi.UpdateOutputKafkaCompression(kafkaModel.Compression.ValueString())
					return &comp
				}
				return nil
			}(),
			CompressionLevel: func() *float32 {
				if !kafkaModel.CompressionLevel.IsNull() && !kafkaModel.Compression.IsNull() && kafkaModel.Compression.ValueString() == "gzip" {
					val := float32(kafkaModel.CompressionLevel.ValueFloat64())
					return &val
				}
				return nil
			}(),
			ConnectionType: kafkaModel.ConnectionType.ValueStringPointer(),
			Topic:          kafkaModel.Topic.ValueStringPointer(),
			Partition: func() *kbapi.UpdateOutputKafkaPartition {
				if !kafkaModel.Partition.IsNull() {
					part := kbapi.UpdateOutputKafkaPartition(kafkaModel.Partition.ValueString())
					return &part
				}
				return nil
			}(),
			RequiredAcks: func() *kbapi.UpdateOutputKafkaRequiredAcks {
				if !kafkaModel.RequiredAcks.IsNull() {
					acks := kbapi.UpdateOutputKafkaRequiredAcks(kafkaModel.RequiredAcks.ValueInt64())
					return &acks
				}
				return nil
			}(),
			Timeout: func() *float32 {
				if !kafkaModel.Timeout.IsNull() {
					val := float32(kafkaModel.Timeout.ValueFloat64())
					return &val
				}
				return nil
			}(),
			Version:    kafkaModel.Version.ValueStringPointer(),
			Username:   kafkaModel.Username.ValueStringPointer(),
			Password:   kafkaModel.Password.ValueStringPointer(),
			Key:        kafkaModel.Key.ValueStringPointer(),
			Headers:    doHeaders(),
			Hash:       doHash(),
			Random:     doRandom(),
			RoundRobin: doRoundRobin(),
			Sasl:       doSasl(),
		}

		err := union.FromUpdateOutputKafka(body)
		if err != nil {
			diags.AddError(err.Error(), "")
			return
		}

	default:
		diags.AddError(fmt.Sprintf("unhandled output type: %s", outputType), "")
	}

	return
}
