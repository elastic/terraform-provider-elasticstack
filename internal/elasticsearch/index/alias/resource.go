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

package alias

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.Resource                   = newAliasResource()
	_ resource.ResourceWithConfigure      = newAliasResource()
	_ resource.ResourceWithImportState    = newAliasResource()
	_ resource.ResourceWithModifyPlan     = newAliasResource()
	_ resource.ResourceWithValidateConfig = newAliasResource()
)

type aliasResource struct {
	*entitycore.ElasticsearchResource[tfModel]
}

func newAliasResource() *aliasResource {
	return &aliasResource{
		ElasticsearchResource: entitycore.NewElasticsearchResource[tfModel]("index_alias", entitycore.ElasticsearchResourceOptions[tfModel]{
			Schema: getSchemaFactory,
			Read:   readAlias,
			Delete: deleteAlias,
			Create: createAlias,
			Update: updateAlias,
		}),
	}
}

func NewAliasResource() resource.Resource {
	return newAliasResource()
}

func (r *aliasResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *aliasResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config tfModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(config.Validate(ctx)...)
}

func (r *aliasResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	var state, plan tfModel
	if !req.State.Raw.IsNull() {
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	}
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ReadIndices.IsNull() || plan.ReadIndices.IsUnknown() {
		return
	}

	// A newly configured alias can refer to indices created in the same apply,
	// so resolving its read index expressions during planning can return 404.
	// Mark the membership unknown and defer resolution until Create instead.
	if req.State.Raw.IsNull() {
		resp.Diagnostics.Append(plan.modifyPlanReadIndexMembership(ctx, state, nil)...)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
		return
	}

	deferReadIndexResolution := func() {
		resp.Diagnostics.Append(plan.modifyPlanReadIndexMembership(ctx, state, nil)...)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
	}

	if plan.WriteIndex.IsUnknown() {
		deferReadIndexResolution()
		return
	}
	if !plan.WriteIndex.IsNull() {
		var writeIndex indexModel
		resp.Diagnostics.Append(plan.WriteIndex.As(ctx, &writeIndex, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}
		if writeIndex.Name.IsUnknown() {
			deferReadIndexResolution()
			return
		}
	}

	var readIndices []readIndexModel
	resp.Diagnostics.Append(plan.ReadIndices.ElementsAs(ctx, &readIndices, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, readIndex := range readIndices {
		if readIndex.Name.IsUnknown() || readIndex.Filter.IsUnknown() {
			deferReadIndexResolution()
			return
		}
	}

	client, clientDiags := r.Client().GetElasticsearchClient(ctx, plan.ElasticsearchConnection)
	resp.Diagnostics.Append(clientDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	type resolution struct {
		targets elasticsearch.ResolvedIndexTargets
		diags   diag.Diagnostics
	}
	resolutions := make(map[string]resolution)
	resolveIndexExpression := func(ctx context.Context, expression string) (elasticsearch.ResolvedIndexTargets, diag.Diagnostics) {
		if resolution, found := resolutions[expression]; found {
			return resolution.targets, resolution.diags
		}

		targets, resolveDiags := elasticsearch.ResolveIndexExpression(ctx, client, expression, plan.Name.ValueString())
		resolutions[expression] = resolution{targets: targets, diags: resolveDiags}
		return targets, resolveDiags
	}

	_, resolveDiags := plan.resolveAliasConfigs(ctx, resolveIndexExpression)
	resp.Diagnostics.Append(resolveDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(plan.modifyPlanReadIndexMembership(ctx, state, resolveIndexExpression)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
}
