package security_detection_rule

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type SecurityDetectionRuleData struct {
	Id       types.String `tfsdk:"id"`
	SpaceId  types.String `tfsdk:"space_id"`
	RuleId   types.String `tfsdk:"rule_id"`
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Query    types.String `tfsdk:"query"`
	Language types.String `tfsdk:"language"`
	Index    types.List   `tfsdk:"index"`
	Enabled  types.Bool   `tfsdk:"enabled"`
	From     types.String `tfsdk:"from"`
	To       types.String `tfsdk:"to"`
	Interval types.String `tfsdk:"interval"`

	// Rule content
	Description types.String `tfsdk:"description"`
	RiskScore   types.Int64  `tfsdk:"risk_score"`
	Severity    types.String `tfsdk:"severity"`
	Author      types.List   `tfsdk:"author"`
	Tags        types.List   `tfsdk:"tags"`
	License     types.String `tfsdk:"license"`

	// Optional fields
	FalsePositives types.List   `tfsdk:"false_positives"`
	References     types.List   `tfsdk:"references"`
	Note           types.String `tfsdk:"note"`
	Setup          types.String `tfsdk:"setup"`
	MaxSignals     types.Int64  `tfsdk:"max_signals"`
	Version        types.Int64  `tfsdk:"version"`

	// Read-only fields
	CreatedAt types.String `tfsdk:"created_at"`
	CreatedBy types.String `tfsdk:"created_by"`
	UpdatedAt types.String `tfsdk:"updated_at"`
	UpdatedBy types.String `tfsdk:"updated_by"`
	Revision  types.Int64  `tfsdk:"revision"`

	// EQL-specific fields
	TiebreakerField types.String `tfsdk:"tiebreaker_field"`

	// Machine Learning-specific fields
	AnomalyThreshold     types.Int64 `tfsdk:"anomaly_threshold"`
	MachineLearningJobId types.List  `tfsdk:"machine_learning_job_id"`

	// New Terms-specific fields
	NewTermsFields     types.List   `tfsdk:"new_terms_fields"`
	HistoryWindowStart types.String `tfsdk:"history_window_start"`

	// Saved Query-specific fields
	SavedId types.String `tfsdk:"saved_id"`

	// Threat Match-specific fields
	ThreatIndex         types.List   `tfsdk:"threat_index"`
	ThreatQuery         types.String `tfsdk:"threat_query"`
	ThreatMapping       types.List   `tfsdk:"threat_mapping"`
	ThreatFilters       types.List   `tfsdk:"threat_filters"`
	ThreatIndicatorPath types.String `tfsdk:"threat_indicator_path"`
	ConcurrentSearches  types.Int64  `tfsdk:"concurrent_searches"`
	ItemsPerSearch      types.Int64  `tfsdk:"items_per_search"`

	// Threshold-specific fields
	Threshold types.Object `tfsdk:"threshold"`

	// Optional timeline fields (common across multiple rule types)
	TimelineId    types.String `tfsdk:"timeline_id"`
	TimelineTitle types.String `tfsdk:"timeline_title"`

	// Threat field (common across multiple rule types)
	Threat types.List `tfsdk:"threat"`
}

func (d SecurityDetectionRuleData) toCreateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	ruleType := d.Type.ValueString()

	switch ruleType {
	case "query":
		return d.toQueryRuleCreateProps(ctx)
	case "eql":
		return d.toEqlRuleCreateProps(ctx)
	case "esql":
		return d.toEsqlRuleCreateProps(ctx)
	case "machine_learning":
		return d.toMachineLearningRuleCreateProps(ctx)
	case "new_terms":
		return d.toNewTermsRuleCreateProps(ctx)
	case "saved_query":
		return d.toSavedQueryRuleCreateProps(ctx)
	case "threat_match":
		return d.toThreatMatchRuleCreateProps(ctx)
	case "threshold":
		return d.toThresholdRuleCreateProps(ctx)
	default:
		diags.AddError(
			"Unsupported rule type",
			fmt.Sprintf("Rule type '%s' is not supported", ruleType),
		)
		return createProps, diags
	}
}

func (d SecurityDetectionRuleData) toQueryRuleCreateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	queryRuleQuery := kbapi.SecurityDetectionsAPIRuleQuery(d.Query.ValueString())
	queryRule := kbapi.SecurityDetectionsAPIQueryRuleCreateProps{
		Name:        kbapi.SecurityDetectionsAPIRuleName(d.Name.ValueString()),
		Description: kbapi.SecurityDetectionsAPIRuleDescription(d.Description.ValueString()),
		Type:        kbapi.SecurityDetectionsAPIQueryRuleCreatePropsType("query"),
		Query:       &queryRuleQuery,
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	d.setCommonCreateProps(ctx, &queryRule.Actions, &queryRule.RuleId, &queryRule.Enabled, &queryRule.From, &queryRule.To, &queryRule.Interval, &queryRule.Index, &queryRule.Author, &queryRule.Tags, &queryRule.FalsePositives, &queryRule.References, &queryRule.License, &queryRule.Note, &queryRule.Setup, &queryRule.MaxSignals, &queryRule.Version, &diags)

	// Set query-specific fields
	if utils.IsKnown(d.Language) {
		var language kbapi.SecurityDetectionsAPIKqlQueryLanguage
		switch d.Language.ValueString() {
		case "kuery":
			language = "kuery"
		case "lucene":
			language = "lucene"
		default:
			language = "kuery"
		}
		queryRule.Language = &language
	}

	if utils.IsKnown(d.SavedId) {
		savedId := kbapi.SecurityDetectionsAPISavedQueryId(d.SavedId.ValueString())
		queryRule.SavedId = &savedId
	}

	// Convert to union type
	err := createProps.FromSecurityDetectionsAPIQueryRuleCreateProps(queryRule)
	if err != nil {
		diags.AddError(
			"Error building create properties",
			"Could not convert query rule properties: "+err.Error(),
		)
	}

	return createProps, diags
}

