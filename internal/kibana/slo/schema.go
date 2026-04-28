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

package slo

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = getSchema()
}

func getSchema() schema.Schema {
	return schema.Schema{
		Version:             2,
		MarkdownDescription: sloResourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Internal identifier of the resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"slo_id": schema.StringAttribute{
				Description: "An ID (8 to 36 characters) that contains only letters, numbers, hyphens, and underscores. If omitted, a UUIDv1 will be generated server-side.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(8, 36),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9_-]+$`), "must contain only letters, numbers, hyphens, and underscores"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the SLO is enabled in Kibana.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"artifacts": schema.SingleNestedAttribute{
				Description: "Links to related assets (for example dashboards) returned and managed with the SLO.",
				Optional:    true,
				Computed:    true,
				// Default null normalizes an omitted `artifacts` to null; a successful read
				// still maps API dashboard references into state (REQ-020/REQ-039).
				Default: objectdefault.StaticValue(
					types.ObjectNull(tfArtifactsAttrTypes),
				),
				Attributes: map[string]schema.Attribute{
					"dashboards": schema.ListNestedAttribute{
						Description: "Dashboard references attached to the SLO.",
						Optional:    true,
						Computed:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Description: "Dashboard saved object id.",
									Required:    true,
								},
							},
						},
					},
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the SLO.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description for the SLO.",
				Required:    true,
			},
			"budgeting_method": schema.StringAttribute{
				Description: budgetingMethodDescription,
				Required:    true,
				Validators:  []validator.String{stringvalidator.OneOf("occurrences", "timeslices")},
			},
			"space_id": schema.StringAttribute{
				Description: "An identifier for the space. If space_id is not provided, the default space is used.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("default"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"group_by": schema.ListAttribute{
				Description: "Optional group by fields to use to generate an SLO per distinct value.",
				Optional:    true,
				Computed:    true,
				CustomType:  NewGroupByType(),
				Validators: []validator.List{
					listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
				},
			},
			"tags": schema.ListAttribute{
				Description: "The tags for the SLO.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			"kibana_connection": providerschema.GetKbFWConnectionBlock(),
			"settings": schema.SingleNestedBlock{
				Description: "The default settings should be sufficient for most users, but if needed, these properties can be overwritten.",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"sync_delay": schema.StringAttribute{Optional: true, Computed: true},
					"frequency":  schema.StringAttribute{Optional: true, Computed: true},
					"sync_field": schema.StringAttribute{
						Description: "The date field used to identify new documents in the source. When unspecified, the indicator timestamp field is used.",
						Optional:    true,
						Computed:    true,
					},
					"prevent_initial_backfill": schema.BoolAttribute{
						Description: "Prevents the underlying ES transform from attempting to backfill data on start, which can sometimes be resource-intensive or time-consuming and unnecessary",
						Optional:    true,
						Computed:    true,
					},
				},
			},
			"time_window": schema.ListNestedBlock{
				Description: timeWindowDescription,
				Validators:  []validator.List{listvalidator.SizeBetween(1, 1)},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"duration": schema.StringAttribute{Required: true},
						"type": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.OneOf("rolling", "calendarAligned"),
							},
						},
					},
				},
			},
			"objective": schema.ListNestedBlock{
				Description: objectiveDescription,
				Validators:  []validator.List{listvalidator.SizeBetween(1, 1)},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"target":           schema.Float64Attribute{Required: true},
						"timeslice_target": schema.Float64Attribute{Optional: true},
						"timeslice_window": schema.StringAttribute{Optional: true},
					},
				},
			},

			"metric_custom_indicator":    metricCustomIndicatorSchema(),
			"histogram_custom_indicator": histogramCustomIndicatorSchema(),
			"apm_latency_indicator":      apmLatencyIndicatorSchema(),
			"apm_availability_indicator": apmAvailabilityIndicatorSchema(),
			"kql_custom_indicator":       kqlCustomIndicatorSchema(),
			"timeslice_metric_indicator": timesliceMetricIndicatorSchema(),
		},
	}
}

