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

// StringNullIfEmpty returns a plan modifier that converts a known empty-string
// configuration value to null. This resolves inconsistencies where the write
// path omits empty strings from the API request but the read path returns null
// for omitted keys.
func StringNullIfEmpty() planmodifier.String {
	return stringNullIfEmpty{}
}

type stringNullIfEmpty struct{}

func (m stringNullIfEmpty) Description(context.Context) string {
	return "Treat an empty-string value as null"
}

func (m stringNullIfEmpty) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m stringNullIfEmpty) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Avoid overwriting unknown or non-empty planned values.
	if req.PlanValue.IsUnknown() || !req.PlanValue.IsNull() && req.PlanValue.ValueString() != "" {
		return
	}

	// Only normalize an explicitly supplied empty string.
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() || req.ConfigValue.ValueString() != "" {
		return
	}

	resp.PlanValue = types.StringNull()
}
