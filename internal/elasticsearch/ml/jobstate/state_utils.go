package jobstate

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
func (r *mlJobStateResource) getJobState(ctx context.Context, jobID string) (*string, diag.Diagnostics) {
	// Get job stats to check current state
	currentJob, diags := elasticsearch.GetMLJobStats(ctx, r.client, jobID)
	if diags.HasError() {
		return nil, diags
	}

	if currentJob == nil {
		return nil, nil
	}

	return &currentJob.State, nil
}

// waitForJobState waits for a job to reach the desired state
func (r *mlJobStateResource) waitForJobState(ctx context.Context, jobID, desiredState string) diag.Diagnostics {
	stateChecker := func(ctx context.Context) (bool, error) {
		currentState, diags := r.getJobState(ctx, jobID)
		if diags.HasError() {
			return false, diagutil.FwDiagsAsError(diags)
		}

		if currentState == nil {
			return false, errJobNotFound
		}

		return *currentState == desiredState, nil
	}

	err := asyncutils.WaitForStateTransition(ctx, "ml_job", jobID, stateChecker)
	return diagutil.FrameworkDiagFromError(err)
}
