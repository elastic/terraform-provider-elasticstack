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

package calendar

import (
	"context"

	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type CalendarTFModel struct {
	ID                      types.String `tfsdk:"id"`
	ElasticsearchConnection types.List   `tfsdk:"elasticsearch_connection"`
	CalendarID              types.String `tfsdk:"calendar_id"`
	Description             types.String `tfsdk:"description"`
	JobIDs                  types.Set    `tfsdk:"job_ids"`
}

type CalendarCreateAPIModel struct {
	JobIDs      []string `json:"job_ids,omitempty"`
	Description string   `json:"description,omitempty"`
}

type CalendarAPIModel struct {
	CalendarID  string   `json:"calendar_id"`
	Description string   `json:"description,omitempty"`
	JobIDs      []string `json:"job_ids"`
}

func (m *CalendarTFModel) toAPICreateModel(ctx context.Context) (*CalendarCreateAPIModel, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics

	apiModel := &CalendarCreateAPIModel{
		Description: m.Description.ValueString(),
	}

	if !m.JobIDs.IsNull() && !m.JobIDs.IsUnknown() {
		var jobIDs []string
		d := m.JobIDs.ElementsAs(ctx, &jobIDs, false)
		diags.Append(d...)
		apiModel.JobIDs = jobIDs
	}

	return apiModel, diags
}

func (m *CalendarTFModel) fromAPIModel(ctx context.Context, apiModel *CalendarAPIModel) fwdiags.Diagnostics {
	var diags fwdiags.Diagnostics

	m.CalendarID = types.StringValue(apiModel.CalendarID)

	if apiModel.Description != "" {
		m.Description = types.StringValue(apiModel.Description)
	} else {
		m.Description = types.StringNull()
	}

	if len(apiModel.JobIDs) == 0 && m.JobIDs.IsNull() {
		return diags
	}

	if len(apiModel.JobIDs) == 0 {
		emptySet, d := types.SetValueFrom(ctx, types.StringType, []string{})
		diags.Append(d...)
		m.JobIDs = emptySet
	} else {
		jobIDsSet, d := types.SetValueFrom(ctx, types.StringType, apiModel.JobIDs)
		diags.Append(d...)
		m.JobIDs = jobIDsSet
	}

	return diags
}
