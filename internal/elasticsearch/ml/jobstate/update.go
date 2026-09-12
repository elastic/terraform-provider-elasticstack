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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml/statetransition"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func performStateTransition(ctx context.Context, client *clients.ElasticsearchScopedClient, data MLJobStateData, currentState string) diag.Diagnostics {
	jobID := data.JobID.ValueString()
	desiredState := data.State.ValueString()
	force := data.Force.ValueBool()

	timeout, parseErrs := data.Timeout.Parse()
	if parseErrs.HasError() {
		return parseErrs
	}

	_, diags := statetransition.Perform(ctx, statetransition.Params[string]{
		ResourceType:           "ML job",
		ResourceID:             jobID,
		CurrentState:           currentState,
		DesiredState:           desiredState,
		StartState:             "opened",
		StopState:              "closed",
		ValidStatesDescription: "'opened' and 'closed'",
		Start: func(ctx context.Context) diag.Diagnostics {
			return elasticsearch.OpenMLJob(ctx, client, jobID)
		},
		Stop: func(ctx context.Context) diag.Diagnostics {
			return elasticsearch.CloseMLJob(ctx, client, jobID, force, timeout)
		},
		Wait: func(ctx context.Context) (bool, diag.Diagnostics) {
			diags := waitForJobState(ctx, client, data, jobID, desiredState)
			return !diags.HasError(), diags
		},
	})
	return diags
}
