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

package index

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type mappingsPlanModifier struct{}

func (p mappingsPlanModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !typeutils.IsKnown(req.StateValue) {
		return
	}

	if !typeutils.IsKnown(req.ConfigValue) {
		return
	}

	stateStr := req.StateValue.ValueString()
	cfgStr := req.ConfigValue.ValueString()

	var stateMappings map[string]any
	var cfgMappings map[string]any

	// No error checking, schema validation ensures this is valid json
	_ = json.Unmarshal([]byte(stateStr), &stateMappings)
	_ = json.Unmarshal([]byte(cfgStr), &cfgMappings)

	if stateProps, ok := stateMappings["properties"]; ok {
		cfgProps, ok := cfgMappings["properties"]
		if !ok {
			resp.RequiresReplace = true
			return
		}

		requiresReplace, finalMappings, diags := p.modifyMappings(path.Root("mappings").AtMapKey("properties"), stateProps.(map[string]any), cfgProps.(map[string]any))
		resp.RequiresReplace = requiresReplace
		cfgMappings["properties"] = finalMappings
		resp.Diagnostics.Append(diags...)

		planBytes, err := json.Marshal(cfgMappings)
		if err != nil {
			resp.Diagnostics.AddAttributeError(req.Path, "Failed to marshal final mappings", err.Error())
			return
		}

		resp.PlanValue = basetypes.NewStringValue(string(planBytes))
	}
}

func (p mappingsPlanModifier) modifyMappings(initialPath path.Path, oldMappings map[string]any, newMappings map[string]any) (bool, map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	warningDetail := "Elasticsearch will maintain the current field in it's mapping. " +
		"Re-index to remove the field completely"
	for k, v := range oldMappings {
		oldFieldSettings := v.(map[string]any)
		newFieldSettings, ok := newMappings[k]
		currentPath := initialPath.AtMapKey(k)
		// When field is removed, it'll be ignored in elasticsearch
		if !ok {
			diags.AddAttributeWarning(
				path.Root("mappings"),
				fmt.Sprintf("removing field [%s] in mappings is ignored.", currentPath),
				warningDetail,
			)
			newMappings[k] = v
			continue
		}
		newSettings := newFieldSettings.(map[string]any)
		// check if the "type" field exists and match with new one
		if s, ok := oldFieldSettings["type"]; ok {
			if ns, ok := newSettings["type"]; ok {
				if !reflect.DeepEqual(s, ns) {
					return true, newMappings, diags
				}
				continue
			}

			return true, newMappings, diags
		}

		// if we have "mapping" field, let's call ourself to check again
		if s, ok := oldFieldSettings["properties"]; ok {
			currentPath = currentPath.AtMapKey("properties")
			if ns, ok := newSettings["properties"]; ok {
				requiresReplace, newProperties, d := p.modifyMappings(currentPath, s.(map[string]any), ns.(map[string]any))
				diags.Append(d...)
				newSettings["properties"] = newProperties
				if requiresReplace {
					return true, newMappings, diags
				}
			} else {
				diags.AddAttributeWarning(
					path.Root("mappings"),
					fmt.Sprintf("removing field [%s] in mappings is ignored.", currentPath),
					warningDetail,
				)
				newSettings["properties"] = s
			}
		}
	}

	return false, newMappings, diags
}

func (p mappingsPlanModifier) Description(_ context.Context) string {
	return "Preserves existing mappings which don't exist in config"
}

func (p mappingsPlanModifier) MarkdownDescription(ctx context.Context) string {
	return p.Description(ctx)
}
