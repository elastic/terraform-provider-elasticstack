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

package security_entity_store_entity_link

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// ValidateConfig performs cross-attribute validation.  It rejects configurations
// where target_id appears in entity_ids (self-link guard).
func (r *EntityLinkResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data entityLinkModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.TargetID.IsNull() || data.TargetID.IsUnknown() || data.EntityIDs.IsNull() || data.EntityIDs.IsUnknown() {
		return
	}

	targetID := data.TargetID.ValueString()
	var entityIDs []string
	diags := data.EntityIDs.ElementsAs(ctx, &entityIDs, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if slices.Contains(entityIDs, targetID) {
		resp.Diagnostics.AddAttributeError(
			path.Root(attrEntityIDs),
			"Self-link not allowed",
			fmt.Sprintf("target_id %q must not appear in entity_ids", targetID),
		)
		return
	}
}
