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
type SecurityDetectionRuleTfData struct {
	ThreatMapping types.List `tfsdk:"threat_mapping"`
}

type SecurityDetectionRuleTfDataItem struct {
	Entries types.List `tfsdk:"entries"`
}

type SecurityDetectionRuleTfDataItemEntry struct {
	Field types.String `tfsdk:"field"`
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

type ThresholdModel struct {
	Field       types.List  `tfsdk:"field"`
	Value       types.Int64 `tfsdk:"value"`
	Cardinality types.List  `tfsdk:"cardinality"`
}

type CardinalityModel struct {
	Field types.String `tfsdk:"field"`
	Value types.Int64  `tfsdk:"value"`
}

// CommonCreateProps holds all the field pointers for setting common create properties
type CommonCreateProps struct {
	Actions        **[]kbapi.SecurityDetectionsAPIRuleAction
	RuleId         **kbapi.SecurityDetectionsAPIRuleSignatureId
	Enabled        **kbapi.SecurityDetectionsAPIIsRuleEnabled
	From           **kbapi.SecurityDetectionsAPIRuleIntervalFrom
	To             **kbapi.SecurityDetectionsAPIRuleIntervalTo
	Interval       **kbapi.SecurityDetectionsAPIRuleInterval
	Index          **[]string
	Author         **[]string
	Tags           **[]string
	FalsePositives **[]string
	References     **[]string
	License        **kbapi.SecurityDetectionsAPIRuleLicense
	Note           **kbapi.SecurityDetectionsAPIInvestigationGuide
	Setup          **kbapi.SecurityDetectionsAPISetupGuide
	MaxSignals     **kbapi.SecurityDetectionsAPIMaxSignals
	Version        **kbapi.SecurityDetectionsAPIRuleVersion
}

// CommonUpdateProps holds all the field pointers for setting common update properties
type CommonUpdateProps struct {
	Actions        **[]kbapi.SecurityDetectionsAPIRuleAction
	RuleId         **kbapi.SecurityDetectionsAPIRuleSignatureId
	Enabled        **kbapi.SecurityDetectionsAPIIsRuleEnabled
	From           **kbapi.SecurityDetectionsAPIRuleIntervalFrom
	To             **kbapi.SecurityDetectionsAPIRuleIntervalTo
	Interval       **kbapi.SecurityDetectionsAPIRuleInterval
	Index          **[]string
	Author         **[]string
	Tags           **[]string
	FalsePositives **[]string
	References     **[]string
	License        **kbapi.SecurityDetectionsAPIRuleLicense
	Note           **kbapi.SecurityDetectionsAPIInvestigationGuide
	Setup          **kbapi.SecurityDetectionsAPISetupGuide
	MaxSignals     **kbapi.SecurityDetectionsAPIMaxSignals
	Version        **kbapi.SecurityDetectionsAPIRuleVersion
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

// getKQLQueryLanguage maps language string to kbapi.SecurityDetectionsAPIKqlQueryLanguage
func (d SecurityDetectionRuleData) getKQLQueryLanguage() *kbapi.SecurityDetectionsAPIKqlQueryLanguage {
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
		return &language
	}
	return nil
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

	d.setCommonCreateProps(ctx, &CommonCreateProps{
		Actions:        &queryRule.Actions,
		RuleId:         &queryRule.RuleId,
		Enabled:        &queryRule.Enabled,
		From:           &queryRule.From,
		To:             &queryRule.To,
		Interval:       &queryRule.Interval,
		Index:          &queryRule.Index,
		Author:         &queryRule.Author,
		Tags:           &queryRule.Tags,
		FalsePositives: &queryRule.FalsePositives,
		References:     &queryRule.References,
		License:        &queryRule.License,
		Note:           &queryRule.Note,
		Setup:          &queryRule.Setup,
		MaxSignals:     &queryRule.MaxSignals,
		Version:        &queryRule.Version,
	}, &diags)

	// Set query-specific fields
	queryRule.Language = d.getKQLQueryLanguage()

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

	d.setCommonCreateProps(ctx, &CommonCreateProps{
		Actions:        &eqlRule.Actions,
		RuleId:         &eqlRule.RuleId,
		Enabled:        &eqlRule.Enabled,
		From:           &eqlRule.From,
		To:             &eqlRule.To,
		Interval:       &eqlRule.Interval,
		Index:          &eqlRule.Index,
		Author:         &eqlRule.Author,
		Tags:           &eqlRule.Tags,
		FalsePositives: &eqlRule.FalsePositives,
		References:     &eqlRule.References,
		License:        &eqlRule.License,
		Note:           &eqlRule.Note,
		Setup:          &eqlRule.Setup,
		MaxSignals:     &eqlRule.MaxSignals,
		Version:        &eqlRule.Version,
	}, &diags)

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

	d.setCommonCreateProps(ctx, &CommonCreateProps{
		Actions:        &esqlRule.Actions,
		RuleId:         &esqlRule.RuleId,
		Enabled:        &esqlRule.Enabled,
		From:           &esqlRule.From,
		To:             &esqlRule.To,
		Interval:       &esqlRule.Interval,
		Index:          nil, // ESQL rules don't use index patterns
		Author:         &esqlRule.Author,
		Tags:           &esqlRule.Tags,
		FalsePositives: &esqlRule.FalsePositives,
		References:     &esqlRule.References,
		License:        &esqlRule.License,
		Note:           &esqlRule.Note,
		Setup:          &esqlRule.Setup,
		MaxSignals:     &esqlRule.MaxSignals,
		Version:        &esqlRule.Version,
	}, &diags)

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

	d.setCommonCreateProps(ctx, &CommonCreateProps{
		Actions:        &mlRule.Actions,
		RuleId:         &mlRule.RuleId,
		Enabled:        &mlRule.Enabled,
		From:           &mlRule.From,
		To:             &mlRule.To,
		Interval:       &mlRule.Interval,
		Index:          nil, // ML rules don't use index patterns
		Author:         &mlRule.Author,
		Tags:           &mlRule.Tags,
		FalsePositives: &mlRule.FalsePositives,
		References:     &mlRule.References,
		License:        &mlRule.License,
		Note:           &mlRule.Note,
		Setup:          &mlRule.Setup,
		MaxSignals:     &mlRule.MaxSignals,
		Version:        &mlRule.Version,
	}, &diags)

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

	d.setCommonCreateProps(ctx, &CommonCreateProps{
		Actions:        &newTermsRule.Actions,
		RuleId:         &newTermsRule.RuleId,
		Enabled:        &newTermsRule.Enabled,
		From:           &newTermsRule.From,
		To:             &newTermsRule.To,
		Interval:       &newTermsRule.Interval,
		Index:          &newTermsRule.Index,
		Author:         &newTermsRule.Author,
		Tags:           &newTermsRule.Tags,
		FalsePositives: &newTermsRule.FalsePositives,
		References:     &newTermsRule.References,
		License:        &newTermsRule.License,
		Note:           &newTermsRule.Note,
		Setup:          &newTermsRule.Setup,
		MaxSignals:     &newTermsRule.MaxSignals,
		Version:        &newTermsRule.Version,
	}, &diags)

	// Set query language
	newTermsRule.Language = d.getKQLQueryLanguage()

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

	d.setCommonCreateProps(ctx, &CommonCreateProps{
		Actions:        &savedQueryRule.Actions,
		RuleId:         &savedQueryRule.RuleId,
		Enabled:        &savedQueryRule.Enabled,
		From:           &savedQueryRule.From,
		To:             &savedQueryRule.To,
		Interval:       &savedQueryRule.Interval,
		Index:          &savedQueryRule.Index,
		Author:         &savedQueryRule.Author,
		Tags:           &savedQueryRule.Tags,
		FalsePositives: &savedQueryRule.FalsePositives,
		References:     &savedQueryRule.References,
		License:        &savedQueryRule.License,
		Note:           &savedQueryRule.Note,
		Setup:          &savedQueryRule.Setup,
		MaxSignals:     &savedQueryRule.MaxSignals,
		Version:        &savedQueryRule.Version,
	}, &diags)

	// Set optional query for saved query rules
	if utils.IsKnown(d.Query) {
		query := kbapi.SecurityDetectionsAPIRuleQuery(d.Query.ValueString())
		savedQueryRule.Query = &query
	}

	// Set query language
	savedQueryRule.Language = d.getKQLQueryLanguage()

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

	if utils.IsKnown(d.ThreatMapping) && len(d.ThreatMapping.Elements()) > 0 {
		apiThreatMapping, threatMappingDiags := d.threatMappingToApi(ctx)
		if !threatMappingDiags.HasError() {
			threatMatchRule.ThreatMapping = apiThreatMapping
		}
		diags.Append(threatMappingDiags...)
	}

	d.setCommonCreateProps(ctx, &CommonCreateProps{
		Actions:        &threatMatchRule.Actions,
		RuleId:         &threatMatchRule.RuleId,
		Enabled:        &threatMatchRule.Enabled,
		From:           &threatMatchRule.From,
		To:             &threatMatchRule.To,
		Interval:       &threatMatchRule.Interval,
		Index:          &threatMatchRule.Index,
		Author:         &threatMatchRule.Author,
		Tags:           &threatMatchRule.Tags,
		FalsePositives: &threatMatchRule.FalsePositives,
		References:     &threatMatchRule.References,
		License:        &threatMatchRule.License,
		Note:           &threatMatchRule.Note,
		Setup:          &threatMatchRule.Setup,
		MaxSignals:     &threatMatchRule.MaxSignals,
		Version:        &threatMatchRule.Version,
	}, &diags)

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
	threatMatchRule.Language = d.getKQLQueryLanguage()

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
	threshold := d.thresholdToApi(ctx, &diags)
	if threshold != nil {
		thresholdRule.Threshold = *threshold
	}

	d.setCommonCreateProps(ctx, &CommonCreateProps{
		Actions:        &thresholdRule.Actions,
		RuleId:         &thresholdRule.RuleId,
		Enabled:        &thresholdRule.Enabled,
		From:           &thresholdRule.From,
		To:             &thresholdRule.To,
		Interval:       &thresholdRule.Interval,
		Index:          &thresholdRule.Index,
		Author:         &thresholdRule.Author,
		Tags:           &thresholdRule.Tags,
		FalsePositives: &thresholdRule.FalsePositives,
		References:     &thresholdRule.References,
		License:        &thresholdRule.License,
		Note:           &thresholdRule.Note,
		Setup:          &thresholdRule.Setup,
		MaxSignals:     &thresholdRule.MaxSignals,
		Version:        &thresholdRule.Version,
	}, &diags)

	// Set query language
	thresholdRule.Language = d.getKQLQueryLanguage()

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
	props *CommonCreateProps,
	diags *diag.Diagnostics,
) {
	// Set optional rule_id if provided
	if props.RuleId != nil && utils.IsKnown(d.RuleId) {
		id := kbapi.SecurityDetectionsAPIRuleSignatureId(d.RuleId.ValueString())
		*props.RuleId = &id
	}

	// Set enabled status
	if props.Enabled != nil && utils.IsKnown(d.Enabled) {
		isEnabled := kbapi.SecurityDetectionsAPIIsRuleEnabled(d.Enabled.ValueBool())
		*props.Enabled = &isEnabled
	}

	// Set time range
	if props.From != nil && utils.IsKnown(d.From) {
		fromTime := kbapi.SecurityDetectionsAPIRuleIntervalFrom(d.From.ValueString())
		*props.From = &fromTime
	}

	if props.To != nil && utils.IsKnown(d.To) {
		toTime := kbapi.SecurityDetectionsAPIRuleIntervalTo(d.To.ValueString())
		*props.To = &toTime
	}

	// Set interval
	if props.Interval != nil && utils.IsKnown(d.Interval) {
		intervalTime := kbapi.SecurityDetectionsAPIRuleInterval(d.Interval.ValueString())
		*props.Interval = &intervalTime
	}

	// Set index patterns (if index pointer is provided)
	if props.Index != nil && utils.IsKnown(d.Index) {
		indexList := utils.ListTypeAs[string](ctx, d.Index, path.Root("index"), diags)
		if !diags.HasError() && len(indexList) > 0 {
			*props.Index = &indexList
		}
	}

	// Set author
	if props.Author != nil && utils.IsKnown(d.Author) {
		authorList := utils.ListTypeAs[string](ctx, d.Author, path.Root("author"), diags)
		if !diags.HasError() && len(authorList) > 0 {
			*props.Author = &authorList
		}
	}

	// Set tags
	if props.Tags != nil && utils.IsKnown(d.Tags) {
		tagsList := utils.ListTypeAs[string](ctx, d.Tags, path.Root("tags"), diags)
		if !diags.HasError() && len(tagsList) > 0 {
			*props.Tags = &tagsList
		}
	}

	// Set false positives
	if props.FalsePositives != nil && utils.IsKnown(d.FalsePositives) {
		fpList := utils.ListTypeAs[string](ctx, d.FalsePositives, path.Root("false_positives"), diags)
		if !diags.HasError() && len(fpList) > 0 {
			*props.FalsePositives = &fpList
		}
	}

	// Set references
	if props.References != nil && utils.IsKnown(d.References) {
		refList := utils.ListTypeAs[string](ctx, d.References, path.Root("references"), diags)
		if !diags.HasError() && len(refList) > 0 {
			*props.References = &refList
		}
	}

	// Set optional string fields
	if props.License != nil && utils.IsKnown(d.License) {
		ruleLicense := kbapi.SecurityDetectionsAPIRuleLicense(d.License.ValueString())
		*props.License = &ruleLicense
	}

	if props.Note != nil && utils.IsKnown(d.Note) {
		ruleNote := kbapi.SecurityDetectionsAPIInvestigationGuide(d.Note.ValueString())
		*props.Note = &ruleNote
	}

	if props.Setup != nil && utils.IsKnown(d.Setup) {
		ruleSetup := kbapi.SecurityDetectionsAPISetupGuide(d.Setup.ValueString())
		*props.Setup = &ruleSetup
	}

	// Set max signals
	if props.MaxSignals != nil && utils.IsKnown(d.MaxSignals) {
		maxSig := kbapi.SecurityDetectionsAPIMaxSignals(d.MaxSignals.ValueInt64())
		*props.MaxSignals = &maxSig
	}

	// Set version
	if props.Version != nil && utils.IsKnown(d.Version) {
		ruleVersion := kbapi.SecurityDetectionsAPIRuleVersion(d.Version.ValueInt64())
		*props.Version = &ruleVersion
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
	case "esql":
		return d.toEsqlRuleUpdateProps(ctx)
	case "machine_learning":
		return d.toMachineLearningRuleUpdateProps(ctx)
	case "new_terms":
		return d.toNewTermsRuleUpdateProps(ctx)
	case "saved_query":
		return d.toSavedQueryRuleUpdateProps(ctx)
	case "threat_match":
		return d.toThreatMatchRuleUpdateProps(ctx)
	case "threshold":
		return d.toThresholdRuleUpdateProps(ctx)
	default:
		diags.AddError(
			"Unsupported rule type",
			fmt.Sprintf("Rule type '%s' is not supported for updates", ruleType),
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

	d.setCommonUpdateProps(ctx, &CommonUpdateProps{
		Actions:        &queryRule.Actions,
		RuleId:         &queryRule.RuleId,
		Enabled:        &queryRule.Enabled,
		From:           &queryRule.From,
		To:             &queryRule.To,
		Interval:       &queryRule.Interval,
		Index:          &queryRule.Index,
		Author:         &queryRule.Author,
		Tags:           &queryRule.Tags,
		FalsePositives: &queryRule.FalsePositives,
		References:     &queryRule.References,
		License:        &queryRule.License,
		Note:           &queryRule.Note,
		Setup:          &queryRule.Setup,
		MaxSignals:     &queryRule.MaxSignals,
		Version:        &queryRule.Version,
	}, &diags)

	// Set query-specific fields
	queryRule.Language = d.getKQLQueryLanguage()

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

	d.setCommonUpdateProps(ctx, &CommonUpdateProps{
		Actions:        &eqlRule.Actions,
		RuleId:         &eqlRule.RuleId,
		Enabled:        &eqlRule.Enabled,
		From:           &eqlRule.From,
		To:             &eqlRule.To,
		Interval:       &eqlRule.Interval,
		Index:          &eqlRule.Index,
		Author:         &eqlRule.Author,
		Tags:           &eqlRule.Tags,
		FalsePositives: &eqlRule.FalsePositives,
		References:     &eqlRule.References,
		License:        &eqlRule.License,
		Note:           &eqlRule.Note,
		Setup:          &eqlRule.Setup,
		MaxSignals:     &eqlRule.MaxSignals,
		Version:        &eqlRule.Version,
	}, &diags)

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

func (d SecurityDetectionRuleData) toEsqlRuleUpdateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
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

	esqlRule := kbapi.SecurityDetectionsAPIEsqlRuleUpdateProps{
		Id:          &id,
		Name:        kbapi.SecurityDetectionsAPIRuleName(d.Name.ValueString()),
		Description: kbapi.SecurityDetectionsAPIRuleDescription(d.Description.ValueString()),
		Type:        kbapi.SecurityDetectionsAPIEsqlRuleUpdatePropsType("esql"),
		Query:       kbapi.SecurityDetectionsAPIRuleQuery(d.Query.ValueString()),
		Language:    kbapi.SecurityDetectionsAPIEsqlQueryLanguage("esql"),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// For updates, we need to include the rule_id if it's set
	if utils.IsKnown(d.RuleId) {
		ruleId := kbapi.SecurityDetectionsAPIRuleSignatureId(d.RuleId.ValueString())
		esqlRule.RuleId = &ruleId
		esqlRule.Id = nil // if rule_id is set, we cant send id
	}

	d.setCommonUpdateProps(ctx, &CommonUpdateProps{
		Actions:        &esqlRule.Actions,
		RuleId:         &esqlRule.RuleId,
		Enabled:        &esqlRule.Enabled,
		From:           &esqlRule.From,
		To:             &esqlRule.To,
		Interval:       &esqlRule.Interval,
		Index:          nil, // ESQL rules don't use index patterns
		Author:         &esqlRule.Author,
		Tags:           &esqlRule.Tags,
		FalsePositives: &esqlRule.FalsePositives,
		References:     &esqlRule.References,
		License:        &esqlRule.License,
		Note:           &esqlRule.Note,
		Setup:          &esqlRule.Setup,
		MaxSignals:     &esqlRule.MaxSignals,
		Version:        &esqlRule.Version,
	}, &diags)

	// ESQL rules don't use index patterns as they use FROM clause in the query

	// Convert to union type
	err = updateProps.FromSecurityDetectionsAPIEsqlRuleUpdateProps(esqlRule)
	if err != nil {
		diags.AddError(
			"Error building update properties",
			"Could not convert ESQL rule properties: "+err.Error(),
		)
	}

	return updateProps, diags
}

func (d SecurityDetectionRuleData) toMachineLearningRuleUpdateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
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

	mlRule := kbapi.SecurityDetectionsAPIMachineLearningRuleUpdateProps{
		Id:               &id,
		Name:             kbapi.SecurityDetectionsAPIRuleName(d.Name.ValueString()),
		Description:      kbapi.SecurityDetectionsAPIRuleDescription(d.Description.ValueString()),
		Type:             kbapi.SecurityDetectionsAPIMachineLearningRuleUpdatePropsType("machine_learning"),
		AnomalyThreshold: kbapi.SecurityDetectionsAPIAnomalyThreshold(d.AnomalyThreshold.ValueInt64()),
		RiskScore:        kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:         kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// For updates, we need to include the rule_id if it's set
	if utils.IsKnown(d.RuleId) {
		ruleId := kbapi.SecurityDetectionsAPIRuleSignatureId(d.RuleId.ValueString())
		mlRule.RuleId = &ruleId
		mlRule.Id = nil // if rule_id is set, we cant send id
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

	d.setCommonUpdateProps(ctx, &CommonUpdateProps{
		Actions:        &mlRule.Actions,
		RuleId:         &mlRule.RuleId,
		Enabled:        &mlRule.Enabled,
		From:           &mlRule.From,
		To:             &mlRule.To,
		Interval:       &mlRule.Interval,
		Index:          nil, // ML rules don't use index patterns
		Author:         &mlRule.Author,
		Tags:           &mlRule.Tags,
		FalsePositives: &mlRule.FalsePositives,
		References:     &mlRule.References,
		License:        &mlRule.License,
		Note:           &mlRule.Note,
		Setup:          &mlRule.Setup,
		MaxSignals:     &mlRule.MaxSignals,
		Version:        &mlRule.Version,
	}, &diags)

	// ML rules don't use index patterns or query

	// Convert to union type
	err = updateProps.FromSecurityDetectionsAPIMachineLearningRuleUpdateProps(mlRule)
	if err != nil {
		diags.AddError(
			"Error building update properties",
			"Could not convert ML rule properties: "+err.Error(),
		)
	}

	return updateProps, diags
}

func (d SecurityDetectionRuleData) toNewTermsRuleUpdateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
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

	newTermsRule := kbapi.SecurityDetectionsAPINewTermsRuleUpdateProps{
		Id:                 &id,
		Name:               kbapi.SecurityDetectionsAPIRuleName(d.Name.ValueString()),
		Description:        kbapi.SecurityDetectionsAPIRuleDescription(d.Description.ValueString()),
		Type:               kbapi.SecurityDetectionsAPINewTermsRuleUpdatePropsType("new_terms"),
		Query:              kbapi.SecurityDetectionsAPIRuleQuery(d.Query.ValueString()),
		HistoryWindowStart: kbapi.SecurityDetectionsAPIHistoryWindowStart(d.HistoryWindowStart.ValueString()),
		RiskScore:          kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:           kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// For updates, we need to include the rule_id if it's set
	if utils.IsKnown(d.RuleId) {
		ruleId := kbapi.SecurityDetectionsAPIRuleSignatureId(d.RuleId.ValueString())
		newTermsRule.RuleId = &ruleId
		newTermsRule.Id = nil // if rule_id is set, we cant send id
	}

	// Set new terms fields
	if utils.IsKnown(d.NewTermsFields) {
		newTermsFields := utils.ListTypeAs[string](ctx, d.NewTermsFields, path.Root("new_terms_fields"), &diags)
		if !diags.HasError() {
			newTermsRule.NewTermsFields = newTermsFields
		}
	}

	d.setCommonUpdateProps(ctx, &CommonUpdateProps{
		Actions:        &newTermsRule.Actions,
		RuleId:         &newTermsRule.RuleId,
		Enabled:        &newTermsRule.Enabled,
		From:           &newTermsRule.From,
		To:             &newTermsRule.To,
		Interval:       &newTermsRule.Interval,
		Index:          &newTermsRule.Index,
		Author:         &newTermsRule.Author,
		Tags:           &newTermsRule.Tags,
		FalsePositives: &newTermsRule.FalsePositives,
		References:     &newTermsRule.References,
		License:        &newTermsRule.License,
		Note:           &newTermsRule.Note,
		Setup:          &newTermsRule.Setup,
		MaxSignals:     &newTermsRule.MaxSignals,
		Version:        &newTermsRule.Version,
	}, &diags)

	// Set query language
	newTermsRule.Language = d.getKQLQueryLanguage()

	// Convert to union type
	err = updateProps.FromSecurityDetectionsAPINewTermsRuleUpdateProps(newTermsRule)
	if err != nil {
		diags.AddError(
			"Error building update properties",
			"Could not convert new terms rule properties: "+err.Error(),
		)
	}

	return updateProps, diags
}

func (d SecurityDetectionRuleData) toSavedQueryRuleUpdateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
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

	savedQueryRule := kbapi.SecurityDetectionsAPISavedQueryRuleUpdateProps{
		Id:          &id,
		Name:        kbapi.SecurityDetectionsAPIRuleName(d.Name.ValueString()),
		Description: kbapi.SecurityDetectionsAPIRuleDescription(d.Description.ValueString()),
		Type:        kbapi.SecurityDetectionsAPISavedQueryRuleUpdatePropsType("saved_query"),
		SavedId:     kbapi.SecurityDetectionsAPISavedQueryId(d.SavedId.ValueString()),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// For updates, we need to include the rule_id if it's set
	if utils.IsKnown(d.RuleId) {
		ruleId := kbapi.SecurityDetectionsAPIRuleSignatureId(d.RuleId.ValueString())
		savedQueryRule.RuleId = &ruleId
		savedQueryRule.Id = nil // if rule_id is set, we cant send id
	}

	d.setCommonUpdateProps(ctx, &CommonUpdateProps{
		Actions:        &savedQueryRule.Actions,
		RuleId:         &savedQueryRule.RuleId,
		Enabled:        &savedQueryRule.Enabled,
		From:           &savedQueryRule.From,
		To:             &savedQueryRule.To,
		Interval:       &savedQueryRule.Interval,
		Index:          &savedQueryRule.Index,
		Author:         &savedQueryRule.Author,
		Tags:           &savedQueryRule.Tags,
		FalsePositives: &savedQueryRule.FalsePositives,
		References:     &savedQueryRule.References,
		License:        &savedQueryRule.License,
		Note:           &savedQueryRule.Note,
		Setup:          &savedQueryRule.Setup,
		MaxSignals:     &savedQueryRule.MaxSignals,
		Version:        &savedQueryRule.Version,
	}, &diags)

	// Set optional query for saved query rules
	if utils.IsKnown(d.Query) {
		query := kbapi.SecurityDetectionsAPIRuleQuery(d.Query.ValueString())
		savedQueryRule.Query = &query
	}

	// Set query language
	savedQueryRule.Language = d.getKQLQueryLanguage()

	// Convert to union type
	err = updateProps.FromSecurityDetectionsAPISavedQueryRuleUpdateProps(savedQueryRule)
	if err != nil {
		diags.AddError(
			"Error building update properties",
			"Could not convert saved query rule properties: "+err.Error(),
		)
	}

	return updateProps, diags
}

func (d SecurityDetectionRuleData) toThreatMatchRuleUpdateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
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

	threatMatchRule := kbapi.SecurityDetectionsAPIThreatMatchRuleUpdateProps{
		Id:          &id,
		Name:        kbapi.SecurityDetectionsAPIRuleName(d.Name.ValueString()),
		Description: kbapi.SecurityDetectionsAPIRuleDescription(d.Description.ValueString()),
		Type:        kbapi.SecurityDetectionsAPIThreatMatchRuleUpdatePropsType("threat_match"),
		Query:       kbapi.SecurityDetectionsAPIRuleQuery(d.Query.ValueString()),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// For updates, we need to include the rule_id if it's set
	if utils.IsKnown(d.RuleId) {
		ruleId := kbapi.SecurityDetectionsAPIRuleSignatureId(d.RuleId.ValueString())
		threatMatchRule.RuleId = &ruleId
		threatMatchRule.Id = nil // if rule_id is set, we cant send id
	}

	// Set threat index
	if utils.IsKnown(d.ThreatIndex) {
		threatIndex := utils.ListTypeAs[string](ctx, d.ThreatIndex, path.Root("threat_index"), &diags)
		if !diags.HasError() {
			threatMatchRule.ThreatIndex = threatIndex
		}
	}

	// TODO consolidate w/ create props
	if utils.IsKnown(d.ThreatMapping) && len(d.ThreatMapping.Elements()) > 0 {
		apiThreatMapping, threatMappingDiags := d.threatMappingToApi(ctx)
		if !threatMappingDiags.HasError() {
			threatMatchRule.ThreatMapping = apiThreatMapping
		}
		diags.Append(threatMappingDiags...)
	}

	d.setCommonUpdateProps(ctx, &CommonUpdateProps{
		Actions:        &threatMatchRule.Actions,
		RuleId:         &threatMatchRule.RuleId,
		Enabled:        &threatMatchRule.Enabled,
		From:           &threatMatchRule.From,
		To:             &threatMatchRule.To,
		Interval:       &threatMatchRule.Interval,
		Index:          &threatMatchRule.Index,
		Author:         &threatMatchRule.Author,
		Tags:           &threatMatchRule.Tags,
		FalsePositives: &threatMatchRule.FalsePositives,
		References:     &threatMatchRule.References,
		License:        &threatMatchRule.License,
		Note:           &threatMatchRule.Note,
		Setup:          &threatMatchRule.Setup,
		MaxSignals:     &threatMatchRule.MaxSignals,
		Version:        &threatMatchRule.Version,
	}, &diags)

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
	threatMatchRule.Language = d.getKQLQueryLanguage()

	if utils.IsKnown(d.SavedId) {
		savedId := kbapi.SecurityDetectionsAPISavedQueryId(d.SavedId.ValueString())
		threatMatchRule.SavedId = &savedId
	}

	// Convert to union type
	err = updateProps.FromSecurityDetectionsAPIThreatMatchRuleUpdateProps(threatMatchRule)
	if err != nil {
		diags.AddError(
			"Error building update properties",
			"Could not convert threat match rule properties: "+err.Error(),
		)
	}

	return updateProps, diags
}

func (d SecurityDetectionRuleData) toThresholdRuleUpdateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
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

	thresholdRule := kbapi.SecurityDetectionsAPIThresholdRuleUpdateProps{
		Id:          &id,
		Name:        kbapi.SecurityDetectionsAPIRuleName(d.Name.ValueString()),
		Description: kbapi.SecurityDetectionsAPIRuleDescription(d.Description.ValueString()),
		Type:        kbapi.SecurityDetectionsAPIThresholdRuleUpdatePropsType("threshold"),
		Query:       kbapi.SecurityDetectionsAPIRuleQuery(d.Query.ValueString()),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// For updates, we need to include the rule_id if it's set
	if utils.IsKnown(d.RuleId) {
		ruleId := kbapi.SecurityDetectionsAPIRuleSignatureId(d.RuleId.ValueString())
		thresholdRule.RuleId = &ruleId
		thresholdRule.Id = nil // if rule_id is set, we cant send id
	}

	// Set threshold - this is required for threshold rules
	threshold := d.thresholdToApi(ctx, &diags)
	if threshold != nil {
		thresholdRule.Threshold = *threshold
	}

	d.setCommonUpdateProps(ctx, &CommonUpdateProps{
		Actions:        &thresholdRule.Actions,
		RuleId:         &thresholdRule.RuleId,
		Enabled:        &thresholdRule.Enabled,
		From:           &thresholdRule.From,
		To:             &thresholdRule.To,
		Interval:       &thresholdRule.Interval,
		Index:          &thresholdRule.Index,
		Author:         &thresholdRule.Author,
		Tags:           &thresholdRule.Tags,
		FalsePositives: &thresholdRule.FalsePositives,
		References:     &thresholdRule.References,
		License:        &thresholdRule.License,
		Note:           &thresholdRule.Note,
		Setup:          &thresholdRule.Setup,
		MaxSignals:     &thresholdRule.MaxSignals,
		Version:        &thresholdRule.Version,
	}, &diags)

	// Set query language
	thresholdRule.Language = d.getKQLQueryLanguage()

	if utils.IsKnown(d.SavedId) {
		savedId := kbapi.SecurityDetectionsAPISavedQueryId(d.SavedId.ValueString())
		thresholdRule.SavedId = &savedId
	}

	// Convert to union type
	err = updateProps.FromSecurityDetectionsAPIThresholdRuleUpdateProps(thresholdRule)
	if err != nil {
		diags.AddError(
			"Error building update properties",
			"Could not convert threshold rule properties: "+err.Error(),
		)
	}

	return updateProps, diags
}

// Helper function to set common update properties across all rule types
func (d SecurityDetectionRuleData) setCommonUpdateProps(
	ctx context.Context,
	props *CommonUpdateProps,
	diags *diag.Diagnostics,
) {
	// Set enabled status
	if props.Enabled != nil && utils.IsKnown(d.Enabled) {
		isEnabled := kbapi.SecurityDetectionsAPIIsRuleEnabled(d.Enabled.ValueBool())
		*props.Enabled = &isEnabled
	}

	// Set time range
	if props.From != nil && utils.IsKnown(d.From) {
		fromTime := kbapi.SecurityDetectionsAPIRuleIntervalFrom(d.From.ValueString())
		*props.From = &fromTime
	}

	if props.To != nil && utils.IsKnown(d.To) {
		toTime := kbapi.SecurityDetectionsAPIRuleIntervalTo(d.To.ValueString())
		*props.To = &toTime
	}

	// Set interval
	if props.Interval != nil && utils.IsKnown(d.Interval) {
		intervalTime := kbapi.SecurityDetectionsAPIRuleInterval(d.Interval.ValueString())
		*props.Interval = &intervalTime
	}

	// Set index patterns (if index pointer is provided)
	if props.Index != nil && utils.IsKnown(d.Index) {
		indexList := utils.ListTypeAs[string](ctx, d.Index, path.Root("index"), diags)
		if !diags.HasError() {
			*props.Index = &indexList
		}
	}

	// Set author
	if props.Author != nil && utils.IsKnown(d.Author) {
		authorList := utils.ListTypeAs[string](ctx, d.Author, path.Root("author"), diags)
		if !diags.HasError() {
			*props.Author = &authorList
		}
	}

	// Set tags
	if props.Tags != nil && utils.IsKnown(d.Tags) {
		tagsList := utils.ListTypeAs[string](ctx, d.Tags, path.Root("tags"), diags)
		if !diags.HasError() {
			*props.Tags = &tagsList
		}
	}

	// Set false positives
	if props.FalsePositives != nil && utils.IsKnown(d.FalsePositives) {
		fpList := utils.ListTypeAs[string](ctx, d.FalsePositives, path.Root("false_positives"), diags)
		if !diags.HasError() {
			*props.FalsePositives = &fpList
		}
	}

	// Set references
	if props.References != nil && utils.IsKnown(d.References) {
		refList := utils.ListTypeAs[string](ctx, d.References, path.Root("references"), diags)
		if !diags.HasError() {
			*props.References = &refList
		}
	}

	// Set optional string fields
	if props.License != nil && utils.IsKnown(d.License) {
		ruleLicense := kbapi.SecurityDetectionsAPIRuleLicense(d.License.ValueString())
		*props.License = &ruleLicense
	}

	if props.Note != nil && utils.IsKnown(d.Note) {
		ruleNote := kbapi.SecurityDetectionsAPIInvestigationGuide(d.Note.ValueString())
		*props.Note = &ruleNote
	}

	if props.Setup != nil && utils.IsKnown(d.Setup) {
		ruleSetup := kbapi.SecurityDetectionsAPISetupGuide(d.Setup.ValueString())
		*props.Setup = &ruleSetup
	}

	// Set max signals
	if props.MaxSignals != nil && utils.IsKnown(d.MaxSignals) {
		maxSig := kbapi.SecurityDetectionsAPIMaxSignals(d.MaxSignals.ValueInt64())
		*props.MaxSignals = &maxSig
	}

	// Set version
	if props.Version != nil && utils.IsKnown(d.Version) {
		ruleVersion := kbapi.SecurityDetectionsAPIRuleVersion(d.Version.ValueInt64())
		*props.Version = &ruleVersion
	}
}

func (d *SecurityDetectionRuleData) updateFromRule(ctx context.Context, response *kbapi.SecurityDetectionsAPIRuleResponse) diag.Diagnostics {
	var diags diag.Diagnostics

	rule, err := response.ValueByDiscriminator()
	if err != nil {
		diags.AddError(
			"Error determining rule type",
			"Could not determine the type of the security detection rule from the API response: "+err.Error(),
		)
		return diags
	}

	switch r := rule.(type) {
	case kbapi.SecurityDetectionsAPIQueryRule:
		return d.updateFromQueryRule(ctx, &r)
	case kbapi.SecurityDetectionsAPIEqlRule:
		return d.updateFromEqlRule(ctx, &r)
	case kbapi.SecurityDetectionsAPIEsqlRule:
		return d.updateFromEsqlRule(ctx, &r)
	case kbapi.SecurityDetectionsAPIMachineLearningRule:
		return d.updateFromMachineLearningRule(ctx, &r)
	case kbapi.SecurityDetectionsAPINewTermsRule:
		return d.updateFromNewTermsRule(ctx, &r)
	case kbapi.SecurityDetectionsAPISavedQueryRule:
		return d.updateFromSavedQueryRule(ctx, &r)
	case kbapi.SecurityDetectionsAPIThreatMatchRule:
		return d.updateFromThreatMatchRule(ctx, &r)
	case kbapi.SecurityDetectionsAPIThresholdRule:
		return d.updateFromThresholdRule(ctx, &r)
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
	d.CreatedAt = utils.TimeToStringValue(rule.CreatedAt)
	d.CreatedBy = types.StringValue(rule.CreatedBy)
	d.UpdatedAt = utils.TimeToStringValue(rule.UpdatedAt)
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
	d.CreatedAt = utils.TimeToStringValue(rule.CreatedAt)
	d.CreatedBy = types.StringValue(rule.CreatedBy)
	d.UpdatedAt = utils.TimeToStringValue(rule.UpdatedAt)
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

func (d *SecurityDetectionRuleData) updateFromEsqlRule(ctx context.Context, rule *kbapi.SecurityDetectionsAPIEsqlRule) diag.Diagnostics {
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

	// ESQL rules don't use index patterns
	d.Index = types.ListValueMust(types.StringType, []attr.Value{})

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

func (d *SecurityDetectionRuleData) updateFromMachineLearningRule(ctx context.Context, rule *kbapi.SecurityDetectionsAPIMachineLearningRule) diag.Diagnostics {
	var diags diag.Diagnostics

	compId := clients.CompositeId{
		ClusterId:  d.SpaceId.ValueString(),
		ResourceId: rule.Id.String(),
	}
	d.Id = types.StringValue(compId.String())

	d.RuleId = types.StringValue(string(rule.RuleId))
	d.Name = types.StringValue(string(rule.Name))
	d.Type = types.StringValue(string(rule.Type))
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

	// ML rules don't use index patterns or query
	d.Index = types.ListValueMust(types.StringType, []attr.Value{})
	d.Query = types.StringNull()
	d.Language = types.StringNull()

	// ML-specific fields
	d.AnomalyThreshold = types.Int64Value(int64(rule.AnomalyThreshold))

	// Handle ML job ID(s) - can be single string or array
	// Try to extract as single job ID first, then as array
	if singleJobId, err := rule.MachineLearningJobId.AsSecurityDetectionsAPIMachineLearningJobId0(); err == nil {
		// Single job ID
		d.MachineLearningJobId = utils.ListValueFrom(ctx, []string{string(singleJobId)}, types.StringType, path.Root("machine_learning_job_id"), &diags)
	} else if multipleJobIds, err := rule.MachineLearningJobId.AsSecurityDetectionsAPIMachineLearningJobId1(); err == nil {
		// Multiple job IDs
		jobIdStrings := make([]string, len(multipleJobIds))
		for i, jobId := range multipleJobIds {
			jobIdStrings[i] = string(jobId)
		}
		d.MachineLearningJobId = utils.ListValueFrom(ctx, jobIdStrings, types.StringType, path.Root("machine_learning_job_id"), &diags)
	} else {
		d.MachineLearningJobId = types.ListValueMust(types.StringType, []attr.Value{})
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

func (d *SecurityDetectionRuleData) updateFromNewTermsRule(ctx context.Context, rule *kbapi.SecurityDetectionsAPINewTermsRule) diag.Diagnostics {
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

	// New Terms-specific fields
	d.HistoryWindowStart = types.StringValue(string(rule.HistoryWindowStart))
	if len(rule.NewTermsFields) > 0 {
		d.NewTermsFields = utils.ListValueFrom(ctx, rule.NewTermsFields, types.StringType, path.Root("new_terms_fields"), &diags)
	} else {
		d.NewTermsFields = types.ListValueMust(types.StringType, []attr.Value{})
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

func (d *SecurityDetectionRuleData) updateFromSavedQueryRule(ctx context.Context, rule *kbapi.SecurityDetectionsAPISavedQueryRule) diag.Diagnostics {
	var diags diag.Diagnostics

	compId := clients.CompositeId{
		ClusterId:  d.SpaceId.ValueString(),
		ResourceId: rule.Id.String(),
	}
	d.Id = types.StringValue(compId.String())

	d.RuleId = types.StringValue(string(rule.RuleId))
	d.Name = types.StringValue(string(rule.Name))
	d.Type = types.StringValue(string(rule.Type))
	d.SavedId = types.StringValue(string(rule.SavedId))
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

	// Optional query for saved query rules
	if rule.Query != nil {
		d.Query = types.StringValue(*rule.Query)
	} else {
		d.Query = types.StringNull()
	}

	// Language for saved query rules (not a pointer)
	d.Language = types.StringValue(string(rule.Language))

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

func (d *SecurityDetectionRuleData) updateFromThreatMatchRule(ctx context.Context, rule *kbapi.SecurityDetectionsAPIThreatMatchRule) diag.Diagnostics {
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

	// Threat Match-specific fields
	d.ThreatQuery = types.StringValue(string(rule.ThreatQuery))
	if len(rule.ThreatIndex) > 0 {
		d.ThreatIndex = utils.ListValueFrom(ctx, rule.ThreatIndex, types.StringType, path.Root("threat_index"), &diags)
	} else {
		d.ThreatIndex = types.ListValueMust(types.StringType, []attr.Value{})
	}

	if rule.ThreatIndicatorPath != nil {
		d.ThreatIndicatorPath = types.StringValue(string(*rule.ThreatIndicatorPath))
	} else {
		d.ThreatIndicatorPath = types.StringNull()
	}

	if rule.ConcurrentSearches != nil {
		d.ConcurrentSearches = types.Int64Value(int64(*rule.ConcurrentSearches))
	} else {
		d.ConcurrentSearches = types.Int64Null()
	}

	if rule.ItemsPerSearch != nil {
		d.ItemsPerSearch = types.Int64Value(int64(*rule.ItemsPerSearch))
	} else {
		d.ItemsPerSearch = types.Int64Null()
	}

	// Optional saved query ID
	if rule.SavedId != nil {
		d.SavedId = types.StringValue(string(*rule.SavedId))
	} else {
		d.SavedId = types.StringNull()
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

	// Convert threat mapping
	if len(rule.ThreatMapping) > 0 {
		listValue, threatMappingDiags := convertThreatMappingToModel(ctx, rule.ThreatMapping)
		diags.Append(threatMappingDiags...)
		if !threatMappingDiags.HasError() {
			d.ThreatMapping = listValue
		}
	}

	return diags
}

func (d *SecurityDetectionRuleData) updateFromThresholdRule(ctx context.Context, rule *kbapi.SecurityDetectionsAPIThresholdRule) diag.Diagnostics {
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

	// Threshold-specific fields
	thresholdObj, thresholdDiags := convertThresholdToModel(ctx, rule.Threshold)
	diags.Append(thresholdDiags...)
	if !thresholdDiags.HasError() {
		d.Threshold = thresholdObj
	}

	// Optional saved query ID
	if rule.SavedId != nil {
		d.SavedId = types.StringValue(string(*rule.SavedId))
	} else {
		d.SavedId = types.StringNull()
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

// Helper function to extract rule ID from any rule type
func extractId(response *kbapi.SecurityDetectionsAPIRuleResponse) (string, diag.Diagnostics) {
	var diags diag.Diagnostics

	rule, err := response.ValueByDiscriminator()
	if err != nil {
		diags.AddError(
			"Error determining rule type",
			"Could not determine the type of the security detection rule from the API response: "+err.Error(),
		)
		return "", diags
	}

	var id string
	switch r := rule.(type) {
	case kbapi.SecurityDetectionsAPIQueryRule:
		id = r.Id.String()
	case kbapi.SecurityDetectionsAPIEqlRule:
		id = r.Id.String()
	case kbapi.SecurityDetectionsAPIEsqlRule:
		id = r.Id.String()
	case kbapi.SecurityDetectionsAPIMachineLearningRule:
		id = r.Id.String()
	case kbapi.SecurityDetectionsAPINewTermsRule:
		id = r.Id.String()
	case kbapi.SecurityDetectionsAPISavedQueryRule:
		id = r.Id.String()
	case kbapi.SecurityDetectionsAPIThreatMatchRule:
		id = r.Id.String()
	case kbapi.SecurityDetectionsAPIThresholdRule:
		id = r.Id.String()
	default:
		diags.AddError(
			"Unsupported rule type for ID extraction",
			fmt.Sprintf("Cannot extract ID from unsupported rule type: %T", r),
		)
		return "", diags
	}

	return id, diags
}

// Helper function to initialize fields that should be set to default values for all rule types
func (d *SecurityDetectionRuleData) initializeAllFieldsToDefaults(ctx context.Context, diags *diag.Diagnostics) {

	// Initialize fields that should be empty lists for all rule types initially
	if !utils.IsKnown(d.Author) {
		d.Author = types.ListNull(types.StringType)
	}
	if !utils.IsKnown(d.Tags) {
		d.Tags = types.ListNull(types.StringType)
	}
	if !utils.IsKnown(d.FalsePositives) {
		d.FalsePositives = types.ListNull(types.StringType)
	}
	if !utils.IsKnown(d.References) {
		d.References = types.ListNull(types.StringType)
	}

	// Initialize all type-specific fields to null/empty by default
	d.initializeTypeSpecificFieldsToDefaults(ctx, diags)
}

// Helper function to initialize type-specific fields to default/null values
func (d *SecurityDetectionRuleData) initializeTypeSpecificFieldsToDefaults(ctx context.Context, diags *diag.Diagnostics) {
	// EQL-specific fields
	if !utils.IsKnown(d.TiebreakerField) {
		d.TiebreakerField = types.StringNull()
	}

	// Machine Learning-specific fields
	if !utils.IsKnown(d.AnomalyThreshold) {
		d.AnomalyThreshold = types.Int64Null()
	}
	if !utils.IsKnown(d.MachineLearningJobId) {
		d.MachineLearningJobId = types.ListNull(types.StringType)
	}

	// New Terms-specific fields
	if !utils.IsKnown(d.NewTermsFields) {
		d.NewTermsFields = types.ListNull(types.StringType)
	}
	if !utils.IsKnown(d.HistoryWindowStart) {
		d.HistoryWindowStart = types.StringNull()
	}

	// Saved Query-specific fields
	if !utils.IsKnown(d.SavedId) {
		d.SavedId = types.StringNull()
	}

	// Threat Match-specific fields
	if !utils.IsKnown(d.ThreatIndex) {
		d.ThreatIndex = types.ListNull(types.StringType)
	}
	if !utils.IsKnown(d.ThreatQuery) {
		d.ThreatQuery = types.StringNull()
	}
	if !utils.IsKnown(d.ThreatMapping) {
		d.ThreatMapping = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"entries": types.ListType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"field": types.StringType,
							"type":  types.StringType,
							"value": types.StringType,
						},
					},
				},
			},
		})
	}
	if !utils.IsKnown(d.ThreatFilters) {
		d.ThreatFilters = types.ListNull(types.StringType)
	}
	if !utils.IsKnown(d.ThreatIndicatorPath) {
		d.ThreatIndicatorPath = types.StringNull()
	}
	if !utils.IsKnown(d.ConcurrentSearches) {
		d.ConcurrentSearches = types.Int64Null()
	}
	if !utils.IsKnown(d.ItemsPerSearch) {
		d.ItemsPerSearch = types.Int64Null()
	}

	// Threshold-specific fields
	if !utils.IsKnown(d.Threshold) {
		d.Threshold = types.ObjectNull(map[string]attr.Type{
			"value": types.Int64Type,
			"field": types.ListType{ElemType: types.StringType},
			"cardinality": types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"field": types.StringType,
						"value": types.Int64Type,
					},
				},
			},
		})
	}

	// Timeline fields (common across multiple rule types)
	if !utils.IsKnown(d.TimelineId) {
		d.TimelineId = types.StringNull()
	}
	if !utils.IsKnown(d.TimelineTitle) {
		d.TimelineTitle = types.StringNull()
	}

	// Threat field (common across multiple rule types) - MITRE ATT&CK framework
	if !utils.IsKnown(d.Threat) {
		d.Threat = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"framework": types.StringType,
				"tactic": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":        types.StringType,
						"name":      types.StringType,
						"reference": types.StringType,
					},
				},
				"technique": types.ListType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"id":        types.StringType,
							"name":      types.StringType,
							"reference": types.StringType,
							"subtechnique": types.ListType{
								ElemType: types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"id":        types.StringType,
										"name":      types.StringType,
										"reference": types.StringType,
									},
								},
							},
						},
					},
				},
			},
		})
	}
}

