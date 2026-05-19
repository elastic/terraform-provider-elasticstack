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
	var es types.Object
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("elasticsearch"), &es)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if es.IsNull() {
		resp.Diagnostics.AddError(
			"Invalid elasticsearch configuration",
			"The `elasticsearch` block is required.",
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
		_, _, baseLen, featureLen := kibanaPrivilegeCounts(obj)
		resp.Diagnostics.Append(validateKibanaPrivileges(baseLen, featureLen)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
}
