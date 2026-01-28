package alerting_rule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state alertingRuleModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	compositeID, diags := clients.CompositeIdFromStrFw(state.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, diags := r.readRule(ctx, compositeID.ClusterId, compositeID.ResourceId)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if rule == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, rule)
	resp.Diagnostics.Append(diags...)
}

func (r *Resource) readRule(ctx context.Context, spaceID, ruleID string) (*alertingRuleModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		diags.AddError("Failed to get Kibana client", err.Error())
		return nil, diags
	}

	ruleResp, readDiags := kibana_oapi.GetAlertingRule(ctx, client, ruleID)
	diags.Append(readDiags...)
	if diags.HasError() {
		return nil, diags
	}

	if ruleResp == nil {
		return nil, nil
	}

	model := &alertingRuleModel{}
	diags.Append(model.populateFromAPI(ctx, ruleResp, spaceID)...)

	return model, diags
}
