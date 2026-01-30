package alerting_rule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan tfModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	client, diags := clients.MaybeNewApiClientFromFrameworkResource(ctx, plan.KibanaConnection, r.client)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	serverVersion, sdkDiags := client.ServerVersion(ctx)
	if sdkDiags.HasError() {
		for _, d := range sdkDiags {
			response.Diagnostics.AddError(d.Summary, d.Detail)
		}
		return
	}

	// Version check for alert_delay (only supported from v8.13.0)
	if utils.IsKnown(plan.AlertDelay) && !plan.AlertDelay.IsNull() {
		if serverVersion.LessThan(alertDelayMinSupportedVersion) {
			response.Diagnostics.AddError(
				"alert_delay is only supported for Elasticsearch v8.13 or higher",
				"alert_delay is only supported for Elasticsearch v8.13 or higher",
			)
			return
		}
	}

	rule, ruleDiags := plan.toAPIModel(ctx)
	response.Diagnostics.Append(ruleDiags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Get the rule ID from the current state
	compositeID, idDiags := plan.GetID()
	response.Diagnostics.Append(idDiags...)
	if response.Diagnostics.HasError() {
		return
	}
	rule.RuleID = compositeID.ResourceId

	// Version check for actions features
	for _, action := range rule.Actions {
		if action.Frequency != nil && serverVersion.LessThan(frequencyMinSupportedVersion) {
			response.Diagnostics.AddError(
				"actions.frequency is only supported for Elasticsearch v8.6 or higher",
				"actions.frequency is only supported for Elasticsearch v8.6 or higher",
			)
			return
		}

		if action.AlertsFilter != nil && serverVersion.LessThan(alertsFilterMinSupportedVersion) {
			response.Diagnostics.AddError(
				"actions.alerts_filter is only supported for Elasticsearch v8.9 or higher",
				"actions.alerts_filter is only supported for Elasticsearch v8.9 or higher",
			)
			return
		}
	}

	res, sdkDiags := kibana.UpdateAlertingRule(ctx, client, rule)
	if sdkDiags.HasError() {
		for _, d := range sdkDiags {
			response.Diagnostics.AddError(d.Summary, d.Detail)
		}
		return
	}

	newCompositeID := &clients.CompositeId{ClusterId: rule.SpaceID, ResourceId: res.RuleID}
	plan.ID = types.StringValue(newCompositeID.String())

	// Read back the rule to populate all computed fields
	exists, readDiags := r.readRuleFromAPI(ctx, client, &plan)
	response.Diagnostics.Append(readDiags...)
	if response.Diagnostics.HasError() {
		return
	}

	if !exists {
		response.Diagnostics.AddError("Rule not found after update", "The rule was updated but could not be found afterward")
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
}
