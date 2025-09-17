package security_detection_rule

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *securityDetectionRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetSchema()
}

func GetSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Creates or updates a Kibana security detection rule. See the [rules API documentation](https://www.elastic.co/guide/en/security/current/rules-api-create.html) for more details.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				MarkdownDescription: "An identifier for the space. If space_id is not provided, the default space is used.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("default"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"rule_id": schema.StringAttribute{
				MarkdownDescription: "A stable unique identifier for the rule object. If omitted, a UUID is generated.",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "A human-readable name for the rule.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Rule type. Supported types: query, eql, esql, machine_learning, new_terms, saved_query, threat_match, threshold.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("query", "eql", "esql", "machine_learning", "new_terms", "saved_query", "threat_match", "threshold"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"query": schema.StringAttribute{
				MarkdownDescription: "The query language definition.",
				Optional:            true,
			},
			"language": schema.StringAttribute{
				MarkdownDescription: "The query language (KQL or Lucene).",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("kuery", "lucene", "eql", "esql"),
				},
			},
			"index": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Indices on which the rule functions.",
				Optional:            true,
				Computed:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Determines whether the rule is enabled.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"from": schema.StringAttribute{
				MarkdownDescription: "Time from which data is analyzed each time the rule runs, using a date math range.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("now-6m"),
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^now-\d+[smhd]$`), "must be a valid date math expression like 'now-6m'"),
				},
			},
			"to": schema.StringAttribute{
				MarkdownDescription: "Time to which data is analyzed each time the rule runs, using a date math range.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("now"),
			},
			"interval": schema.StringAttribute{
				MarkdownDescription: "Frequency of rule execution, using a date math range.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("5m"),
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^\d+[smhd]$`), "must be a valid interval like '5m'"),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The rule's description.",
				Required:            true,
			},
			"risk_score": schema.Int64Attribute{
				MarkdownDescription: "A numerical representation of the alert's severity from 0 to 100.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(50),
				Validators: []validator.Int64{
					int64validator.Between(0, 100),
				},
			},
			"severity": schema.StringAttribute{
				MarkdownDescription: "Severity level of alerts produced by the rule.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("medium"),
				Validators: []validator.String{
					stringvalidator.OneOf("low", "medium", "high", "critical"),
				},
			},
			"author": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The rule's author.",
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "String array containing words and phrases to help categorize, filter, and search rules.",
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"license": schema.StringAttribute{
				MarkdownDescription: "The rule's license.",
				Optional:            true,
			},
			"false_positives": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "String array used to describe common reasons why the rule may issue false-positive alerts.",
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"references": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "String array containing references and URLs to sources of additional information.",
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"note": schema.StringAttribute{
				MarkdownDescription: "Notes to help investigate alerts produced by the rule.",
				Optional:            true,
			},
			"setup": schema.StringAttribute{
				MarkdownDescription: "Setup guide with instructions on rule prerequisites.",
				Optional:            true,
			},
			"max_signals": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of alerts the rule can create during a single run.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(100),
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"version": schema.Int64Attribute{
				MarkdownDescription: "The rule's version number.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			// Read-only fields
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The time the rule was created.",
				Computed:            true,
			},
			"created_by": schema.StringAttribute{
				MarkdownDescription: "The user who created the rule.",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "The time the rule was last updated.",
				Computed:            true,
			},
			"updated_by": schema.StringAttribute{
				MarkdownDescription: "The user who last updated the rule.",
				Computed:            true,
			},
			"revision": schema.Int64Attribute{
				MarkdownDescription: "The rule's revision number.",
				Computed:            true,
			},

			// EQL-specific fields
			"tiebreaker_field": schema.StringAttribute{
				MarkdownDescription: "Sets the tiebreaker field. Required for EQL rules when event.dataset is not provided.",
				Optional:            true,
			},

			// Machine Learning-specific fields
			"anomaly_threshold": schema.Int64Attribute{
				MarkdownDescription: "Anomaly score threshold above which the rule creates an alert. Valid values are from 0 to 100. Required for machine_learning rules.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.Between(0, 100),
				},
			},
			"machine_learning_job_id": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Machine learning job ID(s) the rule monitors for anomaly scores. Required for machine_learning rules.",
				Optional:            true,
			},

			// New Terms-specific fields
			"new_terms_fields": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Field names containing the new terms. Required for new_terms rules.",
				Optional:            true,
			},
			"history_window_start": schema.StringAttribute{
				MarkdownDescription: "Start date to use when checking if a term has been seen before. Supports relative dates like 'now-30d'. Required for new_terms rules.",
				Optional:            true,
			},

			// Saved Query-specific fields
			"saved_id": schema.StringAttribute{
				MarkdownDescription: "Identifier of the saved query used for the rule. Required for saved_query rules.",
				Optional:            true,
			},

			// Threat Match-specific fields
			"threat_index": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Array of index patterns for the threat intelligence indices. Required for threat_match rules.",
				Optional:            true,
			},
			"threat_query": schema.StringAttribute{
				MarkdownDescription: "Query used to filter threat intelligence data. Optional for threat_match rules.",
				Optional:            true,
			},
			"threat_mapping": schema.ListNestedAttribute{
				MarkdownDescription: "Array of threat mappings that specify how to match events with threat intelligence. Required for threat_match rules.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"entries": schema.ListNestedAttribute{
							MarkdownDescription: "Array of mapping entries.",
							Required:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"field": schema.StringAttribute{
										MarkdownDescription: "Event field to match.",
										Required:            true,
									},
									"type": schema.StringAttribute{
										MarkdownDescription: "Type of match (mapping).",
										Required:            true,
										Validators: []validator.String{
											stringvalidator.OneOf("mapping"),
										},
									},
									"value": schema.StringAttribute{
										MarkdownDescription: "Threat intelligence field to match against.",
										Required:            true,
									},
								},
							},
						},
					},
				},
			},
			"threat_filters": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Additional filters for threat intelligence data. Optional for threat_match rules.",
				Optional:            true,
			},
			"threat_indicator_path": schema.StringAttribute{
				MarkdownDescription: "Path to the threat indicator in the indicator documents. Optional for threat_match rules.",
				Optional:            true,
				Computed:            true,
			},
			"concurrent_searches": schema.Int64Attribute{
				MarkdownDescription: "Number of concurrent searches for threat intelligence. Optional for threat_match rules.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"items_per_search": schema.Int64Attribute{
				MarkdownDescription: "Number of items to search for in each concurrent search. Optional for threat_match rules.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},

			// Threshold-specific fields
			"threshold": schema.SingleNestedAttribute{
				MarkdownDescription: "Threshold settings for the rule. Required for threshold rules.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"field": schema.ListAttribute{
						ElementType:         types.StringType,
						MarkdownDescription: "Field(s) to use for threshold aggregation.",
						Optional:            true,
					},
					"value": schema.Int64Attribute{
						MarkdownDescription: "The threshold value from which an alert is generated.",
						Required:            true,
						Validators: []validator.Int64{
							int64validator.AtLeast(1),
						},
					},
					"cardinality": schema.ListNestedAttribute{
						MarkdownDescription: "Cardinality settings for threshold rule.",
						Optional:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"field": schema.StringAttribute{
									MarkdownDescription: "The field on which to calculate and compare the cardinality.",
									Required:            true,
								},
								"value": schema.Int64Attribute{
									MarkdownDescription: "The threshold cardinality value.",
									Required:            true,
									Validators: []validator.Int64{
										int64validator.AtLeast(1),
									},
								},
							},
						},
					},
				},
			},

			// Optional timeline fields (common across multiple rule types)
			"timeline_id": schema.StringAttribute{
				MarkdownDescription: "Timeline template ID for the rule.",
				Optional:            true,
			},
			"timeline_title": schema.StringAttribute{
				MarkdownDescription: "Timeline template title for the rule.",
				Optional:            true,
			},

			// Threat field (common across multiple rule types)
			"threat": schema.ListNestedAttribute{
				MarkdownDescription: "MITRE ATT&CK framework threat information.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"framework": schema.StringAttribute{
							MarkdownDescription: "Threat framework (typically 'MITRE ATT&CK').",
							Required:            true,
						},
						"tactic": schema.SingleNestedAttribute{
							MarkdownDescription: "MITRE ATT&CK tactic information.",
							Required:            true,
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									MarkdownDescription: "MITRE ATT&CK tactic ID.",
									Required:            true,
								},
								"name": schema.StringAttribute{
									MarkdownDescription: "MITRE ATT&CK tactic name.",
									Required:            true,
								},
								"reference": schema.StringAttribute{
									MarkdownDescription: "MITRE ATT&CK tactic reference URL.",
									Required:            true,
								},
							},
						},
						"technique": schema.ListNestedAttribute{
							MarkdownDescription: "MITRE ATT&CK technique information.",
							Optional:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										MarkdownDescription: "MITRE ATT&CK technique ID.",
										Required:            true,
									},
									"name": schema.StringAttribute{
										MarkdownDescription: "MITRE ATT&CK technique name.",
										Required:            true,
									},
									"reference": schema.StringAttribute{
										MarkdownDescription: "MITRE ATT&CK technique reference URL.",
										Required:            true,
									},
									"subtechnique": schema.ListNestedAttribute{
										MarkdownDescription: "MITRE ATT&CK sub-technique information.",
										Optional:            true,
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"id": schema.StringAttribute{
													MarkdownDescription: "MITRE ATT&CK sub-technique ID.",
													Required:            true,
												},
												"name": schema.StringAttribute{
													MarkdownDescription: "MITRE ATT&CK sub-technique name.",
													Required:            true,
												},
												"reference": schema.StringAttribute{
													MarkdownDescription: "MITRE ATT&CK sub-technique reference URL.",
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
		},
	}
}
