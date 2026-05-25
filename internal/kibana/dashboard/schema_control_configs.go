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

package dashboard

import (
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework-validators/float32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// pinnedPanelControlConfigNames lists typed control blocks allowed on a pinned panel entry (dashboard control bar).
var pinnedPanelControlConfigNames = []string{
	controlBlockTimeSlider,
	controlBlockESQL,
	controlBlockOptionsList,
	controlBlockRangeSlider,
}

func pinnedPlacementPreface() string {
	n := strings.TrimSpace(pinnedPanelControlNote)
	if n == "" {
		return ""
	}
	return n + "\n\n"
}

func timeSliderControlConfigInnerAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		attrStartPercentageOfTimeRange: schema.Float32Attribute{
			MarkdownDescription: "Start of the visible time window as a fraction of the dashboard global range (0.0–1.0). " +
				"Float32 in state matches the Kibana API and avoids refresh drift.",
			Optional: true,
			Validators: []validator.Float32{
				float32validator.Between(0.0, 1.0),
			},
		},
		attrEndPercentageOfTimeRange: schema.Float32Attribute{
			MarkdownDescription: "End of the visible time window as a fraction of the dashboard global range (0.0–1.0). " +
				"Float32 in state matches the Kibana API and avoids refresh drift.",
			Optional: true,
			Validators: []validator.Float32{
				float32validator.Between(0.0, 1.0),
			},
		},
		attrIsAnchored: schema.BoolAttribute{
			MarkdownDescription: "Whether the start of the time window is anchored (fixed), so only the end slides.",
			Optional:            true,
		},
	}
}

func pinnedTimeSliderControlConfigSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: panelkit.PanelConfigDescription(
			pinnedPlacementPreface()+"Configuration for a time slider control. Controls the visible time window within the dashboard's global time range.",
			controlBlockTimeSlider,
			pinnedPanelControlConfigNames,
		),
		Optional:   true,
		Attributes: timeSliderControlConfigInnerAttributes(),
	}
}

func esqlControlConfigInnerAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		attrSelectedOptions: schema.ListAttribute{
			MarkdownDescription: "List of currently selected option values for the control.",
			Required:            true,
			ElementType:         types.StringType,
		},
		attrVariableName: schema.StringAttribute{
			MarkdownDescription: "The ES|QL variable name that this control binds to.",
			Required:            true,
		},
		attrVariableType: schema.StringAttribute{
			MarkdownDescription: "The type of ES|QL variable. Allowed values: `fields`, `values`, `functions`, `time_literal`, `multi_values`.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("fields", "values", "functions", "time_literal", "multi_values"),
			},
		},
		attrESQLQuery: schema.StringAttribute{
			MarkdownDescription: "The ES|QL query used to populate the control's options.",
			Required:            true,
		},
		attrControlType: schema.StringAttribute{
			MarkdownDescription: "The control type. Allowed values: `STATIC_VALUES`, `VALUES_FROM_QUERY`.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("STATIC_VALUES", "VALUES_FROM_QUERY"),
			},
		},
		attrTitle: schema.StringAttribute{
			MarkdownDescription: "A human-readable title displayed above the control widget.",
			Optional:            true,
		},
		attrSingleSelect: schema.BoolAttribute{
			MarkdownDescription: "When true, restricts the control to single-value selection.",
			Optional:            true,
		},
		attrAvailableOptions: schema.ListAttribute{
			MarkdownDescription: "Pre-populated list of available options shown before the query executes.",
			Optional:            true,
			ElementType:         types.StringType,
		},
		attrDisplaySettings: schema.SingleNestedAttribute{
			MarkdownDescription: "Display configuration for the control widget.",
			Optional:            true,
			Attributes: map[string]schema.Attribute{
				attrPlaceholder: schema.StringAttribute{
					MarkdownDescription: "Placeholder text shown when no option is selected.",
					Optional:            true,
				},
				attrHideActionBar: schema.BoolAttribute{
					MarkdownDescription: "Whether to hide the action bar on the control.",
					Optional:            true,
				},
				attrHideExclude: schema.BoolAttribute{
					MarkdownDescription: "Whether to hide the exclude option.",
					Optional:            true,
				},
				attrHideExists: schema.BoolAttribute{
					MarkdownDescription: "Whether to hide the exists filter option.",
					Optional:            true,
				},
				attrHideSort: schema.BoolAttribute{
					MarkdownDescription: "Whether to hide the sort option.",
					Optional:            true,
				},
			},
		},
	}
}

func pinnedEsqlControlConfigSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: panelkit.PanelConfigDescription(
			pinnedPlacementPreface()+"Configuration for an ES|QL control. Binds ES|QL variables for the dashboard.",
			controlBlockESQL,
			pinnedPanelControlConfigNames,
		),
		Optional:   true,
		Attributes: esqlControlConfigInnerAttributes(),
	}
}

func optionsListControlConfigInnerAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		attrDataViewID: schema.StringAttribute{
			MarkdownDescription: "The ID of the data view that the control is tied to.",
			Required:            true,
		},
		attrFieldName: schema.StringAttribute{
			MarkdownDescription: "The name of the field in the data view that the control is tied to.",
			Required:            true,
		},
		attrTitle: schema.StringAttribute{
			MarkdownDescription: "Human-readable label displayed above the control.",
			Optional:            true,
		},
		attrUseGlobalFilters: schema.BoolAttribute{
			MarkdownDescription: "Whether the control applies the dashboard's global filters to its own query.",
			Optional:            true,
		},
		attrIgnoreValidations: schema.BoolAttribute{
			MarkdownDescription: "Whether the control skips field-level validation against the data view.",
			Optional:            true,
		},
		attrSingleSelect: schema.BoolAttribute{
			MarkdownDescription: "When true, only one option may be selected at a time.",
			Optional:            true,
		},
		attrExclude: schema.BoolAttribute{
			MarkdownDescription: "When true, selected options are used as an exclusion filter rather than an inclusion filter.",
			Optional:            true,
		},
		attrExistsSelected: schema.BoolAttribute{
			MarkdownDescription: "When true, the control filters for documents where the field exists.",
			Optional:            true,
		},
		attrRunPastTimeout: schema.BoolAttribute{
			MarkdownDescription: "When true, the control continues to show results even when the underlying query times out.",
			Optional:            true,
		},
		attrSearchTechnique: schema.StringAttribute{
			MarkdownDescription: "The technique used to match suggestions. Must be one of `prefix`, `wildcard`, or `exact` when set.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("prefix", "wildcard", "exact"),
			},
		},
		attrSelectedOptions: schema.ListAttribute{
			MarkdownDescription: "The initially or persistently selected option values. All values are represented as strings.",
			Optional:            true,
			ElementType:         types.StringType,
		},
		attrDisplaySettings: schema.SingleNestedAttribute{
			MarkdownDescription: "Display preferences for the control widget.",
			Optional:            true,
			Attributes: map[string]schema.Attribute{
				attrPlaceholder: schema.StringAttribute{
					MarkdownDescription: "Placeholder text shown when no option is selected.",
					Optional:            true,
				},
				attrHideActionBar: schema.BoolAttribute{
					MarkdownDescription: "When true, hides the action bar on the control.",
					Optional:            true,
				},
				attrHideExclude: schema.BoolAttribute{
					MarkdownDescription: "When true, hides the exclude toggle.",
					Optional:            true,
				},
				attrHideExists: schema.BoolAttribute{
					MarkdownDescription: "When true, hides the exists filter option.",
					Optional:            true,
				},
				attrHideSort: schema.BoolAttribute{
					MarkdownDescription: "When true, hides the sort control.",
					Optional:            true,
				},
			},
		},
		attrSort: schema.SingleNestedAttribute{
			MarkdownDescription: "Default sort configuration for the suggestion list.",
			Optional:            true,
			Attributes: map[string]schema.Attribute{
				"by": schema.StringAttribute{
					MarkdownDescription: "The field or criterion to sort by. Must be one of `_count` or `_key`.",
					Required:            true,
					Validators: []validator.String{
						stringvalidator.OneOf("_count", "_key"),
					},
				},
				"direction": schema.StringAttribute{
					MarkdownDescription: "The sort direction. Must be one of `asc` or `desc`.",
					Required:            true,
					Validators: []validator.String{
						stringvalidator.OneOf("asc", "desc"),
					},
				},
			},
		},
	}
}

func pinnedOptionsListControlConfigSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: panelkit.PanelConfigDescription(
			pinnedPlacementPreface()+"Configuration for an options list control. Provides a dropdown or multi-select filter based on a field in a data view.",
			controlBlockOptionsList,
			pinnedPanelControlConfigNames,
		),
		Optional:   true,
		Attributes: optionsListControlConfigInnerAttributes(),
	}
}

func rangeSliderControlConfigInnerAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		attrTitle: schema.StringAttribute{
			MarkdownDescription: "A human-readable title for the control.",
			Optional:            true,
		},
		attrDataViewID: schema.StringAttribute{
			MarkdownDescription: "The ID of the data view that the control is tied to.",
			Required:            true,
		},
		attrFieldName: schema.StringAttribute{
			MarkdownDescription: "The name of the field in the data view that the control is tied to.",
			Required:            true,
		},
		attrUseGlobalFilters: schema.BoolAttribute{
			MarkdownDescription: "Whether the control respects dashboard-level filters.",
			Optional:            true,
		},
		attrIgnoreValidations: schema.BoolAttribute{
			MarkdownDescription: "Whether to suppress validation errors during intermediate states.",
			Optional:            true,
		},
		attrValue: schema.ListAttribute{
			MarkdownDescription: "Initial range as a list of exactly 2 strings: [min, max].",
			ElementType:         types.StringType,
			Optional:            true,
			Validators: []validator.List{
				listvalidator.SizeAtLeast(2),
				listvalidator.SizeAtMost(2),
			},
		},
		attrStep: schema.Float32Attribute{
			MarkdownDescription: "The step size for the range slider. Stored as float32 to match the Kibana API type and avoid refresh drift.",
			Optional:            true,
		},
	}
}

func pinnedRangeSliderControlConfigSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: panelkit.PanelConfigDescription(
			pinnedPlacementPreface()+"Configuration for a range slider control. Provides a min/max range filter tied to a data view field.",
			controlBlockRangeSlider,
			pinnedPanelControlConfigNames,
		),
		Optional:   true,
		Attributes: rangeSliderControlConfigInnerAttributes(),
	}
}

func pinnedPanelsNestedObject() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Validators: []validator.Object{
			pinnedPanelControlValidator{},
		},
		Attributes: map[string]schema.Attribute{
			attrPanelType: schema.StringAttribute{
				MarkdownDescription: strings.TrimSpace(pinnedPanelTypeDescription),
				Required:            true,
			},
			controlBlockTimeSlider:  pinnedTimeSliderControlConfigSchema(),
			controlBlockESQL:        pinnedEsqlControlConfigSchema(),
			controlBlockOptionsList: pinnedOptionsListControlConfigSchema(),
			controlBlockRangeSlider: pinnedRangeSliderControlConfigSchema(),
		},
	}
}