func (d SecurityDetectionRuleData) toEqlRuleCreateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	eqlRule := kbapi.SecurityDetectionsAPIEqlRuleCreateProps{
		Name:        kbapi.SecurityDetectionsAPIRuleName(d.Name.ValueString()),
		Description: kbapi.SecurityDetectionsAPIRuleDescription(d.Description.ValueString()),
		Type:        kbapi.SecurityDetectionsAPIEqlRuleCreatePropsType("eql"),
		Query:       kbapi.SecurityDetectionsAPIRuleQuery(d.Query.ValueString()),
		Language:    kbapi.SecurityDetectionsAPIEqlQueryLanguage("eql"),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	d.setCommonCreateProps(ctx, &eqlRule.Actions, &eqlRule.RuleId, &eqlRule.Enabled, &eqlRule.From, &eqlRule.To, &eqlRule.Interval, &eqlRule.Index, &eqlRule.Author, &eqlRule.Tags, &eqlRule.FalsePositives, &eqlRule.References, &eqlRule.License, &eqlRule.Note, &eqlRule.Setup, &eqlRule.MaxSignals, &eqlRule.Version, &diags)

	// Set EQL-specific fields
	if utils.IsKnown(d.TiebreakerField) {
		tiebreakerField := kbapi.SecurityDetectionsAPITiebreakerField(d.TiebreakerField.ValueString())
		eqlRule.TiebreakerField = &tiebreakerField
	}

	// Convert to union type
	err := createProps.FromSecurityDetectionsAPIEqlRuleCreateProps(eqlRule)
	if err != nil {
		diags.AddError(
			"Error building create properties",
			"Could not convert EQL rule properties: "+err.Error(),
		)
	}

	return createProps, diags
}

func (d SecurityDetectionRuleData) toEsqlRuleCreateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	esqlRule := kbapi.SecurityDetectionsAPIEsqlRuleCreateProps{
		Name:        kbapi.SecurityDetectionsAPIRuleName(d.Name.ValueString()),
		Description: kbapi.SecurityDetectionsAPIRuleDescription(d.Description.ValueString()),
		Type:        kbapi.SecurityDetectionsAPIEsqlRuleCreatePropsType("esql"),
		Query:       kbapi.SecurityDetectionsAPIRuleQuery(d.Query.ValueString()),
		Language:    kbapi.SecurityDetectionsAPIEsqlQueryLanguage("esql"),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	d.setCommonCreateProps(ctx, &esqlRule.Actions, &esqlRule.RuleId, &esqlRule.Enabled, &esqlRule.From, &esqlRule.To, &esqlRule.Interval, nil, &esqlRule.Author, &esqlRule.Tags, &esqlRule.FalsePositives, &esqlRule.References, &esqlRule.License, &esqlRule.Note, &esqlRule.Setup, &esqlRule.MaxSignals, &esqlRule.Version, &diags)

	// ESQL rules don't use index patterns as they use FROM clause in the query

	// Convert to union type
	err := createProps.FromSecurityDetectionsAPIEsqlRuleCreateProps(esqlRule)
	if err != nil {
		diags.AddError(
			"Error building create properties",
			"Could not convert ESQL rule properties: "+err.Error(),
		)
	}

	return createProps, diags
}

func (d SecurityDetectionRuleData) toMachineLearningRuleCreateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	mlRule := kbapi.SecurityDetectionsAPIMachineLearningRuleCreateProps{
		Name:             kbapi.SecurityDetectionsAPIRuleName(d.Name.ValueString()),
		Description:      kbapi.SecurityDetectionsAPIRuleDescription(d.Description.ValueString()),
		Type:             kbapi.SecurityDetectionsAPIMachineLearningRuleCreatePropsType("machine_learning"),
		AnomalyThreshold: kbapi.SecurityDetectionsAPIAnomalyThreshold(d.AnomalyThreshold.ValueInt64()),
		RiskScore:        kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:         kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// Set ML job ID(s) - can be single string or array
	if utils.IsKnown(d.MachineLearningJobId) {
		jobIds := utils.ListTypeAs[string](ctx, d.MachineLearningJobId, path.Root("machine_learning_job_id"), &diags)
		if !diags.HasError() {
			if len(jobIds) == 1 {
				// Single job ID
				var mlJobId kbapi.SecurityDetectionsAPIMachineLearningJobId
				err := mlJobId.FromSecurityDetectionsAPIMachineLearningJobId0(jobIds[0])
				if err != nil {
					diags.AddError("Error setting ML job ID", err.Error())
				} else {
					mlRule.MachineLearningJobId = mlJobId
				}
			} else if len(jobIds) > 1 {
				// Multiple job IDs
				var mlJobId kbapi.SecurityDetectionsAPIMachineLearningJobId
				err := mlJobId.FromSecurityDetectionsAPIMachineLearningJobId1(jobIds)
				if err != nil {
					diags.AddError("Error setting ML job IDs", err.Error())
				} else {
					mlRule.MachineLearningJobId = mlJobId
				}
			}
		}
	}

	d.setCommonCreateProps(ctx, &mlRule.Actions, &mlRule.RuleId, &mlRule.Enabled, &mlRule.From, &mlRule.To, &mlRule.Interval, nil, &mlRule.Author, &mlRule.Tags, &mlRule.FalsePositives, &mlRule.References, &mlRule.License, &mlRule.Note, &mlRule.Setup, &mlRule.MaxSignals, &mlRule.Version, &diags)

	// ML rules don't use index patterns or query

	// Convert to union type
	err := createProps.FromSecurityDetectionsAPIMachineLearningRuleCreateProps(mlRule)
	if err != nil {
		diags.AddError(
			"Error building create properties",
			"Could not convert ML rule properties: "+err.Error(),
		)
	}

	return createProps, diags
}

