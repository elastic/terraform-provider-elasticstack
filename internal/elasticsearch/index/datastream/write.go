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

package datastream

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// createDataStream PUTs the data stream and sets the composite ID on the model.
// The envelope will call readDataStream after this returns to populate computed fields.
func createDataStream(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, data Data) (Data, diag.Diagnostics) {
	var diags diag.Diagnostics

	id, sdkDiags := client.ID(ctx, resourceID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		var zero Data
		return zero, diags
	}

	diags.Append(diagutil.FrameworkDiagsFromSDK(elasticsearch.PutDataStream(ctx, client, resourceID))...)
	if diags.HasError() {
		var zero Data
		return zero, diags
	}

	data.ID = types.StringValue(id.String())
	return data, diags
}