// convertThreatMappingToModel converts kbapi.SecurityDetectionsAPIThreatMapping to the terraform model
func convertThreatMappingToModel(ctx context.Context, apiThreatMappings kbapi.SecurityDetectionsAPIThreatMapping) (types.List, diag.Diagnostics) {
	var threatMappings []SecurityDetectionRuleTfDataItem

	for _, apiMapping := range apiThreatMappings {
		var entries []SecurityDetectionRuleTfDataItemEntry

		for _, apiEntry := range apiMapping.Entries {
			entries = append(entries, SecurityDetectionRuleTfDataItemEntry{
				Field: types.StringValue(string(apiEntry.Field)),
				Type:  types.StringValue(string(apiEntry.Type)),
				Value: types.StringValue(string(apiEntry.Value)),
			})
		}

		entriesListValue, diags := types.ListValueFrom(ctx, threatMappingEntryElementType(), entries)
		if diags.HasError() {
			return types.ListNull(threatMappingElementType()), diags
		}

		threatMappings = append(threatMappings, SecurityDetectionRuleTfDataItem{
			Entries: entriesListValue,
		})
	}

	listValue, diags := types.ListValueFrom(ctx, threatMappingElementType(), threatMappings)
	return listValue, diags
}

// convertThresholdToModel converts kbapi.SecurityDetectionsAPIThreshold to the terraform model
func convertThresholdToModel(ctx context.Context, apiThreshold kbapi.SecurityDetectionsAPIThreshold) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Handle threshold field - can be single string or array
	var fieldList types.List
	if singleField, err := apiThreshold.Field.AsSecurityDetectionsAPIThresholdField0(); err == nil {
		// Single field
		fieldList = utils.SliceToListType_String(ctx, []string{string(singleField)}, path.Root("threshold").AtName("field"), &diags)
	} else if multipleFields, err := apiThreshold.Field.AsSecurityDetectionsAPIThresholdField1(); err == nil {
		// Multiple fields
		fieldStrings := make([]string, len(multipleFields))
		for i, field := range multipleFields {
			fieldStrings[i] = string(field)
		}
		fieldList = utils.SliceToListType_String(ctx, fieldStrings, path.Root("threshold").AtName("field"), &diags)
	} else {
		fieldList = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Handle cardinality (optional)
	var cardinalityList types.List
	if apiThreshold.Cardinality != nil && len(*apiThreshold.Cardinality) > 0 {
		cardinalityList = utils.SliceToListType(ctx, *apiThreshold.Cardinality, cardinalityElementType(), path.Root("threshold").AtName("cardinality"), &diags,
			func(item struct {
				Field string `json:"field"`
				Value int    `json:"value"`
			}, meta utils.ListMeta) CardinalityModel {
				return CardinalityModel{
					Field: types.StringValue(item.Field),
					Value: types.Int64Value(int64(item.Value)),
				}
			})
	} else {
		cardinalityList = types.ListNull(cardinalityElementType())
	}

	thresholdModel := ThresholdModel{
		Field:       fieldList,
		Value:       types.Int64Value(int64(apiThreshold.Value)),
		Cardinality: cardinalityList,
	}

	thresholdObject, objDiags := types.ObjectValueFrom(ctx, thresholdElementType(), thresholdModel)
	diags.Append(objDiags...)
	return thresholdObject, diags
}

