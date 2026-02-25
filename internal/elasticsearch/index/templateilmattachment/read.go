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

package templateilmattachment

import (
	"context"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const customSuffix = "@custom"

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tfModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	compID, diags := state.GetID()
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	componentTemplateName := compID.ResourceID
	// Derive index_template from component template name for import (component name is <index_template>@custom)
	if !typeutils.IsKnown(state.IndexTemplate) {
		state.IndexTemplate = types.StringValue(strings.TrimSuffix(componentTemplateName, customSuffix))
	}

	found, diags := readILMAttachment(ctx, r.client, &state)
	if !found {
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		tflog.Warn(ctx, "Component template or ILM setting not found, removing from state", map[string]any{
			"name": componentTemplateName,
		})
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
