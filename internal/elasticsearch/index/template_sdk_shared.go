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
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// validateDataStreamOptionsVersion returns an error diagnostic if data_stream_options is configured and the server is too old.
// Used by tests; index template resource logic lives in the Plugin Framework template package.
func validateDataStreamOptionsVersion(serverVersion *version.Version, templ *models.Template) diag.Diagnostics {
	if templ != nil && templ.DataStreamOptions != nil && serverVersion.LessThan(MinSupportedDataStreamOptionsVersion) {
		return diag.FromErr(fmt.Errorf("'data_stream_options' is supported only for Elasticsearch v%s and above", MinSupportedDataStreamOptionsVersion.String()))
	}
	return nil
}

func expandTemplate(config any) (models.Template, bool, diag.Diagnostics) {
	templ := models.Template{}
	// only one template block allowed to be declared
	definedTempl, ok := config.([]any)[0].(map[string]any)
	if !ok {
		return templ, false, nil
	}

	aliases, diags := ExpandIndexAliases(definedTempl["alias"].(*schema.Set))
	if diags.HasError() {
		return templ, false, diags
	}
	templ.Aliases = aliases

	if lc, ok := definedTempl["lifecycle"]; ok {
		lifecycle := ExpandLifecycle(lc.(*schema.Set))
		if lifecycle != nil {
			templ.Lifecycle = lifecycle
		}
	}

	if mappings, ok := definedTempl["mappings"]; ok {
		if mappings.(string) != "" {
			maps := make(map[string]any)
			if err := json.Unmarshal([]byte(mappings.(string)), &maps); err != nil {
				return templ, false, diag.FromErr(err)
			}
			templ.Mappings = maps
		}
	}

	if settings, ok := definedTempl["settings"]; ok {
		if settings.(string) != "" {
			sets := make(map[string]any)
			if err := json.Unmarshal([]byte(settings.(string)), &sets); err != nil {
				return templ, false, diag.FromErr(err)
			}
			templ.Settings = sets
		}
	}

	if dso, ok := definedTempl["data_stream_options"]; ok {
		dsoList, ok := dso.([]any)
		if ok && len(dsoList) > 0 && dsoList[0] != nil {
			dsoMap := dsoList[0].(map[string]any)
			dataStreamOptions := &models.DataStreamOptions{}
			if fs, ok := dsoMap["failure_store"]; ok {
				fsList, ok := fs.([]any)
				if ok && len(fsList) > 0 && fsList[0] != nil {
					fsMap := fsList[0].(map[string]any)
					failureStore := &models.FailureStoreOptions{}
					if enabled, ok := fsMap["enabled"]; ok {
						failureStore.Enabled = enabled.(bool)
					}
					if lc, ok := fsMap["lifecycle"]; ok {
						lcList, ok := lc.([]any)
						if ok && len(lcList) > 0 && lcList[0] != nil {
							lcMap := lcList[0].(map[string]any)
							lifecycle := &models.FailureStoreLifecycle{}
							if dr, ok := lcMap["data_retention"]; ok {
								lifecycle.DataRetention = dr.(string)
							}
							failureStore.Lifecycle = lifecycle
						}
					}
					dataStreamOptions.FailureStore = failureStore
				}
			}
			templ.DataStreamOptions = dataStreamOptions
		}
	}

	return templ, true, nil
}

func flattenTemplateData(template *models.Template, preservedAliasRouting map[string]string) ([]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	tmpl := make(map[string]any)
	if template.Mappings != nil {
		m, err := json.Marshal(template.Mappings)
		if err != nil {
			return nil, diag.FromErr(err)
		}
		tmpl["mappings"] = string(m)
	}
	if template.Settings != nil {
		s, err := json.Marshal(template.Settings)
		if err != nil {
			return nil, diag.FromErr(err)
		}
		tmpl["settings"] = string(s)
	}

	if template.Aliases != nil {
		aliases, diags := FlattenIndexAliases(template.Aliases)
		if diags.HasError() {
			return nil, diags
		}
		aliases = preserveAliasRoutingInFlattenedAliases(aliases, preservedAliasRouting)
		tmpl["alias"] = aliases
	}

	if template.Lifecycle != nil {
		lifecycle := FlattenLifecycle(template.Lifecycle)
		tmpl["lifecycle"] = lifecycle
	}

	if template.DataStreamOptions != nil && template.DataStreamOptions.FailureStore != nil {
		fs := make(map[string]any)
		fs["enabled"] = template.DataStreamOptions.FailureStore.Enabled
		if template.DataStreamOptions.FailureStore.Lifecycle != nil {
			lc := make(map[string]any)
			lc["data_retention"] = template.DataStreamOptions.FailureStore.Lifecycle.DataRetention
			fs["lifecycle"] = []any{lc}
		}
		dso := map[string]any{"failure_store": []any{fs}}
		tmpl["data_stream_options"] = []any{dso}
	}

	return []any{tmpl}, diags
}

func extractAliasRoutingFromTemplateState(rawTemplate any) map[string]string {
	preservedRouting := make(map[string]string)

	templates, ok := rawTemplate.([]any)
	if !ok || len(templates) == 0 || templates[0] == nil {
		return preservedRouting
	}

	tmpl, ok := templates[0].(map[string]any)
	if !ok {
		return preservedRouting
	}

	extractAliasRoutingFromRawAliases(tmpl["alias"], preservedRouting)
	return preservedRouting
}

func extractAliasRoutingFromRawAliases(rawAliases any, preservedRouting map[string]string) {
	switch aliases := rawAliases.(type) {
	case *schema.Set:
		for _, alias := range aliases.List() {
			addAliasRouting(alias, preservedRouting)
		}
	case []any:
		for _, alias := range aliases {
			addAliasRouting(alias, preservedRouting)
		}
	}
}

func addAliasRouting(rawAlias any, preservedRouting map[string]string) {
	aliasMap, ok := rawAlias.(map[string]any)
	if !ok {
		return
	}

	name, _ := aliasMap["name"].(string)
	routing, _ := aliasMap["routing"].(string)
	if name != "" && routing != "" {
		preservedRouting[name] = routing
	}
}

func preserveAliasRoutingInFlattenedAliases(rawAliases any, preservedAliasRouting map[string]string) any {
	if len(preservedAliasRouting) == 0 {
		return rawAliases
	}

	aliases, ok := rawAliases.([]any)
	if !ok {
		return rawAliases
	}

	for _, rawAlias := range aliases {
		aliasMap, ok := rawAlias.(map[string]any)
		if !ok {
			continue
		}

		aliasName, _ := aliasMap["name"].(string)
		if routing, found := preservedAliasRouting[aliasName]; found {
			aliasMap["routing"] = routing
		}
	}

	return aliases
}

func hashAliasByName(v any) int {
	aliasMap, ok := v.(map[string]any)
	if !ok {
		return 0
	}

	name, _ := aliasMap["name"].(string)
	return schema.HashString(name)
}
