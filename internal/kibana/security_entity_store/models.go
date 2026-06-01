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

package security_entity_store

import (
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type historySnapshotModel struct {
	Frequency types.String `tfsdk:"frequency"`
}

type logExtractionModel struct {
	AdditionalIndexPatterns     types.List   `tfsdk:"additional_index_patterns"`
	ExcludedIndexPatterns       types.List   `tfsdk:"excluded_index_patterns"`
	Delay                       types.String `tfsdk:"delay"`
	DocsLimit                   types.Int64  `tfsdk:"docs_limit"`
	FieldHistoryLength          types.Int64  `tfsdk:"field_history_length"`
	Frequency                   types.String `tfsdk:"frequency"`
	LookbackPeriod              types.String `tfsdk:"lookback_period"`
	MaxLogsPerPage              types.Int64  `tfsdk:"max_logs_per_page"`
	MaxLogsPerWindow            types.Int64  `tfsdk:"max_logs_per_window"`
	MaxLogsPerWindowCapBehavior types.String `tfsdk:"max_logs_per_window_cap_behavior"`
	MaxTimeWindowSize           types.String `tfsdk:"max_time_window_size"`
}

type tfModel struct {
	ID                    types.String `tfsdk:"id"`
	KibanaConnection      types.List   `tfsdk:"kibana_connection"`
	SpaceID               types.String `tfsdk:"space_id"`
	EntityTypes           types.Set    `tfsdk:"entity_types"`
	AllowEntityTypeShrink types.Bool   `tfsdk:"allow_entity_type_shrink"`
	Started               types.Bool   `tfsdk:"started"`
	HistorySnapshot       types.Object `tfsdk:"history_snapshot"`
	LogExtraction         types.Object `tfsdk:"log_extraction"`
	StatusJSON            types.String `tfsdk:"status_json"`
}

var _ entitycore.KibanaResourceModel = tfModel{}
var _ entitycore.WithVersionRequirements = (*tfModel)(nil)

func (model tfModel) GetID() types.String             { return model.ID }
func (model tfModel) GetSpaceID() types.String        { return model.SpaceID }
func (model tfModel) GetKibanaConnection() types.List { return model.KibanaConnection }
func (model tfModel) GetResourceID() types.String     { return types.StringValue(resourceID) }

func (*tfModel) GetVersionRequirements() ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{{
		MinVersion:   *MinVersion,
		ErrorMessage: fmt.Sprintf("elasticstack_kibana_security_entity_store is supported only for Kibana v%s and above", MinVersion.String()),
	}}, nil
}
