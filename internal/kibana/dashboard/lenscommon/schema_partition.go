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
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// PartitionLegendSchemaAttributes returns nested-object attributes shared by pie, treemap, and mosaic legends.
func PartitionLegendSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"nested": schema.BoolAttribute{
			MarkdownDescription: "Show nested legend with hierarchical breakdown levels.",
			Optional:            true,
		},
		"size": schema.StringAttribute{
			MarkdownDescription: "Legend size: auto, s, m, l, or xl.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("auto", "s", "m", "l", "xl"),
			},
		},
		"truncate_after_lines": schema.Int64Attribute{
			MarkdownDescription: "Maximum lines before truncating legend items (1-10).",
			Optional:            true,
		},
		"visible": schema.StringAttribute{
			MarkdownDescription: "Legend visibility: auto, visible, or hidden.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("auto", "visible", "hidden"),
			},
		},
	}
}

// PartitionValueDisplaySchemaAttributes returns schema for treemap/mosaic slice value display styling.
func PartitionValueDisplaySchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"mode": schema.StringAttribute{
			MarkdownDescription: "Value display mode: hidden, absolute, or percentage.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("hidden", "absolute", "percentage"),
			},
		},
		"percent_decimals": schema.Float64Attribute{
			MarkdownDescription: "Decimal places for percentage display (0-10).",
			Optional:            true,
		},
	}
}
