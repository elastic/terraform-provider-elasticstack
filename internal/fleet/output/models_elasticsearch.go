package output

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (model *outputModel) fromAPIElasticsearchModel(ctx context.Context, data *kbapi.OutputElasticsearch) (diags diag.Diagnostics) {
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
	return
}

func (model outputModel) toAPICreateElasticsearchModel(ctx context.Context) (kbapi.NewOutputUnion, diag.Diagnostics) {
	ssl, diags := objectValueToSSL(ctx, model.Ssl)
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

func (model outputModel) toAPIUpdateElasticsearchModel(ctx context.Context) (kbapi.UpdateOutputUnion, diag.Diagnostics) {
	ssl, diags := objectValueToSSLUpdate(ctx, model.Ssl)
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