// threatMappingElementType returns the element type for threat mapping
func threatMappingElementType() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"entries": types.ListType{
				ElemType: threatMappingEntryElementType(),
			},
		},
	}
}

// threatMappingEntryElementType returns the element type for threat mapping entries
func threatMappingEntryElementType() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"field": types.StringType,
			"type":  types.StringType,
			"value": types.StringType,
		},
	}
}

// thresholdElementType returns the element type for threshold
func thresholdElementType() map[string]attr.Type {
	return map[string]attr.Type{
		"field": types.ListType{ElemType: types.StringType},
		"value": types.Int64Type,
		"cardinality": types.ListType{
			ElemType: cardinalityElementType(),
		},
	}
}

// cardinalityElementType returns the element type for cardinality
func cardinalityElementType() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"field": types.StringType,
			"value": types.Int64Type,
		},
	}
}

// Helper function to process threshold configuration for threshold rules
func (d SecurityDetectionRuleData) thresholdToApi(ctx context.Context, diags *diag.Diagnostics) *kbapi.SecurityDetectionsAPIThreshold {
	if !utils.IsKnown(d.Threshold) {
		return nil
	}

	threshold := utils.ObjectTypeToStruct(ctx, d.Threshold, path.Root("threshold"), diags,
		func(item ThresholdModel, meta utils.ObjectMeta) kbapi.SecurityDetectionsAPIThreshold {
			threshold := kbapi.SecurityDetectionsAPIThreshold{
				Value: kbapi.SecurityDetectionsAPIThresholdValue(item.Value.ValueInt64()),
			}

			// Handle threshold field(s)
			if utils.IsKnown(item.Field) {
				fieldList := utils.ListTypeToSlice_String(ctx, item.Field, meta.Path.AtName("field"), meta.Diags)
				if len(fieldList) > 0 {
					var thresholdField kbapi.SecurityDetectionsAPIThresholdField
					if len(fieldList) == 1 {
						err := thresholdField.FromSecurityDetectionsAPIThresholdField0(fieldList[0])
						if err != nil {
							meta.Diags.AddError("Error setting threshold field", err.Error())
						} else {
							threshold.Field = thresholdField
						}
					} else {
						err := thresholdField.FromSecurityDetectionsAPIThresholdField1(fieldList)
						if err != nil {
							meta.Diags.AddError("Error setting threshold fields", err.Error())
						} else {
							threshold.Field = thresholdField
						}
					}
				}
			}

			// Handle cardinality (optional)
			if utils.IsKnown(item.Cardinality) {
				cardinalityList := utils.ListTypeToSlice(ctx, item.Cardinality, meta.Path.AtName("cardinality"), meta.Diags,
					func(item CardinalityModel, meta utils.ListMeta) struct {
						Field string `json:"field"`
						Value int    `json:"value"`
					} {
						return struct {
							Field string `json:"field"`
							Value int    `json:"value"`
						}{
							Field: item.Field.ValueString(),
							Value: int(item.Value.ValueInt64()),
						}
					})
				if len(cardinalityList) > 0 {
					threshold.Cardinality = (*kbapi.SecurityDetectionsAPIThresholdCardinality)(&cardinalityList)
				}
			}

			return threshold
		})

	return threshold
}

