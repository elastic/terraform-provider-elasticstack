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
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func writeTransform(ctx context.Context, client *clients.ElasticsearchScopedClient, req entitycore.WriteRequest[tfModel]) (entitycore.WriteResult[tfModel], diag.Diagnostics) {
	isCreate := req.Prior == nil
	resourceID := req.Plan.GetResourceID().ValueString()

	if isCreate {
		createdModel, createDiags := createTransform(ctx, client, resourceID, req.Plan)
		return entitycore.WriteResult[tfModel]{Model: createdModel}, createDiags
	}

	var diags diag.Diagnostics

	serverVersion, sdkDiags := client.ServerVersion(ctx)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return entitycore.WriteResult[tfModel]{Model: req.Plan}, diags
	}

	apiTransform, convDiags := toAPIModel(ctx, req.Plan, serverVersion)
	diags.Append(convDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[tfModel]{Model: req.Plan}, diags
	}
	apiTransform.Pivot = nil
	apiTransform.Latest = nil

	timeout, parseDiags := req.Plan.Timeout.Parse()
	diags.Append(parseDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[tfModel]{Model: req.Plan}, diags
	}

	deferValidation := req.Plan.DeferValidation.ValueBool()

	wasEnabled := req.Prior.Enabled.ValueBool()
	willBeEnabled := req.Plan.Enabled.ValueBool()
	enabledChanged := wasEnabled != willBeEnabled

	sdkDiags = elasticsearch.UpdateTransform(ctx, client, apiTransform, deferValidation, timeout, willBeEnabled, enabledChanged)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return entitycore.WriteResult[tfModel]{Model: req.Plan}, diags
	}

	return entitycore.WriteResult[tfModel]{Model: req.Plan}, diags
}