func (d SecurityDetectionRuleData) toNewTermsRuleCreateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	newTermsRule := kbapi.SecurityDetectionsAPINewTermsRuleCreateProps{
		Name:               kbapi.SecurityDetectionsAPIRuleName(d.Name.ValueString()),
		Description:        kbapi.SecurityDetectionsAPIRuleDescription(d.Description.ValueString()),
		Type:               kbapi.SecurityDetectionsAPINewTermsRuleCreatePropsType("new_terms"),
		Query:              kbapi.SecurityDetectionsAPIRuleQuery(d.Query.ValueString()),
		HistoryWindowStart: kbapi.SecurityDetectionsAPIHistoryWindowStart(d.HistoryWindowStart.ValueString()),
		RiskScore:          kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:           kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// Set new terms fields
	if utils.IsKnown(d.NewTermsFields) {
		newTermsFields := utils.ListTypeAs[string](ctx, d.NewTermsFields, path.Root("new_terms_fields"), &diags)
		if !diags.HasError() {
			newTermsRule.NewTermsFields = newTermsFields
		}
	}

	d.setCommonCreateProps(ctx, &newTermsRule.Actions, &newTermsRule.RuleId, &newTermsRule.Enabled, &newTermsRule.From, &newTermsRule.To, &newTermsRule.Interval, &newTermsRule.Index, &newTermsRule.Author, &newTermsRule.Tags, &newTermsRule.FalsePositives, &newTermsRule.References, &newTermsRule.License, &newTermsRule.Note, &newTermsRule.Setup, &newTermsRule.MaxSignals, &newTermsRule.Version, &diags)

	// Set query language
	if utils.IsKnown(d.Language) {
		var language kbapi.SecurityDetectionsAPIKqlQueryLanguage
		switch d.Language.ValueString() {
		case "kuery":
			language = "kuery"
		case "lucene":
			language = "lucene"
		default:
			language = "kuery"
		}
		newTermsRule.Language = &language
	}

	// Convert to union type
	err := createProps.FromSecurityDetectionsAPINewTermsRuleCreateProps(newTermsRule)
	if err != nil {
		diags.AddError(
			"Error building create properties",
			"Could not convert new terms rule properties: "+err.Error(),
		)
	}

	return createProps, diags
}

func (d SecurityDetectionRuleData) toSavedQueryRuleCreateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	savedQueryRule := kbapi.SecurityDetectionsAPISavedQueryRuleCreateProps{
		Name:        kbapi.SecurityDetectionsAPIRuleName(d.Name.ValueString()),
		Description: kbapi.SecurityDetectionsAPIRuleDescription(d.Description.ValueString()),
		Type:        kbapi.SecurityDetectionsAPISavedQueryRuleCreatePropsType("saved_query"),
		SavedId:     kbapi.SecurityDetectionsAPISavedQueryId(d.SavedId.ValueString()),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	d.setCommonCreateProps(ctx, &savedQueryRule.Actions, &savedQueryRule.RuleId, &savedQueryRule.Enabled, &savedQueryRule.From, &savedQueryRule.To, &savedQueryRule.Interval, &savedQueryRule.Index, &savedQueryRule.Author, &savedQueryRule.Tags, &savedQueryRule.FalsePositives, &savedQueryRule.References, &savedQueryRule.License, &savedQueryRule.Note, &savedQueryRule.Setup, &savedQueryRule.MaxSignals, &savedQueryRule.Version, &diags)

	// Set optional query for saved query rules
	if utils.IsKnown(d.Query) {
		query := kbapi.SecurityDetectionsAPIRuleQuery(d.Query.ValueString())
		savedQueryRule.Query = &query
	}

	// Set query language
	if utils.IsKnown(d.Language) {
		var language kbapi.SecurityDetectionsAPIKqlQueryLanguage
		switch d.Language.ValueString() {
		case "kuery":
			language = "kuery"
		case "lucene":
			language = "lucene"
		default:
			language = "kuery"
		}
		savedQueryRule.Language = &language
	}

	// Convert to union type
	err := createProps.FromSecurityDetectionsAPISavedQueryRuleCreateProps(savedQueryRule)
	if err != nil {
		diags.AddError(
			"Error building create properties",
			"Could not convert saved query rule properties: "+err.Error(),
		)
	}

	return createProps, diags
}

