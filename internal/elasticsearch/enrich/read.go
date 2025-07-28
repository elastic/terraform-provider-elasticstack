package enrich

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *enrichPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EnrichPolicyData
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	compId, diags := clients.CompositeIdFromStrFw(data.Id.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	policyName := compId.ResourceId

	client, diags := clients.MaybeNewApiClientFromFrameworkResource(ctx, data.ElasticsearchConnection, r.client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, sdkDiags := elasticsearch.GetEnrichPolicy(ctx, client, policyName)
	resp.Diagnostics.Append(utils.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if policy == nil {
		tflog.Warn(ctx, fmt.Sprintf(`Enrich policy "%s" not found, removing from state`, policyName))
		resp.State.RemoveResource(ctx)
		return
	}

	// Convert model to framework types
	data.Name = types.StringValue(policy.Name)
	data.PolicyType = types.StringValue(policy.Type)
	data.MatchField = types.StringValue(policy.MatchField)

	if policy.Query != "" && policy.Query != "null" {
		data.Query = types.StringValue(policy.Query)
	} else {
		data.Query = types.StringNull()
	}

	// Convert string slices to List
	data.Indices = utils.SliceToListType_String(ctx, policy.Indices, path.Empty(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	data.EnrichFields = utils.SliceToListType_String(ctx, policy.EnrichFields, path.Empty(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
