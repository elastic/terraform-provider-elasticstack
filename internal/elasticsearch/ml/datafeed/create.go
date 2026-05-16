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
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// createDatafeed creates the datafeed and sets the composite ID on the
// returned model. The envelope handles read-after-write and state persistence.
func createDatafeed(ctx context.Context, client *clients.ElasticsearchScopedClient, req entitycore.WriteRequest[Datafeed]) (entitycore.WriteResult[Datafeed], fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics
	plan := req.Plan
	datafeedID := req.WriteID

	if datafeedID == "" {
		diags.AddError("Invalid Configuration", "datafeed_id cannot be empty")
		return entitycore.WriteResult[Datafeed]{Model: plan}, diags
	}

	createBody, convDiags := plan.toAPICreateModel(ctx)
	diags.Append(convDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[Datafeed]{Model: plan}, diags
	}

	createDiags := elasticsearch.PutDatafeed(ctx, client, datafeedID, createBody)
	diags.Append(createDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[Datafeed]{Model: plan}, diags
	}

	compID, sdkDiags := client.ID(ctx, datafeedID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return entitycore.WriteResult[Datafeed]{Model: plan}, diags
	}

	plan.ID = types.StringValue(compID.String())
	return entitycore.WriteResult[Datafeed]{Model: plan}, diags
}
