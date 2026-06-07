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

package securityuser

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func readUser(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, state Data) (Data, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	user, userDiags := elasticsearch.GetUser(ctx, client, resourceID)
	diags.Append(userDiags...)
	if diags.HasError() {
		return state, false, diags
	}

	if user == nil {
		tflog.Warn(ctx, fmt.Sprintf(`User "%s" not found, removing from state`, resourceID))
		return state, false, nil
	}

	// Set the fields
	state.Username = types.StringValue(resourceID)
	if user.Email != nil {
		state.Email = types.StringValue(*user.Email)
	} else {
		state.Email = types.StringValue("")
	}
	if user.FullName != nil {
		state.FullName = types.StringValue(*user.FullName)
	} else {
		state.FullName = types.StringValue("")
	}
	state.Enabled = types.BoolValue(user.Enabled)

	// Handle metadata. The Elasticsearch API treats "no metadata set" and an
	// empty metadata object as equivalent, so when the API returns empty we
	// preserve an incoming "{}" state value instead of overwriting it with
	// null. Without this, configurations using `metadata = jsonencode({})`
	// trip a "Provider produced inconsistent result after apply" error after
	// the entitycore read-after-write step.
	if len(user.Metadata) > 0 {
		metadata, err := json.Marshal(user.Metadata)
		if err != nil {
			diags.AddError("Failed to marshal metadata", err.Error())
			return state, false, diags
		}
		state.Metadata = jsontypes.NewNormalizedValue(string(metadata))
	} else if state.Metadata.IsNull() || state.Metadata.IsUnknown() || !typeutils.IsEmptyJSONObject(state.Metadata.ValueString()) {
		state.Metadata = jsontypes.NewNormalizedNull()
	}

	// Convert roles slice to set
	rolesSet, roleDiags := types.SetValueFrom(ctx, types.StringType, user.Roles)
	diags.Append(roleDiags...)
	if diags.HasError() {
		return state, false, diags
	}
	state.Roles = rolesSet

	return state, true, diags
}
