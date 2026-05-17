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
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	sdkdiag "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func PutSettings(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, settings map[string]any) sdkdiag.Diagnostics {
	var diags sdkdiag.Diagnostics
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return sdkdiag.FromErr(err)
	}

	req := typedClient.Cluster.PutSettings()

	if persistent, ok := settings["persistent"].(map[string]any); ok {
		raw, err := toRawMessageMap(persistent)
		if err != nil {
			return sdkdiag.FromErr(err)
		}
		req.Persistent(raw)
	}
	if transient, ok := settings["transient"].(map[string]any); ok {
		raw, err := toRawMessageMap(transient)
		if err != nil {
			return sdkdiag.FromErr(err)
		}
		req.Transient(raw)
	}

	_, err = req.Do(ctx)
	if err != nil {
		return sdkdiag.FromErr(err)
	}
	return diags
}

func toRawMessageMap(m map[string]any) (map[string]json.RawMessage, error) {
	result := make(map[string]json.RawMessage, len(m))
	for k, v := range m {
		data, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal setting %q: %w", k, err)
		}
		result[k] = data
	}
	return result, nil
}

func GetSettings(ctx context.Context, apiClient *clients.ElasticsearchScopedClient) (map[string]any, sdkdiag.Diagnostics) {
	var diags sdkdiag.Diagnostics
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, sdkdiag.FromErr(err)
	}
	resp, err := typedClient.Cluster.GetSettings().FlatSettings(true).Do(ctx)
	if err != nil {
		return nil, sdkdiag.FromErr(err)
	}

	result := make(map[string]any)
	result["persistent"], err = flattenRawMessageMap(resp.Persistent)
	if err != nil {
		return nil, sdkdiag.FromErr(err)
	}
	result["transient"], err = flattenRawMessageMap(resp.Transient)
	if err != nil {
		return nil, sdkdiag.FromErr(err)
	}
	result["defaults"], err = flattenRawMessageMap(resp.Defaults)
	if err != nil {
		return nil, sdkdiag.FromErr(err)
	}
	return result, diags
}

func flattenRawMessageMap(m map[string]json.RawMessage) (map[string]any, error) {
	result := make(map[string]any, len(m))
	for k, v := range m {
		var val any
		if err := json.Unmarshal(v, &val); err != nil {
			return nil, fmt.Errorf("failed to unmarshal setting %q: %w", k, err)
		}
		result[k] = val
	}
	return result, nil
}
