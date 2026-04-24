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

package customintegration

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *customIntegrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state customIntegrationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.SkipDestroy.ValueBool() {
		// User opted out of uninstall-on-destroy.
		return
	}

	name := state.PackageName.ValueString()
	version := state.PackageVersion.ValueString()
	if name == "" || version == "" {
		// State is incomplete — nothing reliable to uninstall. Let the
		// framework drop the resource from state.
		return
	}

	kibanaClient, diags := r.client.GetKibanaClient(ctx, state.KibanaConnection)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	fleetClient, err := kibanaClient.GetFleetClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to obtain Fleet client", err.Error())
		return
	}

	resp.Diagnostics.Append(
		fleet.Uninstall(ctx, fleetClient, name, version, state.SpaceID.ValueString(), false)...,
	)
}
