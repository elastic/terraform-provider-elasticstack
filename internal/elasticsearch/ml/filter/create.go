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

package filter

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func createFilter(ctx context.Context, client *clients.ElasticsearchScopedClient, req entitycore.WriteRequest[TFModel]) (entitycore.WriteResult[TFModel], fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics
	plan := req.Plan
	filterID := req.WriteID

	if filterID == "" {
		diags.AddError("Invalid resource ID", "filter_id cannot be empty")
		return entitycore.WriteResult[TFModel]{Model: plan}, diags
	}

	tflog.Debug(ctx, fmt.Sprintf("Creating ML filter: %s", filterID))

	typedClient := client.GetESClient()

	put := typedClient.Ml.PutFilter(filterID)
	if typeutils.IsKnown(plan.Description) && plan.Description.ValueString() != "" {
		put = put.Description(plan.Description.ValueString())
	}

	items, itemDiags := itemsFromPlan(ctx, plan)
	diags.Append(itemDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[TFModel]{Model: plan}, diags
	}
	if items != nil {
		put = put.Items(items...)
	}

	_, err := put.Do(ctx)
	if err != nil {
		diags.AddError("Failed to create ML filter", fmt.Sprintf("Unable to create ML filter: %s — %s", filterID, err.Error()))
		return entitycore.WriteResult[TFModel]{Model: plan}, diags
	}

	compID, idDiags := client.ID(ctx, filterID)
	diags.Append(idDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[TFModel]{Model: plan}, diags
	}

	plan.ID = types.StringValue(compID.String())

	tflog.Debug(ctx, fmt.Sprintf("Successfully created ML filter: %s", filterID))
	return entitycore.WriteResult[TFModel]{Model: plan}, diags
}
