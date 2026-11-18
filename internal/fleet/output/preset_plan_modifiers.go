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

package output

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// presetPlanModifier plans the computed preset value when it is not configured.
//
// Fleet only returns a preset for elasticsearch and remote_elasticsearch
// outputs, so other types always plan null. For preset-capable outputs the
// stored preset is kept when the output type and config_yaml are unchanged.
// The planned value is then sent in the update request, because whether Fleet
// re-derives an omitted preset varies across versions. Otherwise the preset is
// left unknown so Fleet can derive it from config_yaml.
func presetPlanModifier() planmodifier.String {
	return presetModifier{}
}

type presetModifier struct{}

func (presetModifier) Description(_ context.Context) string {
	return "Plans preset as null for output types without presets, and keeps the stored preset when it is not configured and neither type nor config_yaml changes."
}

func (m presetModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (presetModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.ConfigValue.IsNull() || !req.PlanValue.IsUnknown() {
		return
	}

	var planType types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root(attrType), &planType)...)
	if resp.Diagnostics.HasError() || planType.IsUnknown() {
		return
	}

	stateType := types.StringNull()
	configYamlUnchanged := false
	if !req.State.Raw.IsNull() {
		resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root(attrType), &stateType)...)

		var planConfigYaml, stateConfigYaml customtypes.NormalizedYamlValue
		resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("config_yaml"), &planConfigYaml)...)
		resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("config_yaml"), &stateConfigYaml)...)
		if resp.Diagnostics.HasError() {
			return
		}

		if !planConfigYaml.IsUnknown() {
			equal, d := stateConfigYaml.StringSemanticEquals(ctx, planConfigYaml)
			resp.Diagnostics.Append(d...)
			configYamlUnchanged = equal
		}
	}

	resp.PlanValue = presetPlanValue(planType.ValueString(), stateType, req.StateValue, configYamlUnchanged)
}

// presetPlanValue returns the planned preset for an unconfigured preset.
// stateType is null when the resource is being created.
func presetPlanValue(planType string, stateType, stateValue types.String, configYamlUnchanged bool) types.String {
	if planType != outputTypeElasticsearch && planType != outputTypeRemoteElasticsearch {
		return types.StringNull()
	}

	if stateType.ValueString() != planType || !configYamlUnchanged || stateValue.IsNull() || stateValue.IsUnknown() {
		return types.StringUnknown()
	}

	return stateValue
}
