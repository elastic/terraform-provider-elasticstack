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

package slo

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// sloPlannedEnabledExplicitlySet reports whether the practitioner set a known
// `enabled` value in the plan (as opposed to omitting it or leaving it unknown).
func sloPlannedEnabledExplicitlySet(plan types.Bool) bool {
	return !plan.IsNull() && typeutils.IsKnown(plan)
}

// reconcileSloEnabledAfterWrite reconciles the SLO enabled flag using dedicated
// Kibana enable/disable APIs when `enabled` is set in the Terraform
// configuration and differs from the value returned by the most recent read.
// When `enabled` is omitted (null in plan) or the planned value is still unknown
// (e.g. mid-plan for a new dependency), the function is a no-op and does not
// call the enable or disable APIs.
func (r *Resource) reconcileSloEnabledAfterWrite(
	ctx context.Context,
	apiClient *clients.KibanaScopedClient,
	oapi *kibanaoapi.Client,
	spaceID, sloID string,
	planEnabled types.Bool,
	m *tfModel,
	diags *diag.Diagnostics,
) {
	if !sloPlannedEnabledExplicitlySet(planEnabled) {
		return
	}
	if !typeutils.IsKnown(m.Enabled) {
		return
	}
	if planEnabled.ValueBool() == m.Enabled.ValueBool() {
		return
	}
	if planEnabled.ValueBool() {
		diags.Append(kibanaoapi.EnableSlo(ctx, oapi, spaceID, sloID)...)
	} else {
		diags.Append(kibanaoapi.DisableSlo(ctx, oapi, spaceID, sloID)...)
	}
	if diags.HasError() {
		return
	}
	exists, readDiags := r.readSloFromAPI(ctx, apiClient, m)
	diags.Append(readDiags...)
	if diags.HasError() {
		return
	}
	if !exists {
		diags.AddError("SLO not found", "SLO could not be read after changing enabled state")
	}
}
