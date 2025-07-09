package output

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	Ssl                  types.List   `tfsdk:"ssl"` //> outputSslModel
	ConfigYaml           types.String `tfsdk:"config_yaml"`
	Kafka                types.List   `tfsdk:"kafka"` //> kafkaModel
}

type outputSslModel struct {
	CertificateAuthorities types.List   `tfsdk:"certificate_authorities"` //> string
	Certificate            types.String `tfsdk:"certificate"`
	Key                    types.String `tfsdk:"key"`
}

type kafkaSaslModel struct {
	Mechanism types.String `tfsdk:"mechanism"`
	Username  types.String `tfsdk:"username"`
	Password  types.String `tfsdk:"password"`
}

type kafkaModel struct {
	Topic       types.String `tfsdk:"topic"`
	ClientID    types.String `tfsdk:"client_id"`
	Version     types.String `tfsdk:"version"`
	Compression types.String `tfsdk:"compression"`
	Sasl        types.List   `tfsdk:"sasl"` //> kafkaSaslModel
}

func (model *outputModel) populateFromAPI(ctx context.Context, union *kbapi.OutputUnion) (diags diag.Diagnostics) {
	if union == nil {
		return
	}

	doSsl := func(ssl *kbapi.OutputSsl) types.List {
		if ssl != nil {
			p := path.Root("ssl")
			sslModels := []outputSslModel{{
				CertificateAuthorities: utils.SliceToListType_String(ctx, utils.Deref(ssl.CertificateAuthorities), p.AtName("certificate_authorities"), &diags),
				Certificate:            types.StringPointerValue(ssl.Certificate),
				Key:                    types.StringPointerValue(ssl.Key),
			}}
			list, nd := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: getSslAttrs()}, sslModels)
			diags.Append(nd...)
			return list
		} else {
			return types.ListNull(types.ObjectType{AttrTypes: getSslAttrs()})
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

		var saslList types.List
		if data.Sasl != nil {
			saslModels := []kafkaSaslModel{{
				Mechanism: types.StringValue(string(*data.Sasl.Mechanism)),
				Username:  types.StringValue(data.Username),
				Password:  types.StringValue(data.Password),
			}}
			var nd diag.Diagnostics
			saslList, nd = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: getKafkaSaslAttrTypes()}, saslModels)
			diags.Append(nd...)
		} else {
			saslList = types.ListNull(types.ObjectType{AttrTypes: getKafkaSaslAttrTypes()})
		}

		kafkaModels := []kafkaModel{{
			Topic:       types.StringPointerValue(data.Topic),
			ClientID:    types.StringPointerValue(data.ClientId),
			Version:     types.StringPointerValue(data.Version),
			Compression: types.StringValue(string(*data.Compression)),
			Sasl:        saslList,
		}}
		list, nd := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: getKafkaAttrTypes()}, kafkaModels)
		diags.Append(nd...)
		model.Kafka = list

	default:
		diags.AddError(fmt.Sprintf("unhandled output type: %s", discriminator), "")
	}

	return
}

