package anomaly_detection_job

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
)

func (r *anomalyDetectionJobResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetSchema()
}

func GetSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Creates and manages Machine Learning anomaly detection jobs. See the [ML Job API documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/ml-put-job.html) for more details.",
		Blocks: map[string]schema.Block{
			"elasticsearch_connection": providerschema.GetEsFWConnectionBlock("elasticsearch_connection", false),
		},
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"job_id": schema.StringAttribute{
				MarkdownDescription: "The identifier for the anomaly detection job. This identifier can contain lowercase alphanumeric characters (a-z and 0-9), hyphens, and underscores. It must start and end with alphanumeric characters.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*[a-z0-9]$|^[a-z0-9]$`), "must contain lowercase alphanumeric characters (a-z and 0-9), hyphens, and underscores. It must start and end with alphanumeric characters"),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A description of the job.",
				Optional:            true,
			},
			"groups": schema.SetAttribute{
				MarkdownDescription: "A set of job groups. A job can belong to no groups or many.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"analysis_config": schema.SingleNestedAttribute{
				MarkdownDescription: "Specifies how to analyze the data. After you create a job, you cannot change the analysis configuration; all the properties are informational.",
				Required:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"bucket_span": schema.StringAttribute{
						MarkdownDescription: "The size of the interval that the analysis is aggregated into, typically between 15m and 1h. If the anomaly detector is expecting to see data at near real-time frequency, then the bucket_span should be set to a value around 10 times the time between ingested documents. For example, if data comes every second, bucket_span should be 10s; if data comes every 5 minutes, bucket_span should be 50m. For sparse or batch data, use larger bucket_span values.",
						Default:             stringdefault.StaticString("5m"),
						Computed:            true,
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.RegexMatches(regexp.MustCompile(`^\d+[nsumdh]$`), "must be a valid time interval (e.g., 15m, 1h)"),
						},
					},
					"categorization_field_name": schema.StringAttribute{
						MarkdownDescription: "For categorization jobs only. The name of the field to categorize.",
						Optional:            true,
					},
					"categorization_filters": schema.ListAttribute{
						MarkdownDescription: "For categorization jobs only. An array of regular expressions. A categorization message is matched against each regex in the order they are listed in the array.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"detectors": schema.ListNestedAttribute{
						MarkdownDescription: "Detector configuration objects. Detectors identify the anomaly detection functions and the fields on which they operate.",
						Required:            true,
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
						},
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"function": schema.StringAttribute{
									MarkdownDescription: "The analysis function that is used. For example, count, rare, mean, min, max, sum.",
									Required:            true,
									Validators: []validator.String{
										stringvalidator.OneOf("count", "high_count", "low_count", "non_zero_count", "high_non_zero_count", "low_non_zero_count", "distinct_count", "high_distinct_count", "low_distinct_count", "info_content", "high_info_content", "low_info_content", "min", "max", "median", "high_median", "low_median", "mean", "high_mean", "low_mean", "metric", "varp", "high_varp", "low_varp", "sum", "high_sum", "low_sum", "non_null_sum", "high_non_null_sum", "low_non_null_sum", "rare", "freq_rare", "time_of_day", "time_of_week", "lat_long"),
									},
								},
								"field_name": schema.StringAttribute{
									MarkdownDescription: "The field that the detector function analyzes. Some functions require a field. Functions that don't require a field are count, rare, and freq_rare.",
									Optional:            true,
								},
								"by_field_name": schema.StringAttribute{
									MarkdownDescription: "The field used to split the data. In particular, this property is used for analyzing the splits with respect to their own history. It is used for finding unusual values in the context of the split.",
									Optional:            true,
								},
								"over_field_name": schema.StringAttribute{
									MarkdownDescription: "The field used to split the data. In particular, this property is used for analyzing the splits with respect to the history of all splits. It is used for finding unusual values in the population of all splits.",
									Optional:            true,
								},
								"partition_field_name": schema.StringAttribute{
									MarkdownDescription: "The field used to segment the analysis. When you use this property, you have completely independent baselines for each value of this field.",
									Optional:            true,
								},
								"detector_description": schema.StringAttribute{
									MarkdownDescription: "A description of the detector.",
									Optional:            true,
								},
								"exclude_frequent": schema.StringAttribute{
									MarkdownDescription: "Contains one of the following values: all, none, by, or over.",
									Optional:            true,
									Validators: []validator.String{
										stringvalidator.OneOf("all", "none", "by", "over"),
									},
								},
								"use_null": schema.BoolAttribute{
									MarkdownDescription: "Defines whether a new series is used as the null series when there is no value for the by or partition fields.",
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.Bool{
										boolplanmodifier.UseStateForUnknown(),
									},
								},
								"custom_rules": schema.ListNestedAttribute{
									MarkdownDescription: "Custom rules enable you to customize the way detectors operate.",
									Optional:            true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"actions": schema.ListAttribute{
												MarkdownDescription: "The set of actions to be triggered when the rule applies. If more than one action is specified the effects of all actions are combined.",
												Optional:            true,
												ElementType:         types.StringType,
												Validators: []validator.List{
													listvalidator.ValueStringsAre(
														stringvalidator.OneOf("skip_result", "skip_model_update"),
													),
												},
											},
											"conditions": schema.ListNestedAttribute{
												MarkdownDescription: "An array of numeric conditions when the rule applies.",
												Optional:            true,
												NestedObject: schema.NestedAttributeObject{
													Attributes: map[string]schema.Attribute{
														"applies_to": schema.StringAttribute{
															MarkdownDescription: "Specifies the result property to which the condition applies.",
															Required:            true,
															Validators: []validator.String{
																stringvalidator.OneOf("actual", "typical", "diff_from_typical", "time"),
															},
														},
														"operator": schema.StringAttribute{
															MarkdownDescription: "Specifies the condition operator.",
															Required:            true,
															Validators: []validator.String{
																stringvalidator.OneOf("gt", "gte", "lt", "lte"),
															},
														},
														"value": schema.Float64Attribute{
															MarkdownDescription: "The value that is compared against the applies_to field using the operator.",
															Required:            true,
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
						MarkdownDescription: "A comma separated list of influencer field names. Typically these can be the by, over, or partition fields that are used in the detector configuration.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"latency": schema.StringAttribute{
						MarkdownDescription: "The size of the window in which to expect data that is out of time order. If you specify a non-zero value, it must be greater than or equal to one second.",
						Optional:            true,
					},
					"model_prune_window": schema.StringAttribute{
						MarkdownDescription: "Advanced configuration option. The time interval (in days) between pruning the model.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"multivariate_by_fields": schema.BoolAttribute{
						MarkdownDescription: "This functionality is reserved for internal use. It is not supported for use in customer environments and is not subject to the support SLA of official GA features.",
						Optional:            true,
					},
					"per_partition_categorization": schema.SingleNestedAttribute{
						MarkdownDescription: "Settings related to how categorization interacts with partition fields.",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"enabled": schema.BoolAttribute{
								MarkdownDescription: "To enable this setting, you must also set the partition_field_name property to the same value in every detector that uses the keyword mlcategory. Otherwise, job creation fails.",
								Optional:            true,
								Computed:            true,
								PlanModifiers: []planmodifier.Bool{
									boolplanmodifier.UseStateForUnknown(),
								},
							},
							"stop_on_warn": schema.BoolAttribute{
								MarkdownDescription: "This setting can be set to true only if per-partition categorization is enabled.",
								Optional:            true,
							},
						},
					},
					"summary_count_field_name": schema.StringAttribute{
						MarkdownDescription: "If this property is specified, the data that is fed to the job is expected to be pre-summarized.",
						Optional:            true,
					},
				},
			},
			"analysis_limits": schema.SingleNestedAttribute{
				MarkdownDescription: "Limits can be applied for the resources required to hold the mathematical models in memory.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"categorization_examples_limit": schema.Int64Attribute{
						MarkdownDescription: "The maximum number of examples stored per category in memory and in the results data store.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						Validators: []validator.Int64{
							int64validator.AtLeast(0),
						},
					},
					"model_memory_limit": schema.StringAttribute{
						MarkdownDescription: "The approximate maximum amount of memory resources that are required for analytical processing.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.RegexMatches(regexp.MustCompile(`^\d+[kmgtKMGT]?[bB]?$`), "must be a valid memory size (e.g., 10mb, 1gb)"),
						},
					},
				},
			},
			"data_description": schema.SingleNestedAttribute{
				MarkdownDescription: "Defines the format of the input data when you send data to the job by using the post data API.",
				Required:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"time_field": schema.StringAttribute{
						MarkdownDescription: "The name of the field that contains the timestamp.",
						Optional:            true,
					},
					"time_format": schema.StringAttribute{
						MarkdownDescription: "The time format, which can be epoch, epoch_ms, or a custom pattern.",
						Optional:            true,
					},
				},
			},
			"model_plot_config": schema.SingleNestedAttribute{
				MarkdownDescription: "This advanced configuration option stores model information along with the results. It provides a more detailed view into anomaly detection.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "If true, enables calculation and storage of the model bounds for each entity that is being analyzed.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"annotations_enabled": schema.BoolAttribute{
						MarkdownDescription: "If true, enables calculation and storage of the model change annotations for each entity that is being analyzed.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"terms": schema.StringAttribute{
						MarkdownDescription: "Limits data collection to this comma separated list of partition or by field values. If terms are not specified or it is an empty string, no filtering is applied.",
						Optional:            true,
					},
				},
			},
			"allow_lazy_open": schema.BoolAttribute{
				MarkdownDescription: "Advanced configuration option. Specifies whether this job can open when there is insufficient machine learning node capacity for it to be immediately assigned to a node.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"background_persist_interval": schema.StringAttribute{
				MarkdownDescription: "Advanced configuration option. The time between each periodic persistence of the model.",
				Optional:            true,
			},
			"custom_settings": schema.StringAttribute{
				MarkdownDescription: "Advanced configuration option. Contains custom meta data about the job. For example, it can contain custom URL information.",
				Optional:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"daily_model_snapshot_retention_after_days": schema.Int64Attribute{
				MarkdownDescription: "Advanced configuration option, which affects the automatic removal of old model snapshots for this job.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"model_snapshot_retention_days": schema.Int64Attribute{
				MarkdownDescription: "Advanced configuration option, which affects the automatic removal of old model snapshots for this job.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"renormalization_window_days": schema.Int64Attribute{
				MarkdownDescription: "Advanced configuration option. The period over which adjustments to the score are applied, as new data is seen.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"results_index_name": schema.StringAttribute{
				MarkdownDescription: "A text string that affects the name of the machine learning results index.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"results_retention_days": schema.Int64Attribute{
				MarkdownDescription: "Advanced configuration option. The period of time (in days) that results are retained.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},

			// Read-only computed attributes
			"create_time": schema.StringAttribute{
				MarkdownDescription: "The time the job was created.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"job_type": schema.StringAttribute{
				MarkdownDescription: "Reserved for future use, currently set to anomaly_detector.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"job_version": schema.StringAttribute{
				MarkdownDescription: "The version of Elasticsearch when the job was created.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"model_snapshot_id": schema.StringAttribute{
				MarkdownDescription: "A numerical character string that uniquely identifies the model snapshot.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func getAnalysisConfigAttrTypes() map[string]attr.Type {
	return GetSchema().Attributes["analysis_config"].GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}

func getDetectorAttrTypes() map[string]attr.Type {
	analysisConfigAttrs := getAnalysisConfigAttrTypes()
	detectorsList := analysisConfigAttrs["detectors"].(types.ListType)
	detectorsObj := detectorsList.ElemType.(types.ObjectType)
	return detectorsObj.AttrTypes
}

func getCustomRuleAttrTypes() map[string]attr.Type {
	detectorAttrs := getDetectorAttrTypes()
	customRulesList := detectorAttrs["custom_rules"].(types.ListType)
	customRulesObj := customRulesList.ElemType.(types.ObjectType)
	return customRulesObj.AttrTypes
}

func getRuleConditionAttrTypes() map[string]attr.Type {
	customRuleAttrs := getCustomRuleAttrTypes()
	conditionsList := customRuleAttrs["conditions"].(types.ListType)
	conditionsObj := conditionsList.ElemType.(types.ObjectType)
	return conditionsObj.AttrTypes
}

func getPerPartitionCategorizationAttrTypes() map[string]attr.Type {
	analysisConfigAttrs := getAnalysisConfigAttrTypes()
	perPartitionObj := analysisConfigAttrs["per_partition_categorization"].(types.ObjectType)
	return perPartitionObj.AttrTypes
}

func getAnalysisLimitsAttrTypes() map[string]attr.Type {
	return GetSchema().Attributes["analysis_limits"].GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}

func getDataDescriptionAttrTypes() map[string]attr.Type {
	return GetSchema().Attributes["data_description"].GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}

func getModelPlotConfigAttrTypes() map[string]attr.Type {
	return GetSchema().Attributes["model_plot_config"].GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}
