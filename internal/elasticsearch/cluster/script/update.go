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

package script

import (
	"context"
	"encoding/json"

	estypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/scriptlanguage"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// writeScript handles both Create and Update; the script PUT API is
// idempotent so the same callback serves both lifecycle methods.
func writeScript(ctx context.Context, client *clients.ElasticsearchScopedClient, req entitycore.WriteRequest[Data]) (entitycore.WriteResult[Data], diag.Diagnostics) {
	var diags diag.Diagnostics
	data := req.Plan
	scriptID := req.WriteID

	id, sdkDiags := client.ID(ctx, scriptID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return entitycore.WriteResult[Data]{}, diags
	}

	script := estypes.StoredScript{
		Lang:   scriptlanguage.ScriptLanguage{Name: data.Lang.ValueString()},
		Source: data.Source.ValueString(),
	}

	var params map[string]any
	if typeutils.IsKnown(data.Params) {
		paramsStr := data.Params.ValueString()
		if paramsStr != "" {
			err := json.Unmarshal([]byte(paramsStr), &params)
			if err != nil {
				diags.AddError("Error unmarshaling script params", err.Error())
				return entitycore.WriteResult[Data]{}, diags
			}
		}
	}

	scriptContext := ""
	if typeutils.IsKnown(data.Context) {
		scriptContext = data.Context.ValueString()
	}

	diags.Append(elasticsearch.PutScript(ctx, client, scriptID, scriptContext, &script, params)...)
	if diags.HasError() {
		return entitycore.WriteResult[Data]{}, diags
	}

	data.ID = types.StringValue(id.String())
	return entitycore.WriteResult[Data]{Model: data}, diags
}
