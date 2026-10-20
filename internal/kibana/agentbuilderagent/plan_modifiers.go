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

package agentbuilderagent

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// useStateForUnknownIncludingNull behaves like stringplanmodifier.UseStateForUnknown
// but also carries forward a null prior state value. The plugin framework marks
// every computed attribute with a null configuration as unknown whenever any
// other computed attribute (here access_control) differs from prior state before
// its own plan modifiers run, which would otherwise turn a stable null avatar
// into a perpetual "known after apply" diff.
type useStateForUnknownIncludingNull struct{}

func (useStateForUnknownIncludingNull) Description(context.Context) string {
	return "Once set (including to null), the value of this attribute in state will not change."
}

func (m useStateForUnknownIncludingNull) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (useStateForUnknownIncludingNull) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.State.Raw.IsNull() || !req.PlanValue.IsUnknown() || !req.ConfigValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

var _ planmodifier.String = useStateForUnknownIncludingNull{}

func avatarPlanModifiers() []planmodifier.String {
	return []planmodifier.String{useStateForUnknownIncludingNull{}}
}
