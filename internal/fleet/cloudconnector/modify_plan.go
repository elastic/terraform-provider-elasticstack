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

package cloudconnector

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ModifyPlan detects drift on write-only secret attributes by comparing config
// values against bcrypt hashes stored in resource private state.
func (r *Resource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}
	if req.State.Raw.IsNull() {
		return
	}

	var config cloudConnectorModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var priv privateData
	if req.Private != nil {
		priv = req.Private
	}
	hasher := cloudConnectorHasher()
	results, driftDiags := evaluateWriteOnlyDrift(ctx, hasher, config, priv)
	resp.Diagnostics.Append(driftDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(results) == 0 {
		return
	}

	for _, result := range results {
		resp.Diagnostics.Append(driftWarningDiagnostic(result))
	}

	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root(attrUpdatedAt), types.StringUnknown())...)
}
