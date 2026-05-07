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

package datafeed

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
)

// readDatafeed fetches the datafeed from Elasticsearch and populates the model.
// It satisfies the entitycore elasticsearchReadFunc[Datafeed] signature.
func readDatafeed(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, state Datafeed) (Datafeed, bool, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics

	datafeedID := resourceID
	if datafeedID == "" {
		diags.AddError("Invalid Configuration", "datafeed_id cannot be empty")
		return state, false, diags
	}

	// Get the datafeed from Elasticsearch
	apiModel, getDiags := elasticsearch.GetDatafeed(ctx, client, datafeedID)
	diags.Append(getDiags...)
	if diags.HasError() {
		return state, false, diags
	}

	if apiModel == nil {
		// Datafeed not found
		return state, false, diags
	}

	// Convert API model to TF model
	convertDiags := state.FromAPIModel(ctx, apiModel)
	diags.Append(convertDiags...)
	if diags.HasError() {
		return state, false, diags
	}

	return state, true, diags
}
