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
	"slices"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func splitCalendarJobResourcePath(resourcePath string) (calendarID, jobID string, diags fwdiags.Diagnostics) {
	parts := strings.SplitN(resourcePath, "|", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		diags.AddError("Invalid ID format", "Expected resource segment format: <calendar_id>|<job_id>")
		return "", "", diags
	}
	return parts[0], parts[1], diags
}

func readCalendarJob(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, state TFModel) (TFModel, bool, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics

	calendarID, jobID, splitDiags := splitCalendarJobResourcePath(resourceID)
	diags.Append(splitDiags...)
	if diags.HasError() {
		return state, false, diags
	}

	tflog.Debug(ctx, fmt.Sprintf("Reading ML calendar job assignment: calendar=%s job=%s", calendarID, jobID))

	typedClient, err := client.GetESClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return state, false, diags
	}

	res, err := typedClient.Ml.GetCalendars().CalendarId(calendarID).Do(ctx)
	if err != nil {
		// Missing calendar: typed client returns *types.ElasticsearchError with status 404
		// (see go-elasticsearch typedapi/ml/getcalendars GetCalendars.Do). Treat as gone so
		// refresh removes the assignment from state when the calendar is deleted out-of-band.
		if elasticsearch.IsNotFoundElasticsearchError(err) {
			return state, false, nil
		}
		diags.AddError("Failed to get ML calendar", fmt.Sprintf("Unable to get ML calendar %q: %s", calendarID, err.Error()))
		return state, false, diags
	}

	if len(res.Calendars) == 0 {
		return state, false, nil
	}

	cal := res.Calendars[0]
	if !slices.Contains(cal.JobIds, jobID) {
		return state, false, nil
	}

	compID, idDiags := client.ID(ctx, calendarID+"|"+jobID)
	diags.Append(idDiags...)
	if diags.HasError() {
		return state, false, diags
	}

	out := TFModel{
		ElasticsearchConnection: state.ElasticsearchConnection,
		CalendarID:              types.StringValue(calendarID),
		JobID:                   types.StringValue(jobID),
		ID:                      types.StringValue(compID.String()),
	}

	tflog.Debug(ctx, fmt.Sprintf("Successfully read ML calendar job assignment: calendar=%s job=%s", calendarID, jobID))
	return out, true, diags
}
