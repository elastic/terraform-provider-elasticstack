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
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data Data
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	compID, diags := clients.CompositeIDFromStrFw(data.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	usernameID := compID.ResourceID

	client, diags := clients.MaybeNewAPIClientFromFrameworkResource(ctx, data.ElasticsearchConnection, r.client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, sdkDiags := elasticsearch.GetUser(ctx, client, usernameID)
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if user == nil {
		tflog.Warn(ctx, fmt.Sprintf(`User "%s" not found, removing from state`, compID.ResourceID))
		resp.State.RemoveResource(ctx)
		return
	}

	// Set the fields
	data.Username = types.StringValue(usernameID)
	data.Email = types.StringValue(user.Email)
	data.FullName = types.StringValue(user.FullName)
	data.Enabled = types.BoolValue(user.Enabled)

	// Handle metadata
	if len(user.Metadata) > 0 {
		metadata, err := json.Marshal(user.Metadata)
		if err != nil {
			resp.Diagnostics.AddError("Failed to marshal metadata", err.Error())
			return
		}
		data.Metadata = jsontypes.NewNormalizedValue(string(metadata))
	} else {
		data.Metadata = jsontypes.NewNormalizedNull()
	}

	// Convert roles slice to set
	rolesSet, diags := types.SetValueFrom(ctx, types.StringType, user.Roles)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Roles = rolesSet

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
