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

package entitycore

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ResolveDataSourceID resolves the composite resource ID from the Elasticsearch
// scoped client and assigns it to target. The composite ID is formed from the
// cluster UUID and resourceID. target is only modified when no error occurs.
//
// This is a shared helper for entitycore-based Elasticsearch data source
// readFunc implementations that eliminates repeated client.ID boilerplate.
func ResolveDataSourceID(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, target *types.String) diag.Diagnostics {
	id, diags := client.ID(ctx, resourceID)
	if !diags.HasError() {
		*target = types.StringValue(id.String())
	}
	return diags
}
