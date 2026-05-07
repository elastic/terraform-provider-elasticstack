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

package watch

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func updateWatch(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, plan Data) (Data, diag.Diagnostics) {
	var diags diag.Diagnostics

	put, modelDiags := plan.toPutModel(ctx)
	diags.Append(modelDiags...)
	if diags.HasError() {
		return plan, diags
	}

	// When transform is not configured, include an empty object so Elasticsearch
	// clears any existing transform (update semantics differ from create).
	if plan.Transform.IsNull() || plan.Transform.IsUnknown() {
		put.Body.Transform = map[string]any{}
	}

	diags.Append(elasticsearch.PutWatch(ctx, client, put)...)
	if diags.HasError() {
		return plan, diags
	}

	id, sdkDiags := client.ID(ctx, resourceID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return plan, diags
	}

	plan.ID = types.StringValue(id.String())
	return plan, diags
}
