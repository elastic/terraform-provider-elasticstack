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
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure tfModel satisfies the entitycore.ElasticsearchResourceModel interface.
var _ interface {
	GetID() types.String
	GetResourceID() types.String
	GetElasticsearchConnection() types.List
} = tfModel{}

var _ entitycore.WithVersionRequirements = tfModel{}

// nameKey is the lifecycle "name" setting key in component template
// index.lifecycle.name. It is shared between models.go and delete.go.
const nameKey = "name"

// tfModel represents the Terraform state model for this resource.
type tfModel struct {
	ID                      types.String `tfsdk:"id"`
	IndexTemplate           types.String `tfsdk:"index_template"`
	LifecycleName           types.String `tfsdk:"lifecycle_name"`
	ElasticsearchConnection types.List   `tfsdk:"elasticsearch_connection"`
}

// getComponentTemplateName returns the name of the @custom component template.
func (m tfModel) getComponentTemplateName() string {
	return m.GetResourceID().ValueString()
}

// GetID returns the composite ID string for the resource.
func (m tfModel) GetID() types.String { return m.ID }

// GetResourceID returns the derived component template name used as the write identity.
func (m tfModel) GetResourceID() types.String {
	if !typeutils.IsKnown(m.IndexTemplate) || m.IndexTemplate.ValueString() == "" {
		return types.StringUnknown()
	}
	return types.StringValue(m.IndexTemplate.ValueString() + "@custom")
}

// GetElasticsearchConnection returns the Elasticsearch connection configuration.
func (m tfModel) GetElasticsearchConnection() types.List { return m.ElasticsearchConnection }

// GetVersionRequirements satisfies [entitycore.WithVersionRequirements] and enforces
// the ES >= 8.2.0 minimum required by the Put Component Template API used by this resource.
func (m tfModel) GetVersionRequirements(_ context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{{
		MinVersion: *MinVersion,
		ErrorMessage: fmt.Sprintf(
			"This resource requires Elasticsearch %s or later. "+
				"This resource is not supported on this Elasticsearch version.",
			MinVersion,
		),
	}}, nil
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
	lifecycle[nameKey] = lifecycleName
	return existingSettings
}

// removeILMSetting removes the index.lifecycle.name setting from the settings map.
func removeILMSetting(settings map[string]any) map[string]any {
	if settings == nil {
		return nil
	}
	if indexSettings, ok := settings["index"].(map[string]any); ok {
		if lifecycle, ok := indexSettings["lifecycle"].(map[string]any); ok {
			delete(lifecycle, nameKey)
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
	if v, ok := lifecycle[nameKey].(string); ok {
		return v
	}
	return ""
}
