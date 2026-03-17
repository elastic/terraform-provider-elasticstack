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

package apikey

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func (r *Resource) UpgradeState(context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: new(r.getSchema(0)),
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var model tfModel
				resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
				if resp.Diagnostics.HasError() {
					return
				}

				if typeutils.IsKnown(model.Expiration) && model.Expiration.ValueString() == "" {
					model.Expiration = basetypes.NewStringNull()
				}

				resp.State.Set(ctx, model)
			},
		},
		1: {
			PriorSchema: new(r.getSchema(1)),
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var model tfModel
				resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
				if resp.Diagnostics.HasError() {
					return
				}

				model.Type = basetypes.NewStringValue(defaultAPIKeyType)

				resp.State.Set(ctx, model)
			},
		},
	}
}
