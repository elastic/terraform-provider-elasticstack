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

package systemuser

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *systemUserResource) update(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State) diag.Diagnostics {
	var data Data
	var diags diag.Diagnostics
	diags.Append(plan.Get(ctx, &data)...)
	if diags.HasError() {
		return diags
	}

	usernameID := data.Username.ValueString()
	id, sdkDiags := r.client.ID(ctx, usernameID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return diags
	}

	client, diags := clients.MaybeNewAPIClientFromFrameworkResource(ctx, data.ElasticsearchConnection, r.client)
	diags.Append(diags...)
	if diags.HasError() {
		return diags
	}

	user, sdkDiags := elasticsearch.GetUser(ctx, client, usernameID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return diags
	}
	if user == nil || !user.IsSystemUser() {
		diags.AddError("", fmt.Sprintf(`System user "%s" not found`, usernameID))
		return diags
	}

	var userPassword models.UserPassword
	if typeutils.IsKnown(data.Password) && (user.Password == nil || data.Password.ValueString() != *user.Password) {
		userPassword.Password = data.Password.ValueStringPointer()
	}
	if typeutils.IsKnown(data.PasswordHash) && (user.PasswordHash == nil || data.PasswordHash.ValueString() != *user.PasswordHash) {
		userPassword.PasswordHash = data.PasswordHash.ValueStringPointer()
	}
	if userPassword.Password != nil || userPassword.PasswordHash != nil {
		diags.Append(elasticsearch.ChangeUserPassword(ctx, r.client, usernameID, &userPassword)...)
		if diags.HasError() {
			return diags
		}
	}

	if typeutils.IsKnown(data.Enabled) && !data.Enabled.IsNull() && data.Enabled.ValueBool() != user.Enabled {
		if data.Enabled.ValueBool() {
			diags.Append(elasticsearch.EnableUser(ctx, r.client, usernameID)...)
		} else {
			diags.Append(elasticsearch.DisableUser(ctx, r.client, usernameID)...)
		}
		if diags.HasError() {
			return diags
		}
	}

	data.ID = types.StringValue(id.String())
	diags.Append(state.Set(ctx, &data)...)
	return diags
}

func (r *systemUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	diags := r.update(ctx, req.Plan, &resp.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
