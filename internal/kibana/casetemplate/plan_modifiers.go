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

package casetemplate

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
)

// suppressSemanticYAMLDiff keeps the prior state value when the planned YAML
// parses to the same structure. Every update creates a new template version
// server-side, so a reformat-only edit must not plan an update at all.
func suppressSemanticYAMLDiff() planmodifier.String {
	return semanticYAMLDiffSuppressor{}
}

type semanticYAMLDiffSuppressor struct{}

func (semanticYAMLDiffSuppressor) Description(context.Context) string {
	return "Keeps the prior state value when the planned YAML is semantically equal to it."
}

func (s semanticYAMLDiffSuppressor) MarkdownDescription(ctx context.Context) string {
	return s.Description(ctx)
}

func (semanticYAMLDiffSuppressor) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() || req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}

	equal, diags := customtypes.NewNormalizedYamlValue(req.StateValue.ValueString()).
		StringSemanticEquals(ctx, customtypes.NewNormalizedYamlValue(req.PlanValue.ValueString()))
	resp.Diagnostics.Append(diags...)
	if equal {
		resp.PlanValue = types.StringValue(req.StateValue.ValueString())
	}
}

// keepComputedWhenOnlyFormattingChanged keeps a computed attribute's prior
// state value when the only difference between config and state is a
// semantically equal definition_yaml. The framework marks every computed
// attribute unknown as soon as config differs from state, which would
// otherwise turn a suppressed reformat-only edit back into an update plan.
func keepComputedWhenOnlyFormattingChanged() computedKeeper {
	return computedKeeper{}
}

type computedKeeper struct{}

func (computedKeeper) Description(context.Context) string {
	return "Keeps the prior state value when only the formatting of definition_yaml changed."
}

func (c computedKeeper) MarkdownDescription(ctx context.Context) string {
	return c.Description(ctx)
}

func (computedKeeper) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if req.PlanValue.IsUnknown() && onlyFormattingChanged(ctx, req.Config, req.Plan, req.State, &resp.Diagnostics) {
		resp.PlanValue = req.StateValue
	}
}

func (computedKeeper) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	if req.PlanValue.IsUnknown() && onlyFormattingChanged(ctx, req.Config, req.Plan, req.State, &resp.Diagnostics) {
		resp.PlanValue = req.StateValue
	}
}

// onlyFormattingChanged reports whether config differs from state solely in a
// semantically equal definition_yaml.
func onlyFormattingChanged(ctx context.Context, config tfsdk.Config, plan tfsdk.Plan, state tfsdk.State, diags *diag.Diagnostics) bool {
	if state.Raw.IsNull() || plan.Raw.IsNull() {
		return false
	}

	var configYAML, stateYAML customtypes.NormalizedYamlValue
	diags.Append(config.GetAttribute(ctx, path.Root("definition_yaml"), &configYAML)...)
	diags.Append(state.GetAttribute(ctx, path.Root("definition_yaml"), &stateYAML)...)
	if diags.HasError() || configYAML.IsUnknown() || configYAML.IsNull() || stateYAML.IsNull() {
		return false
	}
	equal, d := customtypes.NewNormalizedYamlValue(stateYAML.ValueString()).
		StringSemanticEquals(ctx, customtypes.NewNormalizedYamlValue(configYAML.ValueString()))
	diags.Append(d...)
	if !equal {
		return false
	}

	return unchangedAttr[types.String](ctx, config, state, "owner", false, diags) &&
		unchangedAttr[types.String](ctx, config, state, "name", false, diags) &&
		unchangedAttr[types.String](ctx, config, state, "description", false, diags) &&
		unchangedAttr[types.List](ctx, config, state, "tags", false, diags) &&
		unchangedAttr[types.String](ctx, config, state, "space_id", true, diags) &&
		unchangedAttr[types.Bool](ctx, config, state, "is_enabled", true, diags)
}

// unchangedAttr reports whether the configured value of the named attribute
// equals its state value. When ignoreNull is set, an unset config value counts
// as unchanged because the attribute is Optional+Computed.
func unchangedAttr[T attr.Value](ctx context.Context, config tfsdk.Config, state tfsdk.State, name string, ignoreNull bool, diags *diag.Diagnostics) bool {
	var configValue, stateValue T
	diags.Append(config.GetAttribute(ctx, path.Root(name), &configValue)...)
	diags.Append(state.GetAttribute(ctx, path.Root(name), &stateValue)...)
	if diags.HasError() {
		return false
	}
	if ignoreNull && configValue.IsNull() {
		return true
	}
	return configValue.Equal(stateValue)
}
