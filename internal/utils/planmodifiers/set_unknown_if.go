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

// StringShouldSetUnknownFunc is invoked by StringSetUnknownIf to determine
// whether the planned value should be replaced with Unknown. Implementations
// may append diagnostics to resp; any error diagnostics short-circuit the
// modifier without replacing the plan value.
type StringShouldSetUnknownFunc func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) bool

// StringSetUnknownIf returns a planmodifier.String that sets the plan value to
// Unknown when shouldSetUnknown returns true. The modifier is a no-op when
// either state or config is absent (create / destroy), so callers don't need
// to guard for those cases inside shouldSetUnknown.
func StringSetUnknownIf(description string, shouldSetUnknown StringShouldSetUnknownFunc) planmodifier.String {
	return stringSetUnknownIf{description: description, shouldSetUnknown: shouldSetUnknown}
}

type stringSetUnknownIf struct {
	description      string
	shouldSetUnknown StringShouldSetUnknownFunc
}

func (s stringSetUnknownIf) Description(context.Context) string { return s.description }

func (s stringSetUnknownIf) MarkdownDescription(ctx context.Context) string {
	return s.Description(ctx)
}

func (s stringSetUnknownIf) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.State.Raw.IsNull() || req.Config.Raw.IsNull() {
		return
	}

	if !s.shouldSetUnknown(ctx, req, resp) || resp.Diagnostics.HasError() {
		return
	}

	resp.PlanValue = types.StringUnknown()
}