func metricCustomIndicatorSchema() schema.Block {
	return schema.ListNestedBlock{
		Validators: []validator.List{listvalidator.SizeBetween(1, 1)},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"index":           schema.StringAttribute{Required: true},
				"data_view_id":    schema.StringAttribute{Optional: true, Description: "Optional data view id to use for this indicator."},
				"filter":          schema.StringAttribute{Optional: true},
				"timestamp_field": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("@timestamp")},
			},
			Blocks: map[string]schema.Block{
				"good": schema.ListNestedBlock{
					Validators: []validator.List{listvalidator.SizeBetween(1, 1)},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"equation": schema.StringAttribute{Required: true},
						},
						Blocks: map[string]schema.Block{
							"metrics": schema.ListNestedBlock{
								Validators: []validator.List{listvalidator.SizeAtLeast(1)},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Required: true,
											Validators: []validator.String{
												stringvalidator.RegexMatches(regexp.MustCompile(`^[A-Z]$`), "must be a single Latin letter (A–Z)"),
											},
										},
										"aggregation": schema.StringAttribute{
											Required: true,
											Validators: []validator.String{
												stringvalidator.OneOf("sum", "doc_count"),
											},
										},
										"field": schema.StringAttribute{
											Optional:    true,
											Description: "Field to aggregate. Required for all aggregations except doc_count. Must NOT be set for doc_count.",
											Validators: []validator.String{
												validators.RequiredIfDependentPathExpressionOneOf(
													path.MatchRelative().AtParent().AtName("aggregation"),
													[]string{"sum"},
												),
												validators.ForbiddenIfDependentPathExpressionOneOf(
													path.MatchRelative().AtParent().AtName("aggregation"),
													[]string{"doc_count"},
												),
											},
										},
										"filter": schema.StringAttribute{Optional: true},
									},
								},
							},
						},
					},
				},
				"total": schema.ListNestedBlock{
					Validators: []validator.List{listvalidator.SizeBetween(1, 1)},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"equation": schema.StringAttribute{Required: true},
						},
						Blocks: map[string]schema.Block{
							"metrics": schema.ListNestedBlock{
								Validators: []validator.List{listvalidator.SizeAtLeast(1)},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Required: true,
											Validators: []validator.String{
												stringvalidator.RegexMatches(regexp.MustCompile(`^[A-Z]$`), "must be a single Latin letter (A–Z)"),
											},
										},
										"aggregation": schema.StringAttribute{
											Required: true,
											Validators: []validator.String{
												stringvalidator.OneOf("sum", "doc_count"),
											},
										},
										"field": schema.StringAttribute{
											Optional:    true,
											Description: "Field to aggregate. Required for all aggregations except doc_count. Must NOT be set for doc_count.",
											Validators: []validator.String{
												validators.RequiredIfDependentPathExpressionOneOf(
													path.MatchRelative().AtParent().AtName("aggregation"),
													[]string{"sum"},
												),
												validators.ForbiddenIfDependentPathExpressionOneOf(
													path.MatchRelative().AtParent().AtName("aggregation"),
													[]string{"doc_count"},
												),
											},
										},
										"filter": schema.StringAttribute{Optional: true},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func histogramCustomIndicatorSchema() schema.Block {
	return schema.ListNestedBlock{
		Validators: []validator.List{listvalidator.SizeBetween(1, 1)},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"index":           schema.StringAttribute{Required: true},
				"data_view_id":    schema.StringAttribute{Optional: true, Description: "Optional data view id to use for this indicator."},
				"filter":          schema.StringAttribute{Optional: true},
				"timestamp_field": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("@timestamp")},
			},
			Blocks: map[string]schema.Block{
				"good": schema.ListNestedBlock{
					Validators: []validator.List{listvalidator.SizeBetween(1, 1)},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"aggregation": schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.OneOf("value_count", "range")}},
							"field":       schema.StringAttribute{Required: true},
							"filter":      schema.StringAttribute{Optional: true},
							"from": schema.Float64Attribute{
								Optional: true,
								Validators: []validator.Float64{
									validators.RequiredIfDependentPathExpressionOneOf(
										path.MatchRelative().AtParent().AtName("aggregation"),
										[]string{"range"},
									),
									validators.ForbiddenIfDependentPathExpressionOneOf(
										path.MatchRelative().AtParent().AtName("aggregation"),
										[]string{"value_count"},
									),
								},
							},
							"to": schema.Float64Attribute{
								Optional: true,
								Validators: []validator.Float64{
									validators.RequiredIfDependentPathExpressionOneOf(
										path.MatchRelative().AtParent().AtName("aggregation"),
										[]string{"range"},
									),
									validators.ForbiddenIfDependentPathExpressionOneOf(
										path.MatchRelative().AtParent().AtName("aggregation"),
										[]string{"value_count"},
									),
								},
							},
						},
					},
				},
				"total": schema.ListNestedBlock{
					Validators: []validator.List{listvalidator.SizeBetween(1, 1)},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"aggregation": schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.OneOf("value_count", "range")}},
							"field":       schema.StringAttribute{Required: true},
							"filter":      schema.StringAttribute{Optional: true},
							"from": schema.Float64Attribute{
								Optional: true,
								Validators: []validator.Float64{
									validators.RequiredIfDependentPathExpressionOneOf(
										path.MatchRelative().AtParent().AtName("aggregation"),
										[]string{"range"},
									),
									validators.ForbiddenIfDependentPathExpressionOneOf(
										path.MatchRelative().AtParent().AtName("aggregation"),
										[]string{"value_count"},
									),
								},
							},
							"to": schema.Float64Attribute{
								Optional: true,
								Validators: []validator.Float64{
									validators.RequiredIfDependentPathExpressionOneOf(
										path.MatchRelative().AtParent().AtName("aggregation"),
										[]string{"range"},
									),
									validators.ForbiddenIfDependentPathExpressionOneOf(
										path.MatchRelative().AtParent().AtName("aggregation"),
										[]string{"value_count"},
									),
								},
							},
						},
					},
				},
			},
		},
	}
}

