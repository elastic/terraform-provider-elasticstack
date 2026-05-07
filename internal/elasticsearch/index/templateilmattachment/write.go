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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// writeILMAttachment reads the existing component template, merges the ILM
// lifecycle setting into its settings, and writes it back. The isCreate flag
// controls whether an extra warning is emitted when the template already has
// an ILM setting.
func writeILMAttachment(ctx context.Context, client *clients.ElasticsearchScopedClient, componentTemplateName string, plan tfModel, isCreate bool) (tfModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	existingRaw, sdkDiags := elasticsearch.GetComponentTemplate(ctx, client, componentTemplateName)
	if sdkDiags.HasError() {
		return plan, diagutil.FrameworkDiagsFromSDK(sdkDiags)
	}

	existing := toModelComponentTemplateResponse(existingRaw)

	var componentTemplate models.ComponentTemplate
	if existing != nil {
		componentTemplate = existing.ComponentTemplate

		if componentTemplate.Version != nil {
			tflog.Warn(ctx,
				"Existing component template has a version field. This resource does not update the version when "+
					"modifying the template. If you rely on version tracking for change detection, consider using "+
					"elasticstack_elasticsearch_component_template instead.",
				map[string]any{
					"component_template": componentTemplateName,
					"existing_version":   *componentTemplate.Version,
				})
		}

		if isCreate {
			existingILM := extractILMSetting(componentTemplate.Template)
			if existingILM != "" {
				tflog.Warn(ctx,
					"Component template already has an ILM policy configured. This resource will overwrite it. "+
						"If this is unexpected, another process may be managing this setting.",
					map[string]any{
						"component_template": componentTemplateName,
						"existing_ilm":       existingILM,
						"new_ilm":            plan.LifecycleName.ValueString(),
					})
			}
		}
	}
	componentTemplate.Name = componentTemplateName

	if componentTemplate.Template == nil {
		componentTemplate.Template = &models.Template{}
	}

	componentTemplate.Template.Settings = mergeILMSetting(
		componentTemplate.Template.Settings,
		plan.LifecycleName.ValueString(),
	)

	if sdkDiags := elasticsearch.PutComponentTemplate(ctx, client, &componentTemplate); sdkDiags.HasError() {
		return plan, diagutil.FrameworkDiagsFromSDK(sdkDiags)
	}

	if isCreate {
		id, sdkDiags := client.ID(ctx, componentTemplateName)
		if sdkDiags.HasError() {
			return plan, diagutil.FrameworkDiagsFromSDK(sdkDiags)
		}
		plan.ID = types.StringValue(id.String())
	}

	return plan, diags
}
