package detection_rule

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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

// getResourceSchema returns the schema for the Kibana security detection rule resource.
func getResourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Creates and manages a Kibana security detection rule. " +
			"Detection rules run periodically and search for suspicious activity across your data. " +
			"When a rule's conditions are met, it creates an alert.",
		MarkdownDescription: "Creates and manages a Kibana security detection rule. " +
			"Detection rules run periodically and search for suspicious activity across your data. " +
			"When a rule's conditions are met, it creates an alert." +
			"\n\n~> **Note:** This resource requires Kibana and the Security app to be available.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "The ID of the rule (UUID format).",
				MarkdownDescription: "The ID of the rule (UUID format).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"rule_id": schema.StringAttribute{
				Description: "A stable unique identifier for the rule object. " +
					"If not specified, a UUID will be generated automatically. " +
					"This should be unique across spaces and environments.",
				MarkdownDescription: "A stable unique identifier for the rule object. " +
					"If not specified, a UUID will be generated automatically. " +
					"This should be unique across spaces and environments.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				Description:         "The Kibana space ID where the rule should be created. Defaults to 'default'.",
				MarkdownDescription: "The Kibana space ID where the rule should be created. Defaults to 'default'.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("default"),
			},
			"name": schema.StringAttribute{
				Description:         "A human-readable name for the rule.",
				MarkdownDescription: "A human-readable name for the rule.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				Description:         "The rule's description.",
				MarkdownDescription: "The rule's description.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				Description: "The rule type. Valid values are: 'query', 'eql', 'threshold', " +
					"'threat_match', 'machine_learning', 'new_terms', 'esql'.",
				MarkdownDescription: "The rule type. Valid values are: `query`, `eql`, `threshold`, " +
					"`threat_match`, `machine_learning`, `new_terms`, `esql`.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"query",
						"eql",
						"threshold",
						"threat_match",
						"machine_learning",
						"new_terms",
						"esql",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description:         "Determines whether the rule is enabled. Defaults to true.",
				MarkdownDescription: "Determines whether the rule is enabled. Defaults to `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"risk_score": schema.Int64Attribute{
				Description: "A numerical representation of the alert's severity from 0 to 100, where: " +
					"0-21 represents low severity, 22-47 represents medium severity, " +
					"48-73 represents high severity, 74-100 represents critical severity.",
				MarkdownDescription: "A numerical representation of the alert's severity from 0 to 100, where: " +
					"`0-21` represents low severity, `22-47` represents medium severity, " +
					"`48-73` represents high severity, `74-100` represents critical severity.",
				Required:   true,
				Validators: []validator.Int64{
					// Add range validator 0-100
				},
			},
			"severity": schema.StringAttribute{
				Description: "Severity level of alerts produced by the rule. " +
					"Valid values are: 'low', 'medium', 'high', 'critical'.",
				MarkdownDescription: "Severity level of alerts produced by the rule. " +
					"Valid values are: `low`, `medium`, `high`, `critical`.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("low", "medium", "high", "critical"),
				},
			},
			"tags": schema.ListAttribute{
				Description:         "String array containing words and phrases to help categorize, filter, and search rules.",
				MarkdownDescription: "String array containing words and phrases to help categorize, filter, and search rules.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"references": schema.ListAttribute{
				Description:         "Array containing notes about or references to relevant information about the rule.",
				MarkdownDescription: "Array containing notes about or references to relevant information about the rule.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"false_positives": schema.ListAttribute{
				Description:         "String array used to describe common reasons why the rule may issue false-positive alerts.",
				MarkdownDescription: "String array used to describe common reasons why the rule may issue false-positive alerts.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"author": schema.ListAttribute{
				Description:         "The rule's author.",
				MarkdownDescription: "The rule's author.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"license": schema.StringAttribute{
				Description:         "The rule's license.",
				MarkdownDescription: "The rule's license.",
				Optional:            true,
			},
			"version": schema.Int64Attribute{
				Description:         "The rule's version number. Defaults to 1 for new rules.",
				MarkdownDescription: "The rule's version number. Defaults to `1` for new rules.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
			},
			"max_signals": schema.Int64Attribute{
				Description: "Maximum number of alerts the rule can create during a single run. " +
					"Defaults to 100.",
				MarkdownDescription: "Maximum number of alerts the rule can create during a single run. " +
					"Defaults to `100`.",
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(100),
			},
			"interval": schema.StringAttribute{
				Description: "Frequency of rule execution using date math notation (e.g., '5m', '1h'). " +
					"Defaults to '5m'.",
				MarkdownDescription: "Frequency of rule execution using date math notation (e.g., `5m`, `1h`). " +
					"Defaults to `5m`.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("5m"),
			},
			"from": schema.StringAttribute{
				Description: "Time from which data is analyzed each time the rule runs, using date math notation " +
					"(e.g., 'now-6m'). Defaults to 'now-6m'.",
				MarkdownDescription: "Time from which data is analyzed each time the rule runs, using date math notation " +
					"(e.g., `now-6m`). Defaults to `now-6m`.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("now-6m"),
			},
			"to": schema.StringAttribute{
				Description: "Time range end for rule execution using date math notation. " +
					"Defaults to 'now'.",
				MarkdownDescription: "Time range end for rule execution using date math notation. " +
					"Defaults to `now`.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("now"),
			},

			// Rule-type specific fields
			"query": schema.StringAttribute{
				Description: "The query used by the rule to create alerts. " +
					"Required for query, eql, and esql rule types.",
				MarkdownDescription: "The query used by the rule to create alerts. " +
					"Required for `query`, `eql`, and `esql` rule types.",
				Optional: true,
			},
			"language": schema.StringAttribute{
				Description: "The query language. Valid values are 'kuery', 'lucene', 'eql', 'esql'. " +
					"Required for query, eql, and esql rule types.",
				MarkdownDescription: "The query language. Valid values are `kuery`, `lucene`, `eql`, `esql`. " +
					"Required for `query`, `eql`, and `esql` rule types.",
				Optional: true,
				Validators: []validator.String{
					stringvalidator.OneOf("kuery", "lucene", "eql", "esql"),
				},
			},
			"index": schema.ListAttribute{
				Description: "Indices on which the rule functions. " +
					"Defaults to the Security Solution indices defined in Kibana settings.",
				MarkdownDescription: "Indices on which the rule functions. " +
					"Defaults to the Security Solution indices defined in Kibana settings.",
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"data_view_id": schema.StringAttribute{
				Description:         "Data view ID for the rule.",
				MarkdownDescription: "Data view ID for the rule.",
				Optional:            true,
			},

			// Computed fields from API
			"created_at": schema.StringAttribute{
				Description:         "The date and time the rule was created.",
				MarkdownDescription: "The date and time the rule was created.",
				Computed:            true,
			},
			"created_by": schema.StringAttribute{
				Description:         "The user who created the rule.",
				MarkdownDescription: "The user who created the rule.",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				Description:         "The date and time the rule was last updated.",
				MarkdownDescription: "The date and time the rule was last updated.",
				Computed:            true,
			},
			"updated_by": schema.StringAttribute{
				Description:         "The user who last updated the rule.",
				MarkdownDescription: "The user who last updated the rule.",
				Computed:            true,
			},
			"revision": schema.Int64Attribute{
				Description:         "The rule's revision number (incremented on each update).",
				MarkdownDescription: "The rule's revision number (incremented on each update).",
				Computed:            true,
			},
		},
	}
}
