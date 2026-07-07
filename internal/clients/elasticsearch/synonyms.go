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
	"errors"
	"fmt"

	"github.com/elastic/go-elasticsearch/v9/typedapi/synonyms/putsynonym"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

const synonymPageSize = 500

// GetSynonymSet retrieves all synonym rules for a synonym set, paginating through
// all results using the specified page size.
func GetSynonymSet(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, synonymSetID string) ([]types.SynonymRuleRead, fwdiag.Diagnostics) {
	typedClient := apiClient.GetESClient()

	allRules := make([]types.SynonymRuleRead, 0)
	from := 0

	for {
		res, err := typedClient.Synonyms.GetSynonym(synonymSetID).From(from).Size(synonymPageSize).Do(ctx)
		if err != nil {
			if IsNotFoundElasticsearchError(err) {
				return nil, nil
			}
			return nil, diagutil.FrameworkDiagFromError(err)
		}

		allRules = append(allRules, res.SynonymsSet...)

		if len(res.SynonymsSet) == 0 || from+len(res.SynonymsSet) >= res.Count {
			break
		}

		from += len(res.SynonymsSet)
	}

	return allRules, nil
}

// PutSynonymSet creates or replaces a synonym set with the provided rules.
func PutSynonymSet(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, synonymSetID string, rules []types.SynonymRule) fwdiag.Diagnostics {
	typedClient := apiClient.GetESClient()

	_, err := typedClient.Synonyms.PutSynonym(synonymSetID).Request(&putsynonym.Request{
		SynonymsSet: rules,
	}).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	return nil
}

// DeleteSynonymSet deletes a synonym set. If the set is referenced by one or more
// index analyzers, the API returns HTTP 400 and a descriptive diagnostic is returned.
func DeleteSynonymSet(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, synonymSetID string) fwdiag.Diagnostics {
	typedClient := apiClient.GetESClient()

	_, err := typedClient.Synonyms.DeleteSynonym(synonymSetID).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil
		}

		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr != nil && esErr.Status == 400 {
			return fwdiag.Diagnostics{
				fwdiag.NewErrorDiagnostic(
					fmt.Sprintf("Cannot delete synonym set '%s'", synonymSetID),
					"The synonym set is referenced by one or more index analyzers. Remove the synonym set from all analyzer configurations before deleting it.",
				),
			}
		}

		return diagutil.FrameworkDiagFromError(err)
	}

	return nil
}
