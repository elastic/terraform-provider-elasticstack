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

package integration

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *integrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var stateModel integrationModel

	diags := req.State.Get(ctx, &stateModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GetFleetClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	name := stateModel.Name.ValueString()
	version := stateModel.Version.ValueString()
	force := stateModel.Force.ValueBool()
	skipDestroy := stateModel.SkipDestroy.ValueBool()
	if skipDestroy {
		tflog.Debug(ctx, "Skipping uninstall of integration package", map[string]any{"name": name, "version": version})
		return
	}

	// If space_id is set, use space-aware uninstallation
	var spaceID string
	if typeutils.IsKnown(stateModel.SpaceID) {
		spaceID = stateModel.SpaceID.ValueString()
	}

	diags = fleet.Uninstall(ctx, client, name, version, spaceID, force)
	resp.Diagnostics.Append(diags...)
}
