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

package settings

import (
	"context"
	"fmt"

	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.Set = settingNameUniqueValidator{}

// settingNameUniqueValidator ensures that setting names are unique within a
// single persistent or transient block.
type settingNameUniqueValidator struct{}

func (v settingNameUniqueValidator) Description(_ context.Context) string {
	return "Ensures that setting names are unique within this block."
}

func (v settingNameUniqueValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v settingNameUniqueValidator) ValidateSet(_ context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	names := make(map[string]struct{}, len(req.ConfigValue.Elements()))
	for _, elem := range req.ConfigValue.Elements() {
		obj, ok := elem.(types.Object)
		if !ok {
			continue
		}

		nameVal, ok := obj.Attributes()["name"].(types.String)
		if !ok {
			continue
		}
		if nameVal.IsNull() || nameVal.IsUnknown() {
			continue
		}

		n := nameVal.ValueString()
		if _, exists := names[n]; exists {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				fmt.Sprintf(`Duplicate setting name "%s"`, n),
				fmt.Sprintf(`Setting name "%s" has already been configured within this block.`, n),
			)
			return
		}
		names[n] = struct{}{}
	}
}

// validateConfigModel implements the rule that at least one of persistent or
// transient must be a non-empty block. Extracted so it can be unit-tested
// without constructing a tfsdk.Config.
func validateConfigModel(config tfModel) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	if categoryBlockEmpty(config.Persistent) && categoryBlockEmpty(config.Transient) {
		diags.AddError(
			"No cluster settings configured",
			`At least one of "persistent" or "transient" must contain at least one "setting" block.`,
		)
	}
	return diags
}

// categoryBlockEmpty reports whether the given persistent/transient block is
// effectively empty: null, or contains a setting set with no elements.
// An unknown block (or unknown nested set) is NOT treated as empty,
// because the value has not yet been evaluated at validate time.
func categoryBlockEmpty(block types.Object) bool {
	if block.IsNull() {
		return true
	}
	if block.IsUnknown() {
		return false
	}
	settingAttr, ok := block.Attributes()["setting"]
	if !ok {
		return true
	}
	settingSet, ok := settingAttr.(types.Set)
	if !ok {
		return true
	}
	if settingSet.IsUnknown() {
		return false
	}
	return settingSet.IsNull() || len(settingSet.Elements()) == 0
}
