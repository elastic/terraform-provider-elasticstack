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

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/enrichpolicyphase"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func GetEnrichPolicy(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, policyName string) (*models.EnrichPolicy, diag.Diagnostics) {
	var diags diag.Diagnostics

	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}

	res, err := typedClient.Enrich.GetPolicy().Name(policyName).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil, nil
		}
		return nil, diag.FromErr(err)
	}

	if len(res.Policies) == 0 {
		return nil, diags
	}

	if len(res.Policies) > 1 {
		tflog.Warn(ctx, fmt.Sprintf(`Somehow found more than one policy for policy named %s`, policyName))
	}

	summary := res.Policies[0]

	var policyType string
	var policy types.EnrichPolicy
	for pt, p := range summary.Config {
		policyType = pt.String()
		policy = p
		break
	}
	if policyType == "" {
		return nil, diag.Errorf("enrich policy %s has no recognized policy type", policyName)
	}

	name := policyName
	if policy.Name != nil {
		name = *policy.Name
	}

	var queryStr string
	if policy.Query != nil {
		queryBytes, err := json.Marshal(policy.Query)
		if err != nil {
			return nil, diag.FromErr(err)
		}
		// The typed client can return a non-nil *types.Query that still marshals to JSON null.
		// Avoid storing the literal string "null" in state, which would trigger replacement.
		if string(queryBytes) != "null" {
			queryStr = string(queryBytes)
		}
	}

	return &models.EnrichPolicy{
		Type:         policyType,
		Name:         name,
		Indices:      policy.Indices,
		MatchField:   policy.MatchField,
		EnrichFields: policy.EnrichFields,
		Query:        queryStr,
	}, diags
}

func PutEnrichPolicy(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, policy *models.EnrichPolicy) diag.Diagnostics {
	var diags diag.Diagnostics

	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}

	enrichPolicy := &types.EnrichPolicy{
		Indices:      policy.Indices,
		MatchField:   policy.MatchField,
		EnrichFields: policy.EnrichFields,
	}

	if policy.Query != "" {
		var query types.Query
		if err := json.Unmarshal([]byte(policy.Query), &query); err != nil {
			return diag.FromErr(err)
		}
		enrichPolicy.Query = &query
	}

	req := typedClient.Enrich.PutPolicy(policy.Name)
	switch policy.Type {
	case "geo_match":
		req.GeoMatch(enrichPolicy)
	case "match":
		req.Match(enrichPolicy)
	case "range":
		req.Range(enrichPolicy)
	default:
		return diag.Errorf("unsupported enrich policy type: %s", policy.Type)
	}

	_, err = req.Do(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func DeleteEnrichPolicy(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, policyName string) diag.Diagnostics {
	var diags diag.Diagnostics

	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = typedClient.Enrich.DeletePolicy(policyName).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return diags
		}
		return diag.FromErr(err)
	}

	return diags
}

func ExecuteEnrichPolicy(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, policyName string) diag.Diagnostics {
	var diags diag.Diagnostics

	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := typedClient.Enrich.ExecutePolicy(policyName).WaitForCompletion(true).Do(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	if res.Status == nil {
		return diag.Errorf(`Unexpected response to executing enrich policy: no status`)
	}

	if res.Status.Phase != enrichpolicyphase.COMPLETE {
		return diag.Errorf(`Unexpected response to executing enrich policy: %s`, res.Status.Phase.String())
	}

	return diags
}
