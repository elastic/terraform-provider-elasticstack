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

package inferenceendpoint

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func createInferenceEndpoint(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, data Data) (Data, diag.Diagnostics) {
	var diags diag.Diagnostics

	id, sdkDiags := client.ID(ctx, resourceID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return data, diags
	}

	supported, sdkDiags := client.EnforceMinVersion(ctx, MinSupportedVersion)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return data, diags
	}
	if !supported {
		diags.AddError("Unsupported Feature", fmt.Sprintf("inference endpoints require Elasticsearch v%s or above", MinSupportedVersion.String()))
		return data, diags
	}

	endpoint, modelDiags := data.toAPIModel(ctx)
	diags.Append(modelDiags...)
	if diags.HasError() {
		return data, diags
	}

	diags.Append(elasticsearch.PutInferenceEndpoint(ctx, client, resourceID, data.TaskType.ValueString(), endpoint)...)
	if diags.HasError() {
		return data, diags
	}

	data.ID = types.StringValue(id.String())

	return data, diags
}
