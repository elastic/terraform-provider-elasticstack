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

package lenscommon

import (
	"maps"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// ByValueChartNestedAttribute wraps the per-chart attribute map in the standard `vis_config.by_value`
// chart-block envelope used by every typed Lens converter. chartConfigName is the snake_case
// chart-config field name (for example "xy_chart_config") and is interpolated into the markdown.
func ByValueChartNestedAttribute(chartConfigName string, attrs map[string]schema.Attribute) schema.Attribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Typed Lens visualization inside `vis_config.by_value`. " +
			"Mutually exclusive with the other chart blocks in the same `by_value` block. " +
			"Shares the attribute shape with `lens_dashboard_app_config.by_value." + chartConfigName + "`.",
		Optional:   true,
		Attributes: attrs,
	}
}

// DataSourceJSONAttribute returns the canonical `data_source_json` schema attribute used by
// most Lens chart blocks: a required normalized JSON string with the supplied markdown.
func DataSourceJSONAttribute(markdown string) schema.Attribute {
	return schema.StringAttribute{
		MarkdownDescription: markdown,
		CustomType:          jsontypes.NormalizedType{},
		Required:            true,
	}
}

// QueryAttribute returns the canonical `query` schema attribute (optional SingleNestedAttribute
// over LensChartFilterSimpleAttributes) with the supplied markdown.
func QueryAttribute(markdown string) schema.Attribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: markdown,
		Optional:            true,
		Attributes:          LensChartFilterSimpleAttributes(),
	}
}

// AxisTitleAttribute returns the shared `{ value: string optional, visible: bool optional[/computed] }`
// nested attribute used for chart-axis titles in XY and heatmap. computedVisible toggles the
// Computed flag on the visible field (XY uses computed; heatmap does not).
func AxisTitleAttribute(computedVisible bool) schema.Attribute {
	visible := schema.BoolAttribute{
		MarkdownDescription: "Whether to show the title.",
		Optional:            true,
	}
	if computedVisible {
		visible.Computed = true
	}
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Axis title configuration.",
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"value": schema.StringAttribute{
				MarkdownDescription: "Axis title text.",
				Optional:            true,
			},
			"visible": visible,
		},
	}
}

// MutuallyExclusiveStringValidator returns a single-element validator slice declaring this
// string attribute conflicts with the named sibling on the parent object.
func MutuallyExclusiveStringValidator(siblingName string) []validator.String {
	return []validator.String{
		stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName(siblingName)),
	}
}

// MutuallyExclusiveListValidator is the list-attribute counterpart to MutuallyExclusiveStringValidator.
func MutuallyExclusiveListValidator(siblingName string) []validator.List {
	return []validator.List{
		listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName(siblingName)),
	}
}

// MutuallyExclusiveObjectValidator is the object-attribute counterpart to MutuallyExclusiveStringValidator.
func MutuallyExclusiveObjectValidator(siblingName string) []validator.Object {
	return []validator.Object{
		objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName(siblingName)),
	}
}

// MetricJSONAttribute returns the canonical chart `metric_json` schema attribute: a string
// holding JSON normalized through the supplied JSONWithDefaults populator. When esqlSiblingName
// is non-empty, a ConflictsWith validator is attached so practitioners can't set both `metric_json`
// and the typed ES|QL counterpart at the same time.
func MetricJSONAttribute[T any](markdown string, defaults customtypes.PopulateDefaultsFunc[T], required bool, esqlSiblingName string) schema.Attribute {
	attr := schema.StringAttribute{
		MarkdownDescription: markdown,
		CustomType:          customtypes.NewJSONWithDefaultsType(defaults),
	}
	if required {
		attr.Required = true
	} else {
		attr.Optional = true
	}
	if esqlSiblingName != "" {
		attr.Validators = MutuallyExclusiveStringValidator(esqlSiblingName)
	}
	return attr
}

// JSONConfigItemList returns the canonical list-of-`{ config_json }` schema attribute used by
// pie metrics, pie group_by, waffle metrics, waffle group_by, and similar charts. Each item is a
// `{ config_json: <JSONWithDefaults string> }` nested object.
func JSONConfigItemList[T any](markdown, configMarkdown string, defaults customtypes.PopulateDefaultsFunc[T], required bool, sizeValidators ...validator.List) schema.Attribute {
	attr := schema.ListNestedAttribute{
		MarkdownDescription: markdown,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"config_json": schema.StringAttribute{
					MarkdownDescription: configMarkdown,
					CustomType:          customtypes.NewJSONWithDefaultsType(defaults),
					Required:            true,
				},
			},
		},
		Validators: sizeValidators,
	}
	if required {
		attr.Required = true
	} else {
		attr.Optional = true
	}
	return attr
}

// PartitionChartBaseAttributes returns the shared base attribute map used by mosaic, treemap,
// and waffle: LensChartBaseAttributes plus a partition-flavored data_source_json and query, and
// optionally LensChartPresentationAttributes when includePresentation is true.
func PartitionChartBaseAttributes(includePresentation bool) map[string]schema.Attribute {
	attrs := LensChartBaseAttributes()
	attrs["data_source_json"] = DataSourceJSONAttribute(
		"Dataset configuration as JSON. For non-ES|QL, this specifies the data view or index; for ES|QL, this specifies the ES|QL query dataset.",
	)
	attrs["query"] = QueryAttribute(
		"Query configuration for filtering data. Required for non-ES|QL partition charts.",
	)
	if includePresentation {
		maps.Copy(attrs, LensChartPresentationAttributes())
	}
	return attrs
}
