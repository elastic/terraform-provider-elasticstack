package alerting_rule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// readRuleFromAPI fetches a rule from the API and populates the given model
// Returns true if the rule was found, false if it doesn't exist
func (r *Resource) readRuleFromAPI(ctx context.Context, client *clients.ApiClient, model *tfModel) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	compositeID, idDiags := model.GetID()
	diags.Append(idDiags...)
	if diags.HasError() {
		return false, diags
	}

	rule, sdkDiags := kibana.GetAlertingRule(ctx, client, compositeID.ResourceId, compositeID.ClusterId)
	if rule == nil && sdkDiags == nil {
		// Resource not found
		return false, diags
	}
	if sdkDiags.HasError() {
		for _, d := range sdkDiags {
			diags.AddError(d.Summary, d.Detail)
		}
		return false, diags
	}

	diags.Append(model.populateFromAPI(ctx, rule, compositeID)...)
	return true, diags
}

func (r *Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state tfModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	client, diags := clients.MaybeNewApiClientFromFrameworkResource(ctx, state.KibanaConnection, r.client)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	exists, readDiags := r.readRuleFromAPI(ctx, client, &state)
	response.Diagnostics.Append(readDiags...)
	if response.Diagnostics.HasError() {
		return
	}

	if !exists {
		response.State.RemoveResource(ctx)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}
