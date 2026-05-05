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
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// getDatatableSchema returns the schema for datatable chart configuration
func getDatatableSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"no_esql": schema.SingleNestedAttribute{
			MarkdownDescription: "Datatable configuration for standard (non-ES|QL) queries.",
			Optional:            true,
			Attributes:          getDatatableNoESQLSchema(),
			Validators: []validator.Object{
				objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("esql")),
			},
		},
		"esql": schema.SingleNestedAttribute{
			MarkdownDescription: "Datatable configuration for ES|QL queries.",
			Optional:            true,
			Attributes:          getDatatableESQLSchema(),
			Validators: []validator.Object{
				objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("no_esql")),
			},
		},
	}
}

func getDatatableNoESQLSchema() map[string]schema.Attribute {
	attrs := lensChartBaseAttributes()
	attrs["data_source_json"] = schema.StringAttribute{
		MarkdownDescription: "Dataset configuration as JSON. For standard datatables, this specifies the data view and query.",
		CustomType:          jsontypes.NormalizedType{},
		Required:            true,
	}
	attrs["query"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Query configuration for filtering data.",
		Required:            true,
		Attributes:          getFilterSimple(),
	}
	attrs["metrics"] = schema.ListNestedAttribute{
		MarkdownDescription: "Array of metric configurations as JSON. Each entry defines a datatable metric column.",
		Required:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"config_json": schema.StringAttribute{
					MarkdownDescription: "Metric configuration as JSON.",
					CustomType:          jsontypes.NormalizedType{},
					Required:            true,
				},
			},
		},
	}
	attrs["rows"] = schema.ListNestedAttribute{
		MarkdownDescription: "Array of row configurations as JSON. Each entry defines a row split operation.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"config_json": schema.StringAttribute{
					MarkdownDescription: "Row configuration as JSON.",
					CustomType:          jsontypes.NormalizedType{},
					Required:            true,
				},
			},
		},
	}
	attrs["split_metrics_by"] = schema.ListNestedAttribute{
		MarkdownDescription: "Array of split-metrics configurations as JSON. Each entry defines a split operation for metric columns.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"config_json": schema.StringAttribute{
					MarkdownDescription: "Split metrics configuration as JSON.",
					CustomType:          jsontypes.NormalizedType{},
					Required:            true,
				},
			},
		},
	}
	attrs["styling"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Datatable styling and display configuration.",
		Required:            true,
		Attributes:          getDatatableStylingSchema(),
	}
	return attrs
}

func getDatatableESQLSchema() map[string]schema.Attribute {
	attrs := lensChartBaseAttributes()
	attrs["data_source_json"] = schema.StringAttribute{
		MarkdownDescription: "Dataset configuration as JSON. For ES|QL, this specifies the ES|QL query.",
		CustomType:          jsontypes.NormalizedType{},
		Required:            true,
	}
	attrs["metrics"] = schema.ListNestedAttribute{
		MarkdownDescription: "Array of metric configurations as JSON. Each entry defines a datatable metric column.",
		Required:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"config_json": schema.StringAttribute{
					MarkdownDescription: "Metric configuration as JSON.",
					CustomType:          jsontypes.NormalizedType{},
					Required:            true,
				},
			},
		},
	}
	attrs["rows"] = schema.ListNestedAttribute{
		MarkdownDescription: "Array of row configurations as JSON. Each entry defines a row split operation.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"config_json": schema.StringAttribute{
					MarkdownDescription: "Row configuration as JSON.",
					CustomType:          jsontypes.NormalizedType{},
					Required:            true,
				},
			},
		},
	}
	attrs["split_metrics_by"] = schema.ListNestedAttribute{
		MarkdownDescription: "Array of split-metrics configurations as JSON. Each entry defines a split operation for metric columns.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"config_json": schema.StringAttribute{
					MarkdownDescription: "Split metrics configuration as JSON.",
					CustomType:          jsontypes.NormalizedType{},
					Required:            true,
				},
			},
		},
	}
	attrs["styling"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Datatable styling and display configuration.",
		Required:            true,
		Attributes:          getDatatableStylingSchema(),
	}
	return attrs
}

func getDatatableStylingSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"density": schema.SingleNestedAttribute{
			MarkdownDescription: "Density configuration for the datatable.",
			Required:            true,
			Attributes:          getDatatableDensitySchema(),
		},
		"sort_by_json": schema.StringAttribute{
			MarkdownDescription: "Sort configuration as JSON. Only one column can be sorted at a time.",
			CustomType:          jsontypes.NormalizedType{},
			Optional:            true,
		},
		"paging": schema.Int64Attribute{
			MarkdownDescription: "Enables pagination and sets the number of rows to display per page.",
			Optional:            true,
		},
	}
}

func getDatatableDensitySchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"mode": schema.StringAttribute{
			MarkdownDescription: "Density mode. Valid values: 'compact', 'default', 'expanded'.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("compact", "default", "expanded"),
			},
		},
		"height": schema.SingleNestedAttribute{
			MarkdownDescription: "Header and value height configuration.",
			Optional:            true,
			Attributes: map[string]schema.Attribute{
				"header": schema.SingleNestedAttribute{
					MarkdownDescription: "Header height configuration.",
					Optional:            true,
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							MarkdownDescription: "Header height type. Valid values: 'auto', 'custom'.",
							Optional:            true,
							Validators: []validator.String{
								stringvalidator.OneOf(dashboardValueAuto, "custom"),
							},
						},
						"max_lines": schema.Float64Attribute{
							MarkdownDescription: "Maximum number of lines to use before header is truncated (for custom header height).",
							Optional:            true,
						},
					},
				},
				"value": schema.SingleNestedAttribute{
					MarkdownDescription: "Value height configuration.",
					Optional:            true,
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							MarkdownDescription: "Value height type. Valid values: 'auto', 'custom'.",
							Optional:            true,
							Validators: []validator.String{
								stringvalidator.OneOf(dashboardValueAuto, "custom"),
							},
						},
						"lines": schema.Float64Attribute{
							MarkdownDescription: "Number of lines to display per table body cell (for custom value height).",
							Optional:            true,
						},
					},
				},
			},
		},
	}
}
