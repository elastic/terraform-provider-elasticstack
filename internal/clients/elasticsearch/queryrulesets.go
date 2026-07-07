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

	"github.com/elastic/go-elasticsearch/v9/typedapi/queryrules/getruleset"
	"github.com/elastic/go-elasticsearch/v9/typedapi/queryrules/putruleset"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

// PutQueryRuleset creates or replaces a query ruleset with the provided rules.
func PutQueryRuleset(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, rulesetID string, rules []types.QueryRule) fwdiag.Diagnostics {
	typedClient := apiClient.GetESClient()

	_, err := typedClient.QueryRules.PutRuleset(rulesetID).Request(&putruleset.Request{
		Rules: rules,
	}).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	return nil
}

// GetQueryRuleset retrieves a query ruleset by ID. Returns nil, nil on 404.
func GetQueryRuleset(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, rulesetID string) (*getruleset.Response, fwdiag.Diagnostics) {
	typedClient := apiClient.GetESClient()

	res, err := typedClient.QueryRules.GetRuleset(rulesetID).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil, nil
		}
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	return res, nil
}

// DeleteQueryRuleset deletes a query ruleset. Returns nil on 404 (idempotent).
func DeleteQueryRuleset(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, rulesetID string) fwdiag.Diagnostics {
	typedClient := apiClient.GetESClient()

	_, err := typedClient.QueryRules.DeleteRuleset(rulesetID).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil
		}
		return diagutil.FrameworkDiagFromError(err)
	}

	return nil
}
