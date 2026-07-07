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
	"fmt"
	"io"
	"net/http"

	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

type SlmPolicy struct {
	Name       string        `json:"name"`
	Schedule   string        `json:"schedule"`
	Repository string        `json:"repository"`
	Config     *SlmConfig    `json:"config"`
	Retention  *SlmRetention `json:"retention"`
}

type SlmConfig struct {
	FeatureStates      []string       `json:"feature_states"`
	ExpandWildcards    string         `json:"expand_wildcards"`
	IgnoreUnavailable  *bool          `json:"ignore_unavailable"`
	IncludeGlobalState *bool          `json:"include_global_state"`
	Indices            []string       `json:"indices"`
	Metadata           types.Metadata `json:"metadata"`
	Partial            *bool          `json:"partial"`
}

type SlmRetention struct {
	ExpireAfter *string `json:"expire_after"`
	MaxCount    *int    `json:"max_count"`
	MinCount    *int    `json:"min_count"`
}

func PutSlm(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, policyID string, slm *SlmPolicy) fwdiag.Diagnostics {
	typedClient := apiClient.GetESClient()

	req := typedClient.Slm.PutLifecycle(policyID)

	// Build request body manually to support expand_wildcards and accurate
	// Retention field omission (types.Retention lacks omitempty on MaxCount/MinCount).
	body := map[string]any{
		"name":       slm.Name,
		"repository": slm.Repository,
		"schedule":   slm.Schedule,
	}
	if slm.Config != nil {
		config := map[string]any{}
		if len(slm.Config.FeatureStates) > 0 {
			config["feature_states"] = slm.Config.FeatureStates
		}
		if slm.Config.IgnoreUnavailable != nil {
			config["ignore_unavailable"] = *slm.Config.IgnoreUnavailable
		}
		if slm.Config.IncludeGlobalState != nil {
			config["include_global_state"] = *slm.Config.IncludeGlobalState
		}
		if len(slm.Config.Indices) > 0 {
			config["indices"] = slm.Config.Indices
		}
		if slm.Config.Metadata != nil {
			meta := make(map[string]any)
			for k, v := range slm.Config.Metadata {
				var val any
				if err := json.Unmarshal(v, &val); err != nil {
					return diagutil.FrameworkDiagFromError(fmt.Errorf("failed to unmarshal metadata key %q: %w", k, err))
				}
				meta[k] = val
			}
			config["metadata"] = meta
		}
		if slm.Config.Partial != nil {
			config["partial"] = *slm.Config.Partial
		}
		if slm.Config.ExpandWildcards != "" {
			config["expand_wildcards"] = slm.Config.ExpandWildcards
		}
		if len(config) > 0 {
			body["config"] = config
		}
	}
	if slm.Retention != nil {
		retention := map[string]any{}
		if slm.Retention.ExpireAfter != nil {
			retention["expire_after"] = *slm.Retention.ExpireAfter
		}
		if slm.Retention.MaxCount != nil {
			retention["max_count"] = *slm.Retention.MaxCount
		}
		if slm.Retention.MinCount != nil {
			retention["min_count"] = *slm.Retention.MinCount
		}
		if len(retention) > 0 {
			body["retention"] = retention
		}
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	req.Raw(bytes.NewReader(bodyBytes))

	_, err = req.Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func GetSlm(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, slmName string) (*SlmPolicy, fwdiag.Diagnostics) {
	typedClient := apiClient.GetESClient()
	res, err := typedClient.Slm.GetLifecycle().PolicyId(slmName).Perform(ctx)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if d := diagutil.CheckHTTPErrorFromFW(res, fmt.Sprintf("Unable to get SLM policy: %s", slmName)); d.HasError() {
		return nil, d
	}
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	var rawResp map[string]struct {
		Policy SlmPolicy `json:"policy"`
	}
	if err := json.Unmarshal(bodyBytes, &rawResp); err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	if slm, ok := rawResp[slmName]; ok {
		return &slm.Policy, nil
	}
	return nil, fwdiag.Diagnostics{
		fwdiag.NewErrorDiagnostic(
			"Unable to find the SLM policy in the response",
			fmt.Sprintf(`Unable to find "%s" policy in the ES API response.`, slmName),
		),
	}
}

func DeleteSlm(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, slmName string) fwdiag.Diagnostics {
	typedClient := apiClient.GetESClient()
	_, err := typedClient.Slm.DeleteLifecycle(slmName).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil
		}
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}
