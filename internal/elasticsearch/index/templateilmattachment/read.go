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
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const customSuffix = "@custom"

// readILMAttachment derives index_template from the component template name
// during import (when state.IndexTemplate is unknown) and returns found=false
// when the template or the ILM setting is absent.
func readILMAttachment(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, state tfModel) (tfModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !typeutils.IsKnown(state.IndexTemplate) {
		state.IndexTemplate = types.StringValue(strings.TrimSuffix(resourceID, customSuffix))
	}

	tpl, sdkDiags := elasticsearch.GetComponentTemplate(ctx, client, resourceID)
	if sdkDiags.HasError() {
		diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
		return state, false, diags
	}

	modelTpl := toModelComponentTemplateResponse(tpl)
	if modelTpl == nil {
		return state, false, nil
	}

	lifecycleName := extractILMSetting(modelTpl.ComponentTemplate.Template)
	if lifecycleName == "" {
		return state, false, nil
	}

	state.LifecycleName = types.StringValue(lifecycleName)
	return state, true, nil
}
