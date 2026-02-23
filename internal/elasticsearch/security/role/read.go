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

package role

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *roleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data Data
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readData, diags := r.read(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if readData == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(req.State.Set(ctx, readData)...)
}

func (r *roleResource) read(ctx context.Context, data Data) (*Data, diag.Diagnostics) {
	compID, diags := clients.CompositeIDFromStrFw(data.ID.ValueString())
	if diags.HasError() {
		return nil, diags
	}
	roleID := compID.ResourceID

	client, clientDiags := clients.MaybeNewAPIClientFromFrameworkResource(ctx, data.ElasticsearchConnection, r.client)
	diags.Append(clientDiags...)
	if diags.HasError() {
		return nil, diags
	}

	role, sdkDiags := elasticsearch.GetRole(ctx, client, roleID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return nil, diags
	}

	if role == nil {
		tflog.Warn(ctx, fmt.Sprintf(`Role "%s" not found`, roleID))
		return nil, diags
	}

	// Convert from API model
	diags.Append(data.fromAPIModel(ctx, role)...)
	if diags.HasError() {
		return nil, diags
	}

	// Set the name to the roleID we extracted to ensure consistency
	data.Name = types.StringValue(roleID)

	return &data, diags
}
