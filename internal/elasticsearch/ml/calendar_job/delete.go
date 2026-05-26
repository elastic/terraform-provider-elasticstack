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
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func deleteCalendarJob(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, _ TFModel) fwdiags.Diagnostics {
	var diags fwdiags.Diagnostics

	calendarID, jobID, splitDiags := ml.SplitCalendarResourcePath(resourceID, "<job_id>")
	diags.Append(splitDiags...)
	if diags.HasError() {
		return diags
	}

	tflog.Debug(ctx, fmt.Sprintf("Deleting ML calendar job assignment: calendar=%s job=%s", calendarID, jobID))

	typedClient := client.GetESClient()

	_, err := typedClient.Ml.DeleteCalendarJob(calendarID, jobID).Do(ctx)
	if err != nil {
		if elasticsearch.IsNotFoundElasticsearchError(err) {
			tflog.Debug(ctx, fmt.Sprintf("ML calendar job assignment already removed: calendar=%s job=%s", calendarID, jobID))
			return diags
		}
		diags.AddError(
			"Failed to remove ML job from calendar",
			fmt.Sprintf("Unable to remove job %q from calendar %q: %s", jobID, calendarID, err.Error()),
		)
		return diags
	}

	tflog.Debug(ctx, fmt.Sprintf("Successfully removed ML job from calendar: calendar=%s job=%s", calendarID, jobID))
	return diags
}
