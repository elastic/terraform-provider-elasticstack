package output

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type outputDataSourceModel struct {
	ID                   types.String `tfsdk:"id"`
	OutputID             types.String `tfsdk:"output_id"`
	SpaceID              types.String `tfsdk:"space_id"`
	Name                 types.String `tfsdk:"name"`
	Type                 types.String `tfsdk:"type"`
	Hosts                types.List   `tfsdk:"hosts"`
	CaSha256             types.String `tfsdk:"ca_sha256"`
	CaTrustedFingerprint types.String `tfsdk:"ca_trusted_fingerprint"`
	DefaultIntegrations  types.Bool   `tfsdk:"default_integrations"`
	DefaultMonitoring    types.Bool   `tfsdk:"default_monitoring"`
	ConfigYaml           types.String `tfsdk:"config_yaml"`
	Ssl                  types.Object `tfsdk:"ssl"`
	Kafka                types.Object `tfsdk:"kafka"`
}

func (model *outputDataSourceModel) populateFromAPI(ctx context.Context, union *kbapi.OutputUnion) diag.Diagnostics {
	var diags diag.Diagnostics
	var resourceModel outputModel

	diags.Append(resourceModel.populateFromAPI(ctx, union)...)
	if diags.HasError() {
		return diags
	}

	model.OutputID = resourceModel.OutputID
	model.Name = resourceModel.Name
	model.Type = resourceModel.Type
	model.Hosts = resourceModel.Hosts
	model.CaSha256 = resourceModel.CaSha256
	model.CaTrustedFingerprint = resourceModel.CaTrustedFingerprint
	model.DefaultIntegrations = resourceModel.DefaultIntegrations
	model.DefaultMonitoring = resourceModel.DefaultMonitoring
	model.ConfigYaml = resourceModel.ConfigYaml
	model.Ssl = resourceModel.Ssl
	model.Kafka = resourceModel.Kafka

	return diags
}
