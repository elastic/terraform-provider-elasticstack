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

	esv8 "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/ml/getfilters"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
)

func readFilter(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, state TFModel) (TFModel, bool, fwdiags.Diagnostics) {
	filterID := resourceID
	if filterID == "" {
		var diags fwdiags.Diagnostics
		diags.AddError("Invalid resource ID", "filter_id cannot be empty")
		return state, false, diags
	}

	return entitycore.SimpleElasticsearchRead[TFModel, getfilters.Response](
		"ML filter",
		func(ctx context.Context, typedClient *esv8.TypedClient, filterID string) (*getfilters.Response, error) {
			return typedClient.Ml.GetFilters().FilterId(filterID).Do(ctx)
		},
		func(model *TFModel, ctx context.Context, res *getfilters.Response) (bool, fwdiags.Diagnostics) {
			if len(res.Filters) == 0 {
				return false, nil
			}
			out := *model
			diags := (&out).fromMLFilter(ctx, &res.Filters[0])
			if diags.HasError() {
				return false, diags
			}
			*model = out
			return true, diags
		},
	)(ctx, client, filterID, state)
}
