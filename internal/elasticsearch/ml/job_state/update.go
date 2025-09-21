package job_state

import (
	"context"
	"fmt"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
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

	client, fwDiags := clients.MaybeNewApiClientFromFrameworkResource(ctx, data.ElasticsearchConnection, r.client)
	diags.Append(fwDiags...)
	if diags.HasError() {
		return diags
	}

	jobId := data.JobId.ValueString()
	desiredState := data.State.ValueString()

	// First, get the current job stats to check if the job exists and its current state
	currentJob, fwDiags := elasticsearch.GetMLJobStats(ctx, client, jobId)
	diags.Append(fwDiags...)
	if diags.HasError() {
		return diags
	}

	if currentJob == nil {
		diags.AddError(
			"ML Job not found",
			fmt.Sprintf("ML job %s does not exist", jobId),
		)
		return diags
	}

	currentState := currentJob.State

	// Perform state transition if needed
	fwDiags = r.performStateTransition(ctx, client, data, currentState, operationTimeout)
	diags.Append(fwDiags...)
	if diags.HasError() {
		return diags
	}

	// Generate composite ID
	compId, sdkDiags := client.ID(ctx, jobId)
	if len(sdkDiags) > 0 {
		for _, d := range sdkDiags {
			diags.AddError(d.Summary, d.Detail)
		}
		return diags
	}

	// Set the response state
	data.Id = types.StringValue(compId.String())
	data.JobId = types.StringValue(jobId)
	data.State = types.StringValue(desiredState)

	diags.Append(state.Set(ctx, data)...)
	return diags
}

// waitForJobStateTransition waits for an ML job to reach the desired state
func waitForJobStateTransition(ctx context.Context, client *clients.ApiClient, jobId, desiredState string) error {
	const pollInterval = 2 * time.Second
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			currentJob, fwDiags := elasticsearch.GetMLJobStats(ctx, client, jobId)
			if fwDiags.HasError() {
				return fmt.Errorf("failed to get job stats during state transition check")
			}

			if currentJob == nil {
				return fmt.Errorf("job not found during state transition check")
			}

			if currentJob.State == desiredState {
				return nil // Successfully reached desired state
			}
			tflog.Debug(ctx, fmt.Sprintf("ML job %s current state: %s, waiting for: %s", jobId, currentJob.State, desiredState))
		}
	}
}

// performStateTransition handles the ML job state transition process
func (r *mlJobStateResource) performStateTransition(ctx context.Context, client *clients.ApiClient, data MLJobStateData, currentState string, operationTimeout time.Duration) diag.Diagnostics {
	jobId := data.JobId.ValueString()
	desiredState := data.State.ValueString()
	force := data.Force.ValueBool()

	// Parse timeout duration
	timeout, parseErrs := data.Timeout.Parse()
	if parseErrs.HasError() {
		return parseErrs
	}

	// Return early if no state change is needed
	if currentState == desiredState {
		tflog.Debug(ctx, fmt.Sprintf("ML job %s is already in desired state %s", jobId, desiredState))
		return nil
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()

	// Initiate the state change
	switch desiredState {
	case "opened":
		diags := elasticsearch.OpenMLJob(ctx, client, jobId)
		if diags.HasError() {
			return diags
		}
	case "closed":
		diags := elasticsearch.CloseMLJob(ctx, client, jobId, force, timeout) // Always allow no match
		if diags.HasError() {
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
	err := waitForJobStateTransition(ctx, client, jobId, desiredState)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"State transition timeout",
				fmt.Sprintf("ML job %s did not transition to state %s within timeout: %s", jobId, desiredState, err.Error()),
			),
		}
	}

	tflog.Info(ctx, fmt.Sprintf("ML job %s successfully transitioned to state %s", jobId, desiredState))
	return nil
}
