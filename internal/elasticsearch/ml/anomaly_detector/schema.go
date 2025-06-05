package anomaly_detector

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *anomalyDetectorResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates or updates an Elasticsearch machine learning job.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Internal identifier of the resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"job_id": schema.StringAttribute{
				Description: "The ID of the job to create.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "A description of the job.",
				Optional:    true,
			},
			"groups": schema.ListAttribute{
				Description: "A list of job groups. A job can belong to no groups or many.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"analysis_config": schema.SingleNestedAttribute{
				Description: "The analysis configuration, which specifies how to analyze the data.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"bucket_span": schema.StringAttribute{
						Description: "The span of the analysis time bucket.",
						Required:    true,
					},
					"detectors": schema.ListNestedAttribute{
						Description: "An array of detector objects, which define the statistical analysis that will be performed.",
						Required:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"function": schema.StringAttribute{
									Description: "The analysis function that is used. For example, count, rare, mean, median, min, max, sum, lat_long, time_of_day, time_of_week, metric, text_field.",
									Required:    true,
								},
								"field_name": schema.StringAttribute{
									Description: "The field that the detector analyzes.",
									Optional:    true,
								},
								"by_field_name": schema.StringAttribute{
									Description: "The field used to split the data. In particular, this property is used for analyzing the splits with respect to the history of all splits. Refer to performing multifactor analysis.",
									Optional:    true,
								},
								"partition_field_name": schema.StringAttribute{
									Description: "The field used to segment the analysis. When you use this property, you have completely independent baselines for each value of this field.",
									Optional:    true,
								},
								"detector_description": schema.StringAttribute{
									Description: "A description of the detector.",
									Optional:    true,
								},
								"use_null": schema.BoolAttribute{
									Description: "If true, null values are not ignored when calculating the detector statistics.",
									Optional:    true,
								},
								"exclude_frequent": schema.StringAttribute{
									Description: "One of by, over, all, or none. If set, frequent entities are excluded from influencing the anomaly results. Entities can be considered frequent by virtue of their appearance in a by field or over field.",
									Optional:    true,
								},
								"custom_rules": schema.ListNestedAttribute{
									Description: "A list of custom rules that apply to this detector.",
									Optional:    true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"actions": schema.ListAttribute{
												Description: "A list of actions to take when the conditions are met.",
												ElementType: types.StringType,
												Required:    true,
											},
											"scope": schema.StringAttribute{
												Description: "The scope to which this rule applies (e.g., anomaly_score, typical, actual). This scope applies to all conditions under this rule.",
												Optional:    true, // A rule might have no conditions, or scope might be implicit in some API versions/contexts
											},
											"conditions": schema.ListNestedAttribute{
												Description: "A list of conditions that must be met for the actions to be taken.",
												Optional:    true, // Can have rules with actions only, though rare
												NestedObject: schema.NestedAttributeObject{
													Attributes: map[string]schema.Attribute{
														"operator": schema.StringAttribute{
															Description: "The operator for the condition (e.g., gt, lt, eq).",
															Required:    true,
														},
														"value": schema.Float64Attribute{
															Description: "The value used in the condition (numeric, e.g., for anomaly_score).",
															Required:    true,
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					"influencers": schema.ListAttribute{
						Description: "A list of strings that influence the anomaly results.",
						ElementType: types.StringType,
						Optional:    true,
					},
					"categorization_field_name": schema.StringAttribute{
						Description: "If the job is a categorization job, this field contains the name of the field that is used to segment the analysis.",
						Optional:    true,
					},
					"summary_count_field_name": schema.StringAttribute{
						Description: "The name of the field that contains the count of events for summarization.",
						Optional:    true,
					},
					"latency": schema.StringAttribute{
						Description: "The latency period during which data is not analyzed. After this period, the data is analyzed and potential anomalies are identified.",
						Optional:    true,
					},
				},
			},
			"data_description": schema.SingleNestedAttribute{
				Description: "The data description, which specifies how to interpret the data.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"time_field": schema.StringAttribute{
						Description: "The name of the field that contains the timestamp.",
						Required:    true,
					},
					"time_format": schema.StringAttribute{
						Description: "The time format that is used. Can be epoch, epoch_ms, or a Joda time format string.",
						Optional:    true,
					},
				},
			},
			"model_plot_config": schema.SingleNestedAttribute{
				Description: "Controls the locations of model plot documents. It is optional.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "If true, model plot is enabled.",
						Required:    true,
					},
				},
			},
			"analysis_limits": schema.SingleNestedAttribute{
				Description: "Limits can be applied to an anomaly detection job to fight memory outage. It is optional.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"model_memory_limit": schema.StringAttribute{
						Description: "The approved memory usage limit, expressed as a string. For example, 100mb.",
						Optional:    true, // Often has a default in ES, making it optional here
					},
					"categorization_examples_limit": schema.Int64Attribute{
						Description: "A limit to the number of examples that are stored for categorization.",
						Optional:    true,
					},
				},
			},
			"model_snapshot_retention_days": schema.Int64Attribute{
				Description: "The number of days to keep model snapshots.",
				Optional:    true,
			},
			"results_retention_days": schema.Int64Attribute{
				Description: "The number of days to keep anomaly detection results.",
				Optional:    true,
			},
			"allow_lazy_open": schema.BoolAttribute{
				Description: "Advanced configuration option. Specifies whether this job can be opened lazily.",
				Optional:    true,
			},
			"daily_model_snapshot_retention_after_days": schema.Int64Attribute{
				Description: "The number of days after which old daily snapshots are deleted.",
				Optional:    true,
			},
			"custom_settings": schema.MapAttribute{
				Description: "Advanced configuration option. Contains custom meta data about the job.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"results_index_name": schema.StringAttribute{
				Description: "The name of the index in which to store the machine learning results.",
				Optional:    true,
				Computed:    true, // It can be auto-generated by ES
			},
		},
	}
}
