package security_detection_rule

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SecurityDetectionRuleData struct {
	Id        types.String `tfsdk:"id"`
	SpaceId   types.String `tfsdk:"space_id"`
	RuleId    types.String `tfsdk:"rule_id"`
	Name      types.String `tfsdk:"name"`
	Type      types.String `tfsdk:"type"`
	Query     types.String `tfsdk:"query"`
	Language  types.String `tfsdk:"language"`
	Index     types.List   `tfsdk:"index"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	From      types.String `tfsdk:"from"`
	To        types.String `tfsdk:"to"`
	Interval  types.String `tfsdk:"interval"`
	
	// Rule content
	Description types.String `tfsdk:"description"`
	RiskScore   types.Int64  `tfsdk:"risk_score"`
	Severity    types.String `tfsdk:"severity"`
	Author      types.List   `tfsdk:"author"`
	Tags        types.List   `tfsdk:"tags"`
	License     types.String `tfsdk:"license"`
	
	// Optional fields
	FalsePositives  types.List   `tfsdk:"false_positives"`
	References      types.List   `tfsdk:"references"`
	Note            types.String `tfsdk:"note"`
	Setup           types.String `tfsdk:"setup"`
	MaxSignals      types.Int64  `tfsdk:"max_signals"`
	Version         types.Int64  `tfsdk:"version"`
	
	// Read-only fields
	CreatedAt types.String `tfsdk:"created_at"`
	CreatedBy types.String `tfsdk:"created_by"`
	UpdatedAt types.String `tfsdk:"updated_at"`
	UpdatedBy types.String `tfsdk:"updated_by"`
	Revision  types.Int64  `tfsdk:"revision"`
}