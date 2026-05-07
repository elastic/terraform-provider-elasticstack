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
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// createDatafeed creates the datafeed and sets the composite ID on the returned
// model. It satisfies the entitycore ElasticsearchCreateFunc[Datafeed] signature.
// The envelope handles read-after-write and state persistence.
func createDatafeed(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, plan Datafeed) (Datafeed, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics

	datafeedID := resourceID
	if datafeedID == "" {
		diags.AddError("Invalid Configuration", "datafeed_id cannot be empty")
		return plan, diags
	}

	// Convert to API create model (raw JSON to preserve query form)
	createBody, convDiags := plan.toAPICreateModel(ctx)
	diags.Append(convDiags...)
	if diags.HasError() {
		return plan, diags
	}

	createDiags := elasticsearch.PutDatafeed(ctx, client, datafeedID, createBody)
	diags.Append(createDiags...)
	if diags.HasError() {
		return plan, diags
	}

	// Set the composite ID so the envelope and readFunc can carry it through.
	compID, sdkDiags := client.ID(ctx, datafeedID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return plan, diags
	}

	plan.ID = types.StringValue(compID.String())
	return plan, diags
}
