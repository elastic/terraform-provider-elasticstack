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

	estypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(r.update(ctx, req.Plan, req.Config, &resp.State)...)
}

func (r *userResource) update(ctx context.Context, plan tfsdk.Plan, config tfsdk.Config, state *tfsdk.State) diag.Diagnostics {
	var planData Data
	var diags diag.Diagnostics
	diags.Append(plan.Get(ctx, &planData)...)
	if diags.HasError() {
		return diags
	}

	writeID := planData.GetResourceID()
	if !typeutils.IsKnown(writeID) || writeID.ValueString() == "" {
		diags.AddError(
			"Invalid resource identifier",
			"The resource write identity from configuration is unknown or empty; cannot create or update.",
		)
		return diags
	}
	usernameID := writeID.ValueString()

	hasState := false
	var stateData Data
	if state != nil && !state.Raw.IsNull() {
		hasState = true
		diags.Append(state.Get(ctx, &stateData)...)
		if diags.HasError() {
			return diags
		}
	}

	client, connDiags := r.Client().GetElasticsearchClient(ctx, planData.ElasticsearchConnection)
	diags.Append(connDiags...)
	if diags.HasError() {
		return diags
	}

	id, sdkDiags := client.ID(ctx, usernameID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return diags
	}

	user := &estypes.User{
		Username: usernameID,
		Enabled:  planData.Enabled.ValueBool(),
	}

	// Handle password fields - only set password if it's in the plan AND (it's a create OR it has changed from state)
	// Priority: password_wo > password > password_hash
	// Read password_wo from config as per Terraform write-only attribute guidelines
	var passwordWoFromConfig types.String
	diags.Append(config.GetAttribute(ctx, path.Root("password_wo"), &passwordWoFromConfig)...)
	if diags.HasError() {
		return diags
	}

	var password, passwordHash *string
	switch {
	case typeutils.IsKnown(passwordWoFromConfig) && (!hasState || !planData.PasswordWoVersion.Equal(stateData.PasswordWoVersion)):
		// Use write-only password - changes triggered by version change
		pw := passwordWoFromConfig.ValueString()
		password = &pw
	case typeutils.IsKnown(planData.Password) && (!hasState || !planData.Password.Equal(stateData.Password)):
		pw := planData.Password.ValueString()
		password = &pw
	case typeutils.IsKnown(planData.PasswordHash) && (!hasState || !planData.PasswordHash.Equal(stateData.PasswordHash)):
		ph := planData.PasswordHash.ValueString()
		passwordHash = &ph
	}

	if typeutils.IsKnown(planData.Email) {
		email := planData.Email.ValueString()
		user.Email = &email
	}
	if typeutils.IsKnown(planData.FullName) {
		fullName := planData.FullName.ValueString()
		user.FullName = &fullName
	}

	roles := make([]string, 0, len(planData.Roles.Elements()))
	diags.Append(planData.Roles.ElementsAs(ctx, &roles, false)...)
	if diags.HasError() {
		return diags
	}
	user.Roles = roles

	if !planData.Metadata.IsNull() && !planData.Metadata.IsUnknown() {
		var metadataMap map[string]any
		err := json.Unmarshal([]byte(planData.Metadata.ValueString()), &metadataMap)
		if err != nil {
			diags.AddError("Failed to decode metadata", err.Error())
			return diags
		}
		metadata := make(estypes.Metadata, len(metadataMap))
		for k, v := range metadataMap {
			b, err := json.Marshal(v)
			if err != nil {
				diags.AddError("Failed to marshal metadata", err.Error())
				return diags
			}
			metadata[k] = b
		}
		user.Metadata = metadata
	}

	diags.Append(elasticsearch.PutUser(ctx, client, user, password, passwordHash)...)
	if diags.HasError() {
		return diags
	}

	// Read the user back to get computed fields like metadata
	readUser, sdkDiags := elasticsearch.GetUser(ctx, client, usernameID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return diags
	}

	if readUser == nil {
		diags.AddError("Failed to read user after update", "The user was not found after the update operation.")
		return diags
	}

	planData.ID = types.StringValue(id.String())

	// Set computed fields from the API response
	if len(readUser.Metadata) > 0 {
		metadata, err := json.Marshal(readUser.Metadata)
		if err != nil {
			diags.AddError("Failed to marshal metadata", err.Error())
			return diags
		}
		planData.Metadata = jsontypes.NewNormalizedValue(string(metadata))
	} else {
		planData.Metadata = jsontypes.NewNormalizedNull()
	}

	diags.Append(state.Set(ctx, &planData)...)
	return diags
}
