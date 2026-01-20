package security_detection_rule

import (
	"context"
	"regexp"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
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
			"data_view_id": schema.StringAttribute{
				MarkdownDescription: "Data view ID for the rule. Not supported for esql and machine_learning rule types.",
				Optional:            true,
				Validators: []validator.String{
					// Enforce that data_view_id is not set if the rule type is ml or esql
					validators.ForbiddenIfDependentPathOneOf(
						path.Root("type"),
						[]string{"machine_learning", "esql"},
					),
				},
			},
			"namespace": schema.StringAttribute{
				MarkdownDescription: "Alerts index namespace. Available for all rule types.",
				Optional:            true,
			},
			"rule_name_override": schema.StringAttribute{
				MarkdownDescription: "Override the rule name in Kibana. Available for all rule types.",
				Optional:            true,
			},
			"timestamp_override": schema.StringAttribute{
				MarkdownDescription: "Field name to use for timestamp override. Available for all rule types.",
				Optional:            true,
			},
			"timestamp_override_fallback_disabled": schema.BoolAttribute{
				MarkdownDescription: "Disables timestamp override fallback. Available for all rule types.",
				Optional:            true,
			},
			"query": schema.StringAttribute{
				MarkdownDescription: "The query language definition.",
				Optional:            true,
				Computed:            true,
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
				Validators: []validator.List{
					// Enforce that index is not set if the rule type is ml or esql
					validators.ForbiddenIfDependentPathOneOf(
						path.Root("type"),
						[]string{"machine_learning", "esql"},
					),
				},
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
			"risk_score_mapping": schema.ListNestedAttribute{
				MarkdownDescription: "Array of risk score mappings to override the default risk score based on source event field values.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"field": schema.StringAttribute{
							MarkdownDescription: "Source event field used to override the default risk_score.",
							Required:            true,
						},
						"operator": schema.StringAttribute{
							MarkdownDescription: "Operator to use for field value matching. Currently only 'equals' is supported.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("equals"),
							},
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "Value to match against the field.",
							Required:            true,
						},
						"risk_score": schema.Int64Attribute{
							MarkdownDescription: "Risk score to use when the field matches the value (0-100). If omitted, uses the rule's default risk_score.",
							Optional:            true,
							Validators: []validator.Int64{
								int64validator.Between(0, 100),
							},
						},
					},
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
			"severity_mapping": schema.ListNestedAttribute{
				MarkdownDescription: "Array of severity mappings to override the default severity based on source event field values.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"field": schema.StringAttribute{
							MarkdownDescription: "Source event field used to override the default severity.",
							Required:            true,
						},
						"operator": schema.StringAttribute{
							MarkdownDescription: "Operator to use for field value matching. Currently only 'equals' is supported.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("equals"),
							},
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "Value to match against the field.",
							Required:            true,
						},
						"severity": schema.StringAttribute{
							MarkdownDescription: "Severity level to use when the field matches the value.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("low", "medium", "high", "critical"),
							},
						},
					},
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
			"related_integrations": schema.ListNestedAttribute{
				MarkdownDescription: "Array of related integrations that provide additional context for the rule.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"package": schema.StringAttribute{
							MarkdownDescription: "Name of the integration package.",
							Required:            true,
						},
						"version": schema.StringAttribute{
							MarkdownDescription: "Version of the integration package.",
							Required:            true,
						},
						"integration": schema.StringAttribute{
							MarkdownDescription: "Name of the specific integration.",
							Optional:            true,
						},
					},
				},
			},
			"required_fields": schema.ListNestedAttribute{
				MarkdownDescription: "Array of Elasticsearch fields and types that must be present in source indices for the rule to function properly.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the Elasticsearch field.",
							Required:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Type of the Elasticsearch field.",
							Required:            true,
						},
						"ecs": schema.BoolAttribute{
							MarkdownDescription: "Indicates whether the field is ECS-compliant. This is computed by the backend based on the field name and type.",
							Computed:            true,
						},
					},
				},
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
			"investigation_fields": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Array of field names to include in alert investigation. Available for all rule types.",
				Optional:            true,
			},
			"filters": schema.StringAttribute{
				MarkdownDescription: "Query and filter context array to define alert conditions as JSON. Supports complex filter structures including bool queries, term filters, range filters, etc. Available for all rule types.",
				Optional:            true,
				CustomType:          jsontypes.NormalizedType{},
				Validators: []validator.String{
					validators.ForbiddenIfDependentPathOneOf(
						path.Root("type"),
						[]string{"machine_learning", "esql"},
					),
				},
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

			// Actions field (common across all rule types)
			"actions": schema.ListNestedAttribute{
				MarkdownDescription: "Array of automated actions taken when alerts are generated by the rule.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"action_type_id": schema.StringAttribute{
							MarkdownDescription: "The action type used for sending notifications (e.g., .slack, .email, .webhook, .pagerduty, etc.).",
							Required:            true,
						},
						"id": schema.StringAttribute{
							MarkdownDescription: "The connector ID.",
							Required:            true,
						},
						"params": schema.MapAttribute{
							ElementType:         types.StringType,
							MarkdownDescription: "Object containing the allowed connector fields, which varies according to the connector type. Simple string values can be specified directly. For nested objects or arrays, use `jsonencode()` to encode them as JSON strings - they will be automatically converted back to objects when sent to the API.",
							Required:            true,
						},
						"group": schema.StringAttribute{
							MarkdownDescription: "Optionally groups actions by use cases. Use 'default' for alert notifications.",
							Optional:            true,
						},
						"uuid": schema.StringAttribute{
							MarkdownDescription: "A unique identifier for the action.",
							Optional:            true,
							Computed:            true,
						},
						"alerts_filter": schema.MapAttribute{
							ElementType:         types.StringType,
							MarkdownDescription: "Object containing an action's conditional filters.",
							Optional:            true,
						},
						"frequency": schema.SingleNestedAttribute{
							MarkdownDescription: "The action frequency defines when the action runs.",
							Optional:            true,
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"notify_when": schema.StringAttribute{
									MarkdownDescription: "Defines how often rules run actions. Valid values: onActionGroupChange, onActiveAlert, onThrottleInterval.",
									Required:            true,
									Validators: []validator.String{
										stringvalidator.OneOf("onActionGroupChange", "onActiveAlert", "onThrottleInterval"),
									},
								},
								"summary": schema.BoolAttribute{
									MarkdownDescription: "Action summary indicates whether we will send a summary notification about all the generated alerts or notification per individual alert.",
									Required:            true,
								},
								"throttle": schema.StringAttribute{
									MarkdownDescription: "Time interval for throttling actions (e.g., '1h', '30m', 'no_actions', 'rule').",
									Required:            true,
								},
							},
						},
					},
				},
			},

			// Response actions field (common across all rule types)
			"response_actions": schema.ListNestedAttribute{
				MarkdownDescription: "Array of response actions to take when alerts are generated by the rule.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"action_type_id": schema.StringAttribute{
							MarkdownDescription: "The action type used for response actions (.osquery, .endpoint).",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf(".osquery", ".endpoint"),
							},
						},
						"params": schema.SingleNestedAttribute{
							MarkdownDescription: "Parameters for the response action. Structure varies based on action_type_id.",
							Required:            true,
							Attributes: map[string]schema.Attribute{
								// Osquery params
								"query": schema.StringAttribute{
									MarkdownDescription: "SQL query to run (osquery only). Example: 'SELECT * FROM processes;'",
									Optional:            true,
								},
								"pack_id": schema.StringAttribute{
									MarkdownDescription: "Query pack identifier (osquery only).",
									Optional:            true,
								},
								"saved_query_id": schema.StringAttribute{
									MarkdownDescription: "Saved query identifier (osquery only).",
									Optional:            true,
								},
								"timeout": schema.Int64Attribute{
									MarkdownDescription: "Timeout period in seconds (osquery only). Min: 60, Max: 900.",
									Optional:            true,
									Validators: []validator.Int64{
										int64validator.Between(60, 900),
									},
								},
								"ecs_mapping": schema.MapAttribute{
									ElementType:         types.StringType,
									MarkdownDescription: "Map Osquery results columns to ECS fields (osquery only).",
									Optional:            true,
								},
								"queries": schema.ListNestedAttribute{
									MarkdownDescription: "Array of queries to run (osquery only).",
									Optional:            true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"id": schema.StringAttribute{
												MarkdownDescription: "Query ID.",
												Required:            true,
											},
											"query": schema.StringAttribute{
												MarkdownDescription: "Query to run.",
												Required:            true,
											},
											"platform": schema.StringAttribute{
												MarkdownDescription: "Platform to run the query on.",
												Optional:            true,
											},
											"version": schema.StringAttribute{
												MarkdownDescription: "Query version.",
												Optional:            true,
											},
											"removed": schema.BoolAttribute{
												MarkdownDescription: "Whether the query is removed.",
												Optional:            true,
											},
											"snapshot": schema.BoolAttribute{
												MarkdownDescription: "Whether this is a snapshot query.",
												Optional:            true,
											},
											"ecs_mapping": schema.MapAttribute{
												ElementType:         types.StringType,
												MarkdownDescription: "ECS field mappings for this query.",
												Optional:            true,
											},
										},
									},
								},
								// Endpoint params - common command and comment
								"command": schema.StringAttribute{
									MarkdownDescription: "Command to run (endpoint only). Valid values: isolate, kill-process, suspend-process.",
									Optional:            true,
									Validators: []validator.String{
										stringvalidator.OneOf("isolate", "kill-process", "suspend-process"),
									},
								},
								"comment": schema.StringAttribute{
									MarkdownDescription: "Comment describing the action (endpoint only).",
									Optional:            true,
								},
								// Endpoint process params - for kill-process and suspend-process commands
								"config": schema.SingleNestedAttribute{
									MarkdownDescription: "Configuration for process commands (endpoint only).",
									Optional:            true,
									Attributes: map[string]schema.Attribute{
										"field": schema.StringAttribute{
											MarkdownDescription: "Field to use instead of process.pid.",
											Required:            true,
										},
										"overwrite": schema.BoolAttribute{
											MarkdownDescription: "Whether to overwrite field with process.pid.",
											Optional:            true,
											Computed:            true,
											Default:             booldefault.StaticBool(true),
										},
									},
								},
							},
						},
					},
				},
			},

			// Exceptions list field (common across all rule types)
			"exceptions_list": schema.ListNestedAttribute{
				MarkdownDescription: "Array of exception containers to prevent the rule from generating alerts.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The exception container ID.",
							Required:            true,
						},
						"list_id": schema.StringAttribute{
							MarkdownDescription: "The exception container's list ID.",
							Required:            true,
						},
						"namespace_type": schema.StringAttribute{
							MarkdownDescription: "The namespace type for the exception container.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("single", "agnostic"),
							},
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "The type of exception container.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("detection", "endpoint", "endpoint_events", "endpoint_host_isolation_exceptions", "endpoint_blocklists", "endpoint_trusted_apps"),
							},
						},
					},
				},
			},

			// Alert suppression field (common across all rule types)
			"alert_suppression": schema.SingleNestedAttribute{
				MarkdownDescription: "Defines alert suppression configuration to reduce duplicate alerts.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"group_by": schema.ListAttribute{
						MarkdownDescription: "Array of field names to group alerts by for suppression.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"duration": schema.StringAttribute{
						Description: "Duration for which alerts are suppressed.",
						Optional:    true,
						CustomType:  customtypes.DurationType{},
					},
					"missing_fields_strategy": schema.StringAttribute{
						MarkdownDescription: "Strategy for handling missing fields in suppression grouping: 'suppress' - only one alert will be created per suppress by bucket, 'doNotSuppress' - per each document a separate alert will be created.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("suppress", "doNotSuppress"),
						},
					},
				},
			},

			// Building block type field (common across all rule types)
			"building_block_type": schema.StringAttribute{
				MarkdownDescription: "Determines if the rule acts as a building block. If set, value must be `default`. Building-block alerts are not displayed in the UI by default and are used as a foundation for other rules.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("default"),
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
					validators.RequiredIfDependentPathEquals(
						path.Root("type"),
						"machine_learning",
					),
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
				Computed:            true,
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

// func getCardinalityType() map[string]attr.Type {
func getCardinalityType() attr.Type {
	return GetSchema().Attributes["threshold"].(schema.SingleNestedAttribute).Attributes["cardinality"].GetType().(attr.TypeWithElementType).ElementType()
}

// getThresholdType returns the attribute types for threshold objects
func getThresholdType() map[string]attr.Type {
	return GetSchema().Attributes["threshold"].GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}

// getAlertSuppressionType returns the attribute types for alert suppression objects
func getAlertSuppressionType() map[string]attr.Type {
	return GetSchema().Attributes["alert_suppression"].GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}

// getThreatElementType returns the element type for threat objects (MITRE ATT&CK framework)
func getThreatElementType() attr.Type {
	return GetSchema().Attributes["threat"].GetType().(attr.TypeWithElementType).ElementType()
}

func getThreatMappingElementType() attr.Type {
	return GetSchema().Attributes["threat_mapping"].GetType().(attr.TypeWithElementType).ElementType()
}

func getThreatMappingEntryElementType() attr.Type {
	threatMappingType := GetSchema().Attributes["threat_mapping"].GetType().(attr.TypeWithElementType).ElementType().(attr.TypeWithAttributeTypes)
	return threatMappingType.AttributeTypes()["entries"].(attr.TypeWithElementType).ElementType()
}

func getResponseActionElementType() attr.Type {
	return GetSchema().Attributes["response_actions"].GetType().(attr.TypeWithElementType).ElementType()
}

func getResponseActionParamsType() map[string]attr.Type {
	responseActionType := GetSchema().Attributes["response_actions"].GetType().(attr.TypeWithElementType).ElementType().(attr.TypeWithAttributeTypes)
	return responseActionType.AttributeTypes()["params"].(attr.TypeWithAttributeTypes).AttributeTypes()
}

func getOsqueryQueryElementType() attr.Type {
	responseActionType := GetSchema().Attributes["response_actions"].GetType().(attr.TypeWithElementType).ElementType().(attr.TypeWithAttributeTypes)
	paramsType := responseActionType.AttributeTypes()["params"].(attr.TypeWithAttributeTypes)
	return paramsType.AttributeTypes()["queries"].(attr.TypeWithElementType).ElementType()
}

func getEndpointProcessConfigType() map[string]attr.Type {
	responseActionType := GetSchema().Attributes["response_actions"].GetType().(attr.TypeWithElementType).ElementType().(attr.TypeWithAttributeTypes)
	paramsType := responseActionType.AttributeTypes()["params"].(attr.TypeWithAttributeTypes)
	return paramsType.AttributeTypes()["config"].(attr.TypeWithAttributeTypes).AttributeTypes()
}

func getActionElementType() attr.Type {
	return GetSchema().Attributes["actions"].GetType().(attr.TypeWithElementType).ElementType()
}

func getActionFrequencyType() map[string]attr.Type {
	actionType := GetSchema().Attributes["actions"].GetType().(attr.TypeWithElementType).ElementType().(attr.TypeWithAttributeTypes)
	return actionType.AttributeTypes()["frequency"].(attr.TypeWithAttributeTypes).AttributeTypes()
}

func getExceptionsListElementType() attr.Type {
	return GetSchema().Attributes["exceptions_list"].GetType().(attr.TypeWithElementType).ElementType()
}

func getRiskScoreMappingElementType() attr.Type {
	return GetSchema().Attributes["risk_score_mapping"].GetType().(attr.TypeWithElementType).ElementType()
}

func getRelatedIntegrationElementType() attr.Type {
	return GetSchema().Attributes["related_integrations"].GetType().(attr.TypeWithElementType).ElementType()
}

func getRequiredFieldElementType() attr.Type {
	return GetSchema().Attributes["required_fields"].GetType().(attr.TypeWithElementType).ElementType()
}

func getSeverityMappingElementType() attr.Type {
	return GetSchema().Attributes["severity_mapping"].GetType().(attr.TypeWithElementType).ElementType()
}

func getThreatTacticType() map[string]attr.Type {
	threatType := GetSchema().Attributes["threat"].GetType().(attr.TypeWithElementType).ElementType().(attr.TypeWithAttributeTypes)
	return threatType.AttributeTypes()["tactic"].(attr.TypeWithAttributeTypes).AttributeTypes()
}

func getThreatTechniqueElementType() attr.Type {
	threatType := GetSchema().Attributes["threat"].GetType().(attr.TypeWithElementType).ElementType().(attr.TypeWithAttributeTypes)
	return threatType.AttributeTypes()["technique"].(attr.TypeWithElementType).ElementType()
}

func getThreatSubtechniqueElementType() attr.Type {
	threatType := GetSchema().Attributes["threat"].GetType().(attr.TypeWithElementType).ElementType().(attr.TypeWithAttributeTypes)
	techniqueType := threatType.AttributeTypes()["technique"].(attr.TypeWithElementType).ElementType().(attr.TypeWithAttributeTypes)
	return techniqueType.AttributeTypes()["subtechnique"].(attr.TypeWithElementType).ElementType()
}

// ValidateConfig validates the configuration for a security detection rule resource.
// It ensures that the configuration meets the following requirements:
//
// - For rule types "esql" and "machine_learning", no additional validation is performed
// - For other rule types, exactly one of 'index' or 'data_view_id' must be specified
// - Both 'index' and 'data_view_id' cannot be set simultaneously
//
// The function adds appropriate error diagnostics if validation fails.
func (r securityDetectionRuleResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data SecurityDetectionRuleData

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.Type.ValueString() == "esql" || data.Type.ValueString() == "machine_learning" {
		return
	}

	if utils.IsKnown(data.Index) && utils.IsKnown(data.DataViewId) {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Both 'index' and 'data_view_id' cannot be set at the same time.",
		)

	}

	if data.Index.IsNull() && data.DataViewId.IsNull() {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"One of 'index' or 'data_view_id' must be set.",
		)
	}
}
