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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func readMLJobState(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, state MLJobStateData) (MLJobStateData, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Get job stats to check current state
	currentJob, getDiags := elasticsearch.GetMLJobStats(ctx, client, resourceID)
	diags.Append(getDiags...)
	if diags.HasError() {
		return state, false, diags
	}

	if currentJob == nil {
		tflog.Warn(ctx, fmt.Sprintf(`ML job "%s" not found, removing from state`, resourceID))
		return state, false, diags
	}

	// Update the state with current job information
	state.JobID = types.StringValue(resourceID)
	state.State = types.StringValue(currentJob.State.String())

	// Set defaults for computed attributes if they're not already set (e.g., during import)
	if state.Force.IsNull() {
		state.Force = types.BoolValue(false)
	}
	if state.Timeout.IsNull() {
		state.Timeout = customtypes.NewDurationValue("30s")
	}

	return state, true, diags
}
