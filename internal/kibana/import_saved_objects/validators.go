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

package importsavedobjects

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// bothTrueConflict returns a ConfigValidator that errors only when both named
// boolean attributes are explicitly set to true. This is narrower than
// resourcevalidator.Conflicting, which fires whenever both attributes are
// non-null regardless of their values.
func bothTrueConflict(attr1, attr2 string) resource.ConfigValidator {
	return &bothTrueValidator{attr1: attr1, attr2: attr2}
}

type bothTrueValidator struct {
	attr1 string
	attr2 string
}

func (v *bothTrueValidator) Description(_ context.Context) string {
	return fmt.Sprintf("%s and %s cannot both be true", v.attr1, v.attr2)
}

func (v *bothTrueValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v *bothTrueValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var val1, val2 types.Bool
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(v.attr1), &val1)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(v.attr2), &val2)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !val1.IsNull() && !val1.IsUnknown() && val1.ValueBool() &&
		!val2.IsNull() && !val2.IsUnknown() && val2.ValueBool() {
		resp.Diagnostics.AddAttributeError(
			path.Root(v.attr1),
			"Invalid attribute combination",
			fmt.Sprintf("%s and %s cannot both be set to true", v.attr1, v.attr2),
		)
	}
}
