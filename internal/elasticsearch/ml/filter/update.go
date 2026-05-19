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
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func updateFilter(ctx context.Context, client *clients.ElasticsearchScopedClient, req entitycore.WriteRequest[TFModel]) (entitycore.WriteResult[TFModel], fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics
	filterID := req.WriteID
	plan := req.Plan
	prior := req.Prior

	tflog.Debug(ctx, fmt.Sprintf("Updating ML filter: %s", filterID))

	typedClient, err := client.GetESClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return entitycore.WriteResult[TFModel]{Model: plan}, diags
	}

	getRes, err := typedClient.Ml.GetFilters().FilterId(filterID).Do(ctx)
	notFound := false
	switch {
	case err != nil:
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			notFound = true
		} else {
			diags.AddError("Failed to get current ML filter", err.Error())
			return entitycore.WriteResult[TFModel]{Model: plan}, diags
		}
	case len(getRes.Filters) == 0:
		notFound = true
	}
	if notFound {
		diags.AddError("Filter not found", fmt.Sprintf("Filter %s not found during update", filterID))
		return entitycore.WriteResult[TFModel]{Model: plan}, diags
	}

	current := getRes.Filters[0]

	currentItemSet := make(map[string]struct{})
	for _, item := range current.Items {
		currentItemSet[item] = struct{}{}
	}

	var planItems []string
	if !plan.Items.IsNull() && !plan.Items.IsUnknown() {
		d := plan.Items.ElementsAs(ctx, &planItems, false)
		diags.Append(d...)
		if diags.HasError() {
			return entitycore.WriteResult[TFModel]{Model: plan}, diags
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
	if !plan.Description.Equal(prior.Description) {
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
		diags.AddError(
			"Failed to update ML filter",
			fmt.Sprintf("Unable to update ML filter: %s — %s", filterID, err.Error()),
		)
		return entitycore.WriteResult[TFModel]{Model: plan}, diags
	}

	tflog.Debug(ctx, fmt.Sprintf("Successfully updated ML filter: %s", filterID))
	return entitycore.WriteResult[TFModel]{Model: plan}, diags
}
