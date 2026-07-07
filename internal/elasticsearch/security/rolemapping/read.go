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

package rolemapping

import (
	"context"
	"encoding/json"

	esapiTypes "github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// readRoleMapping reads role mapping data from Elasticsearch and returns Data
func readRoleMapping(ctx context.Context, stateData Data, roleMappingName string, client *clients.ElasticsearchScopedClient) (*Data, diag.Diagnostics) {
	var diags diag.Diagnostics

	roleMapping, apiDiags := elasticsearch.GetRoleMapping(ctx, client, roleMappingName)
	diags.Append(apiDiags...)
	if diags.HasError() {
		return nil, diags
	}

	if roleMapping == nil {
		return nil, diags
	}

	data := &Data{}

	// Set basic fields
	compID, compDiags := client.ID(ctx, roleMappingName)
	diags.Append(compDiags...)
	if diags.HasError() {
		return nil, diags
	}
	data.ID = types.StringValue(compID.String())
	data.ElasticsearchConnection = stateData.ElasticsearchConnection
	data.Name = types.StringValue(roleMappingName)
	data.Enabled = types.BoolValue(roleMapping.Enabled)

	// Handle rules — store the typed client's JSON as-is (Elasticsearch may
	// return single-element field values as strings or arrays). StringSemanticEquals
	// on NormalizedRulesValue treats both forms as equal during plan comparison.
	rulesJSON, err := json.Marshal(roleMapping.Rules)
	if err != nil {
		diags.AddAttributeError(path.Root("rules"), "Failed to marshal rules", err.Error())
		return nil, diags
	}
	data.Rules = NewNormalizedRulesValue(string(rulesJSON))

	// Handle roles
	data.Roles = typeutils.SetValueFrom(ctx, roleMapping.Roles, types.StringType, path.Root("roles"), &diags)
	if diags.HasError() {
		return nil, diags
	}

	// Handle role templates
	// Preserve planned/state value when known to avoid representation drift
	// caused by the typed client's Script type normalizing strings to objects.
	switch {
	case typeutils.IsKnown(stateData.RoleTemplates):
		data.RoleTemplates = stateData.RoleTemplates
	case len(roleMapping.RoleTemplates) > 0:
		templatesJSON, err := roleTemplatesToJSON(roleMapping.RoleTemplates)
		if err != nil {
			diags.AddError("Failed to serialize role templates", err.Error())
			return nil, diags
		}
		data.RoleTemplates = jsontypes.NewNormalizedValue(templatesJSON)
	default:
		data.RoleTemplates = jsontypes.NewNormalizedNull()
	}

	// Handle metadata
	metadataJSON, err := json.Marshal(roleMapping.Metadata)
	if err != nil {
		diags.AddError("Failed to marshal metadata", err.Error())
		return nil, diags
	}
	data.Metadata = jsontypes.NewNormalizedValue(string(metadataJSON))

	return data, diags
}

func readRoleMappingResource(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, state Data) (Data, bool, diag.Diagnostics) {
	readData, diags := readRoleMapping(ctx, state, resourceID, client)
	if diags.HasError() {
		return state, false, diags
	}
	if readData == nil {
		return state, false, nil
	}
	return *readData, true, diags
}

// roleTemplatesToJSON serializes typed role templates back to JSON by
// directly extracting the Format and Template.Source fields. This avoids
// round-trip drift caused by the typed client's Script type which may
// normalize a plain template string into {"source":"..."} on marshal.
func roleTemplatesToJSON(templates []esapiTypes.RoleTemplate) (string, error) {
	items := make([]map[string]any, len(templates))
	for i, t := range templates {
		item := map[string]any{}
		if t.Format != nil {
			item["format"] = t.Format.String()
		}
		if t.Template.Source != nil {
			item["template"] = *t.Template.Source
		}
		items[i] = item
	}
	out, err := json.Marshal(items)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
