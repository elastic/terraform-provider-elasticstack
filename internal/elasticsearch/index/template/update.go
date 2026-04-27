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
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var prior Model
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	if resp.Diagnostics.HasError() {
		return
	}

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

	applyAllowCustomRouting8xWorkaround(ctx, prior, config, indexTemplate)

	resp.Diagnostics.Append(elasticsearch.PutIndexTemplate(ctx, client, indexTemplate)...)
	if resp.Diagnostics.HasError() {
		return
	}

	refreshed, found, diags := readIndexTemplate(ctx, client, plan.Name.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		tflog.Warn(ctx, "Index template missing after update readback", map[string]any{"template_name": plan.Name.ValueString()})
		resp.Diagnostics.AddError("Index template missing after update", plan.Name.ValueString())
		return
	}

	resp.Diagnostics.Append(applyTemplateAliasReconciliationFromReference(ctx, &refreshed, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(canonicalizeTemplateAliasSetInModel(ctx, &refreshed)...)
	if resp.Diagnostics.HasError() {
		return
	}

	refreshed.ElasticsearchConnection = plan.ElasticsearchConnection
	refreshed.ID = prior.ID

	resp.Diagnostics.Append(resp.State.Set(ctx, &refreshed)...)
}

// applyAllowCustomRouting8xWorkaround mirrors resourceIndexTemplatePut in the SDK: when prior state had
// allow_custom_routing=true and configuration does not explicitly set that attribute to true, re-send
// allow_custom_routing=false in the PUT body. toAPIModel only emits the field when true, so without this
// pass Elasticsearch 8.x can keep the old true value when practitioners remove the attribute from HCL.
func applyAllowCustomRouting8xWorkaround(ctx context.Context, prior, config Model, indexTemplate *models.IndexTemplate) {
	if !dataStreamAllowCustomRoutingWasTrue(ctx, prior.DataStream) {
		return
	}
	if configSetsAllowCustomRoutingTrue(config.DataStream) {
		return
	}
	f := false
	if indexTemplate.DataStream == nil {
		indexTemplate.DataStream = &models.DataStreamSettings{}
	}
	indexTemplate.DataStream.AllowCustomRouting = &f
}

func configSetsAllowCustomRoutingTrue(configDataStream types.Object) bool {
	if configDataStream.IsNull() || configDataStream.IsUnknown() {
		return false
	}
	v, ok := configDataStream.Attributes()["allow_custom_routing"]
	if !ok || v.IsNull() || v.IsUnknown() {
		return false
	}
	b, ok := v.(types.Bool)
	if !ok {
		return false
	}
	return b.ValueBool()
}

func dataStreamAllowCustomRoutingWasTrue(ctx context.Context, ds types.Object) bool {
	if ds.IsNull() || ds.IsUnknown() {
		return false
	}
	var m DataStreamModel
	diags := ds.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return false
	}
	return !m.AllowCustomRouting.IsNull() && !m.AllowCustomRouting.IsUnknown() && m.AllowCustomRouting.ValueBool()
}
