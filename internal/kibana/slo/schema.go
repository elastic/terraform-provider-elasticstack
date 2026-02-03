package slo

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
				Description: "An ID (8 to 48 characters) that contains only letters, numbers, hyphens, and underscores. If omitted, a UUIDv1 will be generated server-side.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(8, 48),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9_-]+$`), "must contain only letters, numbers, hyphens, and underscores"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
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
				Description: "An `occurrences` budgeting method uses the number of good and total events during the time window. A `timeslices` budgeting method uses the number of good slices and total slices during the time window. A slice is an arbitrary time window (smaller than the overall SLO time window) that is either considered good or bad, calculated from the timeslice threshold and the ratio of good over total events that happened during the slice window. A budgeting method is required and must be either occurrences or timeslices.",
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
				ElementType: types.StringType,
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
			"settings": schema.SingleNestedBlock{
				Description: "The default settings should be sufficient for most users, but if needed, these properties can be overwritten.",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"sync_delay": schema.StringAttribute{Optional: true, Computed: true},
					"frequency":  schema.StringAttribute{Optional: true, Computed: true},
					"prevent_initial_backfill": schema.BoolAttribute{
						Description: "Prevents the underlying ES transform from attempting to backfill data on start, which can sometimes be resource-intensive or time-consuming and unnecessary",
						Optional:    true,
						Computed:    true,
					},
				},
			},
			"time_window": schema.ListNestedBlock{
				Description: "Currently support `calendarAligned` and `rolling` time windows. Any duration greater than 1 day can be used: days, weeks, months, quarters, years. Rolling time window requires a duration, e.g. `1w` for one week, and type: `rolling`. SLOs defined with such time window, will only consider the SLI data from the last duration period as a moving window. Calendar aligned time window requires a duration, limited to `1M` for monthly or `1w` for weekly, and type: `calendarAligned`.",
				Validators:  []validator.List{listvalidator.SizeBetween(1, 1)},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"duration": schema.StringAttribute{Required: true},
						"type":     schema.StringAttribute{Required: true},
					},
				},
			},
			"objective": schema.ListNestedBlock{
				Description: "The target objective is the value the SLO needs to meet during the time window. If a timeslices budgeting method is used, we also need to define the timesliceTarget which can be different than the overall SLO target.",
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
										"name":        schema.StringAttribute{Required: true},
										"aggregation": schema.StringAttribute{Required: true},
										"field":       schema.StringAttribute{Required: true},
										"filter":      schema.StringAttribute{Optional: true},
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
										"name":        schema.StringAttribute{Required: true},
										"aggregation": schema.StringAttribute{Required: true},
										"field":       schema.StringAttribute{Required: true},
										"filter":      schema.StringAttribute{Optional: true},
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
							"from":        schema.Float64Attribute{Optional: true},
							"to":          schema.Float64Attribute{Optional: true},
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
							"from":        schema.Float64Attribute{Optional: true},
							"to":          schema.Float64Attribute{Optional: true},
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
				"index":           schema.StringAttribute{Required: true},
				"data_view_id":    schema.StringAttribute{Optional: true, Description: "Optional data view id to use for this indicator."},
				"filter":          schema.StringAttribute{Optional: true},
				"good":            schema.StringAttribute{Optional: true},
				"total":           schema.StringAttribute{Optional: true},
				"timestamp_field": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("@timestamp")},
			},
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
										"name":        schema.StringAttribute{Required: true, Description: "The unique name for this metric. Used as a variable in the equation field."},
										"aggregation": schema.StringAttribute{Required: true, Description: "The aggregation type for this metric. One of: sum, avg, min, max, value_count, percentile, doc_count. Determines which other fields are required:"},
										"field":       schema.StringAttribute{Optional: true, Description: "Field to aggregate. Required for aggregations: sum, avg, min, max, value_count, percentile. Must NOT be set for doc_count."},
										"percentile":  schema.Float64Attribute{Optional: true, Description: "Percentile value (e.g., 99). Required if aggregation is 'percentile'. Must NOT be set for other aggregations."},
										"filter":      schema.StringAttribute{Optional: true, Description: "Optional KQL filter for this metric. Supported for all aggregations except doc_count."},
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
