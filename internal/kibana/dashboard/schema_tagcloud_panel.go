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

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// populateTagcloudMetricDefaults populates default values for tagcloud metric configuration
func populateTagcloudMetricDefaults(model map[string]any) map[string]any {
	if model == nil {
		return model
	}
	// Set defaults for all field metric operations
	if operation, ok := model["operation"].(string); ok && isFieldMetricOperation(operation) {
		if _, exists := model["empty_as_null"]; !exists {
			model["empty_as_null"] = false
		}
		if _, exists := model["show_metric_label"]; !exists {
			model["show_metric_label"] = true
		}
		if _, exists := model["color"]; !exists {
			model["color"] = map[string]any{"type": "auto"}
		}
	}
	return model
}

// populateTagcloudTagByDefaults populates default values for tagcloud tag_by configuration
func populateTagcloudTagByDefaults(model map[string]any) map[string]any {
	if model == nil {
		return model
	}
	// Set defaults for terms operation
	if operation, ok := model["operation"].(string); ok && operation == operationTerms {
		if _, exists := model["rank_by"]; !exists {
			model["rank_by"] = map[string]any{
				"type":         "metric",
				"metric_index": float64(0),
				"direction":    "desc",
			}
		}
		if _, exists := model["color"]; !exists {
			model["color"] = map[string]any{
				"mode":    "categorical",
				"palette": "default",
				"mapping": []any{},
			}
		}
	}
	return model
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
		MarkdownDescription: "Query configuration for filtering data.",
		Required:            true,
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
		MarkdownDescription: tagcloudMetricDescription,
		CustomType:          customtypes.NewJSONWithDefaultsType(populateTagcloudMetricDefaults),
		Required:            true,
	}
	attrs["tag_by_json"] = schema.StringAttribute{
		MarkdownDescription: "Tag grouping configuration as JSON. Can be a date histogram, terms, histogram, range, or filters operation. This determines how tags are grouped and displayed.",
		CustomType:          customtypes.NewJSONWithDefaultsType(populateTagcloudTagByDefaults),
		Required:            true,
	}
	if includePresentation {
		maps.Copy(attrs, lensChartPresentationAttributes())
	}
	return attrs
}
