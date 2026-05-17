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

package transform

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// createTransform is the [entitycore.WriteFunc] wired as the resource Create
// callback. It resolves the composite ID before issuing Put Transform so a
// failure cannot leave an orphaned remote resource, then optionally starts the
// transform when `enabled` is true.
func createTransform(ctx context.Context, client *clients.ElasticsearchScopedClient, req entitycore.WriteRequest[tfModel]) (entitycore.WriteResult[tfModel], diag.Diagnostics) {
	plan := req.Plan

	apiTransform, timeout, diags := buildTransformAPIRequest(ctx, client, plan)
	if diags.HasError() {
		return entitycore.WriteResult[tfModel]{Model: plan}, diags
	}

	id, idDiags := client.ID(ctx, plan.GetResourceID().ValueString())
	diags.Append(idDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[tfModel]{Model: plan}, diags
	}

	diags.Append(elasticsearch.PutTransform(ctx, client, apiTransform, plan.DeferValidation.ValueBool(), timeout, plan.Enabled.ValueBool())...)
	if diags.HasError() {
		return entitycore.WriteResult[tfModel]{Model: plan}, diags
	}

	plan.ID = types.StringValue(id.String())
	return entitycore.WriteResult[tfModel]{Model: plan}, diags
}
