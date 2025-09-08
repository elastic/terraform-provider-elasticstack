package detection_rule

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SecurityDetectionRuleData struct {
	Id                 types.String `tfsdk:"id"`
	KibanaConnection   types.List   `tfsdk:"kibana_connection"`
	SpaceId            types.String `tfsdk:"space_id"`
	RuleId             types.String `tfsdk:"rule_id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	Type               types.String `tfsdk:"type"`
	Query              types.String `tfsdk:"query"`
	Language           types.String `tfsdk:"language"`
	Index              types.List   `tfsdk:"index"`
	Severity           types.String `tfsdk:"severity"`
	Risk               types.Int64  `tfsdk:"risk"`
	Enabled            types.Bool   `tfsdk:"enabled"`
	Tags               types.List   `tfsdk:"tags"`
	From               types.String `tfsdk:"from"`
	To                 types.String `tfsdk:"to"`
	Interval           types.String `tfsdk:"interval"`
	Meta               types.String `tfsdk:"meta"`
	Author             types.List   `tfsdk:"author"`
	License            types.String `tfsdk:"license"`
	RuleNameOverride   types.String `tfsdk:"rule_name_override"`
	TimestampOverride  types.String `tfsdk:"timestamp_override"`
	Note               types.String `tfsdk:"note"`
	References         types.List   `tfsdk:"references"`
	FalsePositives     types.List   `tfsdk:"false_positives"`
	ExceptionsList     types.List   `tfsdk:"exceptions_list"`
	Version            types.Int64  `tfsdk:"version"`
	MaxSignals         types.Int64  `tfsdk:"max_signals"`
}