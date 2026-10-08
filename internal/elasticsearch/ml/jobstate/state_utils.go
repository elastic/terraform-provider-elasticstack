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

	"github.com/elastic/terraform-provider-elasticstack/internal/asyncutils"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// terminalJobStates are the ML job states that do not transition on their
// own. If a job settles into one of these while desiredState is something
// else, continued polling would just wait out the context deadline instead of
// surfacing the mismatch, so waitForJobState fails fast instead.
var terminalJobStates = map[string]struct{}{
	"opened": {},
	"closed": {},
	"failed": {},
}

func getJobState(ctx context.Context, client *clients.ElasticsearchScopedClient, _ MLJobStateData, jobID string) (*string, diag.Diagnostics) {
	var diags diag.Diagnostics

	currentJob, getDiags := elasticsearch.GetMLJobStats(ctx, client, jobID)
	diags.Append(getDiags...)
	if diags.HasError() {
		return nil, diags
	}

	if currentJob == nil {
		return nil, diags
	}

	stateStr := currentJob.State.String()
	return &stateStr, diags
}

func waitForJobState(ctx context.Context, client *clients.ElasticsearchScopedClient, data MLJobStateData, jobID, desiredState string) diag.Diagnostics {
	getState := func(ctx context.Context) (*string, error) {
		currentState, diags := getJobState(ctx, client, data, jobID)
		if diags.HasError() {
			return nil, diagutil.FwDiagsAsError(diags)
		}
		return currentState, nil
	}

	err := asyncutils.WaitForTerminalOrDesiredState(ctx, "ml_job", jobID, desiredState, terminalJobStates, getState)
	return diagutil.FrameworkDiagFromError(err)
}
