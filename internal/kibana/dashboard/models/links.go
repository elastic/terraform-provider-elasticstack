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

package models

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	linksAttrTitle        = "title"
	linksAttrDescription  = "description"
	linksAttrHideTitle    = "hide_title"
	linksAttrHideBorder   = "hide_border"
	linksAttrLayout       = "layout"
	linksAttrLinks        = "links"
	linksAttrRefID        = "ref_id"
	linksAttrType         = "type"
	linksAttrDestination  = "destination"
	linksAttrLabel        = "label"
	linksAttrOpenInNewTab = "open_in_new_tab"
	linksAttrUseFilters   = "use_filters"
	linksAttrUseTimeRange = "use_time_range"
	linksAttrEncodeURL    = "encode_url"
)

type LinksPanelConfigModel struct {
	Title       types.String                `tfsdk:"title"`
	Description types.String                `tfsdk:"description"`
	HideTitle   types.Bool                  `tfsdk:"hide_title"`
	HideBorder  types.Bool                  `tfsdk:"hide_border"`
	ByValue     *LinksPanelByValueModel     `tfsdk:"by_value"`
	ByReference *LinksPanelByReferenceModel `tfsdk:"by_reference"`
}

type LinksPanelByValueModel struct {
	Layout      types.String    `tfsdk:"layout"`
	Title       types.String    `tfsdk:"title"`
	Description types.String    `tfsdk:"description"`
	HideTitle   types.Bool      `tfsdk:"hide_title"`
	HideBorder  types.Bool      `tfsdk:"hide_border"`
	Links       []LinkItemModel `tfsdk:"links"`
}

type LinksPanelByReferenceModel struct {
	RefID       types.String `tfsdk:"ref_id"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	HideTitle   types.Bool   `tfsdk:"hide_title"`
	HideBorder  types.Bool   `tfsdk:"hide_border"`
}

type LinkItemModel struct {
	Type         types.String `tfsdk:"type"`
	Destination  types.String `tfsdk:"destination"`
	Label        types.String `tfsdk:"label"`
	OpenInNewTab types.Bool   `tfsdk:"open_in_new_tab"`
	UseFilters   types.Bool   `tfsdk:"use_filters"`
	UseTimeRange types.Bool   `tfsdk:"use_time_range"`
	EncodeURL    types.Bool   `tfsdk:"encode_url"`
}

func (m LinksPanelConfigModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		linksAttrTitle:       types.StringType,
		linksAttrDescription: types.StringType,
		linksAttrHideTitle:   types.BoolType,
		linksAttrHideBorder:  types.BoolType,
		"by_value": types.ObjectType{
			AttrTypes: LinksPanelByValueModel{}.AttrTypes(),
		},
		"by_reference": types.ObjectType{
			AttrTypes: LinksPanelByReferenceModel{}.AttrTypes(),
		},
	}
}

func (m LinksPanelByValueModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		linksAttrLayout:      types.StringType,
		linksAttrTitle:       types.StringType,
		linksAttrDescription: types.StringType,
		linksAttrHideTitle:   types.BoolType,
		linksAttrHideBorder:  types.BoolType,
		linksAttrLinks: types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: LinkItemModel{}.AttrTypes(),
			},
		},
	}
}

func (m LinksPanelByReferenceModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		linksAttrRefID:       types.StringType,
		linksAttrTitle:       types.StringType,
		linksAttrDescription: types.StringType,
		linksAttrHideTitle:   types.BoolType,
		linksAttrHideBorder:  types.BoolType,
	}
}

func (m LinkItemModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		linksAttrType:         types.StringType,
		linksAttrDestination:  types.StringType,
		linksAttrLabel:        types.StringType,
		linksAttrOpenInNewTab: types.BoolType,
		linksAttrUseFilters:   types.BoolType,
		linksAttrUseTimeRange: types.BoolType,
		linksAttrEncodeURL:    types.BoolType,
	}
}
