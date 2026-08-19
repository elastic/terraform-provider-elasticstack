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

// SimpleElasticsearchDelete returns an [ElasticsearchDeleteFunc] for the
// common "delete-then-treat-404-as-gone" shape shared by simple Elasticsearch
// resources: resolve the typed client, call apiDelete for resourceID, and
// treat a transport 404 as an already-successful idempotent delete. It owns
// the start/success/already-deleted tflog.Debug messages so callers only
// supply the per-resource API call.
//
// resourceLabel is a human-readable resource name (for example "ML filter")
// used in log messages and error diagnostics.
//
// Resources with extra pre/post steps (composite ID parsing) can still use
// this helper by resolving resourceID first and invoking the returned
// function with the resolved ID, for example:
//
//	func deleteCalendarJob(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, model TFModel) diag.Diagnostics {
//		calendarID, jobID, diags := ml.SplitCalendarResourcePath(resourceID, "<job_id>")
//		if diags.HasError() {
//			return diags
//		}
//		return entitycore.SimpleElasticsearchDelete[TFModel](
//			"ML calendar job assignment",
//			func(ctx context.Context, typedClient *elasticsearch.TypedClient, _ string) error {
//				_, err := typedClient.Ml.DeleteCalendarJob(calendarID, jobID).Do(ctx)
//				return err
//			},
//		)(ctx, client, resourceID, model)
//	}
func SimpleElasticsearchDelete[T ElasticsearchResourceModel](
	resourceLabel string,
	apiDelete func(ctx context.Context, typedClient *elasticsearch.TypedClient, resourceID string) error,
) ElasticsearchDeleteFunc[T] {
	return func(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, _ T) diag.Diagnostics {
		var diags diag.Diagnostics

		tflog.Debug(ctx, fmt.Sprintf("Deleting %s: %s", resourceLabel, resourceID))

		typedClient := client.GetESClient()

		if err := apiDelete(ctx, typedClient, resourceID); err != nil {
			if esclient.IsNotFoundElasticsearchError(err) {
				tflog.Debug(ctx, fmt.Sprintf("%s already deleted: %s", resourceLabel, resourceID))
				return diags
			}
			diags.AddError(
				fmt.Sprintf("Failed to delete %s", resourceLabel),
				fmt.Sprintf("Unable to delete %s: %s — %s", resourceLabel, resourceID, err.Error()),
			)
			return diags
		}

		tflog.Debug(ctx, fmt.Sprintf("Successfully deleted %s: %s", resourceLabel, resourceID))
		return diags
	}
}
