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
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func readFollowerIndex(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	resourceID string,
	state Model,
) (Model, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	follower, getDiags := elasticsearch.GetFollowerIndex(ctx, client, resourceID)
	diags.Append(getDiags...)
	if diags.HasError() {
		return state, false, diags
	}

	if follower == nil {
		tflog.Warn(ctx, fmt.Sprintf("CCR follower index %q not found, removing from state", resourceID))
		return state, false, diags
	}

	model := mapFollowerIndexToModel(follower, state)
	model.Name = types.StringValue(resourceID)

	id, idDiags := client.ID(ctx, resourceID)
	diags.Append(idDiags...)
	if diags.HasError() {
		return state, false, diags
	}
	model.ID = types.StringValue(id.String())

	return model, true, diags
}
