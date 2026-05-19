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
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// PartitionESQLMetricNestedObject returns the shared ES|QL metric nested schema used by treemap, mosaic, and waffle.
func PartitionESQLMetricNestedObject() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"column": schema.StringAttribute{
				MarkdownDescription: "ES|QL column name for the metric.",
				Required:            true,
			},
			"label": schema.StringAttribute{
				MarkdownDescription: "Optional label for the metric.",
				Optional:            true,
			},
			"format_json": schema.StringAttribute{
				MarkdownDescription: "Number or other format configuration as JSON (`formatType` union).",
				CustomType:          jsontypes.NormalizedType{},
				Required:            true,
			},
			"color": schema.SingleNestedAttribute{
				MarkdownDescription: "Static color for the metric.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "Color type; use `static` for partition chart ES|QL metrics.",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("static"),
						},
					},
					"color": schema.StringAttribute{
						MarkdownDescription: "Color value (e.g. hex).",
						Required:            true,
					},
				},
			},
		},
	}
}

// PartitionESQLGroupByNestedObject returns the shared ES|QL group-by nested schema used by treemap, mosaic, and waffle.
func PartitionESQLGroupByNestedObject() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"column": schema.StringAttribute{
				MarkdownDescription: "ES|QL column for the breakdown.",
				Required:            true,
			},
			"collapse_by": schema.StringAttribute{
				MarkdownDescription: "Collapse function when multiple rows map to the same bucket.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("avg", "max", "min", "sum"),
				},
			},
			"color_json": schema.StringAttribute{
				MarkdownDescription: "Color mapping as JSON (`colorMapping` union).",
				CustomType:          jsontypes.NormalizedType{},
				Required:            true,
			},
			"format_json": schema.StringAttribute{
				MarkdownDescription: "Column format as JSON (e.g. `{\"type\":\"number\"}`). Defaults to numeric format when omitted.",
				CustomType:          jsontypes.NormalizedType{},
				Optional:            true,
			},
			"label": schema.StringAttribute{
				MarkdownDescription: "Optional label for the group-by column.",
				Optional:            true,
			},
		},
	}
}

// MosaicESQLMetricNestedObject returns the ES|QL metric nested schema for mosaic (single metric, no color block).
func MosaicESQLMetricNestedObject() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"column": schema.StringAttribute{
				MarkdownDescription: "ES|QL column name for the metric.",
				Required:            true,
			},
			"label": schema.StringAttribute{
				MarkdownDescription: "Optional label for the metric.",
				Optional:            true,
			},
			"format_json": schema.StringAttribute{
				MarkdownDescription: "Number or other format configuration as JSON (`formatType` union).",
				CustomType:          jsontypes.NormalizedType{},
				Required:            true,
			},
		},
	}
}