func apmLatencyIndicatorSchema() schema.Block {
	return schema.ListNestedBlock{
		Validators: []validator.List{listvalidator.SizeBetween(1, 1)},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"index":            schema.StringAttribute{Required: true},
				"filter":           schema.StringAttribute{Optional: true},
				"service":          schema.StringAttribute{Required: true},
				"environment":      schema.StringAttribute{Required: true},
				"transaction_type": schema.StringAttribute{Required: true},
				"transaction_name": schema.StringAttribute{Required: true},
				"threshold":        schema.Int64Attribute{Required: true},
			},
		},
	}
}

func apmAvailabilityIndicatorSchema() schema.Block {
	return schema.ListNestedBlock{
		Validators: []validator.List{listvalidator.SizeBetween(1, 1)},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"index":            schema.StringAttribute{Required: true},
				"filter":           schema.StringAttribute{Optional: true},
				"service":          schema.StringAttribute{Required: true},
				"environment":      schema.StringAttribute{Required: true},
				"transaction_type": schema.StringAttribute{Required: true},
				"transaction_name": schema.StringAttribute{Required: true},
			},
		},
	}
}

func kqlCustomIndicatorSchema() schema.Block {
	return schema.ListNestedBlock{
		Validators: []validator.List{listvalidator.SizeBetween(1, 1)},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"index":        schema.StringAttribute{Required: true},
				"data_view_id": schema.StringAttribute{Optional: true, Description: "Optional data view id to use for this indicator."},
				"filter": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("filter_kql")),
					},
				},
				"filter_kql": kqlWithFiltersObjectSchema("filter"),
				"good": schema.StringAttribute{
					Optional: true,
					Computed: true,
					Default:  stringdefault.StaticString(""),
					Validators: []validator.String{
						stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("good_kql")),
					},
				},
				"good_kql": kqlWithFiltersObjectSchema("good"),
				"total": schema.StringAttribute{
					Optional: true,
					Computed: true,
					Default:  stringdefault.StaticString(""),
					Validators: []validator.String{
						stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("total_kql")),
					},
				},
				"total_kql":       kqlWithFiltersObjectSchema("total"),
				"timestamp_field": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("@timestamp")},
			},
		},
	}
}