func (d SecurityDetectionRuleData) toThreatMatchRuleCreateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	threatMatchRule := kbapi.SecurityDetectionsAPIThreatMatchRuleCreateProps{
		Name:        kbapi.SecurityDetectionsAPIRuleName(d.Name.ValueString()),
		Description: kbapi.SecurityDetectionsAPIRuleDescription(d.Description.ValueString()),
		Type:        kbapi.SecurityDetectionsAPIThreatMatchRuleCreatePropsType("threat_match"),
		Query:       kbapi.SecurityDetectionsAPIRuleQuery(d.Query.ValueString()),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// Set threat index
	if utils.IsKnown(d.ThreatIndex) {
		threatIndex := utils.ListTypeAs[string](ctx, d.ThreatIndex, path.Root("threat_index"), &diags)
		if !diags.HasError() {
			threatMatchRule.ThreatIndex = threatIndex
		}
	}

	d.setCommonCreateProps(ctx, &threatMatchRule.Actions, &threatMatchRule.RuleId, &threatMatchRule.Enabled, &threatMatchRule.From, &threatMatchRule.To, &threatMatchRule.Interval, &threatMatchRule.Index, &threatMatchRule.Author, &threatMatchRule.Tags, &threatMatchRule.FalsePositives, &threatMatchRule.References, &threatMatchRule.License, &threatMatchRule.Note, &threatMatchRule.Setup, &threatMatchRule.MaxSignals, &threatMatchRule.Version, &diags)

	// Set threat-specific fields
	if utils.IsKnown(d.ThreatQuery) {
		threatMatchRule.ThreatQuery = kbapi.SecurityDetectionsAPIThreatQuery(d.ThreatQuery.ValueString())
	}

	if utils.IsKnown(d.ThreatIndicatorPath) {
		threatIndicatorPath := kbapi.SecurityDetectionsAPIThreatIndicatorPath(d.ThreatIndicatorPath.ValueString())
		threatMatchRule.ThreatIndicatorPath = &threatIndicatorPath
	}

	if utils.IsKnown(d.ConcurrentSearches) {
		concurrentSearches := kbapi.SecurityDetectionsAPIConcurrentSearches(d.ConcurrentSearches.ValueInt64())
		threatMatchRule.ConcurrentSearches = &concurrentSearches
	}

	if utils.IsKnown(d.ItemsPerSearch) {
		itemsPerSearch := kbapi.SecurityDetectionsAPIItemsPerSearch(d.ItemsPerSearch.ValueInt64())
		threatMatchRule.ItemsPerSearch = &itemsPerSearch
	}

	// Set query language
	if utils.IsKnown(d.Language) {
		var language kbapi.SecurityDetectionsAPIKqlQueryLanguage
		switch d.Language.ValueString() {
		case "kuery":
			language = "kuery"
		case "lucene":
			language = "lucene"
		default:
			language = "kuery"
		}
		threatMatchRule.Language = &language
	}

	if utils.IsKnown(d.SavedId) {
		savedId := kbapi.SecurityDetectionsAPISavedQueryId(d.SavedId.ValueString())
		threatMatchRule.SavedId = &savedId
	}

	// Convert to union type
	err := createProps.FromSecurityDetectionsAPIThreatMatchRuleCreateProps(threatMatchRule)
	if err != nil {
		diags.AddError(
			"Error building create properties",
			"Could not convert threat match rule properties: "+err.Error(),
		)
	}

	return createProps, diags
}

func (d SecurityDetectionRuleData) toThresholdRuleCreateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	thresholdRule := kbapi.SecurityDetectionsAPIThresholdRuleCreateProps{
		Name:        kbapi.SecurityDetectionsAPIRuleName(d.Name.ValueString()),
		Description: kbapi.SecurityDetectionsAPIRuleDescription(d.Description.ValueString()),
		Type:        kbapi.SecurityDetectionsAPIThresholdRuleCreatePropsType("threshold"),
		Query:       kbapi.SecurityDetectionsAPIRuleQuery(d.Query.ValueString()),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// Set threshold - this is required for threshold rules
	if utils.IsKnown(d.Threshold) {
		// Parse threshold object
		var thresholdAttrs map[string]attr.Value
		diag := d.Threshold.As(ctx, &thresholdAttrs, basetypes.ObjectAsOptions{})
		diags.Append(diag...)
		if !diags.HasError() {
			threshold := kbapi.SecurityDetectionsAPIThreshold{}

			if valueAttr, ok := thresholdAttrs["value"]; ok && utils.IsKnown(valueAttr.(types.Int64)) {
				threshold.Value = kbapi.SecurityDetectionsAPIThresholdValue(valueAttr.(types.Int64).ValueInt64())
			}

			if fieldAttr, ok := thresholdAttrs["field"]; ok && utils.IsKnown(fieldAttr.(types.List)) {
				fieldList := utils.ListTypeAs[string](ctx, fieldAttr.(types.List), path.Root("threshold").AtName("field"), &diags)
				if !diags.HasError() && len(fieldList) > 0 {
					var thresholdField kbapi.SecurityDetectionsAPIThresholdField
					if len(fieldList) == 1 {
						err := thresholdField.FromSecurityDetectionsAPIThresholdField0(fieldList[0])
						if err != nil {
							diags.AddError("Error setting threshold field", err.Error())
						} else {
							threshold.Field = thresholdField
						}
					} else {
						err := thresholdField.FromSecurityDetectionsAPIThresholdField1(fieldList)
						if err != nil {
							diags.AddError("Error setting threshold fields", err.Error())
						} else {
							threshold.Field = thresholdField
						}
					}
				}
			}

			thresholdRule.Threshold = threshold
		}
	}

	d.setCommonCreateProps(ctx, &thresholdRule.Actions, &thresholdRule.RuleId, &thresholdRule.Enabled, &thresholdRule.From, &thresholdRule.To, &thresholdRule.Interval, &thresholdRule.Index, &thresholdRule.Author, &thresholdRule.Tags, &thresholdRule.FalsePositives, &thresholdRule.References, &thresholdRule.License, &thresholdRule.Note, &thresholdRule.Setup, &thresholdRule.MaxSignals, &thresholdRule.Version, &diags)

	// Set query language
	if utils.IsKnown(d.Language) {
		var language kbapi.SecurityDetectionsAPIKqlQueryLanguage
		switch d.Language.ValueString() {
		case "kuery":
			language = "kuery"
		case "lucene":
			language = "lucene"
		default:
			language = "kuery"
		}
		thresholdRule.Language = &language
	}

	if utils.IsKnown(d.SavedId) {
		savedId := kbapi.SecurityDetectionsAPISavedQueryId(d.SavedId.ValueString())
		thresholdRule.SavedId = &savedId
	}

	// Convert to union type
	err := createProps.FromSecurityDetectionsAPIThresholdRuleCreateProps(thresholdRule)
	if err != nil {
		diags.AddError(
			"Error building create properties",
			"Could not convert threshold rule properties: "+err.Error(),
		)
	}

	return createProps, diags
}

