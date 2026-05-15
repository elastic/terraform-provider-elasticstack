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
	"maps"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// populateTagcloudMetricDefaults populates default values for tagcloud metric configuration
func populateTagcloudMetricDefaults(model map[string]any) map[string]any {
	return lenscommon.PopulateTagcloudMetricDefaults(model)
}

// populateTagcloudTagByDefaults populates default values for tagcloud tag_by configuration
func populateTagcloudTagByDefaults(model map[string]any) map[string]any {
	return lenscommon.PopulateTagcloudTagByDefaults(model)
}

// getTagcloudSchema returns the schema for tagcloud chart configuration.
// includePresentation merges REQ-037 fields for vis panels only.
func getTagcloudSchema(includePresentation bool) map[string]schema.Attribute {
	attrs := lensChartBaseAttributes()
	attrs["data_source_json"] = schema.StringAttribute{
		MarkdownDescription: "Dataset configuration as JSON. For standard layers, this specifies the data view and query.",
		CustomType:          jsontypes.NormalizedType{},
		Required:            true,
	}
	attrs["query"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Query configuration for filtering data. Required for non-ES|QL tagclouds; omit for ES|QL mode.",
		Optional:            true,
		Attributes:          getFilterSimple(),
	}
	attrs["orientation"] = schema.StringAttribute{
		MarkdownDescription: "Orientation of the tagcloud. Valid values: 'horizontal', 'vertical', 'angled'.",
		Optional:            true,
		Validators: []validator.String{
			stringvalidator.OneOf("horizontal", "vertical", "angled"),
		},
	}
	attrs["font_size"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Minimum and maximum font size for the tags.",
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"min": schema.Float64Attribute{
				MarkdownDescription: "Minimum font size (default: 18, minimum: 1).",
				Optional:            true,
			},
			"max": schema.Float64Attribute{
				MarkdownDescription: "Maximum font size (default: 72, maximum: 120).",
				Optional:            true,
			},
		},
	}
	attrs["metric_json"] = schema.StringAttribute{
		MarkdownDescription: tagcloudMetricDescription + " Required for non-ES|QL tagclouds; mutually exclusive with `esql_metric`.",
		CustomType:          customtypes.NewJSONWithDefaultsType(populateTagcloudMetricDefaults),
		Optional:            true,
		Validators: []validator.String{
			stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("esql_metric")),
		},
	}
	attrs["tag_by_json"] = schema.StringAttribute{
		MarkdownDescription: "Tag grouping configuration as JSON. Can be a date histogram, terms, histogram, range, or filters operation. " +
			"This determines how tags are grouped and displayed. Required for non-ES|QL tagclouds; mutually exclusive with `esql_tag_by`.",
		CustomType: customtypes.NewJSONWithDefaultsType(populateTagcloudTagByDefaults),
		Optional:   true,
		Validators: []validator.String{
			stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("esql_tag_by")),
		},
	}
	attrs["esql_metric"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Typed metric column for ES|QL tagclouds. Mutually exclusive with `metric_json`.",
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"column": schema.StringAttribute{
				MarkdownDescription: "ES|QL column name for the metric.",
				Required:            true,
			},
			"format_json": schema.StringAttribute{
				MarkdownDescription: "Number or other format configuration as JSON (`formatType` union).",
				CustomType:          jsontypes.NormalizedType{},
				Required:            true,
			},
			"label": schema.StringAttribute{
				MarkdownDescription: "Optional label for the metric.",
				Optional:            true,
			},
		},
	}
	attrs["esql_tag_by"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Typed tag-by column for ES|QL tagclouds. Mutually exclusive with `tag_by_json`.",
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"column": schema.StringAttribute{
				MarkdownDescription: "ES|QL column for the tag dimension.",
				Required:            true,
			},
			"format_json": schema.StringAttribute{
				MarkdownDescription: "Column format as JSON (`formatType` union).",
				CustomType:          jsontypes.NormalizedType{},
				Required:            true,
			},
			"color_json": schema.StringAttribute{
				MarkdownDescription: "Color mapping as JSON (`colorMapping` union).",
				CustomType:          jsontypes.NormalizedType{},
				Required:            true,
			},
			"label": schema.StringAttribute{
				MarkdownDescription: "Optional label for the tag-by column.",
				Optional:            true,
			},
		},
	}
	if includePresentation {
		maps.Copy(attrs, lensChartPresentationAttributes())
	}
	return attrs
}
