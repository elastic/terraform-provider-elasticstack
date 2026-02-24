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

package securityenablerule

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *EnableRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(r.upsert(ctx, req.Plan, &resp.State)...)
}

func (r *EnableRuleResource) upsert(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State) diag.Diagnostics {
	var model enableRuleModel

	diags := plan.Get(ctx, &model)
	if diags.HasError() {
		return diags
	}

	serverVersion, sdkDiags := r.client.ServerVersion(ctx)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return diags
	}

	if serverVersion.LessThan(minSupportedVersion) {
		diags.AddError("Unsupported server version", "Security detection rules bulk actions are not supported until Elastic Stack v8.11.0. Upgrade the target server to use this resource")
		return diags
	}

	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		diags.AddError(err.Error(), "Failed to get Kibana client")
		return diags
	}

	spaceID := model.SpaceID.ValueString()
	key := model.Key.ValueString()
	value := model.Value.ValueString()

	if model.DisableOnDestroy.IsNull() {
		model.DisableOnDestroy = types.BoolValue(true)
	}

	model.ID = types.StringValue(fmt.Sprintf("%s/%s:%s", spaceID, key, value))

	diags.Append(kibanaoapi.EnableRulesByTag(ctx, client, spaceID, key, value)...)
	if diags.HasError() {
		return diags
	}

	model.AllRulesEnabled = types.BoolValue(true)

	diags.Append(state.Set(ctx, model)...)
	return diags
}
