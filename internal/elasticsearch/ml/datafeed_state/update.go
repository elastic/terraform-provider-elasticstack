package datafeed_state

import (
	"context"
	"fmt"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml/datafeed"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *mlDatafeedStateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MLDatafeedStateData
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get update timeout
	updateTimeout, fwDiags := data.Timeouts.Update(ctx, 5*time.Minute) // Default 5 minutes
	resp.Diagnostics.Append(fwDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = r.update(ctx, req.Plan, &resp.State, updateTimeout)
	if diagutil.ContainsContextDeadlineExceeded(ctx, diags) {
		diags.AddError("Operation timed out", fmt.Sprintf("The operation to update the ML datafeed state timed out after %s. You may need to allocate more free memory within ML nodes by either closing other jobs, or increasing the overall ML memory. You may retry the operation.", updateTimeout))
	}

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *mlDatafeedStateResource) update(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State, operationTimeout time.Duration) diag.Diagnostics {
	var data MLDatafeedStateData
	diags := plan.Get(ctx, &data)
	if diags.HasError() {
		return diags
	}

	client, fwDiags := clients.MaybeNewApiClientFromFrameworkResource(ctx, data.ElasticsearchConnection, r.client)
	diags.Append(fwDiags...)
	if diags.HasError() {
		return diags
	}

	datafeedId := data.DatafeedId.ValueString()
	desiredState := data.State.ValueString()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()

	// First, get the current datafeed stats to check if the datafeed exists and its current state
	datafeedStats, fwDiags := elasticsearch.GetDatafeedStats(ctx, client, datafeedId)
	diags.Append(fwDiags...)
	if diags.HasError() {
		return diags
	}

	if datafeedStats == nil {
		diags.AddError(
			"ML Datafeed not found",
			fmt.Sprintf("ML datafeed %s does not exist", datafeedId),
		)
		return diags
	}

	// Perform state transition if needed
	inDesiredState, fwDiags := r.performStateTransition(ctx, client, data, datafeed.State(datafeedStats.State))
	diags.Append(fwDiags...)
	if diags.HasError() {
		return diags
	}

	// Generate composite ID
	compId, sdkDiags := client.ID(ctx, datafeedId)
	if len(sdkDiags) > 0 {
		for _, d := range sdkDiags {
			diags.AddError(d.Summary, d.Detail)
		}
		return diags
	}

	// Set the response state
	data.Id = types.StringValue(compId.String())

	var finalData *MLDatafeedStateData
	if inDesiredState {
		var getDiags diag.Diagnostics
		finalData, getDiags = r.read(ctx, data)
		diags.Append(getDiags...)
		if diags.HasError() {
			return diags
		}
	} else {
		var updateDiags diag.Diagnostics
		finalData, updateDiags = r.updateAfterMissedTransition(ctx, client, data, datafeedStats)
		diags.Append(updateDiags...)
		if diags.HasError() {
			return diags
		}
	}

	if finalData == nil {
		diags.AddError("Failed to read datafeed stats after update", fmt.Sprintf("The datafeed was successfully transitioned to the %s state, but could not be read after this change", desiredState))
		return diags
	}

	diags.Append(state.Set(ctx, finalData)...)
	return diags
}

func (r *mlDatafeedStateResource) updateAfterMissedTransition(ctx context.Context, client *clients.ApiClient, data MLDatafeedStateData, datafeedStats *models.DatafeedStats) (*MLDatafeedStateData, diag.Diagnostics) {
	datafeedId := data.DatafeedId.ValueString()
	statsAfterUpdate, diags := elasticsearch.GetDatafeedStats(ctx, client, datafeedId)
	if diags.HasError() {
		return nil, diags
	}

	if statsAfterUpdate == nil {
		diags.AddError(
			"ML Datafeed not found",
			fmt.Sprintf("ML datafeed %s does not exist after successful update", datafeedId),
		)
		return nil, diags
	}

	// It's possible that the datafeed starts, and then immediately stops if there is no (or very little) data to process.
	// In this case, the state transition may occur too quickly to be detected by the wait function.
	// To handle this, we check if the search count has increased to determine if the datafeed actually started since the update.
	if statsAfterUpdate.TimingStats == nil || datafeedStats.TimingStats == nil {
		diags.AddWarning("Expected Datafeed to contain timing stats",
			fmt.Sprintf("Stats for datafeed %s did not contain timing stats either before or after the update. Before %v - After %v", datafeedId, datafeedStats, statsAfterUpdate))
	} else if statsAfterUpdate.TimingStats.SearchCount <= datafeedStats.TimingStats.SearchCount {
		diags.AddError(
			"Datafeed did not successfully transition to the desired state",
			fmt.Sprintf("[%s] datafeed did not settle into the [%s] state. The current state is [%s]", datafeedId, data.State.ValueString(), statsAfterUpdate.State),
		)
		return nil, diags
	}

	if data.Start.IsUnknown() {
		data.Start = timetypes.NewRFC3339Null()
	}

	return &data, nil
}

// performStateTransition handles the ML datafeed state transition process
func (r *mlDatafeedStateResource) performStateTransition(ctx context.Context, client *clients.ApiClient, data MLDatafeedStateData, currentState datafeed.State) (bool, diag.Diagnostics) {
	datafeedId := data.DatafeedId.ValueString()
	desiredState := datafeed.State(data.State.ValueString())
	force := data.Force.ValueBool()

	// Parse timeout duration
	timeout, parseErrs := data.Timeout.Parse()
	if parseErrs.HasError() {
		return false, parseErrs
	}

	// Return early if no state change is needed
	if currentState == desiredState {
		tflog.Debug(ctx, fmt.Sprintf("ML datafeed %s is already in desired state %s", datafeedId, desiredState))
		return true, nil
	}

	// Initiate the state change
	switch desiredState {
	case datafeed.StateStarted:
		start, diags := data.GetStartAsString()
		if diags.HasError() {
			return false, diags
		}
		end, endDiags := data.GetEndAsString()
		diags.Append(endDiags...)
		if diags.HasError() {
			return false, diags
		}

		startDiags := elasticsearch.StartDatafeed(ctx, client, datafeedId, start, end, timeout)
		diags.Append(startDiags...)
		if diags.HasError() {
			return false, diags
		}
	case datafeed.StateStopped:
		if diags := elasticsearch.StopDatafeed(ctx, client, datafeedId, force, timeout); diags.HasError() {
			return false, diags
		}
	default:
		return false, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Invalid state",
				fmt.Sprintf("Invalid state %s. Valid states are 'started' and 'stopped'", desiredState),
			),
		}
	}

	// Wait for state transition to complete
	inDesiredState, diags := datafeed.WaitForDatafeedState(ctx, client, datafeedId, desiredState)
	if diags.HasError() {
		return false, diags
	}

	tflog.Info(ctx, fmt.Sprintf("ML datafeed %s successfully transitioned to state %s", datafeedId, desiredState))
	return inDesiredState, nil
}
