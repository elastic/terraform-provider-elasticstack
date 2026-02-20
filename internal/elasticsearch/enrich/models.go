package enrich

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PolicyData struct {
	ID                      types.String         `tfsdk:"id"`
	ElasticsearchConnection types.List           `tfsdk:"elasticsearch_connection"`
	Name                    types.String         `tfsdk:"name"`
	PolicyType              types.String         `tfsdk:"policy_type"`
	Indices                 types.Set            `tfsdk:"indices"`
	MatchField              types.String         `tfsdk:"match_field"`
	EnrichFields            types.Set            `tfsdk:"enrich_fields"`
	Query                   jsontypes.Normalized `tfsdk:"query"`
}

type PolicyDataWithExecute struct {
	PolicyData
	Execute types.Bool `tfsdk:"execute"`
}

// populateFromPolicy converts models.EnrichPolicy to PolicyData fields
func (data *PolicyData) populateFromPolicy(ctx context.Context, policy *models.EnrichPolicy, diagnostics *diag.Diagnostics) {
	data.Name = types.StringValue(policy.Name)
	data.PolicyType = types.StringValue(policy.Type)
	data.MatchField = types.StringValue(policy.MatchField)

	if policy.Query != "" && policy.Query != "null" {
		data.Query = jsontypes.NewNormalizedValue(policy.Query)
	} else {
		data.Query = jsontypes.NewNormalizedNull()
	}

	// Convert string slices to Set
	data.Indices = typeutils.SetValueFrom(ctx, policy.Indices, types.StringType, path.Empty(), diagnostics)
	if diagnostics.HasError() {
		return
	}

	data.EnrichFields = typeutils.SetValueFrom(ctx, policy.EnrichFields, types.StringType, path.Empty(), diagnostics)
}
