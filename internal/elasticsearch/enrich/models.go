package enrich

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EnrichPolicyData struct {
	Id                      types.String         `tfsdk:"id"`
	ElasticsearchConnection types.List           `tfsdk:"elasticsearch_connection"`
	Name                    types.String         `tfsdk:"name"`
	PolicyType              types.String         `tfsdk:"policy_type"`
	Indices                 types.Set            `tfsdk:"indices"`
	MatchField              types.String         `tfsdk:"match_field"`
	EnrichFields            types.Set            `tfsdk:"enrich_fields"`
	Query                   jsontypes.Normalized `tfsdk:"query"`
}

type EnrichPolicyDataWithExecute struct {
	EnrichPolicyData
	Execute types.Bool `tfsdk:"execute"`
}

// populateFromPolicy converts models.EnrichPolicy to EnrichPolicyData fields
func (data *EnrichPolicyData) populateFromPolicy(ctx context.Context, policy *models.EnrichPolicy, diagnostics *diag.Diagnostics) {
	data.Name = types.StringValue(policy.Name)
	data.PolicyType = types.StringValue(policy.Type)
	data.MatchField = types.StringValue(policy.MatchField)

	if policy.Query != "" && policy.Query != "null" {
		data.Query = jsontypes.NewNormalizedValue(policy.Query)
	} else {
		data.Query = jsontypes.NewNormalizedNull()
	}

	// Convert string slices to Set
	data.Indices = sliceToSetType_String(ctx, policy.Indices, path.Empty(), diagnostics)
	if diagnostics.HasError() {
		return
	}

	data.EnrichFields = sliceToSetType_String(ctx, policy.EnrichFields, path.Empty(), diagnostics)
}

// setTypeToSlice_String converts a types.Set to []string
func setTypeToSlice_String(ctx context.Context, value types.Set, p path.Path, diags *diag.Diagnostics) []string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}

	var elements []types.String
	d := value.ElementsAs(ctx, &elements, false)
	diags.Append(utils.ConvertToAttrDiags(d, p)...)
	if diags.HasError() {
		return nil
	}

	result := make([]string, len(elements))
	for i, elem := range elements {
		result[i] = elem.ValueString()
	}
	return result
}

// sliceToSetType_String converts a []string to types.Set
func sliceToSetType_String(ctx context.Context, value []string, p path.Path, diags *diag.Diagnostics) types.Set {
	if value == nil {
		return types.SetNull(types.StringType)
	}

	// Convert []string to []attr.Value
	elements := make([]attr.Value, len(value))
	for i, v := range value {
		elements[i] = types.StringValue(v)
	}

	set, d := types.SetValue(types.StringType, elements)
	diags.Append(utils.ConvertToAttrDiags(d, p)...)
	return set
}
