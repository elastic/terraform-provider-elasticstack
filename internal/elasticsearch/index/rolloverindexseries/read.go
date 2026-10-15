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
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func readRolloverIndexSeries(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	resourceID string,
	state Model,
) (Model, bool, diag.Diagnostics) {
	aliasName := resourceID
	if aliasName == "" {
		aliasName = state.RolloverAlias.ValueString()
	}
	series, diags := discoverAlias(ctx, client, aliasName)
	if diags.HasError() {
		return state, false, diags
	}
	if !series.Found {
		return state, false, nil
	}
	if streamDiags := rejectDataStreamTargets(aliasName, series); streamDiags.HasError() {
		return state, false, streamDiags
	}

	id, idDiags := client.ID(ctx, aliasName)
	diags.Append(idDiags...)
	if diags.HasError() {
		return state, false, diags
	}
	return state.withObservedSeries(id.String(), series), true, nil
}
