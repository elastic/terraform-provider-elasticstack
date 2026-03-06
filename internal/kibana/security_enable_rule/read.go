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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *EnableRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model enableRuleModel

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverVersion, sdkDiags := r.client.ServerVersion(ctx)
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if serverVersion.LessThan(minSupportedVersion) {
		resp.Diagnostics.AddError("Unsupported server version", "Security detection rules bulk actions are not supported until Elastic Stack v8.11.0. Upgrade the target server to use this resource")
		return
	}

	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "Failed to get Kibana client")
		return
	}

	spaceID := model.SpaceID.ValueString()
	key := model.Key.ValueString()
	value := model.Value.ValueString()

	allEnabled, checkDiags := kibanaoapi.CheckRulesEnabledByTag(ctx, client, spaceID, key, value)
	resp.Diagnostics.Append(checkDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Read rules enabled status", map[string]any{
		"space_id":          spaceID,
		"key":               key,
		"value":             value,
		"all_rules_enabled": allEnabled,
	})

	model.AllRulesEnabled = types.BoolValue(allEnabled)

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
