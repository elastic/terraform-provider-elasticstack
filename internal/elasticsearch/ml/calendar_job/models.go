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

package calendar_job

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TFModel is the Terraform state for a single calendar–job (or calendar–group) assignment.
type TFModel struct {
	ID                      types.String `tfsdk:"id"`
	ElasticsearchConnection types.List   `tfsdk:"elasticsearch_connection"`
	CalendarID              types.String `tfsdk:"calendar_id"`
	JobID                   types.String `tfsdk:"job_id"`
}

// GetID implements entitycore.ElasticsearchResourceModel.
func (m TFModel) GetID() types.String { return m.ID }

// GetResourceID implements entitycore.ElasticsearchResourceModel.
// It returns "<calendar_id>/<job_id>" for the composite Elasticsearch resource ID segment
// (the part after the cluster UUID). The second segment is the same string sent to
// Elasticsearch as `job_id` (job identifier or job group name).
func (m TFModel) GetResourceID() types.String {
	if !typeutils.IsKnown(m.CalendarID) || !typeutils.IsKnown(m.JobID) {
		return types.StringUnknown()
	}
	if m.CalendarID.IsNull() || m.JobID.IsNull() {
		return types.StringNull()
	}
	c := m.CalendarID.ValueString()
	j := m.JobID.ValueString()
	if c == "" || j == "" {
		return types.StringNull()
	}
	return types.StringValue(c + "/" + j)
}

// GetElasticsearchConnection implements entitycore.ElasticsearchResourceModel.
func (m TFModel) GetElasticsearchConnection() types.List { return m.ElasticsearchConnection }
