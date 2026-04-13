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
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *inferenceEndpointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data Data
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, clientDiags := clients.MaybeNewAPIClientFromFrameworkResource(ctx, data.ElasticsearchConnection, r.client)
	resp.Diagnostics.Append(clientDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	inferenceID := data.InferenceID.ValueString()
	id, sdkDiags := client.ID(ctx, inferenceID)
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	supported, sdkDiags := client.EnforceMinVersion(ctx, MinSupportedVersion)
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !supported {
		resp.Diagnostics.AddError("Unsupported Feature", fmt.Sprintf("inference endpoints require Elasticsearch v%s or above", MinSupportedVersion.String()))
		return
	}

	endpoint, modelDiags := data.toAPIModel(ctx)
	resp.Diagnostics.Append(modelDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	putDiags := elasticsearch.PutInferenceEndpoint(ctx, client, endpoint)
	resp.Diagnostics.Append(putDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(id.String())

	readData, readDiags := r.read(ctx, data)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if readData == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Inference endpoint %q was not found after create", inferenceID))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, readData)...)
}
