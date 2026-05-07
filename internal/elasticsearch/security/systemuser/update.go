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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func writeSystemUser(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, data Data) (Data, diag.Diagnostics) {
	var diags diag.Diagnostics
	usernameID := resourceID

	id, sdkDiags := client.ID(ctx, usernameID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		var zero Data
		return zero, diags
	}

	user, sdkDiags := elasticsearch.GetUser(ctx, client, usernameID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		var zero Data
		return zero, diags
	}
	if user == nil || !isSystemUser(user) {
		diags.AddError("Not Found", fmt.Sprintf(`System user "%s" not found`, usernameID))
		var zero Data
		return zero, diags
	}

	var password, passwordHash *string
	if typeutils.IsKnown(data.Password) {
		password = data.Password.ValueStringPointer()
	}
	if typeutils.IsKnown(data.PasswordHash) {
		passwordHash = data.PasswordHash.ValueStringPointer()
	}
	if password != nil || passwordHash != nil {
		diags.Append(elasticsearch.ChangeUserPassword(ctx, client, usernameID, password, passwordHash)...)
		if diags.HasError() {
			var zero Data
			return zero, diags
		}
	}

	if typeutils.IsKnown(data.Enabled) && !data.Enabled.IsNull() && data.Enabled.ValueBool() != user.Enabled {
		if data.Enabled.ValueBool() {
			diags.Append(elasticsearch.EnableUser(ctx, client, usernameID)...)
		} else {
			diags.Append(elasticsearch.DisableUser(ctx, client, usernameID)...)
		}
		if diags.HasError() {
			var zero Data
			return zero, diags
		}
	}

	data.ID = types.StringValue(id.String())
	return data, diags
}
