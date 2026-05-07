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

package index

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// readIndex is the envelope read callback. It fetches the index identified by
// resourceID and populates the returned model. Returning found==false signals
// the resource should be removed from state.
func readIndex(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, stateModel tfModel) (tfModel, bool, diag.Diagnostics) {
	apiModel, diags := elasticsearch.GetIndex(ctx, client, resourceID)
	if diags.HasError() {
		return tfModel{}, false, diags
	}

	if apiModel == nil {
		return tfModel{}, false, nil
	}

	diags = stateModel.populateFromAPI(ctx, resourceID, *apiModel)
	if diags.HasError() {
		return tfModel{}, false, diags
	}

	return stateModel, true, nil
}
