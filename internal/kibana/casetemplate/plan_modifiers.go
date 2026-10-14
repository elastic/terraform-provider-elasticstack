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

	"github.com/hashicorp/terraform-plugin-framework/diag"
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
// semantically equal definition_yaml. The cheap attribute comparisons run
// before the YAML parse.
func onlyFormattingChanged(ctx context.Context, config tfsdk.Config, plan tfsdk.Plan, state tfsdk.State, diags *diag.Diagnostics) bool {
	if state.Raw.IsNull() || plan.Raw.IsNull() {
		return false
	}

	var configModel, stateModel templateModel
	diags.Append(config.Get(ctx, &configModel)...)
	diags.Append(state.Get(ctx, &stateModel)...)
	if diags.HasError() {
		return false
	}

	// Optional+Computed attributes count as unchanged when unset in config.
	unchanged := configModel.Owner.Equal(stateModel.Owner) &&
		configModel.Name.Equal(stateModel.Name) &&
		configModel.Description.Equal(stateModel.Description) &&
		configModel.Tags.Equal(stateModel.Tags) &&
		(configModel.SpaceID.IsNull() || configModel.SpaceID.Equal(stateModel.SpaceID)) &&
		(configModel.IsEnabled.IsNull() || configModel.IsEnabled.Equal(stateModel.IsEnabled)) &&
		// Envelope-injected attributes: a change to either still triggers Update,
		// whose PUT bumps template_version, so the keeper must not pin it.
		configModel.Timeouts.Equal(stateModel.Timeouts) &&
		configModel.KibanaConnection.Equal(stateModel.KibanaConnection)
	if !unchanged || configModel.DefinitionYAML.IsUnknown() || configModel.DefinitionYAML.IsNull() || stateModel.DefinitionYAML.IsNull() {
		return false
	}

	equal, d := stateModel.DefinitionYAML.StringSemanticEquals(ctx, configModel.DefinitionYAML)
	diags.Append(d...)
	return equal
}
