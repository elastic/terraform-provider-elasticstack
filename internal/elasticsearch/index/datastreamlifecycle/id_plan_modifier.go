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

package datastreamlifecycle

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/planmodifiers"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// idSetUnknownIfNameChanged marks planned id unknown when name changes in
// place. UseStateForUnknown runs first and would otherwise copy the prior
// composite id as a known value; apply then writes <prior-uuid>/<new-name>
// and Terraform reports an inconsistent result. Role/watch/datafeed do not
// need this: their identity keys use RequiresReplace.
func idSetUnknownIfNameChanged() planmodifier.String {
	return planmodifiers.StringSetUnknownIf(
		"Sets id to unknown when name changes in place so apply can write <prior-uuid>/<new-name>",
		func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) bool {
			var stateName, configName types.String
			resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root(attrName), &stateName)...)
			resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(attrName), &configName)...)
			if resp.Diagnostics.HasError() {
				return false
			}
			if !typeutils.IsKnown(stateName) {
				return false
			}
			if !typeutils.IsKnown(configName) {
				return configName.IsUnknown()
			}
			return stateName.ValueString() != configName.ValueString()
		},
	)
}
