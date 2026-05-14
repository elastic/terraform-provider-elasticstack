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
	"errors"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Update overrides the envelope's Update because building the update body
// requires comparing the plan with the prior Terraform state and diffing items.
func (r *filterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.Client() == nil {
		resp.Diagnostics.AddError("Client not configured", "Provider client is not configured")
		return
	}

	var plan TFModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state TFModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	compID, compDiags := clients.CompositeIDFromStrFw(state.ID.ValueString())
	resp.Diagnostics.Append(compDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	filterID := compID.ResourceID
	if filterID == "" {
		resp.Diagnostics.AddError("Invalid resource ID", "Could not determine filter id from composite id")
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Updating ML filter: %s", filterID))

	client, connDiags := r.Client().GetElasticsearchClient(ctx, plan.GetElasticsearchConnection())
	resp.Diagnostics.Append(connDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	typedClient, err := client.GetESClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Elasticsearch client", err.Error())
		return
	}

	getRes, err := typedClient.Ml.GetFilters().FilterId(filterID).Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			resp.Diagnostics.AddError("Filter not found", fmt.Sprintf("Filter %s not found during update", filterID))
			return
		}
		resp.Diagnostics.AddError("Failed to get current ML filter", err.Error())
		return
	}

	if len(getRes.Filters) == 0 {
		resp.Diagnostics.AddError("Filter not found", fmt.Sprintf("Filter %s not found during update", filterID))
		return
	}

	current := getRes.Filters[0]

	currentItemSet := make(map[string]struct{})
	for _, item := range current.Items {
		currentItemSet[item] = struct{}{}
	}

	var planItems []string
	if !plan.Items.IsNull() && !plan.Items.IsUnknown() {
		d := plan.Items.ElementsAs(ctx, &planItems, false)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	planItemSet := make(map[string]struct{})
	for _, item := range planItems {
		planItemSet[item] = struct{}{}
	}

	var addItems []string
	for _, item := range planItems {
		if _, exists := currentItemSet[item]; !exists {
			addItems = append(addItems, item)
		}
	}

	var removeItems []string
	for _, item := range current.Items {
		if _, exists := planItemSet[item]; !exists {
			removeItems = append(removeItems, item)
		}
	}

	updateReq := typedClient.Ml.UpdateFilter(filterID)

	desc := plan.Description.ValueString()
	if !plan.Description.Equal(state.Description) {
		updateReq = updateReq.Description(desc)
	}

	if len(addItems) > 0 {
		updateReq = updateReq.AddItems(addItems...)
	}
	if len(removeItems) > 0 {
		updateReq = updateReq.RemoveItems(removeItems...)
	}

	_, err = updateReq.Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update ML filter", fmt.Sprintf("Unable to update ML filter: %s — %s", filterID, err.Error()))
		return
	}

	resultModel, found, readDiags := readFilter(ctx, client, filterID, plan)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &resultModel)...)

	tflog.Debug(ctx, fmt.Sprintf("Successfully updated ML filter: %s", filterID))
}
