// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
//
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

package security_role

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type configValidator struct{}

func (v configValidator) Description(_ context.Context) string {
	return "Validates elasticsearch and kibana privilege blocks."
}

func (v configValidator) MarkdownDescription(_ context.Context) string {
	return "Validates elasticsearch and kibana privilege blocks."
}

func (v configValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var es types.Set
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("elasticsearch"), &es)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if es.IsNull() || es.IsUnknown() || len(es.Elements()) != 1 {
		resp.Diagnostics.AddError(
			"Invalid elasticsearch configuration",
			"Exactly one elasticsearch block is required.",
		)
		return
	}

	var kibana types.Set
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("kibana"), &kibana)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if kibana.IsNull() || kibana.IsUnknown() {
		return
	}
	for _, elem := range kibana.Elements() {
		obj, ok := elem.(types.Object)
		if !ok {
			resp.Diagnostics.AddError("Invalid kibana block", "unexpected element type")
			return
		}
		attrs := obj.Attributes()
		base := attrs["base"].(types.Set)
		feature := attrs["feature"].(types.Set)
		baseLen := 0
		if !base.IsNull() && !base.IsUnknown() {
			baseLen = len(base.Elements())
		}
		featureLen := 0
		if !feature.IsNull() && !feature.IsUnknown() {
			featureLen = len(feature.Elements())
		}
		if baseLen > 0 && featureLen > 0 {
			resp.Diagnostics.AddError(
				"Invalid kibana privileges",
				"Only one of the `feature` or `base` privileges allowed!",
			)
			return
		}
		if baseLen == 0 && featureLen == 0 {
			resp.Diagnostics.AddError(
				"Invalid kibana privileges",
				"Either on of the `feature` or `base` privileges must be set for kibana role!",
			)
			return
		}
	}
}
