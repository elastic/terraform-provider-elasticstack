package index

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type mappingsPlanModifier struct{}

func (p mappingsPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !utils.IsKnown(req.StateValue) {
		return
	}

	if !utils.IsKnown(req.ConfigValue) {
		return
	}

	stateStr := req.StateValue.ValueString()
	cfgStr := req.ConfigValue.ValueString()

	var stateMappings map[string]interface{}
	var cfgMappings map[string]interface{}

	// No error checking, schema validation ensures this is valid json
	_ = json.Unmarshal([]byte(stateStr), &stateMappings)
	_ = json.Unmarshal([]byte(cfgStr), &cfgMappings)

	if stateProps, ok := stateMappings["properties"]; ok {
		cfgProps, ok := cfgMappings["properties"]
		if !ok {
			resp.RequiresReplace = true
			return
		}

		requiresReplace, finalMappings, diags := p.modifyMappings(ctx, path.Root("mappings").AtMapKey("properties"), stateProps.(map[string]interface{}), cfgProps.(map[string]interface{}))
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

func (p mappingsPlanModifier) modifyMappings(ctx context.Context, initialPath path.Path, old map[string]interface{}, new map[string]interface{}) (bool, map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	for k, v := range old {
		oldFieldSettings := v.(map[string]interface{})
		newFieldSettings, ok := new[k]
		currentPath := initialPath.AtMapKey(k)
		// When field is removed, it'll be ignored in elasticsearch
		if !ok {
			diags.AddAttributeWarning(path.Root("mappings"), fmt.Sprintf("removing field [%s] in mappings is ignored.", currentPath), "Elasticsearch will maintain the current field in it's mapping. Re-index to remove the field completely")
			new[k] = v
			continue
		}
		newSettings := newFieldSettings.(map[string]interface{})
		// check if the "type" field exists and match with new one
		if s, ok := oldFieldSettings["type"]; ok {
			if ns, ok := newSettings["type"]; ok {
				if !reflect.DeepEqual(s, ns) {
					return true, new, diags
				}
				continue
			} else {
				return true, new, diags
			}
		}

		// if we have "mapping" field, let's call ourself to check again
		if s, ok := oldFieldSettings["properties"]; ok {
			currentPath = currentPath.AtMapKey("properties")
			if ns, ok := newSettings["properties"]; ok {
				requiresReplace, newProperties, d := p.modifyMappings(ctx, currentPath, s.(map[string]interface{}), ns.(map[string]interface{}))
				diags.Append(d...)
				newSettings["properties"] = newProperties
				if requiresReplace {
					return true, new, diags
				}
			} else {
				diags.AddAttributeWarning(path.Root("mappings"), fmt.Sprintf("removing field [%s] in mappings is ignored.", currentPath), "Elasticsearch will maintain the current field in it's mapping. Re-index to remove the field completely")
				newSettings["properties"] = s
			}
		}
	}

	return false, new, diags
}

func (p mappingsPlanModifier) Description(_ context.Context) string {
	return "Preserves existing mappings which don't exist in config"
}

func (p mappingsPlanModifier) MarkdownDescription(ctx context.Context) string {
	return p.Description(ctx)
}
