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

package agentdownloadsource

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// writeOnlyConfigPaths maps each write-only attribute path to its framework path.
var writeOnlyConfigPaths = map[string]path.Path{
	sslKeyWOPath:       path.Root("ssl").AtName("key_wo"),
	authPasswordWOPath: path.Root("auth").AtName("password_wo"),
	authAPIKeyWOPath:   path.Root("auth").AtName("api_key_wo"),
}

// ModifyPlan detects in-config edits and removals of the write-only secrets by
// comparing the configured values against the hashes stored in private state
// after the last successful apply. A detected change marks id unknown so
// Terraform plans an update. Stored hashes are never modified here: a removal
// keeps its hash until the clearing update succeeds (see persistSecretHashes).
func (r *Resource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}

	resp.Diagnostics.Append(modifyPlan(ctx, req.Private, req.Config, &resp.Plan)...)
}

func modifyPlan(ctx context.Context, private entitycore.PrivateStateStorage, config tfsdk.Config, plan *tfsdk.Plan) diag.Diagnostics {
	var diags diag.Diagnostics

	configured := make(map[string]types.String, len(writeOnlyConfigPaths))
	for attrPath, configPath := range writeOnlyConfigPaths {
		var value types.String
		diags.Append(config.GetAttribute(ctx, configPath, &value)...)
		configured[attrPath] = value
	}
	if diags.HasError() {
		return diags
	}

	stored, loadDiags := loadStoredHashes(ctx, private)
	diags.Append(loadDiags...)
	if diags.HasError() {
		return diags
	}

	changed := planDriftPaths(configured, stored)
	if len(changed) == 0 {
		return diags
	}
	for _, attrPath := range changed {
		diags.AddWarning("Write-only attribute changed",
			fmt.Sprintf("Detected a change to write-only attribute %s; the resource will be updated.", attrPath))
	}
	diags.Append(plan.SetAttribute(ctx, path.Root("id"), types.StringUnknown())...)
	return diags
}

// planDriftPaths returns the write-only attribute paths whose configuration no
// longer matches the last applied value, in writeOnlyPaths order: a configured
// value whose hash differs from the stored one, or a stored hash whose
// attribute is no longer configured (a removal that must clear the remote
// secret). Without a stored hash there is nothing to compare against, as on
// first apply or after import. Unknown values defer to apply.
func planDriftPaths(configured map[string]types.String, stored map[string][]byte) []string {
	var changed []string
	for _, attrPath := range writeOnlyPaths {
		storedHash := stored[attrPath]
		if len(storedHash) == 0 {
			continue
		}
		value := configured[attrPath]
		switch {
		case value.IsUnknown():
		case value.IsNull():
			changed = append(changed, attrPath)
		case !secretHasher.Matches(value.ValueString(), storedHash):
			changed = append(changed, attrPath)
		}
	}
	return changed
}
