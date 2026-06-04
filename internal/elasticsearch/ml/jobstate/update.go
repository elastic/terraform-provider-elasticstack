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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func performStateTransition(ctx context.Context, client *clients.ElasticsearchScopedClient, data MLJobStateData, currentState string) diag.Diagnostics {
	jobID := data.JobID.ValueString()
	desiredState := data.State.ValueString()
	force := data.Force.ValueBool()

	timeout, parseErrs := data.Timeout.Parse()
	if parseErrs.HasError() {
		return parseErrs
	}

	if currentState == desiredState {
		tflog.Debug(ctx, fmt.Sprintf("ML job %s is already in desired state %s", jobID, desiredState))
		return nil
	}

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

	diags := waitForJobState(ctx, client, data, jobID, desiredState)
	if diags.HasError() {
		return diags
	}

	tflog.Info(ctx, fmt.Sprintf("ML job %s successfully transitioned to state %s", jobID, desiredState))
	return nil
}
