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
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func readScript(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, state Data) (Data, bool, diag.Diagnostics) {
	readData, diags := readScriptPayload(ctx, client, resourceID, state)
	if diags.HasError() {
		return state, false, diags
	}
	if readData.ScriptID.IsNull() {
		tflog.Warn(ctx, fmt.Sprintf(`Script "%s" not found`, resourceID))
		return state, false, diags
	}

	readData.ElasticsearchConnection = state.ElasticsearchConnection
	readData.ID = state.ID
	readData.Context = state.Context
	if readData.Params.IsNull() && !state.Params.IsNull() {
		readData.Params = state.Params
	}

	return readData, true, diags
}

func readScriptPayload(ctx context.Context, client *clients.ElasticsearchScopedClient, scriptID string, stateData Data) (Data, diag.Diagnostics) {
	var data Data
	var diags diag.Diagnostics

	script, frameworkDiags := elasticsearch.GetScript(ctx, client, scriptID)
	diags.Append(frameworkDiags...)
	if diags.HasError() {
		return data, diags
	}

	if script == nil {
		data.ScriptID = types.StringNull()
		return data, diags
	}

	data.ScriptID = types.StringValue(scriptID)
	data.Lang = types.StringValue(script.Lang.Name)
	if script.Lang.Name == "" {
		data.Lang = types.StringValue(script.Lang.String())
	}
	data.Source = types.StringValue(script.Source)

	if stateData.Params.IsNull() {
		data.Params = jsontypes.NewNormalizedNull()
	} else {
		data.Params = stateData.Params
	}

	return data, diags
}
