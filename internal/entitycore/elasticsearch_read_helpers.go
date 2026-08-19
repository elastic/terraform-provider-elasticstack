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

package entitycore

import (
	"context"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// SimpleElasticsearchRead returns an [ElasticsearchReadFunc] for the common
// "get-then-not-found-check" shape shared by simple Elasticsearch resources:
// resolve the typed client, call apiGet for resourceID, treat a transport 404
// (or a nil result) as "not found", and otherwise apply populate to the prior
// model. It owns the start/success tflog.Debug messages and the
// [esclient.IsNotFoundElasticsearchError] branch so callers only supply the
// per-resource API call and payload-to-model mapping.
//
// resourceLabel is a human-readable resource name (for example "ML calendar")
// used in log messages and error diagnostics.
//
// Resources with extra pre/post steps (composite ID parsing, membership
// checks) can still use this helper by resolving resourceID first and
// invoking the returned function with the resolved ID, folding any extra
// state into the populate closure:
//
//	func readCalendarJob(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, state TFModel) (TFModel, bool, diag.Diagnostics) {
//		calendarID, jobID, diags := ml.SplitCalendarResourcePath(resourceID, "<job_id>")
//		if diags.HasError() {
//			return state, false, diags
//		}
//		return entitycore.SimpleElasticsearchRead[TFModel, getcalendars.Response](
//			"ML calendar",
//			func(ctx context.Context, typedClient *elasticsearch.TypedClient, calendarID string) (*getcalendars.Response, error) {
//				return typedClient.Ml.GetCalendars().CalendarId(calendarID).Do(ctx)
//			},
//			func(model *TFModel, ctx context.Context, res *getcalendars.Response) (bool, diag.Diagnostics) {
//				// ... derive found/state from res and jobID ...
//			},
//		)(ctx, client, calendarID, state)
//	}
func SimpleElasticsearchRead[T ElasticsearchResourceModel, R any](
	resourceLabel string,
	apiGet func(ctx context.Context, typedClient *elasticsearch.TypedClient, resourceID string) (*R, error),
	populate func(model *T, ctx context.Context, data *R) (bool, diag.Diagnostics),
) ElasticsearchReadFunc[T] {
	return func(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, state T) (T, bool, diag.Diagnostics) {
		var diags diag.Diagnostics

		tflog.Debug(ctx, fmt.Sprintf("Reading %s: %s", resourceLabel, resourceID))

		typedClient := client.GetESClient()

		data, err := apiGet(ctx, typedClient, resourceID)
		if err != nil {
			if esclient.IsNotFoundElasticsearchError(err) {
				return state, false, nil
			}
			diags.AddError(
				fmt.Sprintf("Failed to get %s", resourceLabel),
				fmt.Sprintf("Unable to get %s: %s — %s", resourceLabel, resourceID, err.Error()),
			)
			return state, false, diags
		}

		if data == nil {
			return state, false, nil
		}

		found, populateDiags := populate(&state, ctx, data)
		diags.Append(populateDiags...)
		if diags.HasError() || !found {
			return state, false, diags
		}

		tflog.Debug(ctx, fmt.Sprintf("Successfully read %s: %s", resourceLabel, resourceID))
		return state, true, diags
	}
}
