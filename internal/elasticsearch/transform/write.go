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
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// writeTransform is the [entitycore.WriteFunc] wired as the resource Update
// callback. It calls Update Transform (which does not accept Pivot/Latest) and
// reconciles the enabled-state delta against the prior state.
func writeTransform(ctx context.Context, client *clients.ElasticsearchScopedClient, req entitycore.WriteRequest[tfModel]) (entitycore.WriteResult[tfModel], diag.Diagnostics) {
	plan := req.Plan

	apiTransform, timeout, diags := buildTransformAPIRequest(ctx, client, plan)
	if diags.HasError() {
		return entitycore.WriteResult[tfModel]{Model: plan}, diags
	}

	// Pivot and Latest are immutable for an existing transform; the Update
	// Transform API rejects them.
	apiTransform.Pivot = nil
	apiTransform.Latest = nil

	willBeEnabled := plan.Enabled.ValueBool()
	enabledChanged := req.Prior.Enabled.ValueBool() != willBeEnabled

	diags.Append(elasticsearch.UpdateTransform(ctx, client, apiTransform, plan.DeferValidation.ValueBool(), timeout, willBeEnabled, enabledChanged)...)
	return entitycore.WriteResult[tfModel]{Model: plan}, diags
}

// buildTransformAPIRequest performs the shared plan-to-API conversion and
// timeout parsing used by both createTransform and writeTransform.
func buildTransformAPIRequest(ctx context.Context, client *clients.ElasticsearchScopedClient, plan tfModel) (*models.Transform, time.Duration, diag.Diagnostics) {
	var diags diag.Diagnostics

	apiTransform, convDiags := toAPIModel(ctx, client, plan)
	diags.Append(convDiags...)
	if diags.HasError() {
		return nil, 0, diags
	}

	timeout, parseDiags := plan.Timeout.Parse()
	diags.Append(parseDiags...)
	if diags.HasError() {
		return nil, 0, diags
	}

	return apiTransform, timeout, diags
}
