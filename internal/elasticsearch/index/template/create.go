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

package template

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var config Model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, diags := r.Client().GetElasticsearchClient(ctx, plan.ElasticsearchConnection)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverVersion, sdkDiags := client.ServerVersion(ctx)
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(validateIgnoreMissingComponentTemplatesVersion(plan, serverVersion)...)
	resp.Diagnostics.Append(validateDataStreamOptionsVersion(plan, serverVersion)...)
	if resp.Diagnostics.HasError() {
		return
	}

	indexTemplate, diags := plan.toAPIModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(elasticsearch.PutIndexTemplate(ctx, client, indexTemplate)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, sdkDiags := client.ID(ctx, plan.Name.ValueString())
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build a prior model carrying the computed ID and connection so readIndexTemplate can copy them.
	// Use configuration (not plan) as the reconciliation reference: plan can carry unknown/Computed
	// placeholders in nested set elements that then differ from non-refresh planning.
	priorForRead := config
	priorForRead.ID = types.StringValue(id.String())
	priorForRead.ElasticsearchConnection = plan.ElasticsearchConnection

	refreshed, found, diags := readIndexTemplate(ctx, client, plan.Name.ValueString(), priorForRead)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		tflog.Warn(ctx, "Index template missing after create readback", map[string]any{"template_name": plan.Name.ValueString()})
		resp.Diagnostics.AddError("Index template missing after create", plan.Name.ValueString())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &refreshed)...)
}
