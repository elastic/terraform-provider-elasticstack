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

package templateilmattachment

import (
	"context"
	"encoding/json"

	estypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// tfModel represents the Terraform state model for this resource.
type tfModel struct {
	ID                      types.String `tfsdk:"id"`
	IndexTemplate           types.String `tfsdk:"index_template"`
	LifecycleName           types.String `tfsdk:"lifecycle_name"`
	ElasticsearchConnection types.List   `tfsdk:"elasticsearch_connection"`
}

// getComponentTemplateName returns the name of the @custom component template.
func (m *tfModel) getComponentTemplateName() string {
	return m.IndexTemplate.ValueString() + "@custom"
}

// GetID parses and returns the composite ID from the model.
func (m *tfModel) GetID() (*clients.CompositeID, diag.Diagnostics) {
	compID, sdkDiags := clients.CompositeIDFromStr(m.ID.ValueString())
	if sdkDiags.HasError() {
		return nil, diagutil.FrameworkDiagsFromSDK(sdkDiags)
	}
	return compID, nil
}

// mergeILMSetting adds the ILM lifecycle.name setting to existing settings.
func mergeILMSetting(existingSettings map[string]any, lifecycleName string) map[string]any {
	if existingSettings == nil {
		existingSettings = make(map[string]any)
	}
	indexSettings, ok := existingSettings["index"].(map[string]any)
	if !ok {
		indexSettings = make(map[string]any)
		existingSettings["index"] = indexSettings
	}
	lifecycle, ok := indexSettings["lifecycle"].(map[string]any)
	if !ok {
		lifecycle = make(map[string]any)
		indexSettings["lifecycle"] = lifecycle
	}
	lifecycle["name"] = lifecycleName
	return existingSettings
}

// removeILMSetting removes the index.lifecycle.name setting from the settings map.
func removeILMSetting(settings map[string]any) map[string]any {
	if settings == nil {
		return nil
	}
	if indexSettings, ok := settings["index"].(map[string]any); ok {
		if lifecycle, ok := indexSettings["lifecycle"].(map[string]any); ok {
			delete(lifecycle, "name")
			if len(lifecycle) == 0 {
				delete(indexSettings, "lifecycle")
			}
		}
		if len(indexSettings) == 0 {
			delete(settings, "index")
		}
	}
	if len(settings) == 0 {
		return nil
	}
	return settings
}

// isComponentTemplateEmpty checks if a component template has no meaningful content.
func isComponentTemplateEmpty(template *models.Template) bool {
	if template == nil {
		return true
	}
	return len(template.Settings) == 0 &&
		len(template.Mappings) == 0 &&
		len(template.Aliases) == 0
}

// extractILMSetting extracts the index.lifecycle.name setting from component template settings.
func extractILMSetting(template *models.Template) string {
	if template == nil || template.Settings == nil {
		return ""
	}
	indexSettings, ok := template.Settings["index"].(map[string]any)
	if !ok {
		return ""
	}
	lifecycle, ok := indexSettings["lifecycle"].(map[string]any)
	if !ok {
		return ""
	}
	if v, ok := lifecycle["name"].(string); ok {
		return v
	}
	return ""
}

// readILMAttachment reads the component template and updates the model with the actual ILM setting.
// It returns (true, nil) on success, (false, diags) on SDK error, and (false, nil) when the
// template or ILM setting is missing. The caller decides how to handle "not found" (e.g. Read
// removes from state, Create/Update report an error).
func readILMAttachment(ctx context.Context, model *tfModel, client *clients.ElasticsearchScopedClient) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	componentTemplateName := model.getComponentTemplateName()

	tpl, sdkDiags := elasticsearch.GetComponentTemplate(ctx, client, componentTemplateName)
	if sdkDiags.HasError() {
		diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
		return false, diags
	}

	modelTpl := toModelComponentTemplateResponse(tpl)

	if modelTpl == nil {
		return false, nil
	}

	lifecycleName := extractILMSetting(modelTpl.ComponentTemplate.Template)
	if lifecycleName == "" {
		return false, nil
	}

	model.LifecycleName = types.StringValue(lifecycleName)
	return true, nil
}

func toModelComponentTemplateResponse(tpl *estypes.ClusterComponentTemplate) *models.ComponentTemplateResponse {
	if tpl == nil {
		return nil
	}

	resp := &models.ComponentTemplateResponse{
		Name: tpl.Name,
		ComponentTemplate: models.ComponentTemplate{
			Name: tpl.Name,
		},
	}

	if tpl.ComponentTemplate.Version != nil {
		version := int(*tpl.ComponentTemplate.Version)
		resp.ComponentTemplate.Version = &version
	}

	if tpl.ComponentTemplate.Meta_ != nil {
		metaBytes, _ := json.Marshal(tpl.ComponentTemplate.Meta_)
		var metaMap map[string]any
		_ = json.Unmarshal(metaBytes, &metaMap)
		resp.ComponentTemplate.Meta = metaMap
	}

	{
		t := &models.Template{}

		if tpl.ComponentTemplate.Template.Settings != nil {
			settingsBytes, _ := json.Marshal(tpl.ComponentTemplate.Template.Settings)
			var settingsMap map[string]any
			_ = json.Unmarshal(settingsBytes, &settingsMap)
			t.Settings = settingsMap
		}

		if tpl.ComponentTemplate.Template.Mappings != nil {
			mappingsBytes, _ := json.Marshal(tpl.ComponentTemplate.Template.Mappings)
			var mappingsMap map[string]any
			_ = json.Unmarshal(mappingsBytes, &mappingsMap)
			t.Mappings = mappingsMap
		}

		if len(tpl.ComponentTemplate.Template.Aliases) > 0 {
			t.Aliases = make(map[string]models.IndexAlias, len(tpl.ComponentTemplate.Template.Aliases))
			for name, alias := range tpl.ComponentTemplate.Template.Aliases {
				ia := models.IndexAlias{Name: name}
				if alias.Filter != nil {
					filterBytes, _ := json.Marshal(alias.Filter)
					var filterMap map[string]any
					_ = json.Unmarshal(filterBytes, &filterMap)
					ia.Filter = filterMap
				}
				if alias.IndexRouting != nil {
					ia.IndexRouting = *alias.IndexRouting
				}
				if alias.IsHidden != nil {
					ia.IsHidden = *alias.IsHidden
				}
				if alias.IsWriteIndex != nil {
					ia.IsWriteIndex = *alias.IsWriteIndex
				}
				if alias.Routing != nil {
					ia.Routing = *alias.Routing
				}
				if alias.SearchRouting != nil {
					ia.SearchRouting = *alias.SearchRouting
				}
				t.Aliases[name] = ia
			}
		}

		resp.ComponentTemplate.Template = t
	}

	return resp
}
