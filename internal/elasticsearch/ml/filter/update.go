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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *filterResource) update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.resourceReady(&resp.Diagnostics) {
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

	filterID := state.FilterID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Updating ML filter: %s", filterID))

	esClient, err := r.client.GetESClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Elasticsearch client", err.Error())
		return
	}

	compID, sdkDiags := clients.CompositeIDFromStr(state.ID.ValueString())
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read current filter to get current items for diffing.
	res, err := esClient.ML.GetFilters(esClient.ML.GetFilters.WithFilterID(compID.ResourceID), esClient.ML.GetFilters.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to get current ML filter", err.Error())
		return
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		resp.Diagnostics.AddError("Filter not found", fmt.Sprintf("Filter %s not found during update", filterID))
		return
	}

	getDiags := diagutil.CheckErrorFromFW(res, fmt.Sprintf("Unable to get ML filter for update: %s", filterID))
	resp.Diagnostics.Append(getDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var currentResponse struct {
		Filters []APIModel `json:"filters"`
		Count   int        `json:"count"`
	}
	if err := json.NewDecoder(res.Body).Decode(&currentResponse); err != nil {
		resp.Diagnostics.AddError("Failed to decode filter response", err.Error())
		return
	}

	if len(currentResponse.Filters) == 0 {
		resp.Diagnostics.AddError("Filter not found", fmt.Sprintf("Filter %s not found during update", filterID))
		return
	}

	currentFilter := currentResponse.Filters[0]

	currentItemSet := make(map[string]struct{})
	for _, item := range currentFilter.Items {
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
	for _, item := range currentFilter.Items {
		if _, exists := planItemSet[item]; !exists {
			removeItems = append(removeItems, item)
		}
	}

	updateBody := UpdateAPIModel{
		AddItems:    addItems,
		RemoveItems: removeItems,
	}

	desc := plan.Description.ValueString()
	if !plan.Description.Equal(state.Description) {
		updateBody.Description = &desc
	}

	body, err := json.Marshal(updateBody)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal filter update", err.Error())
		return
	}

	updateRes, err := esClient.ML.UpdateFilter(bytes.NewReader(body), filterID, esClient.ML.UpdateFilter.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to update ML filter", err.Error())
		return
	}
	defer updateRes.Body.Close()

	diags = diagutil.CheckErrorFromFW(updateRes, fmt.Sprintf("Unable to update ML filter: %s", filterID))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, diags := r.read(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	tflog.Debug(ctx, fmt.Sprintf("Successfully updated ML filter: %s", filterID))
}
