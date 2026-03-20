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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *inferenceEndpointResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data Data
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readData, diags := r.read(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if readData == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, readData)...)
}

func (r *inferenceEndpointResource) read(ctx context.Context, data Data) (*Data, diag.Diagnostics) {
	compID, diags := clients.CompositeIDFromStrFw(data.ID.ValueString())
	if diags.HasError() {
		return nil, diags
	}

	endpoint, endpointDiags := elasticsearch.GetInferenceEndpoint(ctx, r.client, compID.ResourceID)
	diags.Append(endpointDiags...)
	if diags.HasError() {
		return nil, diags
	}

	if endpoint == nil {
		tflog.Warn(ctx, fmt.Sprintf(`Inference endpoint "%s" not found`, compID.ResourceID))
		return nil, nil
	}

	diags.Append(data.fromAPIModel(ctx, endpoint)...)
	if diags.HasError() {
		return nil, diags
	}

	return &data, diags
}