// Helper function to set common properties across all rule types
func (d SecurityDetectionRuleData) setCommonCreateProps(
	ctx context.Context,
	actions **[]kbapi.SecurityDetectionsAPIRuleAction,
	ruleId **kbapi.SecurityDetectionsAPIRuleSignatureId,
	enabled **kbapi.SecurityDetectionsAPIIsRuleEnabled,
	from **kbapi.SecurityDetectionsAPIRuleIntervalFrom,
	to **kbapi.SecurityDetectionsAPIRuleIntervalTo,
	interval **kbapi.SecurityDetectionsAPIRuleInterval,
	index **[]string,
	author **[]string,
	tags **[]string,
	falsePositives **[]string,
	references **[]string,
	license **kbapi.SecurityDetectionsAPIRuleLicense,
	note **kbapi.SecurityDetectionsAPIInvestigationGuide,
	setup **kbapi.SecurityDetectionsAPISetupGuide,
	maxSignals **kbapi.SecurityDetectionsAPIMaxSignals,
	version **kbapi.SecurityDetectionsAPIRuleVersion,
	diags *diag.Diagnostics,
) {
	// Set optional rule_id if provided
	if utils.IsKnown(d.RuleId) {
		id := kbapi.SecurityDetectionsAPIRuleSignatureId(d.RuleId.ValueString())
		*ruleId = &id
	}

	// Set enabled status
	if utils.IsKnown(d.Enabled) {
		isEnabled := kbapi.SecurityDetectionsAPIIsRuleEnabled(d.Enabled.ValueBool())
		*enabled = &isEnabled
	}

	// Set time range
	if utils.IsKnown(d.From) {
		fromTime := kbapi.SecurityDetectionsAPIRuleIntervalFrom(d.From.ValueString())
		*from = &fromTime
	}

	if utils.IsKnown(d.To) {
		toTime := kbapi.SecurityDetectionsAPIRuleIntervalTo(d.To.ValueString())
		*to = &toTime
	}

	// Set interval
	if utils.IsKnown(d.Interval) {
		intervalTime := kbapi.SecurityDetectionsAPIRuleInterval(d.Interval.ValueString())
		*interval = &intervalTime
	}

	// Set index patterns (if index pointer is provided)
	if index != nil && utils.IsKnown(d.Index) {
		indexList := utils.ListTypeAs[string](ctx, d.Index, path.Root("index"), diags)
		if !diags.HasError() && len(indexList) > 0 {
			*index = &indexList
		}
	}

	// Set author
	if author != nil && utils.IsKnown(d.Author) {
		authorList := utils.ListTypeAs[string](ctx, d.Author, path.Root("author"), diags)
		if !diags.HasError() && len(authorList) > 0 {
			*author = &authorList
		}
	}

	// Set tags
	if tags != nil && utils.IsKnown(d.Tags) {
		tagsList := utils.ListTypeAs[string](ctx, d.Tags, path.Root("tags"), diags)
		if !diags.HasError() && len(tagsList) > 0 {
			*tags = &tagsList
		}
	}

	// Set false positives
	if falsePositives != nil && utils.IsKnown(d.FalsePositives) {
		fpList := utils.ListTypeAs[string](ctx, d.FalsePositives, path.Root("false_positives"), diags)
		if !diags.HasError() && len(fpList) > 0 {
			*falsePositives = &fpList
		}
	}

	// Set references
	if references != nil && utils.IsKnown(d.References) {
		refList := utils.ListTypeAs[string](ctx, d.References, path.Root("references"), diags)
		if !diags.HasError() && len(refList) > 0 {
			*references = &refList
		}
	}

	// Set optional string fields
	if license != nil && utils.IsKnown(d.License) {
		ruleLicense := kbapi.SecurityDetectionsAPIRuleLicense(d.License.ValueString())
		*license = &ruleLicense
	}

	if note != nil && utils.IsKnown(d.Note) {
		ruleNote := kbapi.SecurityDetectionsAPIInvestigationGuide(d.Note.ValueString())
		*note = &ruleNote
	}

	if setup != nil && utils.IsKnown(d.Setup) {
		ruleSetup := kbapi.SecurityDetectionsAPISetupGuide(d.Setup.ValueString())
		*setup = &ruleSetup
	}

	// Set max signals
	if maxSignals != nil && utils.IsKnown(d.MaxSignals) {
		maxSig := kbapi.SecurityDetectionsAPIMaxSignals(d.MaxSignals.ValueInt64())
		*maxSignals = &maxSig
	}

	// Set version
	if version != nil && utils.IsKnown(d.Version) {
		ruleVersion := kbapi.SecurityDetectionsAPIRuleVersion(d.Version.ValueInt64())
		*version = &ruleVersion
	}
}

func (d SecurityDetectionRuleData) toUpdateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var updateProps kbapi.SecurityDetectionsAPIRuleUpdateProps

	ruleType := d.Type.ValueString()

	switch ruleType {
	case "query":
		return d.toQueryRuleUpdateProps(ctx)
	case "eql":
		return d.toEqlRuleUpdateProps(ctx)
	default:
		// Other rule types are not yet fully implemented for updates
		diags.AddError(
			"Unsupported rule type for updates",
			fmt.Sprintf("Rule type '%s' is not yet fully implemented for updates. Currently only 'query' and 'eql' rules are supported.", ruleType),
		)
		return updateProps, diags
	}
}

