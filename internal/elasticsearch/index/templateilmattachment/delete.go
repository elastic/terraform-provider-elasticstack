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
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// deleteILMAttachment removes the index.lifecycle.name setting from the @custom
// component template. It never calls Delete Component Template because the
// template may be in use by an index template (e.g. from Fleet).
func deleteILMAttachment(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, _ tfModel) diag.Diagnostics {
	var diags diag.Diagnostics

	existingRaw, sdkDiags := elasticsearch.GetComponentTemplate(ctx, client, resourceID)
	if sdkDiags.HasError() {
		diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
		return diags
	}

	existing := toModelComponentTemplateResponse(existingRaw)
	if existing == nil {
		tflog.Debug(ctx, "Component template already deleted", map[string]any{
			"name": resourceID,
		})
		return nil
	}

	if existing.ComponentTemplate.Template != nil {
		existing.ComponentTemplate.Template.Settings = removeILMSetting(existing.ComponentTemplate.Template.Settings)
	}

	componentTemplate := models.ComponentTemplate{
		Name:     resourceID,
		Template: existing.ComponentTemplate.Template,
		Meta:     existing.ComponentTemplate.Meta,
		Version:  existing.ComponentTemplate.Version,
	}
	if sdkDiags := elasticsearch.PutComponentTemplate(ctx, client, &componentTemplate); sdkDiags.HasError() {
		diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	}
	return diags
}
