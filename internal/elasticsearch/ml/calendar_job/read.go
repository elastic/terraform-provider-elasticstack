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
	"slices"

	esv8 "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/ml/getcalendars"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func readCalendarJob(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, state TFModel) (TFModel, bool, fwdiags.Diagnostics) {
	calendarID, jobID, splitDiags := ml.SplitCalendarResourcePath(resourceID, "<job_id>")
	if splitDiags.HasError() {
		return state, false, splitDiags
	}

	return entitycore.SimpleElasticsearchRead[TFModel, getcalendars.Response](
		"ML calendar",
		func(ctx context.Context, typedClient *esv8.TypedClient, calendarID string) (*getcalendars.Response, error) {
			return typedClient.Ml.GetCalendars().CalendarId(calendarID).Do(ctx)
		},
		func(model *TFModel, ctx context.Context, res *getcalendars.Response) (bool, fwdiags.Diagnostics) {
			if len(res.Calendars) == 0 {
				return false, nil
			}

			cal := res.Calendars[0]
			if !slices.Contains(cal.JobIds, jobID) {
				return false, nil
			}

			var diags fwdiags.Diagnostics
			compID, idDiags := client.ID(ctx, calendarID+"/"+jobID)
			diags.Append(idDiags...)
			if diags.HasError() {
				return false, diags
			}

			model.CalendarID = types.StringValue(calendarID)
			model.JobID = types.StringValue(jobID)
			model.ID = types.StringValue(compID.String())
			return true, diags
		},
	)(ctx, client, calendarID, state)
}
