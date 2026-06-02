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
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dsModel struct {
	entitycore.KibanaConnectionField
	SpaceID           types.String         `tfsdk:"space_id"`
	IncludeComponents types.Bool           `tfsdk:"include_components"`
	Installed         types.Bool           `tfsdk:"installed"`
	OverallStatus     types.String         `tfsdk:"overall_status"`
	Engines           types.List           `tfsdk:"engines"`
	StatusJSON        jsontypes.Normalized `tfsdk:"status_json"`
}

type engineModel struct {
	Type               types.String `tfsdk:"type"`
	Status             types.String `tfsdk:"status"`
	IndexPattern       types.String `tfsdk:"index_pattern"`
	FieldHistoryLength types.Int64  `tfsdk:"field_history_length"`
	Delay              types.String `tfsdk:"delay"`
	Frequency          types.String `tfsdk:"frequency"`
	LookbackPeriod     types.String `tfsdk:"lookback_period"`
	Filter             types.String `tfsdk:"filter"`
	Timeout            types.String `tfsdk:"timeout"`
	TimestampField     types.String `tfsdk:"timestamp_field"`
	ErrorAction        types.String `tfsdk:"error_action"`
	ErrorMessage       types.String `tfsdk:"error_message"`
	Components         types.List   `tfsdk:"components"`
}

type engineComponentModel struct {
	ID        types.String `tfsdk:"id"`
	Installed types.Bool   `tfsdk:"installed"`
	Resource  types.String `tfsdk:"resource"`
	Health    types.String `tfsdk:"health"`
}

var _ entitycore.WithVersionRequirements = (*dsModel)(nil)

func (*dsModel) GetVersionRequirements(ctx context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{{
		MinVersion:   *MinVersion,
		ErrorMessage: fmt.Sprintf("elasticstack_kibana_security_entity_store_status is supported only for Kibana v%s and above", MinVersion.String()),
	}}, nil
}
