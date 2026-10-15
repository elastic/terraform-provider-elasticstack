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

	aliasName := resourceID
	if aliasName == "" {
		aliasName = state.RolloverAlias.ValueString()
	}
	series, diags := discoverAlias(ctx, client, aliasName)
	if diags.HasError() {
		return diags
	}
	if !series.Found {
		return nil
	}
	if streamDiags := rejectDataStreamTargets(aliasName, series); streamDiags.HasError() {
		return streamDiags
	}
	for _, member := range series.Members {
		diags.Append(elasticsearch.DeleteIndex(ctx, client, member.Name)...)
		if diags.HasError() {
			return diags
		}
	}
	return nil
}
