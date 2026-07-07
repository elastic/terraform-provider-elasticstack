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

package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

func GetScript(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, id string) (*types.StoredScript, fwdiag.Diagnostics) {
	typedClient := apiClient.GetESClient()
	resp, err := typedClient.Core.GetScript(id).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil, nil
		}
		var diags fwdiag.Diagnostics
		diags.AddError("Failed to get script", err.Error())
		return nil, diags
	}
	return resp.Script, nil
}

func PutScript(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, id string, context string, script *types.StoredScript, params map[string]any) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics
	typedClient := apiClient.GetESClient()

	req := typedClient.Core.PutScript(id)
	if context != "" {
		req.Context(context)
	}

	// Build request body manually to support params (types.StoredScript lacks Params).
	type storedScriptWithParams struct {
		types.StoredScript
		Params map[string]any `json:"params,omitempty"`
	}
	body := struct {
		Script storedScriptWithParams `json:"script"`
	}{
		Script: storedScriptWithParams{
			StoredScript: *script,
			Params:       params,
		},
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		diags.AddError("Failed to marshal script request", err.Error())
		return diags
	}
	req.Raw(bytes.NewReader(bodyBytes))

	_, err = req.Do(ctx)
	if err != nil {
		diags.AddError("Failed to put script", err.Error())
		return diags
	}
	return diags
}

func DeleteScript(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, id string) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics
	typedClient := apiClient.GetESClient()
	_, err := typedClient.Core.DeleteScript(id).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return diags
		}
		diags.AddError("Failed to delete script", err.Error())
		return diags
	}
	return diags
}
