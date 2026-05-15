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
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// writeUser handles both Create (req.Prior == nil) and Update (req.Prior !=
// nil); the security user PUT API is idempotent and the prior is only used to
// detect password rotation and pick the correct password source.
func writeUser(ctx context.Context, client *clients.ElasticsearchScopedClient, req entitycore.WriteRequest[Data]) (entitycore.WriteResult[Data], diag.Diagnostics) {
	var diags diag.Diagnostics
	plan := req.Plan
	hasState := req.Prior != nil
	var stateData Data
	if hasState {
		stateData = *req.Prior
	}

	usernameID := plan.GetResourceID().ValueString()

	id, sdkDiags := client.ID(ctx, usernameID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return entitycore.WriteResult[Data]{Model: plan}, diags
	}

	user := &estypes.User{
		Username: usernameID,
		Enabled:  plan.Enabled.ValueBool(),
	}

	var passwordWoFromConfig types.String
	diags.Append(req.Config.GetAttribute(ctx, path.Root("password_wo"), &passwordWoFromConfig)...)
	if diags.HasError() {
		return entitycore.WriteResult[Data]{Model: plan}, diags
	}

	var password, passwordHash *string
	switch {
	case typeutils.IsKnown(passwordWoFromConfig) && (!hasState || !plan.PasswordWoVersion.Equal(stateData.PasswordWoVersion)):
		pw := passwordWoFromConfig.ValueString()
		password = &pw
	case typeutils.IsKnown(plan.Password) && (!hasState || !plan.Password.Equal(stateData.Password)):
		pw := plan.Password.ValueString()
		password = &pw
	case typeutils.IsKnown(plan.PasswordHash) && (!hasState || !plan.PasswordHash.Equal(stateData.PasswordHash)):
		ph := plan.PasswordHash.ValueString()
		passwordHash = &ph
	}

	if typeutils.IsKnown(plan.Email) {
		email := plan.Email.ValueString()
		user.Email = &email
	}
	if typeutils.IsKnown(plan.FullName) {
		fullName := plan.FullName.ValueString()
		user.FullName = &fullName
	}

	roles := make([]string, 0, len(plan.Roles.Elements()))
	diags.Append(plan.Roles.ElementsAs(ctx, &roles, false)...)
	if diags.HasError() {
		return entitycore.WriteResult[Data]{Model: plan}, diags
	}
	user.Roles = roles

	if !plan.Metadata.IsNull() && !plan.Metadata.IsUnknown() {
		var metadataMap map[string]any
		err := json.Unmarshal([]byte(plan.Metadata.ValueString()), &metadataMap)
		if err != nil {
			diags.AddError("Failed to decode metadata", err.Error())
			return entitycore.WriteResult[Data]{Model: plan}, diags
		}
		metadata := make(estypes.Metadata, len(metadataMap))
		for k, v := range metadataMap {
			b, err := json.Marshal(v)
			if err != nil {
				diags.AddError("Failed to marshal metadata", err.Error())
				return entitycore.WriteResult[Data]{Model: plan}, diags
			}
			metadata[k] = b
		}
		user.Metadata = metadata
	}

	diags.Append(elasticsearch.PutUser(ctx, client, user, password, passwordHash)...)
	if diags.HasError() {
		return entitycore.WriteResult[Data]{Model: plan}, diags
	}

	plan.ID = types.StringValue(id.String())

	return entitycore.WriteResult[Data]{Model: plan}, diags
}
