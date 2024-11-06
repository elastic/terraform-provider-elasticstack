package output

import (
	"context"
	"fmt"

	fleetapi "github.com/elastic/terraform-provider-elasticstack/generated/fleet"
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
	Ssl                  types.List   `tfsdk:"ssl"` //> outputSslModel
	ConfigYaml           types.String `tfsdk:"config_yaml"`
}

type outputSslModel struct {
	CertificateAuthorities types.List   `tfsdk:"certificate_authorities"` //> string
	Certificate            types.String `tfsdk:"certificate"`
	Key                    types.String `tfsdk:"key"`
}

func (model *outputModel) populateFromAPI(ctx context.Context, union *fleetapi.OutputUnion) (diags diag.Diagnostics) {
	if union == nil {
		return
	}

	doSsl := func(ssl *fleetapi.OutputSsl) types.List {
		if ssl != nil {
			p := path.Root("ssl")
			sslModels := []outputSslModel{{
				CertificateAuthorities: utils.SliceToListType_String(ctx, utils.Deref(ssl.CertificateAuthorities), p.AtName("certificate_authorities"), &diags),
				Certificate:            types.StringPointerValue(ssl.Certificate),
				Key:                    types.StringPointerValue(ssl.Key),
			}}
			list, nd := types.ListValueFrom(ctx, getSslAttrTypes(), sslModels)
			diags.Append(nd...)
			return list
		} else {
			return types.ListNull(getSslAttrTypes())
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

	default:
		diags.AddError(fmt.Sprintf("unhandled output type: %s", discriminator), "")
	}

	return
}

func (model outputModel) toAPICreateModel(ctx context.Context) (union fleetapi.NewOutputUnion, diags diag.Diagnostics) {
	doSsl := func() *fleetapi.NewOutputSsl {
		if utils.IsKnown(model.Ssl) {
			sslModels := utils.ListTypeAs[outputSslModel](ctx, model.Ssl, path.Root("ssl"), &diags)
			if len(sslModels) > 0 {
				return &fleetapi.NewOutputSsl{
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
		body := fleetapi.NewOutputElasticsearch{
			Type:                 fleetapi.NewOutputElasticsearchTypeElasticsearch,
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
		body := fleetapi.NewOutputLogstash{
			Type:                 fleetapi.NewOutputLogstashTypeLogstash,
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

	default:
		diags.AddError(fmt.Sprintf("unhandled output type: %s", outputType), "")
	}

	return
}

func (model outputModel) toAPIUpdateModel(ctx context.Context) (union fleetapi.UpdateOutputUnion, diags diag.Diagnostics) {
	doSsl := func() *fleetapi.UpdateOutputSsl {
		if utils.IsKnown(model.Ssl) {
			sslModels := utils.ListTypeAs[outputSslModel](ctx, model.Ssl, path.Root("ssl"), &diags)
			if len(sslModels) > 0 {
				return &fleetapi.UpdateOutputSsl{
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
		body := fleetapi.UpdateOutputElasticsearch{
			Type:                 utils.Pointer(fleetapi.Elasticsearch),
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
		body := fleetapi.UpdateOutputLogstash{
			Type:                 utils.Pointer(fleetapi.Logstash),
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

	default:
		diags.AddError(fmt.Sprintf("unhandled output type: %s", outputType), "")
	}

	return
}
