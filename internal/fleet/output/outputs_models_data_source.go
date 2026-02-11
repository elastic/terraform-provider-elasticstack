package output

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type outputsDataSourceModel struct {
	ID                  types.String `tfsdk:"id"`
	OutputID            types.String `tfsdk:"output_id"`
	Type                types.String `tfsdk:"type"`
	DefaultIntegrations types.Bool   `tfsdk:"default_integrations"`
	DefaultMonitoring   types.Bool   `tfsdk:"default_monitoring"`
	SpaceID             types.String `tfsdk:"space_id"`
	Items               types.List   `tfsdk:"items"` //> outputItemModel
}

type outputItemModel struct {
	ID                   types.String `tfsdk:"id"`
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

func (model *outputsDataSourceModel) populateFromAPI(ctx context.Context, outputs []kbapi.OutputUnion) diag.Diagnostics {
	var diags diag.Diagnostics
	items := make([]outputItemModel, 0, len(outputs))
	for i := range outputs {
		item, itemDiags := newOutputItemModel(ctx, &outputs[i])
		diags.Append(itemDiags...)
		if diags.HasError() {
			return diags
		}

		if model.matches(item) {
			items = append(items, item)
		}
	}

	model.Items = utils.SliceToListType(ctx, items, types.ObjectType{AttrTypes: getOutputItemAttrTypes()}, path.Root("items"), &diags, func(item outputItemModel, meta utils.ListMeta) outputItemModel {
		return item
	})
	return diags
}

func (model outputsDataSourceModel) matches(item outputItemModel) bool {
	if utils.IsKnown(model.OutputID) {
		if !utils.IsKnown(item.ID) || item.ID.ValueString() != model.OutputID.ValueString() {
			return false
		}
	}

	if utils.IsKnown(model.Type) {
		if !utils.IsKnown(item.Type) || item.Type.ValueString() != model.Type.ValueString() {
			return false
		}
	}

	if utils.IsKnown(model.DefaultIntegrations) {
		if !utils.IsKnown(item.DefaultIntegrations) || item.DefaultIntegrations.ValueBool() != model.DefaultIntegrations.ValueBool() {
			return false
		}
	}

	if utils.IsKnown(model.DefaultMonitoring) {
		if !utils.IsKnown(item.DefaultMonitoring) || item.DefaultMonitoring.ValueBool() != model.DefaultMonitoring.ValueBool() {
			return false
		}
	}

	return true
}

func newOutputItemModel(ctx context.Context, union *kbapi.OutputUnion) (outputItemModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	var resourceModel outputModel

	diags.Append(resourceModel.populateFromAPI(ctx, union)...)
	if diags.HasError() {
		return outputItemModel{}, diags
	}

	return outputItemModel{
		ID:                   resourceModel.OutputID,
		Name:                 resourceModel.Name,
		Type:                 resourceModel.Type,
		Hosts:                resourceModel.Hosts,
		CaSha256:             resourceModel.CaSha256,
		CaTrustedFingerprint: resourceModel.CaTrustedFingerprint,
		DefaultIntegrations:  resourceModel.DefaultIntegrations,
		DefaultMonitoring:    resourceModel.DefaultMonitoring,
		ConfigYaml:           resourceModel.ConfigYaml,
		Ssl:                  resourceModel.Ssl,
		Kafka:                resourceModel.Kafka,
	}, diags
}

func getOutputItemAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                     types.StringType,
		"name":                   types.StringType,
		"type":                   types.StringType,
		"hosts":                  types.ListType{ElemType: types.StringType},
		"ca_sha256":              types.StringType,
		"ca_trusted_fingerprint": types.StringType,
		"default_integrations":   types.BoolType,
		"default_monitoring":     types.BoolType,
		"config_yaml":            types.StringType,
		"ssl":                    types.ObjectType{AttrTypes: getSslAttrTypes()},
		"kafka":                  types.ObjectType{AttrTypes: getKafkaAttrTypes()},
	}
}
