package jobstate

import (
	"context"
	"fmt"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *mlJobStateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MLJobStateData
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
		diags.AddError("Operation timed out", fmt.Sprintf(updateTimeoutErrorMessage, updateTimeout))
	}

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *mlJobStateResource) update(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State, operationTimeout time.Duration) diag.Diagnostics {
	var data MLJobStateData
	diags := plan.Get(ctx, &data)
	if diags.HasError() {
		return diags
	}

	client, fwDiags := clients.MaybeNewAPIClientFromFrameworkResource(ctx, data.ElasticsearchConnection, r.client)
	diags.Append(fwDiags...)
	if diags.HasError() {
		return diags
	}

	jobID := data.JobID.ValueString()
	desiredState := data.State.ValueString()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()

	// First, get the current job stats to check if the job exists and its current state
	currentState, fwDiags := r.getJobState(ctx, jobID)
	diags.Append(fwDiags...)
	if diags.HasError() {
		return diags
	}

	if currentState == nil {
		diags.AddError(
			"ML Job not found",
			fmt.Sprintf("ML job %s does not exist", jobID),
		)
		return diags
	}

	// Perform state transition if needed
	fwDiags = r.performStateTransition(ctx, client, data, *currentState)
	diags.Append(fwDiags...)
	if diags.HasError() {
		return diags
	}

	// Generate composite ID
	compID, sdkDiags := client.ID(ctx, jobID)
	if len(sdkDiags) > 0 {
		for _, d := range sdkDiags {
			diags.AddError(d.Summary, d.Detail)
		}
		return diags
	}

	// Set the response state
	data.ID = types.StringValue(compID.String())
	data.JobID = types.StringValue(jobID)
	data.State = types.StringValue(desiredState)

	diags.Append(state.Set(ctx, data)...)
	return diags
}

// performStateTransition handles the ML job state transition process
func (r *mlJobStateResource) performStateTransition(ctx context.Context, client *clients.APIClient, data MLJobStateData, currentState string) diag.Diagnostics {
	jobID := data.JobID.ValueString()
	desiredState := data.State.ValueString()
	force := data.Force.ValueBool()

	// Parse timeout duration
	timeout, parseErrs := data.Timeout.Parse()
	if parseErrs.HasError() {
		return parseErrs
	}

	// Return early if no state change is needed
	if currentState == desiredState {
		tflog.Debug(ctx, fmt.Sprintf("ML job %s is already in desired state %s", jobID, desiredState))
		return nil
	}

	// Initiate the state change
	switch desiredState {
	case "opened":
		if diags := elasticsearch.OpenMLJob(ctx, client, jobID); diags.HasError() {
			return diags
		}
	case "closed":
		if diags := elasticsearch.CloseMLJob(ctx, client, jobID, force, timeout); diags.HasError() {
			return diags
		}
	default:
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Invalid state",
				fmt.Sprintf("Invalid state %s. Valid states are 'opened' and 'closed'", desiredState),
			),
		}
	}

	// Wait for state transition to complete
	diags := r.waitForJobState(ctx, jobID, desiredState)
	if diags.HasError() {
		return diags
	}

	tflog.Info(ctx, fmt.Sprintf("ML job %s successfully transitioned to state %s", jobID, desiredState))
	return nil
}
