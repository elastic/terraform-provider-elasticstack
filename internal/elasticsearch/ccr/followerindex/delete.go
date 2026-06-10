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

package followerindex

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func deleteFollowerIndex(ctx context.Context, client *clients.ElasticsearchScopedClient, indexName string, state Model) diag.Diagnostics {
	return executeDeleteOperations(ctx, client, indexName, planDeleteOperations(state))
}

func executeDeleteOperations(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	indexName string,
	ops []apiOperation,
) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, op := range ops {
		switch op {
		case opPause:
			diags.Append(elasticsearch.PauseFollowerIndex(ctx, client, indexName)...)
		case opClose:
			diags.Append(elasticsearch.CloseIndex(ctx, client, indexName)...)
		case opUnfollow:
			diags.Append(elasticsearch.UnfollowIndex(ctx, client, indexName)...)
		case opDeleteIndex:
			diags.Append(elasticsearch.DeleteIndex(ctx, client, indexName)...)
		case opOpenIndex:
			diags.Append(elasticsearch.OpenIndex(ctx, client, indexName)...)
		default:
			diags.AddError("Internal error", "Unexpected delete operation: "+op.String())
		}
		if diags.HasError() {
			return diags
		}
	}

	return diags
}
