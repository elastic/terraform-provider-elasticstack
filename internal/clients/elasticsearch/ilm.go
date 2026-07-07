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
	"net/http"
	"strings"

	"github.com/elastic/go-elasticsearch/v9/typedapi/ilm/putlifecycle"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
)

func PutIlm(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, policy *models.Policy) fwdiags.Diagnostics {
	policyBytes, err := json.Marshal(map[string]any{"policy": policy})
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	typedClient := apiClient.GetESClient()
	var req putlifecycle.Request
	if err := json.Unmarshal(policyBytes, &req); err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	_, err = typedClient.Ilm.PutLifecycle(policy.Name).Request(&req).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func GetIlm(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, policyName string) (*types.Lifecycle, fwdiags.Diagnostics) {
	typedClient := apiClient.GetESClient()
	res, err := typedClient.Ilm.GetLifecycle().Policy(policyName).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil, nil
		}
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	if lifecycle, ok := res[policyName]; ok {
		return &lifecycle, nil
	}
	return nil, fwdiags.Diagnostics{
		fwdiags.NewErrorDiagnostic(
			"Unable to find a ILM policy in the cluster",
			fmt.Sprintf(`Unable to find "%s" ILM policy in the cluster`, policyName),
		),
	}
}

// GetIndicesWithILMPolicy returns the names of all indices currently using
// the given ILM policy.
//
// It queries GET /_ilm/policy/<policyName> and reads the
// `<policy>.in_use_by.indices` field, which Elasticsearch maintains per
// policy. This is a single targeted lookup keyed by the policy and avoids
// scanning indices cluster-wide.
//
// The typed client's generated `Lifecycle` struct does not expose
// `in_use_by`, so this function uses Perform to obtain the raw HTTP response
// and decodes the relevant subset of the body itself.
func GetIndicesWithILMPolicy(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, policyName string) ([]string, fwdiags.Diagnostics) {
	typedClient := apiClient.GetESClient()

	res, err := typedClient.Ilm.GetLifecycle().Policy(policyName).Perform(ctx)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if d := diagutil.CheckHTTPErrorFromFW(res, "Unable to fetch ILM policy"); d.HasError() {
		return nil, d
	}

	// The response is shaped as:
	//   { "<policy_name>": { "in_use_by": { "indices": [...], ... }, ... } }
	var decoded map[string]struct {
		InUseBy struct {
			Indices []string `json:"indices"`
		} `json:"in_use_by"`
	}
	if err := json.NewDecoder(res.Body).Decode(&decoded); err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	entry, ok := decoded[policyName]
	if !ok {
		return nil, nil
	}
	return entry.InUseBy.Indices, nil
}

// ClearILMPolicyFromIndices removes the ILM policy reference from the
// provided indices by setting index.lifecycle.name to null.
// It issues PUT /{indices}/_settings.
func ClearILMPolicyFromIndices(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, indices []string) fwdiags.Diagnostics {
	if len(indices) == 0 {
		return nil
	}

	settingsBytes, err := json.Marshal(map[string]any{"index.lifecycle.name": nil})
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	typedClient := apiClient.GetESClient()

	_, err = typedClient.Indices.PutSettings().Indices(strings.Join(indices, ",")).Raw(bytes.NewReader(settingsBytes)).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func DeleteIlm(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, policyName string) fwdiags.Diagnostics {
	typedClient := apiClient.GetESClient()
	_, err := typedClient.Ilm.DeleteLifecycle(policyName).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil
		}
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}
