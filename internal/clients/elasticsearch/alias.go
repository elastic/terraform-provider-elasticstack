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
	"strings"

	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
)

func DeleteIndexAlias(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, index string, aliases []string) fwdiags.Diagnostics {
	typedClient := apiClient.GetESClient()
	_, err := typedClient.Indices.DeleteAlias(index, strings.Join(aliases, ",")).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil
		}
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func UpdateIndexAlias(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, index string, alias *models.IndexAlias) fwdiags.Diagnostics {
	aliasBytes, err := json.Marshal(alias)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	typedClient := apiClient.GetESClient()
	_, err = typedClient.Indices.PutAlias(index, alias.Name).Raw(bytes.NewReader(aliasBytes)).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func GetAlias(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, aliasName string) (map[string]types.IndexAliases, fwdiags.Diagnostics) {
	typedClient := apiClient.GetESClient()
	res, err := typedClient.Indices.GetAlias().Name(aliasName).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil, nil
		}
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	return res, nil
}

// AliasAction represents a single action in an atomic alias update operation
type AliasAction struct {
	Type          string
	Index         string
	Alias         string
	IsWriteIndex  bool
	Filter        map[string]any
	IndexRouting  string
	IsHidden      bool
	Routing       string
	SearchRouting string
}

// UpdateAliasesAtomic performs atomic alias updates using multiple actions
func UpdateAliasesAtomic(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, actions []AliasAction) fwdiags.Diagnostics {
	aliasActions := make([]map[string]any, 0, len(actions))

	for _, action := range actions {
		switch action.Type {
		case "remove":
			aliasActions = append(aliasActions, map[string]any{
				"remove": map[string]any{
					"index": action.Index,
					"alias": action.Alias,
				},
			})
		case "add":
			addDetails := map[string]any{
				"index": action.Index,
				"alias": action.Alias,
			}

			if action.IsWriteIndex {
				addDetails["is_write_index"] = true
			}
			if action.Filter != nil {
				addDetails["filter"] = action.Filter
			}
			if action.IndexRouting != "" {
				addDetails["index_routing"] = action.IndexRouting
			}
			if action.SearchRouting != "" {
				addDetails["search_routing"] = action.SearchRouting
			}
			if action.Routing != "" {
				addDetails["routing"] = action.Routing
			}
			if action.IsHidden {
				addDetails["is_hidden"] = action.IsHidden
			}

			aliasActions = append(aliasActions, map[string]any{
				"add": addDetails,
			})
		}
	}

	requestBody := map[string]any{
		"actions": aliasActions,
	}

	aliasBytes, err := json.Marshal(requestBody)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	typedClient := apiClient.GetESClient()
	_, err = typedClient.Indices.UpdateAliases().Raw(bytes.NewReader(aliasBytes)).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}
