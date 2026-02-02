package alerting_rule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Version thresholds for feature support
var (
	frequencyMinSupportedVersion    = version.Must(version.NewVersion("8.6.0"))
	alertsFilterMinSupportedVersion = version.Must(version.NewVersion("8.9.0"))
	alertDelayMinSupportedVersion   = version.Must(version.NewVersion("8.13.0"))
)

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan alertingRuleModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "Expected configured API client")
		return
	}

	// Get server version to validate version-specific features
	serverVersion, versionDiags := r.client.ServerVersion(ctx)
	if versionDiags.HasError() {
		resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(versionDiags)...)
		return
	}

	// Validate version-specific requirements
	if !utils.IsKnown(plan.NotifyWhen) || plan.NotifyWhen.ValueString() == "" {
		if serverVersion.LessThan(frequencyMinSupportedVersion) {
			resp.Diagnostics.AddError(
				"notify_when is required until v8.6",
				"notify_when is required until v8.6",
			)
			return
		}
	}

	if utils.IsKnown(plan.AlertDelay) && !plan.AlertDelay.IsNull() {
		if serverVersion.LessThan(alertDelayMinSupportedVersion) {
			resp.Diagnostics.AddError(
				"alert_delay is only supported for Elasticsearch v8.13 or higher",
				"alert_delay is only supported for Elasticsearch v8.13 or higher",
			)
			return
		}
	}

	// Validate version-specific requirements for actions
	if utils.IsKnown(plan.Actions) && !plan.Actions.IsNull() {
		var actions []actionModel
		diags := plan.Actions.ElementsAs(ctx, &actions, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, action := range actions {
			// Check frequency version requirement
			if utils.IsKnown(action.Frequency) && !action.Frequency.IsNull() && len(action.Frequency.Elements()) > 0 {
				if serverVersion.LessThan(frequencyMinSupportedVersion) {
					resp.Diagnostics.AddError(
						"actions.frequency is only supported for Kibana v8.6 or higher",
						"actions.frequency is only supported for Kibana v8.6 or higher",
					)
					return
				}
			}

			// Check alerts_filter version requirement
			if utils.IsKnown(action.AlertsFilter) && !action.AlertsFilter.IsNull() && len(action.AlertsFilter.Elements()) > 0 {
				if serverVersion.LessThan(alertsFilterMinSupportedVersion) {
					resp.Diagnostics.AddError(
						"actions.alerts_filter is only supported for Kibana v8.9 or higher",
						"actions.alerts_filter is only supported for Kibana v8.9 or higher",
					)
					return
				}
			}
		}
	}

	// Convert to API model
	rule, diags := plan.toAPIModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the rule
	createdRule, createDiags := kibana.CreateAlertingRule(ctx, r.client, rule)
	if createDiags.HasError() {
		resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(createDiags)...)
		return
	}

	// Store alert_delay from plan before populating (API may not echo it back)
	originalAlertDelay := plan.AlertDelay

	// Populate state directly from the API response to avoid race conditions
	// with eventual consistency (the API response has the authoritative values)
	resp.Diagnostics.Append(plan.populateFromAPI(ctx, createdRule)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve alert_delay from plan if API didn't return it
	// (some Kibana versions don't echo alert_delay in the response)
	if plan.AlertDelay.IsNull() && !originalAlertDelay.IsNull() {
		plan.AlertDelay = originalAlertDelay
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
