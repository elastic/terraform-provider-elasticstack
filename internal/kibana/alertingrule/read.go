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

package alertingrule

import (
	"context"

	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// readRuleFromAPI reads the alerting rule from the API and populates the model.
// Returns (exists, diagnostics).
func (r *Resource) readRuleFromAPI(ctx context.Context, model *alertingRuleModel) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	ruleID, spaceID := model.getRuleIDAndSpaceID()

	oapiClient, err := r.client.GetKibanaOapiClient()
	if err != nil {
		diags.AddError("Failed to get Kibana client", err.Error())
		return false, diags
	}

	rule, readDiags := kibanaoapi.GetAlertingRule(ctx, oapiClient, spaceID, ruleID)
	diags.Append(readDiags...)
	if diags.HasError() {
		return false, diags
	}

	if rule == nil {
		return false, diags
	}

	diags.Append(model.populateFromAPI(ctx, rule)...)
	return true, diags
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state alertingRuleModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "Expected configured API client")
		return
	}

	exists, diags := r.readRuleFromAPI(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !exists {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
