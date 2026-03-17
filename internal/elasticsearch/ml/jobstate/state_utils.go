// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

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
