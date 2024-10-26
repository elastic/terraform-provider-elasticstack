package enrollment_tokens

import (
	"context"

	fleetapi "github.com/elastic/terraform-provider-elasticstack/generated/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (d *enrollmentTokensDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model enrollmentTokensModel

	diags := req.Config.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := d.client.GetFleetClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	var tokens []fleetapi.EnrollmentApiKey
	policyID := model.PolicyID.ValueString()
	if policyID == "" {
		tokens, diags = fleet.GetEnrollmentTokens(ctx, client)
	} else {
		tokens, diags = fleet.GetEnrollmentTokensByPolicy(ctx, client, policyID)
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if policyID != "" {
		model.ID = types.StringValue(policyID)
	} else {
		hash, err := utils.StringToHash(client.URL)
		if err != nil {
			resp.Diagnostics.AddError(err.Error(), "")
			return
		}
		model.ID = types.StringPointerValue(hash)
	}

	diags = model.populateFromAPI(ctx, tokens)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
}
