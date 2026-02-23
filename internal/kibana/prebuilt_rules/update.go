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

package prebuiltrules

import (
	"context"

	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func (r *PrebuiltRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(r.upsert(ctx, req.Plan, &resp.State)...)
}

func (r *PrebuiltRuleResource) upsert(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State) diag.Diagnostics {
	var model prebuiltRuleModel

	diags := plan.Get(ctx, &model)
	if diags.HasError() {
		return diags
	}

	serverVersion, sdkDiags := r.client.ServerVersion(ctx)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return diags
	}

	minVersion := version.Must(version.NewVersion("8.0.0"))
	if serverVersion.LessThan(minVersion) {
		diags.AddError("Unsupported server version", "Prebuilt rules are not supported until Elastic Stack v8.0.0. Upgrade the target server to use this resource")
		return diags
	}

	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		diags.AddError(err.Error(), "Failed to get Kibana client")
		return diags
	}

	spaceID := model.SpaceID.ValueString()
	model.ID = model.SpaceID

	diags.Append(kibanaoapi.InstallPrebuiltRules(ctx, client, spaceID)...)
	if diags.HasError() {
		return diags
	}

	status, statusDiags := kibanaoapi.GetPrebuiltRulesStatus(ctx, client, spaceID)
	diags.Append(statusDiags...)
	if diags.HasError() {
		return diags
	}

	model.populateFromStatus(status)

	diags.Append(state.Set(ctx, model)...)
	return diags
}
