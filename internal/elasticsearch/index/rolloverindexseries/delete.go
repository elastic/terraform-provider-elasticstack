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

package rolloverindexseries

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

const deletionProtectionError = "cannot destroy rollover index series without setting deletion_protection=false and running `terraform apply`"

func deleteRolloverIndexSeries(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	resourceID string,
	state Model,
) diag.Diagnostics {
	if !typeutils.IsKnown(state.DeletionProtection) || state.DeletionProtection.ValueBool() {
		var diags diag.Diagnostics
		diags.AddAttributeError(
			path.Root("deletion_protection"),
			deletionProtectionError,
			deletionProtectionError+". Destroy deletes every index currently attached to the rollover alias, including indices outside prefix.",
		)
		return diags
	}

	for {
		if ctx.Err() != nil {
			return deleteTimeoutDiagnostics(resourceID)
		}
		series, diags := discoverAliasForDelete(ctx, client, resourceID)
		if diags.HasError() {
			return timeoutOr(ctx, resourceID, diags)
		}
		if !series.Found {
			return nil
		}
		if streamDiags := rejectDataStreamTargets(resourceID, series); streamDiags.HasError() {
			return streamDiags
		}
		for _, member := range series.Members {
			if ctx.Err() != nil {
				return deleteTimeoutDiagnostics(resourceID)
			}
			if delDiags := elasticsearch.DeleteIndex(ctx, client, member.Name); delDiags.HasError() {
				return timeoutOr(ctx, resourceID, delDiags)
			}
		}
		if !waitForRediscovery(ctx) {
			return deleteTimeoutDiagnostics(resourceID)
		}
	}
}

// timeoutOr reports the delete timeout when ctx has ended, otherwise diags.
func timeoutOr(ctx context.Context, resourceID string, diags diag.Diagnostics) diag.Diagnostics {
	if ctx.Err() != nil {
		return deleteTimeoutDiagnostics(resourceID)
	}
	return diags
}

func deleteTimeoutDiagnostics(resourceID string) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.AddError(
		"Timed out deleting rollover index series",
		fmt.Sprintf(
			"Deleting indices attached to alias %q did not finish before the delete timeout. "+
				"Observed members may still exist. Terraform kept this resource in state so a later destroy can rediscover current membership.",
			resourceID,
		),
	)
	return diags
}
