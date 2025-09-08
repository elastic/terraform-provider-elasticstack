package detection_rule

import (
	"context"
	"regexp"

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

	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
)

func (r *securityDetectionRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetSchema()
}

func GetSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Creates or updates a Kibana security detection rule. See https://www.elastic.co/guide/en/security/current/rules-api-create.html",
		Blocks: map[string]schema.Block{
			"kibana_connection": providerschema.GetKbFWConnectionBlock(),
		},
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource",
				Computed:            true,
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
				MarkdownDescription: "The identifier for the rule. If not provided, an ID is randomly generated.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the detection rule.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the detection rule.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The rule type. Valid values are: eql, query, machine_learning, threshold, threat_match, new_terms.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("eql", "query", "machine_learning", "threshold", "threat_match", "new_terms"),
				},
			},
			"query": schema.StringAttribute{
				MarkdownDescription: "The query that the rule will use to generate alerts.",
				Optional:            true,
			},
			"language": schema.StringAttribute{
				MarkdownDescription: "The query language. Valid values are: kuery, lucene, eql.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("kuery"),
				Validators: []validator.String{
					stringvalidator.OneOf("kuery", "lucene", "eql"),
				},
			},
			"index": schema.ListAttribute{
				MarkdownDescription: "A list of index patterns to search.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{types.StringValue("*")})),
			},
			"severity": schema.StringAttribute{
				MarkdownDescription: "The severity of the rule. Valid values are: low, medium, high, critical.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("low", "medium", "high", "critical"),
				},
			},
			"risk": schema.Int64Attribute{
				MarkdownDescription: "A numerical representation of the alert's severity from 1-100.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(21),
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Determines whether the rule is enabled.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "String array containing words and phrases to help categorize, filter, and search rules.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"from": schema.StringAttribute{
				MarkdownDescription: "Time from which data is analyzed each time the rule executes, using date math syntax.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("now-6m"),
			},
			"to": schema.StringAttribute{
				MarkdownDescription: "Time to which data is analyzed each time the rule executes, using date math syntax.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("now"),
			},
			"interval": schema.StringAttribute{
				MarkdownDescription: "How often the rule executes.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("5m"),
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^\d+[smhd]$`), "must be a valid duration (e.g., '5m', '1h')"),
				},
			},
			"meta": schema.StringAttribute{
				MarkdownDescription: "Optional metadata about the rule as a JSON string.",
				Optional:            true,
			},
			"author": schema.ListAttribute{
				MarkdownDescription: "String array containing the rule's author(s).",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"license": schema.StringAttribute{
				MarkdownDescription: "The rule's license.",
				Optional:            true,
			},
			"rule_name_override": schema.StringAttribute{
				MarkdownDescription: "Sets the source field for the alert's rule name.",
				Optional:            true,
			},
			"timestamp_override": schema.StringAttribute{
				MarkdownDescription: "Sets the time field used to query indices.",
				Optional:            true,
			},
			"note": schema.StringAttribute{
				MarkdownDescription: "Notes to help investigate alerts produced by the rule.",
				Optional:            true,
			},
			"references": schema.ListAttribute{
				MarkdownDescription: "String array containing notes about or references to relevant information about the rule.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"false_positives": schema.ListAttribute{
				MarkdownDescription: "String array describing common reasons why the rule may issue false-positive alerts.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"exceptions_list": schema.ListAttribute{
				MarkdownDescription: "List of exceptions that prevent alerts from being generated.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"version": schema.Int64Attribute{
				MarkdownDescription: "The rule's version number.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
			},
			"max_signals": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of alerts the rule can produce during a single execution.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(100),
			},
		},
	}
}
