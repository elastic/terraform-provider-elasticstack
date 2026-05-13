// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package models

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SloOverviewConfigModel struct {
	Single *SloOverviewSingleModel `tfsdk:"single"`
	Groups *SloOverviewGroupsModel `tfsdk:"groups"`
}

type SloOverviewSingleModel struct {
	SloID         types.String        `tfsdk:"slo_id"`
	SloInstanceID types.String        `tfsdk:"slo_instance_id"`
	RemoteName    types.String        `tfsdk:"remote_name"`
	Title         types.String        `tfsdk:"title"`
	Description   types.String        `tfsdk:"description"`
	HideTitle     types.Bool          `tfsdk:"hide_title"`
	HideBorder    types.Bool          `tfsdk:"hide_border"`
	Drilldowns    []SloDrilldownModel `tfsdk:"drilldowns"`
}

type SloOverviewGroupsModel struct {
	Title        types.String          `tfsdk:"title"`
	Description  types.String          `tfsdk:"description"`
	HideTitle    types.Bool            `tfsdk:"hide_title"`
	HideBorder   types.Bool            `tfsdk:"hide_border"`
	Drilldowns   []SloDrilldownModel   `tfsdk:"drilldowns"`
	GroupFilters *SloGroupFiltersModel `tfsdk:"group_filters"`
}

type SloGroupFiltersModel struct {
	GroupBy     types.String         `tfsdk:"group_by"`
	Groups      []types.String       `tfsdk:"groups"`
	KQLQuery    types.String         `tfsdk:"kql_query"`
	FiltersJSON jsontypes.Normalized `tfsdk:"filters_json"`
}

type SloDrilldownModel struct {
	URL          types.String `tfsdk:"url"`
	Label        types.String `tfsdk:"label"`
	EncodeURL    types.Bool   `tfsdk:"encode_url"`
	OpenInNewTab types.Bool   `tfsdk:"open_in_new_tab"`
}
