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
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func readFilter(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, state TFModel) (TFModel, bool, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics

	filterID := resourceID
	if filterID == "" {
		diags.AddError("Invalid resource ID", "filter_id cannot be empty")
		return state, false, diags
	}

	tflog.Debug(ctx, fmt.Sprintf("Reading ML filter: %s", filterID))

	typedClient, err := client.GetESClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return state, false, diags
	}

	res, err := typedClient.Ml.GetFilters().FilterId(filterID).Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			return state, false, nil
		}
		diags.AddError("Failed to get ML filter", fmt.Sprintf("Unable to get ML filter: %s — %s", filterID, err.Error()))
		return state, false, diags
	}

	if len(res.Filters) == 0 {
		return state, false, nil
	}

	out := state
	diags.Append((&out).fromMLFilter(ctx, &res.Filters[0])...)
	if diags.HasError() {
		return state, false, diags
	}

	tflog.Debug(ctx, fmt.Sprintf("Successfully read ML filter: %s", filterID))
	return out, true, diags
}