func (model outputModel) toAPICreateModel(ctx context.Context) (union kbapi.NewOutputUnion, diags diag.Diagnostics) {
	doSsl := func() *kbapi.NewOutputSsl {
		if utils.IsKnown(model.Ssl) {
			sslModels := utils.ListTypeAs[outputSslModel](ctx, model.Ssl, path.Root("ssl"), &diags)
			if len(sslModels) > 0 {
				return &kbapi.NewOutputSsl{
					Certificate:            sslModels[0].Certificate.ValueStringPointer(),
					CertificateAuthorities: utils.SliceRef(utils.ListTypeToSlice_String(ctx, sslModels[0].CertificateAuthorities, path.Root("certificate_authorities"), &diags)),
					Key:                    sslModels[0].Key.ValueStringPointer(),
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
		}

		if utils.IsKnown(model.Kafka) {
			kafkaModels := utils.ListTypeAs[kafkaModel](ctx, model.Kafka, path.Root("kafka"), &diags)
			if len(kafkaModels) > 0 {
				k := kafkaModels[0]
				body.Topic = k.Topic.ValueStringPointer()
				body.ClientId = k.ClientID.ValueStringPointer()
				body.Version = k.Version.ValueStringPointer()
				compression := k.Compression.ValueString()
				if compression != "" {
					body.Compression = utils.Pointer(kbapi.NewOutputKafkaCompression(compression))
				}

				if utils.IsKnown(k.Sasl) {
					saslModels := utils.ListTypeAs[kafkaSaslModel](ctx, k.Sasl, path.Root("sasl"), &diags)
					if len(saslModels) > 0 {
						s := saslModels[0]
						body.Username = s.Username.ValueString()
						body.Password = s.Password.ValueString()
						mechanism := s.Mechanism.ValueStringPointer()
						if mechanism != nil {
							body.Sasl = &struct {
								Mechanism *kbapi.NewOutputKafkaSaslMechanism `json:"mechanism,omitempty"`
							}{
								Mechanism: (*kbapi.NewOutputKafkaSaslMechanism)(mechanism),
							}
						}
					}
				}
			}
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

func (model outputModel) toAPIUpdateModel(ctx context.Context) (union kbapi.UpdateOutputUnion, diags diag.Diagnostics) {
	doSsl := func() *kbapi.UpdateOutputSsl {
		if utils.IsKnown(model.Ssl) {
			sslModels := utils.ListTypeAs[outputSslModel](ctx, model.Ssl, path.Root("ssl"), &diags)
			if len(sslModels) > 0 {
				return &kbapi.UpdateOutputSsl{
					Certificate:            sslModels[0].Certificate.ValueStringPointer(),
					CertificateAuthorities: utils.SliceRef(utils.ListTypeToSlice_String(ctx, sslModels[0].CertificateAuthorities, path.Root("certificate_authorities"), &diags)),
					Key:                    sslModels[0].Key.ValueStringPointer(),
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
		}

		if utils.IsKnown(model.Kafka) {
			kafkaModels := utils.ListTypeAs[kafkaModel](ctx, model.Kafka, path.Root("kafka"), &diags)
			if len(kafkaModels) > 0 {
				k := kafkaModels[0]
				body.Topic = k.Topic.ValueStringPointer()
				body.ClientId = k.ClientID.ValueStringPointer()
				body.Version = k.Version.ValueStringPointer()
				compression := k.Compression.ValueString()
				if compression != "" {
					body.Compression = utils.Pointer(kbapi.UpdateOutputKafkaCompression(compression))
				}

				if utils.IsKnown(k.Sasl) {
					saslModels := utils.ListTypeAs[kafkaSaslModel](ctx, k.Sasl, path.Root("sasl"), &diags)
					if len(saslModels) > 0 {
						s := saslModels[0]
						body.Username = s.Username.ValueString()
						body.Password = s.Password.ValueString()
						mechanism := s.Mechanism.ValueStringPointer()
						if mechanism != nil {
							body.Sasl = &struct {
								Mechanism *kbapi.UpdateOutputKafkaSaslMechanism `json:"mechanism,omitempty"`
							}{
								Mechanism: (*kbapi.UpdateOutputKafkaSaslMechanism)(mechanism),
							}
						}
					}
				}
			}
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

func getSslAttrs() map[string]attr.Type {
	return map[string]attr.Type{
		"certificate_authorities": types.ListType{ElemType: types.StringType},
		"certificate":             types.StringType,
		"key":                     types.StringType,
	}
}

func getKafkaSaslAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"mechanism": types.StringType,
		"username":  types.StringType,
		"password":  types.StringType,
	}
}

func getKafkaAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"topic":       types.StringType,
		"client_id":   types.StringType,
		"version":     types.StringType,
		"compression": types.StringType,
		"sasl":        types.ListType{ElemType: types.ObjectType{AttrTypes: getKafkaSaslAttrTypes()}},
	}
}
