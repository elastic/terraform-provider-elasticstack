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

package watch

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *watchResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data Data
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, diags := r.Client().GetElasticsearchClient(ctx, data.ElasticsearchConnection)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	watchID := data.WatchID.ValueString()
	id, sdkDiags := client.ID(ctx, watchID)
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	put, modelDiags := data.toPutModel(ctx)
	resp.Diagnostics.Append(modelDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(elasticsearch.PutWatch(ctx, client, put)...)
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
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Watch %q was not found after create", watchID))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, readData)...)
}

func (r *watchResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data Data
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, diags := r.Client().GetElasticsearchClient(ctx, data.ElasticsearchConnection)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	watchID := data.WatchID.ValueString()
	id, sdkDiags := client.ID(ctx, watchID)
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = types.StringValue(id.String())

	put, modelDiags := data.toPutModel(ctx)
	resp.Diagnostics.Append(modelDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(elasticsearch.PutWatch(ctx, client, put)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readData, readDiags := r.read(ctx, data)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if readData == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Watch %q was not found after update", data.WatchID.ValueString()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, readData)...)
}
