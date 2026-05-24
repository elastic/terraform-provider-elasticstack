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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func readSecurityEnableRule(ctx context.Context, client *clients.KibanaScopedClient, _, _ string, model enableRuleModel) (enableRuleModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	oapiClient, err := client.GetKibanaOapiClient()
	if err != nil {
		diags.AddError(err.Error(), "Failed to get Kibana client")
		return model, false, diags
	}

	spaceID := model.SpaceID.ValueString()
	key := model.Key.ValueString()
	value := model.Value.ValueString()

	allEnabled, checkDiags := kibanaoapi.CheckRulesEnabledByTag(ctx, oapiClient, spaceID, key, value)
	diags.Append(checkDiags...)
	if diags.HasError() {
		return model, false, diags
	}

	tflog.Debug(ctx, "Read rules enabled status", map[string]any{
		"space_id":          spaceID,
		"key":               key,
		"value":             value,
		"all_rules_enabled": allEnabled,
	})

	model.AllRulesEnabled = types.BoolValue(allEnabled)

	return model, true, diags
}