func (d SecurityDetectionRuleData) toQueryRuleUpdateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var updateProps kbapi.SecurityDetectionsAPIRuleUpdateProps

	queryRuleQuery := kbapi.SecurityDetectionsAPIRuleQuery(d.Query.ValueString())

	// Parse ID to get space_id and rule_id
	compId, resourceIdDiags := clients.CompositeIdFromStrFw(d.Id.ValueString())
	diags.Append(resourceIdDiags...)

	uid, err := uuid.Parse(compId.ResourceId)
	if err != nil {
		diags.AddError("ID was not a valid UUID", err.Error())
		return updateProps, diags
	}
	var id = kbapi.SecurityDetectionsAPIRuleObjectId(uid)

	queryRule := kbapi.SecurityDetectionsAPIQueryRuleUpdateProps{
		Id:          &id,
		Name:        kbapi.SecurityDetectionsAPIRuleName(d.Name.ValueString()),
		Description: kbapi.SecurityDetectionsAPIRuleDescription(d.Description.ValueString()),
		Type:        kbapi.SecurityDetectionsAPIQueryRuleUpdatePropsType("query"),
		Query:       &queryRuleQuery,
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// For updates, we need to include the rule_id if it's set
	if utils.IsKnown(d.RuleId) {
		ruleId := kbapi.SecurityDetectionsAPIRuleSignatureId(d.RuleId.ValueString())
		queryRule.RuleId = &ruleId
		queryRule.Id = nil // if rule_id is set, we cant send id
	}

	d.setCommonUpdateProps(ctx, &queryRule.Actions, &queryRule.RuleId, &queryRule.Enabled, &queryRule.From, &queryRule.To, &queryRule.Interval, &queryRule.Index, &queryRule.Author, &queryRule.Tags, &queryRule.FalsePositives, &queryRule.References, &queryRule.License, &queryRule.Note, &queryRule.Setup, &queryRule.MaxSignals, &queryRule.Version, &diags)

	// Set query-specific fields
	if utils.IsKnown(d.Language) {
		var language kbapi.SecurityDetectionsAPIKqlQueryLanguage
		switch d.Language.ValueString() {
		case "kuery":
			language = "kuery"
		case "lucene":
			language = "lucene"
		default:
			language = "kuery"
		}
		queryRule.Language = &language
	}

	if utils.IsKnown(d.SavedId) {
		savedId := kbapi.SecurityDetectionsAPISavedQueryId(d.SavedId.ValueString())
		queryRule.SavedId = &savedId
	}

	// Convert to union type
	err = updateProps.FromSecurityDetectionsAPIQueryRuleUpdateProps(queryRule)
	if err != nil {
		diags.AddError(
			"Error building update properties",
			"Could not convert query rule properties: "+err.Error(),
		)
	}

	return updateProps, diags
}

func (d SecurityDetectionRuleData) toEqlRuleUpdateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var updateProps kbapi.SecurityDetectionsAPIRuleUpdateProps

	// Parse ID to get space_id and rule_id
	compId, resourceIdDiags := clients.CompositeIdFromStrFw(d.Id.ValueString())
	diags.Append(resourceIdDiags...)

	uid, err := uuid.Parse(compId.ResourceId)
	if err != nil {
		diags.AddError("ID was not a valid UUID", err.Error())
		return updateProps, diags
	}
	var id = kbapi.SecurityDetectionsAPIRuleObjectId(uid)

	eqlRule := kbapi.SecurityDetectionsAPIEqlRuleUpdateProps{
		Id:          &id,
		Name:        kbapi.SecurityDetectionsAPIRuleName(d.Name.ValueString()),
		Description: kbapi.SecurityDetectionsAPIRuleDescription(d.Description.ValueString()),
		Type:        kbapi.SecurityDetectionsAPIEqlRuleUpdatePropsType("eql"),
		Query:       kbapi.SecurityDetectionsAPIRuleQuery(d.Query.ValueString()),
		Language:    kbapi.SecurityDetectionsAPIEqlQueryLanguage("eql"),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// For updates, we need to include the rule_id if it's set
	if utils.IsKnown(d.RuleId) {
		ruleId := kbapi.SecurityDetectionsAPIRuleSignatureId(d.RuleId.ValueString())
		eqlRule.RuleId = &ruleId
		eqlRule.Id = nil // if rule_id is set, we cant send id
	}

	d.setCommonUpdateProps(ctx, &eqlRule.Actions, &eqlRule.RuleId, &eqlRule.Enabled, &eqlRule.From, &eqlRule.To, &eqlRule.Interval, &eqlRule.Index, &eqlRule.Author, &eqlRule.Tags, &eqlRule.FalsePositives, &eqlRule.References, &eqlRule.License, &eqlRule.Note, &eqlRule.Setup, &eqlRule.MaxSignals, &eqlRule.Version, &diags)

	// Set EQL-specific fields
	if utils.IsKnown(d.TiebreakerField) {
		tiebreakerField := kbapi.SecurityDetectionsAPITiebreakerField(d.TiebreakerField.ValueString())
		eqlRule.TiebreakerField = &tiebreakerField
	}

	// Convert to union type
	err = updateProps.FromSecurityDetectionsAPIEqlRuleUpdateProps(eqlRule)
	if err != nil {
		diags.AddError(
			"Error building update properties",
			"Could not convert EQL rule properties: "+err.Error(),
		)
	}

	return updateProps, diags
}

