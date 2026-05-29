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

package connector

import (
	"context"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *contentConnectorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importID := strings.TrimSpace(req.ID)
	if importID == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Import ID must be a non-empty connector_id or composite <cluster_uuid>/<connector_id>.")
		return
	}

	var connectorID string
	if strings.Contains(importID, "/") {
		compID, diags := clients.CompositeIDFromStr(importID)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		connectorID = compID.ResourceID
	} else {
		connectorID = importID
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("connector_id"), connectorID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), fwtypes.StringValue(importID))...)
}
