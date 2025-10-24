package job_state

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/asyncutils"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

var errJobNotFound = fmt.Errorf("ML job not found")

// getJobState returns the current state of a job
func (r *mlJobStateResource) getJobState(ctx context.Context, jobId string) (*string, diag.Diagnostics) {
	// Get job stats to check current state
	currentJob, diags := elasticsearch.GetMLJobStats(ctx, r.client, jobId)
	if diags.HasError() {
		return nil, diags
	}

	if currentJob == nil {
		return nil, nil
	}

	return &currentJob.State, nil
}

// waitForJobState waits for a job to reach the desired state
func (r *mlJobStateResource) waitForJobState(ctx context.Context, jobId, desiredState string) diag.Diagnostics {
	stateChecker := func(ctx context.Context) (bool, error) {
		currentState, diags := r.getJobState(ctx, jobId)
		if diags.HasError() {
			return false, diagutil.FwDiagsAsError(diags)
		}

		if currentState == nil {
			return false, errJobNotFound
		}

		return *currentState == desiredState, nil
	}

	err := asyncutils.WaitForStateTransition(ctx, "ml_job", jobId, stateChecker)
	return diagutil.FrameworkDiagFromError(err)
}
