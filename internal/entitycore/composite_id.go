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
//
// On Create (req.Prior == nil) it computes the ID from the live cluster UUID via
// client.ID.
//
// On Update it never calls client.ID, so a stale cluster-UUID prefix in state
// (for example after a cluster recreation behind the same endpoint) is preserved
// and UseStateForUnknown on id is not violated. The write-identity segment is
// taken from req.WriteID:
//
//   - If the prior id's resource segment equals req.WriteID, the prior id is
//     returned unchanged. Safe for RequiresReplace identity keys (role name,
//     watch_id, datafeed_id) and for in-place updates that do not change name.
//   - If the prior id's resource segment differs from req.WriteID, the returned
//     id is <prior cluster UUID>/<WriteID>. This keeps Read/Delete targeting
//     the new name after an in-place name change (data_stream_lifecycle).
//   - If the prior id is not a valid composite id, the prior id is returned
//     unchanged rather than inventing a cluster UUID.
func CompositeIDForWrite[T ElasticsearchResourceModel](
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	req WriteRequest[T],
) (types.String, diag.Diagnostics) {
	if req.Prior == nil {
		id, idDiags := client.ID(ctx, req.WriteID)
		if idDiags.HasError() {
			return types.StringNull(), idDiags
		}
		return types.StringValue(id.String()), nil
	}

	priorID := (*req.Prior).GetID()
	parsed, parseDiags := clients.CompositeIDFromStr(priorID.ValueString())
	if parseDiags.HasError() {
		return priorID, nil
	}
	if parsed.ResourceID == req.WriteID {
		return priorID, nil
	}

	updated := &clients.CompositeID{ClusterID: parsed.ClusterID, ResourceID: req.WriteID}
	return types.StringValue(updated.String()), nil
}
