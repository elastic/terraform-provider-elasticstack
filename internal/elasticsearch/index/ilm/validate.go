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

package ilm

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var c tfModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &c)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if c.Hot.IsUnknown() || c.Warm.IsUnknown() || c.Cold.IsUnknown() || c.Frozen.IsUnknown() || c.Delete.IsUnknown() {
		return
	}
	hasPhase := phaseObjectNonEmpty(c.Hot) || phaseObjectNonEmpty(c.Warm) || phaseObjectNonEmpty(c.Cold) ||
		phaseObjectNonEmpty(c.Frozen) || phaseObjectNonEmpty(c.Delete)
	if !hasPhase {
		resp.Diagnostics.AddError(
			"Missing phase configuration",
			"At least one of `hot`, `warm`, `cold`, `frozen`, or `delete` blocks must be configured.",
		)
	}
}

func phaseObjectNonEmpty(o types.Object) bool {
	return !o.IsNull() && !o.IsUnknown()
}
