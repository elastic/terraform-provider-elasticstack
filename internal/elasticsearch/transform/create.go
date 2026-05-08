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
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// createTransform calls Put Transform and optionally starts the transform when
// enabled is true. It builds the composite ID and returns the updated model
// ready for the envelope to persist via readTransform.
func createTransform(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, model tfModel) (tfModel, fwdiag.Diagnostics) {
	var diags fwdiag.Diagnostics

	// Resolve server version for version-gated fields.
	serverVersion, sdkDiags := client.ServerVersion(ctx)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return model, diags
	}

	// Convert PF model to API request.
	apiTransform, convDiags := toAPIModel(ctx, model, serverVersion)
	diags.Append(convDiags...)
	if diags.HasError() {
		return model, diags
	}

	timeout, parseDiags := model.Timeout.Parse()
	diags.Append(parseDiags...)
	if diags.HasError() {
		return model, diags
	}

	deferValidation := model.DeferValidation.ValueBool()
	enabled := model.Enabled.ValueBool()

	// Resolve composite ID BEFORE creating the remote transform so a failure
	// here cannot leave an orphaned remote resource.
	id, sdkDiags := client.ID(ctx, resourceID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return model, diags
	}

	sdkDiags = elasticsearch.PutTransform(ctx, client, apiTransform, deferValidation, timeout, enabled)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return model, diags
	}

	model.ID = types.StringValue(id.String())
	return model, diags
}
