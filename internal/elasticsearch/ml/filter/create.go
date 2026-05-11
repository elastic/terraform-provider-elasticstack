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
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func createFilter(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, plan TFModel) (TFModel, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics

	filterID := resourceID
	if filterID == "" {
		diags.AddError("Invalid resource ID", "filter_id cannot be empty")
		return plan, diags
	}

	tflog.Debug(ctx, fmt.Sprintf("Creating ML filter: %s", filterID))

	typedClient, err := client.GetESClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return plan, diags
	}

	put := typedClient.Ml.PutFilter(filterID)
	if typeutils.IsKnown(plan.Description) && plan.Description.ValueString() != "" {
		put = put.Description(plan.Description.ValueString())
	}

	if !plan.Items.IsNull() && !plan.Items.IsUnknown() {
		var items []string
		d := plan.Items.ElementsAs(ctx, &items, false)
		diags.Append(d...)
		if diags.HasError() {
			return plan, diags
		}
		put = put.Items(items...)
	}

	_, err = put.Do(ctx)
	if err != nil {
		diags.AddError("Failed to create ML filter", fmt.Sprintf("Unable to create ML filter: %s — %s", filterID, err.Error()))
		return plan, diags
	}

	compID, idDiags := client.ID(ctx, filterID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(idDiags)...)
	if diags.HasError() {
		return plan, diags
	}

	plan.ID = types.StringValue(compID.String())

	tflog.Debug(ctx, fmt.Sprintf("Successfully created ML filter: %s", filterID))
	return plan, diags
}