// Helper function to set common update properties across all rule types
func (d SecurityDetectionRuleData) setCommonUpdateProps(
	ctx context.Context,
	actions **[]kbapi.SecurityDetectionsAPIRuleAction,
	ruleId **kbapi.SecurityDetectionsAPIRuleSignatureId,
	enabled **kbapi.SecurityDetectionsAPIIsRuleEnabled,
	from **kbapi.SecurityDetectionsAPIRuleIntervalFrom,
	to **kbapi.SecurityDetectionsAPIRuleIntervalTo,
	interval **kbapi.SecurityDetectionsAPIRuleInterval,
	index **[]string,
	author **[]string,
	tags **[]string,
	falsePositives **[]string,
	references **[]string,
	license **kbapi.SecurityDetectionsAPIRuleLicense,
	note **kbapi.SecurityDetectionsAPIInvestigationGuide,
	setup **kbapi.SecurityDetectionsAPISetupGuide,
	maxSignals **kbapi.SecurityDetectionsAPIMaxSignals,
	version **kbapi.SecurityDetectionsAPIRuleVersion,
	diags *diag.Diagnostics,
) {
	// Set enabled status
	if utils.IsKnown(d.Enabled) {
		isEnabled := kbapi.SecurityDetectionsAPIIsRuleEnabled(d.Enabled.ValueBool())
		*enabled = &isEnabled
	}

	// Set time range
	if utils.IsKnown(d.From) {
		fromTime := kbapi.SecurityDetectionsAPIRuleIntervalFrom(d.From.ValueString())
		*from = &fromTime
	}

	if utils.IsKnown(d.To) {
		toTime := kbapi.SecurityDetectionsAPIRuleIntervalTo(d.To.ValueString())
		*to = &toTime
	}

	// Set interval
	if utils.IsKnown(d.Interval) {
		intervalTime := kbapi.SecurityDetectionsAPIRuleInterval(d.Interval.ValueString())
		*interval = &intervalTime
	}

	// Set index patterns (if index pointer is provided)
	if index != nil && utils.IsKnown(d.Index) {
		indexList := utils.ListTypeAs[string](ctx, d.Index, path.Root("index"), diags)
		if !diags.HasError() {
			*index = &indexList
		}
	}

	// Set author
	if author != nil && utils.IsKnown(d.Author) {
		authorList := utils.ListTypeAs[string](ctx, d.Author, path.Root("author"), diags)
		if !diags.HasError() {
			*author = &authorList
		}
	}

	// Set tags
	if tags != nil && utils.IsKnown(d.Tags) {
		tagsList := utils.ListTypeAs[string](ctx, d.Tags, path.Root("tags"), diags)
		if !diags.HasError() {
			*tags = &tagsList
		}
	}

	// Set false positives
	if falsePositives != nil && utils.IsKnown(d.FalsePositives) {
		fpList := utils.ListTypeAs[string](ctx, d.FalsePositives, path.Root("false_positives"), diags)
		if !diags.HasError() {
			*falsePositives = &fpList
		}
	}

	// Set references
	if references != nil && utils.IsKnown(d.References) {
		refList := utils.ListTypeAs[string](ctx, d.References, path.Root("references"), diags)
		if !diags.HasError() {
			*references = &refList
		}
	}

	// Set optional string fields
	if license != nil && utils.IsKnown(d.License) {
		ruleLicense := kbapi.SecurityDetectionsAPIRuleLicense(d.License.ValueString())
		*license = &ruleLicense
	}

	if note != nil && utils.IsKnown(d.Note) {
		ruleNote := kbapi.SecurityDetectionsAPIInvestigationGuide(d.Note.ValueString())
		*note = &ruleNote
	}

	if setup != nil && utils.IsKnown(d.Setup) {
		ruleSetup := kbapi.SecurityDetectionsAPISetupGuide(d.Setup.ValueString())
		*setup = &ruleSetup
	}

	// Set max signals
	if maxSignals != nil && utils.IsKnown(d.MaxSignals) {
		maxSig := kbapi.SecurityDetectionsAPIMaxSignals(d.MaxSignals.ValueInt64())
		*maxSignals = &maxSig
	}

	// Set version
	if version != nil && utils.IsKnown(d.Version) {
		ruleVersion := kbapi.SecurityDetectionsAPIRuleVersion(d.Version.ValueInt64())
		*version = &ruleVersion
	}
}

func (d *SecurityDetectionRuleData) updateFromRule(ctx context.Context, rule interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	switch r := rule.(type) {
	case *kbapi.SecurityDetectionsAPIQueryRule:
		return d.updateFromQueryRule(ctx, r)
	case *kbapi.SecurityDetectionsAPIEqlRule:
		return d.updateFromEqlRule(ctx, r)
	default:
		diags.AddError(
			"Unsupported rule type",
			"Cannot update data from unsupported rule type",
		)
		return diags
	}
}

