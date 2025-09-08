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
		MarkdownDescription: "Creates or updates a Kibana security detection rule. See https://www.elastic.co/guide/en/security/current/rules-api-create.html",
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
				MarkdownDescription: "A stable unique identifier for the rule object. If omitted, a UUID is generated.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
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
				MarkdownDescription: "Rule type. Currently only 'query' is supported.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("query"),
				Validators: []validator.String{
					stringvalidator.OneOf("query"),
				},
			},
			"query": schema.StringAttribute{
				MarkdownDescription: "The query language definition.",
				Required:            true,
			},
			"language": schema.StringAttribute{
				MarkdownDescription: "The query language (KQL or Lucene).",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("kuery"),
				Validators: []validator.String{
					stringvalidator.OneOf("kuery", "lucene"),
				},
			},
			"index": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Indices on which the rule functions.",
				Optional:            true,
				Computed:            true,
				// Default to empty list - will use Security Solution default indices
				Default: listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
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
		},
	}
}
