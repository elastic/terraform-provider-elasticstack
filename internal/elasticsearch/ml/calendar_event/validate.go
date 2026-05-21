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

package calendar_event

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func validateEventEndAfterStart(start, end timetypes.RFC3339) diag.Diagnostics {
	var diags diag.Diagnostics

	if start.IsNull() || start.IsUnknown() || end.IsNull() || end.IsUnknown() {
		return diags
	}

	st, d := start.ValueRFC3339Time()
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	et, d := end.ValueRFC3339Time()
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	if !et.After(st) {
		diags.AddAttributeError(
			path.Root("end_time"),
			"Invalid event time range",
			"end_time must be after start_time.",
		)
	}
	return diags
}

func (r *calendarEventResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config CalendarEventTFModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(validateEventEndAfterStart(config.StartTime, config.EndTime)...)
}
