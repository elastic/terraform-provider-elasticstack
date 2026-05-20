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

package calendar_job

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func createCalendarJob(ctx context.Context, client *clients.ElasticsearchScopedClient, req entitycore.WriteRequest[TFModel]) (entitycore.WriteResult[TFModel], fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics
	plan := req.Plan

	calendarID := plan.CalendarID.ValueString()
	jobID := plan.JobID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Creating ML calendar job assignment: calendar=%s job=%s", calendarID, jobID))

	typedClient, err := client.GetESClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return entitycore.WriteResult[TFModel]{Model: plan}, diags
	}

	if _, err := typedClient.Ml.PutCalendarJob(calendarID, jobID).Do(ctx); err != nil {
		diags.AddError(
			"Failed to assign ML job to calendar",
			fmt.Sprintf("Unable to assign job %q to calendar %q: %s", jobID, calendarID, err.Error()),
		)
		return entitycore.WriteResult[TFModel]{Model: plan}, diags
	}

	compID, idDiags := client.ID(ctx, calendarID+"/"+jobID)
	diags.Append(idDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[TFModel]{Model: plan}, diags
	}

	plan.ID = types.StringValue(compID.String())

	tflog.Debug(ctx, fmt.Sprintf("Successfully assigned ML job to calendar: calendar=%s job=%s", calendarID, jobID))
	return entitycore.WriteResult[TFModel]{Model: plan}, diags
}
