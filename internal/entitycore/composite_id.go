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

// CompositeIDForWrite returns the composite resource ID to persist after a write.
// On Create (req.Prior == nil) it computes the ID from the live cluster UUID via
// client.ID. On Update it returns the prior state's ID unchanged so that
// UseStateForUnknown on id is never violated when the connected cluster's UUID
// differs from the prefix stored in state.
func CompositeIDForWrite[T ElasticsearchResourceModel](
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	req WriteRequest[T],
) (types.String, diag.Diagnostics) {
	if req.Prior != nil {
		return (*req.Prior).GetID(), nil
	}

	id, idDiags := client.ID(ctx, req.WriteID)
	if idDiags.HasError() {
		return types.StringNull(), idDiags
	}
	return types.StringValue(id.String()), nil
}