// kqlWithFiltersObjectSchema defines the object-form KQL union (kqlQuery + filters) used for filter, good, and total.
func kqlWithFiltersObjectSchema(parallelStringAttr string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: fmt.Sprintf(
			"Object-form KQL (kqlQuery and filters). Mutually exclusive with the legacy string attribute for the same logical field. "+
				"Use the attribute form in Terraform (e.g. `%[1]s = { kql_query = \"...\" }`), not a nested `%[1]s { ... }` block.",
			parallelStringAttr+"_kql",
		),
		Optional: true,
		Computed: true,
		// ObjectNull default keeps a stable placeholder for the nested _kql block; a read
		// may still set object-form KQL from the API.
		Default: objectdefault.StaticValue(
			types.ObjectNull(tfKqlKqlObjectAttrTypes),
		),
		Attributes: map[string]schema.Attribute{
			"kql_query": schema.StringAttribute{
				Description: "KQL query string when using the object form.",
				Optional:    true,
				Computed:    true,
			},
			"filters": schema.ListNestedAttribute{
				Description: "Optional Kibana filter objects (query JSON) accompanying the KQL object form.",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"query": schema.StringAttribute{
							Description: "Filter query as a JSON object.",
							Optional:    true,
							Computed:    true,
							CustomType:  jsontypes.NormalizedType{},
						},
					},
				},
			},
		},
		Validators: []validator.Object{
			objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName(parallelStringAttr)),
			kqlObjectFormMeaningful{},
		},
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
	}
}

func timesliceMetricIndicatorSchema() schema.Block {
	return schema.ListNestedBlock{
		Description: "Defines a timeslice metric indicator for SLO.",
		Validators:  []validator.List{listvalidator.SizeBetween(1, 1)},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"index":           schema.StringAttribute{Required: true},
				"data_view_id":    schema.StringAttribute{Optional: true, Description: "Optional data view id to use for this indicator."},
				"timestamp_field": schema.StringAttribute{Required: true},
				"filter":          schema.StringAttribute{Optional: true},
			},
			Blocks: map[string]schema.Block{
				"metric": schema.ListNestedBlock{
					Validators: []validator.List{listvalidator.SizeBetween(1, 1)},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"equation":   schema.StringAttribute{Required: true},
							"comparator": schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.OneOf("GT", "GTE", "LT", "LTE")}},
							"threshold":  schema.Float64Attribute{Required: true},
						},
						Blocks: map[string]schema.Block{
							"metrics": schema.ListNestedBlock{
								Validators: []validator.List{listvalidator.SizeAtLeast(1)},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Required:    true,
											Description: "The unique name for this metric. Used as a variable in the equation field. Must be a single letter A–Z.",
											Validators: []validator.String{
												stringvalidator.RegexMatches(regexp.MustCompile(`^[A-Z]$`), "must be a single Latin letter (A–Z)"),
											},
										},
										"aggregation": schema.StringAttribute{
											Required: true,
											Description: fmt.Sprintf(
												"The aggregation type for this metric (kbapi timeslice metric union: no value_count). One of: %s. Determines which other fields are required.",
												strings.Join(timesliceMetricAggregations, ", "),
											),
											Validators: []validator.String{stringvalidator.OneOf(timesliceMetricAggregations...)},
										},
										"field": schema.StringAttribute{
											Optional:    true,
											Description: fmt.Sprintf("Field to aggregate. Required for %s. Must NOT be set for doc_count.", strings.Join(timesliceMetricAggregationsWithField, ", ")),
											Validators: []validator.String{
												validators.RequiredIfDependentPathExpressionOneOf(
													path.MatchRelative().AtParent().AtName("aggregation"),
													timesliceMetricAggregationsWithField,
												),
												validators.ForbiddenIfDependentPathExpressionOneOf(
													path.MatchRelative().AtParent().AtName("aggregation"),
													[]string{timesliceMetricAggregationDocCount},
												),
											},
										},
										"percentile": schema.Float64Attribute{
											Optional:    true,
											Description: "Percentile value (e.g., 99). Required if aggregation is 'percentile'. Must NOT be set for other aggregations.",
											Validators: []validator.Float64{
												validators.RequiredIfDependentPathExpressionOneOf(
													path.MatchRelative().AtParent().AtName("aggregation"),
													[]string{timesliceMetricAggregationPercentile},
												),
												validators.ForbiddenIfDependentPathExpressionOneOf(
													path.MatchRelative().AtParent().AtName("aggregation"),
													timesliceMetricAggregationsWithoutPercentile,
												),
												float64validator.Between(0, 100),
											},
										},
										"filter": schema.StringAttribute{
											Optional: true,
											Description: "Optional KQL filter for this metric. Supported for all timeslice metric aggregation " +
												"kinds, including doc_count, per the Kibana SLO API.",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