// Helper function to process threat mapping configuration for threat match rules
func (d SecurityDetectionRuleData) threatMappingToApi(ctx context.Context) (kbapi.SecurityDetectionsAPIThreatMapping, diag.Diagnostics) {
	var diags diag.Diagnostics

	threatMapping := make([]SecurityDetectionRuleTfDataItem, len(d.ThreatMapping.Elements()))

	threatMappingDiags := d.ThreatMapping.ElementsAs(ctx, &threatMapping, false)
	if threatMappingDiags.HasError() {
		diags.Append(threatMappingDiags...)
		return nil, diags
	}

	apiThreatMapping := make(kbapi.SecurityDetectionsAPIThreatMapping, 0)
	for _, mapping := range threatMapping {
		if mapping.Entries.IsNull() || mapping.Entries.IsUnknown() {
			continue
		}

		entries := make([]SecurityDetectionRuleTfDataItemEntry, len(mapping.Entries.Elements()))
		entryDiag := mapping.Entries.ElementsAs(ctx, &entries, false)
		diags = append(diags, entryDiag...)

		apiThreatMappingEntries := make([]kbapi.SecurityDetectionsAPIThreatMappingEntry, 0)
		for _, entry := range entries {

			apiMapping := kbapi.SecurityDetectionsAPIThreatMappingEntry{
				Field: kbapi.SecurityDetectionsAPINonEmptyString(entry.Field.ValueString()),
				Type:  kbapi.SecurityDetectionsAPIThreatMappingEntryType(entry.Type.ValueString()),
				Value: kbapi.SecurityDetectionsAPINonEmptyString(entry.Value.ValueString()),
			}
			apiThreatMappingEntries = append(apiThreatMappingEntries, apiMapping)

		}

		apiThreatMapping = append(apiThreatMapping, struct {
			Entries []kbapi.SecurityDetectionsAPIThreatMappingEntry `json:"entries"`
		}{Entries: apiThreatMappingEntries})
	}

	return apiThreatMapping, diags
}
