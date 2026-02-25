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
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *scriptResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data Data
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	compID, diags := clients.CompositeIDFromStrFw(data.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	scriptID := compID.ResourceID

	client, diags := clients.MaybeNewAPIClientFromFrameworkResource(ctx, data.ElasticsearchConnection, r.client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use the helper read function
	readData, readDiags := r.read(ctx, scriptID, client)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if script was found
	if readData.ScriptID.IsNull() {
		tflog.Warn(ctx, fmt.Sprintf(`Script "%s" not found, removing from state`, compID.ResourceID))
		resp.State.RemoveResource(ctx)
		return
	}

	// Preserve connection and ID from original state
	readData.ElasticsearchConnection = data.ElasticsearchConnection
	readData.ID = data.ID

	// Preserve context from state as it's not returned by the API
	readData.Context = data.Context

	// Preserve params from state if API didn't return them
	if readData.Params.IsNull() && !data.Params.IsNull() {
		readData.Params = data.Params
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &readData)...)
}

func (r *scriptResource) read(ctx context.Context, scriptID string, client *clients.APIClient) (Data, diag.Diagnostics) {
	var data Data
	var diags diag.Diagnostics

	script, frameworkDiags := elasticsearch.GetScript(ctx, client, scriptID)
	diags.Append(frameworkDiags...)
	if diags.HasError() {
		return data, diags
	}

	if script == nil {
		// Script not found - return empty data with null ScriptId to signal not found
		data.ScriptID = types.StringNull()
		return data, diags
	}

	data.ScriptID = types.StringValue(scriptID)
	data.Lang = types.StringValue(script.Language)
	data.Source = types.StringValue(script.Source)

	// Handle params if returned by the API
	if len(script.Params) > 0 {
		paramsBytes, err := json.Marshal(script.Params)
		if err != nil {
			diags.AddError("Error marshaling script params", err.Error())
			return data, diags
		}
		data.Params = jsontypes.NewNormalizedValue(string(paramsBytes))
	} else {
		data.Params = jsontypes.NewNormalizedNull()
	}
	// Note: If params were set but API doesn't return them, they are preserved from state
	// This maintains backwards compatibility

	// Note: context is not returned by the Elasticsearch API (json:"-" in model)
	// It's only used during script creation, so we preserve it from state
	// This is consistent with the SDKv2 implementation

	return data, diags
}
