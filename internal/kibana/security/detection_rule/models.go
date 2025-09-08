package detection_rule

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SecurityDetectionRuleData struct {
	Id                types.String `tfsdk:"id"`
	KibanaConnection  types.List   `tfsdk:"kibana_connection"`
	SpaceId           types.String `tfsdk:"space_id"`
	RuleId            types.String `tfsdk:"rule_id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	Type              types.String `tfsdk:"type"`
	Query             types.String `tfsdk:"query"`
	Language          types.String `tfsdk:"language"`
	Index             types.List   `tfsdk:"index"`
	Severity          types.String `tfsdk:"severity"`
	Risk              types.Int64  `tfsdk:"risk"`
	Enabled           types.Bool   `tfsdk:"enabled"`
	Tags              types.List   `tfsdk:"tags"`
	From              types.String `tfsdk:"from"`
	To                types.String `tfsdk:"to"`
	Interval          types.String `tfsdk:"interval"`
	Meta              types.String `tfsdk:"meta"`
	Author            types.List   `tfsdk:"author"`
	License           types.String `tfsdk:"license"`
	RuleNameOverride  types.String `tfsdk:"rule_name_override"`
	TimestampOverride types.String `tfsdk:"timestamp_override"`
	Note              types.String `tfsdk:"note"`
	References        types.List   `tfsdk:"references"`
	FalsePositives    types.List   `tfsdk:"false_positives"`
	ExceptionsList    types.List   `tfsdk:"exceptions_list"`
	Version           types.Int64  `tfsdk:"version"`
	MaxSignals        types.Int64  `tfsdk:"max_signals"`
}

// SecurityDetectionRuleRequest represents a security detection rule creation/update request
type SecurityDetectionRuleRequest struct {
	Name              string          `json:"name"`
	Description       string          `json:"description"`
	Type              string          `json:"type"`
	Query             *string         `json:"query,omitempty"`
	Language          *string         `json:"language,omitempty"`
	Index             []string        `json:"index,omitempty"`
	Severity          string          `json:"severity"`
	Risk              int             `json:"risk_score"`
	Enabled           bool            `json:"enabled"`
	Tags              []string        `json:"tags,omitempty"`
	From              string          `json:"from"`
	To                string          `json:"to"`
	Interval          string          `json:"interval"`
	Meta              *map[string]any `json:"meta,omitempty"`
	Author            []string        `json:"author,omitempty"`
	License           *string         `json:"license,omitempty"`
	RuleNameOverride  *string         `json:"rule_name_override,omitempty"`
	TimestampOverride *string         `json:"timestamp_override,omitempty"`
	Note              *string         `json:"note,omitempty"`
	References        []string        `json:"references,omitempty"`
	FalsePositives    []string        `json:"false_positives,omitempty"`
	ExceptionsList    []any           `json:"exceptions_list,omitempty"`
	Version           int             `json:"version"`
	MaxSignals        int             `json:"max_signals"`
}

// SecurityDetectionRuleResponse represents the API response for a security detection rule
type SecurityDetectionRuleResponse struct {
	ID                string          `json:"id"`
	Name              string          `json:"name"`
	Description       string          `json:"description"`
	Type              string          `json:"type"`
	Query             *string         `json:"query,omitempty"`
	Language          *string         `json:"language,omitempty"`
	Index             []string        `json:"index,omitempty"`
	Severity          string          `json:"severity"`
	Risk              int             `json:"risk_score"`
	Enabled           bool            `json:"enabled"`
	Tags              []string        `json:"tags,omitempty"`
	From              string          `json:"from"`
	To                string          `json:"to"`
	Interval          string          `json:"interval"`
	Meta              *map[string]any `json:"meta,omitempty"`
	Author            []string        `json:"author,omitempty"`
	License           *string         `json:"license,omitempty"`
	RuleNameOverride  *string         `json:"rule_name_override,omitempty"`
	TimestampOverride *string         `json:"timestamp_override,omitempty"`
	Note              *string         `json:"note,omitempty"`
	References        []string        `json:"references,omitempty"`
	FalsePositives    []string        `json:"false_positives,omitempty"`
	ExceptionsList    []any           `json:"exceptions_list,omitempty"`
	Version           int             `json:"version"`
	MaxSignals        int             `json:"max_signals"`
	CreatedAt         string          `json:"created_at"`
	CreatedBy         string          `json:"created_by"`
	UpdatedAt         string          `json:"updated_at"`
	UpdatedBy         string          `json:"updated_by"`
}