func (d *SecurityDetectionRuleData) updateFromQueryRule(ctx context.Context, rule *kbapi.SecurityDetectionsAPIQueryRule) diag.Diagnostics {
	var diags diag.Diagnostics

	compId := clients.CompositeId{
		ClusterId:  d.SpaceId.ValueString(),
		ResourceId: rule.Id.String(),
	}
	d.Id = types.StringValue(compId.String())

	d.RuleId = types.StringValue(string(rule.RuleId))
	d.Name = types.StringValue(string(rule.Name))
	d.Type = types.StringValue(string(rule.Type))
	d.Query = types.StringValue(rule.Query)
	d.Language = types.StringValue(string(rule.Language))
	d.Enabled = types.BoolValue(bool(rule.Enabled))
	d.From = types.StringValue(string(rule.From))
	d.To = types.StringValue(string(rule.To))
	d.Interval = types.StringValue(string(rule.Interval))
	d.Description = types.StringValue(string(rule.Description))
	d.RiskScore = types.Int64Value(int64(rule.RiskScore))
	d.Severity = types.StringValue(string(rule.Severity))
	d.MaxSignals = types.Int64Value(int64(rule.MaxSignals))
	d.Version = types.Int64Value(int64(rule.Version))

	// Update read-only fields
	d.CreatedAt = types.StringValue(rule.CreatedAt.Format("2006-01-02T15:04:05.000Z"))
	d.CreatedBy = types.StringValue(rule.CreatedBy)
	d.UpdatedAt = types.StringValue(rule.UpdatedAt.Format("2006-01-02T15:04:05.000Z"))
	d.UpdatedBy = types.StringValue(rule.UpdatedBy)
	d.Revision = types.Int64Value(int64(rule.Revision))

	// Update index patterns
	if rule.Index != nil && len(*rule.Index) > 0 {
		d.Index = utils.ListValueFrom(ctx, *rule.Index, types.StringType, path.Root("index"), &diags)
	} else {
		d.Index = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Update author
	if len(rule.Author) > 0 {
		d.Author = utils.ListValueFrom(ctx, rule.Author, types.StringType, path.Root("author"), &diags)
	} else {
		d.Author = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Update tags
	if len(rule.Tags) > 0 {
		d.Tags = utils.ListValueFrom(ctx, rule.Tags, types.StringType, path.Root("tags"), &diags)
	} else {
		d.Tags = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Update false positives
	if len(rule.FalsePositives) > 0 {
		d.FalsePositives = utils.ListValueFrom(ctx, rule.FalsePositives, types.StringType, path.Root("false_positives"), &diags)
	} else {
		d.FalsePositives = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Update references
	if len(rule.References) > 0 {
		d.References = utils.ListValueFrom(ctx, rule.References, types.StringType, path.Root("references"), &diags)
	} else {
		d.References = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Update optional string fields
	if rule.License != nil {
		d.License = types.StringValue(string(*rule.License))
	} else {
		d.License = types.StringNull()
	}

	if rule.Note != nil {
		d.Note = types.StringValue(string(*rule.Note))
	} else {
		d.Note = types.StringNull()
	}

	// Handle setup field - if empty, set to null to maintain consistency with optional schema
	if string(rule.Setup) != "" {
		d.Setup = types.StringValue(string(rule.Setup))
	} else {
		d.Setup = types.StringNull()
	}

	return diags
}

func (d *SecurityDetectionRuleData) updateFromEqlRule(ctx context.Context, rule *kbapi.SecurityDetectionsAPIEqlRule) diag.Diagnostics {
	var diags diag.Diagnostics

	compId := clients.CompositeId{
		ClusterId:  d.SpaceId.ValueString(),
		ResourceId: rule.Id.String(),
	}
	d.Id = types.StringValue(compId.String())

	d.RuleId = types.StringValue(string(rule.RuleId))
	d.Name = types.StringValue(string(rule.Name))
	d.Type = types.StringValue(string(rule.Type))
	d.Query = types.StringValue(rule.Query)
	d.Language = types.StringValue(string(rule.Language))
	d.Enabled = types.BoolValue(bool(rule.Enabled))
	d.From = types.StringValue(string(rule.From))
	d.To = types.StringValue(string(rule.To))
	d.Interval = types.StringValue(string(rule.Interval))
	d.Description = types.StringValue(string(rule.Description))
	d.RiskScore = types.Int64Value(int64(rule.RiskScore))
	d.Severity = types.StringValue(string(rule.Severity))
	d.MaxSignals = types.Int64Value(int64(rule.MaxSignals))
	d.Version = types.Int64Value(int64(rule.Version))

	// Update read-only fields
	d.CreatedAt = types.StringValue(rule.CreatedAt.Format("2006-01-02T15:04:05.000Z"))
	d.CreatedBy = types.StringValue(rule.CreatedBy)
	d.UpdatedAt = types.StringValue(rule.UpdatedAt.Format("2006-01-02T15:04:05.000Z"))
	d.UpdatedBy = types.StringValue(rule.UpdatedBy)
	d.Revision = types.Int64Value(int64(rule.Revision))

	// Update index patterns
	if rule.Index != nil && len(*rule.Index) > 0 {
		d.Index = utils.ListValueFrom(ctx, *rule.Index, types.StringType, path.Root("index"), &diags)
	} else {
		d.Index = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Update author
	if len(rule.Author) > 0 {
		d.Author = utils.ListValueFrom(ctx, rule.Author, types.StringType, path.Root("author"), &diags)
	} else {
		d.Author = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Update tags
	if len(rule.Tags) > 0 {
		d.Tags = utils.ListValueFrom(ctx, rule.Tags, types.StringType, path.Root("tags"), &diags)
	} else {
		d.Tags = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Update false positives
	if len(rule.FalsePositives) > 0 {
		d.FalsePositives = utils.ListValueFrom(ctx, rule.FalsePositives, types.StringType, path.Root("false_positives"), &diags)
	} else {
		d.FalsePositives = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Update references
	if len(rule.References) > 0 {
		d.References = utils.ListValueFrom(ctx, rule.References, types.StringType, path.Root("references"), &diags)
	} else {
		d.References = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Update optional string fields
	if rule.License != nil {
		d.License = types.StringValue(string(*rule.License))
	} else {
		d.License = types.StringNull()
	}

	if rule.Note != nil {
		d.Note = types.StringValue(string(*rule.Note))
	} else {
		d.Note = types.StringNull()
	}

	// Handle setup field - if empty, set to null to maintain consistency with optional schema
	if string(rule.Setup) != "" {
		d.Setup = types.StringValue(string(rule.Setup))
	} else {
		d.Setup = types.StringNull()
	}

	// EQL-specific fields
	if rule.TiebreakerField != nil {
		d.TiebreakerField = types.StringValue(string(*rule.TiebreakerField))
	} else {
		d.TiebreakerField = types.StringNull()
	}

	return diags
}

// Helper function to extract rule ID from any rule type
func extractRuleId(rule interface{}) (string, error) {
	switch r := rule.(type) {
	case *kbapi.SecurityDetectionsAPIQueryRule:
		return r.Id.String(), nil
	case *kbapi.SecurityDetectionsAPIEqlRule:
		return r.Id.String(), nil
	default:
		return "", fmt.Errorf("unsupported rule type for ID extraction")
	}
}
