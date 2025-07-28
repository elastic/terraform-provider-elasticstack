package enrich

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *enrichPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	diags := r.upsert(ctx, req.Plan, &resp.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *enrichPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	diags := r.upsert(ctx, req.Plan, &resp.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *enrichPolicyResource) upsert(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State) diag.Diagnostics {
	var data EnrichPolicyData
	var diags diag.Diagnostics
	diags.Append(plan.Get(ctx, &data)...)
	if diags.HasError() {
		return diags
	}

	policyName := data.Name.ValueString()
	id, sdkDiags := r.client.ID(ctx, policyName)
	diags.Append(utils.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return diags
	}

	client, diags := clients.MaybeNewApiClientFromFrameworkResource(ctx, data.ElasticsearchConnection, r.client)
	diags.Append(diags...)
	if diags.HasError() {
		return diags
	}

	// Convert framework types to model
	indices := utils.ListTypeToSlice_String(ctx, data.Indices, path.Empty(), &diags)
	if diags.HasError() {
		return diags
	}

	enrichFields := utils.ListTypeToSlice_String(ctx, data.EnrichFields, path.Empty(), &diags)
	if diags.HasError() {
		return diags
	}

	policy := &models.EnrichPolicy{
		Type:         data.PolicyType.ValueString(),
		Name:         policyName,
		Indices:      indices,
		MatchField:   data.MatchField.ValueString(),
		EnrichFields: enrichFields,
	}

	if !data.Query.IsNull() && !data.Query.IsUnknown() {
		policy.Query = data.Query.ValueString()
	}

	if sdkDiags := elasticsearch.PutEnrichPolicy(ctx, client, policy); sdkDiags.HasError() {
		diags.Append(utils.FrameworkDiagsFromSDK(sdkDiags)...)
		return diags
	}

	data.Id = types.StringValue(id.String())

	// Execute policy if requested
	if !data.Execute.IsNull() && !data.Execute.IsUnknown() && data.Execute.ValueBool() {
		if sdkDiags := elasticsearch.ExecuteEnrichPolicy(ctx, client, policyName); sdkDiags.HasError() {
			diags.Append(utils.FrameworkDiagsFromSDK(sdkDiags)...)
			return diags
		}
	}

	diags.Append(state.Set(ctx, &data)...)
	return diags
}