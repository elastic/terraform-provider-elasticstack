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

package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StringNullIfEmpty returns a planmodifier.String that converts an explicit
// empty string ("") in the configuration to null in the plan. This is needed
// for optional string attributes that are not sent to the API when empty
// (via setIfNotEmpty), so the API never echoes them back. Without this
// normalisation, the plan holds "" but the post-apply read returns null,
// causing Terraform to report an inconsistency.
func StringNullIfEmpty() planmodifier.String {
	return stringNullIfEmpty{}
}

type stringNullIfEmpty struct{}

func (stringNullIfEmpty) Description(context.Context) string {
	return "Converts an explicit empty string to null, avoiding post-apply inconsistency when the API omits the field."
}

func (m stringNullIfEmpty) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (stringNullIfEmpty) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Only act when the config explicitly provides an empty string.
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if req.ConfigValue.ValueString() == "" {
		resp.PlanValue = types.StringNull()
	}
}
