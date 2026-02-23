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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *mlJobStateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MLJobStateData
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	compID, diags := clients.CompositeIDFromStrFw(data.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get job stats to check current state
	jobID := compID.ResourceID
	currentState, diags := r.getJobState(ctx, jobID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if currentState == nil {
		tflog.Warn(ctx, fmt.Sprintf(`ML job "%s" not found, removing from state`, jobID))
		resp.State.RemoveResource(ctx)
		return
	}

	// Update the state with current job information
	data.JobID = types.StringValue(jobID)
	data.State = types.StringValue(*currentState)

	// Set defaults for computed attributes if they're not already set (e.g., during import)
	if data.Force.IsNull() {
		data.Force = types.BoolValue(false)
	}
	if data.Timeout.IsNull() {
		data.Timeout = customtypes.NewDurationValue("30s")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
