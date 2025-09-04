package detection_rule

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DetectionRuleModel represents the Terraform data model for a Kibana security detection rule.
type DetectionRuleModel struct {
	// Core attributes
	ID          types.String `tfsdk:"id"`
	RuleID      types.String `tfsdk:"rule_id"`
	SpaceID     types.String `tfsdk:"space_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
	Enabled     types.Bool   `tfsdk:"enabled"`

	// Risk and severity
	RiskScore types.Int64  `tfsdk:"risk_score"`
	Severity  types.String `tfsdk:"severity"`

	// Metadata
	Tags           types.List   `tfsdk:"tags"`
	References     types.List   `tfsdk:"references"`
	FalsePositives types.List   `tfsdk:"false_positives"`
	Author         types.List   `tfsdk:"author"`
	License        types.String `tfsdk:"license"`
	Version        types.Int64  `tfsdk:"version"`

	// Execution settings
	MaxSignals types.Int64  `tfsdk:"max_signals"`
	Interval   types.String `tfsdk:"interval"`
	From       types.String `tfsdk:"from"`
	To         types.String `tfsdk:"to"`

	// Rule-specific fields
	Query      types.String `tfsdk:"query"`
	Language   types.String `tfsdk:"language"`
	Index      types.List   `tfsdk:"index"`
	DataViewID types.String `tfsdk:"data_view_id"`

	// Computed fields
	CreatedAt types.String `tfsdk:"created_at"`
	CreatedBy types.String `tfsdk:"created_by"`
	UpdatedAt types.String `tfsdk:"updated_at"`
	UpdatedBy types.String `tfsdk:"updated_by"`
	Revision  types.Int64  `tfsdk:"revision"`
}
