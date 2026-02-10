package security_detection_rule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// MinVersionResponseActions defines the minimum server version required for response actions
var MinVersionResponseActions = version.Must(version.NewVersion("8.16.0"))

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
	Description         types.String `tfsdk:"description"`
	RiskScore           types.Int64  `tfsdk:"risk_score"`
	RiskScoreMapping    types.List   `tfsdk:"risk_score_mapping"`
	Severity            types.String `tfsdk:"severity"`
	SeverityMapping     types.List   `tfsdk:"severity_mapping"`
	Author              types.List   `tfsdk:"author"`
	Tags                types.List   `tfsdk:"tags"`
	License             types.String `tfsdk:"license"`
	RelatedIntegrations types.List   `tfsdk:"related_integrations"`
	RequiredFields      types.List   `tfsdk:"required_fields"`

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

	// Actions field (common across all rule types)
	Actions types.List `tfsdk:"actions"`

	// Response actions field (common across all rule types)
	ResponseActions types.List `tfsdk:"response_actions"`

	// Exceptions list field (common across all rule types)
	ExceptionsList types.List `tfsdk:"exceptions_list"`

	// Alert suppression field (common across all rule types)
	AlertSuppression types.Object `tfsdk:"alert_suppression"`

	// Building block type field (common across all rule types)
	BuildingBlockType types.String `tfsdk:"building_block_type"`

	// Data view ID field (common across all rule types)
	DataViewId types.String `tfsdk:"data_view_id"`

	// Namespace field (common across all rule types)
	Namespace types.String `tfsdk:"namespace"`

	// Rule name override field (common across all rule types)
	RuleNameOverride types.String `tfsdk:"rule_name_override"`

	// Timestamp override fields (common across all rule types)
	TimestampOverride                 types.String `tfsdk:"timestamp_override"`
	TimestampOverrideFallbackDisabled types.Bool   `tfsdk:"timestamp_override_fallback_disabled"`

	// Investigation fields (common across all rule types)
	InvestigationFields types.List `tfsdk:"investigation_fields"`

	// Filters field (common across all rule types) - Query and filter context array to define alert conditions
	Filters jsontypes.Normalized `tfsdk:"filters"`
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

type AlertSuppressionModel struct {
	GroupBy               types.List           `tfsdk:"group_by"`
	Duration              customtypes.Duration `tfsdk:"duration"`
	MissingFieldsStrategy types.String         `tfsdk:"missing_fields_strategy"`
}

type CardinalityModel struct {
	Field types.String `tfsdk:"field"`
	Value types.Int64  `tfsdk:"value"`
}

type ActionModel struct {
	ActionTypeId types.String `tfsdk:"action_type_id"`
	Id           types.String `tfsdk:"id"`
	Params       types.Map    `tfsdk:"params"` // Map of strings (some may be JSON)
	Group        types.String `tfsdk:"group"`
	Uuid         types.String `tfsdk:"uuid"`
	AlertsFilter types.Map    `tfsdk:"alerts_filter"`
	Frequency    types.Object `tfsdk:"frequency"`
}

type ActionFrequencyModel struct {
	NotifyWhen types.String `tfsdk:"notify_when"`
	Summary    types.Bool   `tfsdk:"summary"`
	Throttle   types.String `tfsdk:"throttle"`
}

type ResponseActionModel struct {
	ActionTypeId types.String `tfsdk:"action_type_id"`
	Params       types.Object `tfsdk:"params"`
}

type ResponseActionParamsModel struct {
	// Osquery params
	Query        types.String `tfsdk:"query"`
	PackId       types.String `tfsdk:"pack_id"`
	SavedQueryId types.String `tfsdk:"saved_query_id"`
	Timeout      types.Int64  `tfsdk:"timeout"`
	EcsMapping   types.Map    `tfsdk:"ecs_mapping"`
	Queries      types.List   `tfsdk:"queries"`

	// Endpoint params
	Command types.String `tfsdk:"command"`
	Comment types.String `tfsdk:"comment"`
	Config  types.Object `tfsdk:"config"`
}

type OsqueryQueryModel struct {
	Id         types.String `tfsdk:"id"`
	Query      types.String `tfsdk:"query"`
	Platform   types.String `tfsdk:"platform"`
	Version    types.String `tfsdk:"version"`
	Removed    types.Bool   `tfsdk:"removed"`
	Snapshot   types.Bool   `tfsdk:"snapshot"`
	EcsMapping types.Map    `tfsdk:"ecs_mapping"`
}

type EndpointProcessConfigModel struct {
	Field     types.String `tfsdk:"field"`
	Overwrite types.Bool   `tfsdk:"overwrite"`
}

type ExceptionsListModel struct {
	Id            types.String `tfsdk:"id"`
	ListId        types.String `tfsdk:"list_id"`
	NamespaceType types.String `tfsdk:"namespace_type"`
	Type          types.String `tfsdk:"type"`
}

type RiskScoreMappingModel struct {
	Field     types.String `tfsdk:"field"`
	Operator  types.String `tfsdk:"operator"`
	Value     types.String `tfsdk:"value"`
	RiskScore types.Int64  `tfsdk:"risk_score"`
}

type RelatedIntegrationModel struct {
	Package     types.String `tfsdk:"package"`
	Version     types.String `tfsdk:"version"`
	Integration types.String `tfsdk:"integration"`
}

type RequiredFieldModel struct {
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
	Ecs  types.Bool   `tfsdk:"ecs"`
}

type SeverityMappingModel struct {
	Field    types.String `tfsdk:"field"`
	Operator types.String `tfsdk:"operator"`
	Value    types.String `tfsdk:"value"`
	Severity types.String `tfsdk:"severity"`
}

type ThreatModel struct {
	Framework types.String `tfsdk:"framework"`
	Tactic    types.Object `tfsdk:"tactic"`
	Technique types.List   `tfsdk:"technique"`
}

type ThreatTacticModel struct {
	Id        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Reference types.String `tfsdk:"reference"`
}

type ThreatTechniqueModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Reference    types.String `tfsdk:"reference"`
	Subtechnique types.List   `tfsdk:"subtechnique"`
}

type ThreatSubtechniqueModel struct {
	Id        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Reference types.String `tfsdk:"reference"`
}

// CommonCreateProps holds all the field pointers for setting common create properties
type CommonCreateProps struct {
	Actions                           **[]kbapi.SecurityDetectionsAPIRuleAction
	ResponseActions                   **[]kbapi.SecurityDetectionsAPIResponseAction
	RuleId                            **kbapi.SecurityDetectionsAPIRuleSignatureId
	Enabled                           **kbapi.SecurityDetectionsAPIIsRuleEnabled
	From                              **kbapi.SecurityDetectionsAPIRuleIntervalFrom
	To                                **kbapi.SecurityDetectionsAPIRuleIntervalTo
	Interval                          **kbapi.SecurityDetectionsAPIRuleInterval
	Index                             **[]string
	Author                            **[]string
	Tags                              **[]string
	FalsePositives                    **[]string
	References                        **[]string
	License                           **kbapi.SecurityDetectionsAPIRuleLicense
	Note                              **kbapi.SecurityDetectionsAPIInvestigationGuide
	Setup                             **kbapi.SecurityDetectionsAPISetupGuide
	MaxSignals                        **kbapi.SecurityDetectionsAPIMaxSignals
	Version                           **kbapi.SecurityDetectionsAPIRuleVersion
	ExceptionsList                    **[]kbapi.SecurityDetectionsAPIRuleExceptionList
	AlertSuppression                  **kbapi.SecurityDetectionsAPIAlertSuppression
	RiskScoreMapping                  **kbapi.SecurityDetectionsAPIRiskScoreMapping
	SeverityMapping                   **kbapi.SecurityDetectionsAPISeverityMapping
	RelatedIntegrations               **kbapi.SecurityDetectionsAPIRelatedIntegrationArray
	RequiredFields                    **[]kbapi.SecurityDetectionsAPIRequiredFieldInput
	BuildingBlockType                 **kbapi.SecurityDetectionsAPIBuildingBlockType
	DataViewId                        **kbapi.SecurityDetectionsAPIDataViewId
	Namespace                         **kbapi.SecurityDetectionsAPIAlertsIndexNamespace
	RuleNameOverride                  **kbapi.SecurityDetectionsAPIRuleNameOverride
	TimestampOverride                 **kbapi.SecurityDetectionsAPITimestampOverride
	TimestampOverrideFallbackDisabled **kbapi.SecurityDetectionsAPITimestampOverrideFallbackDisabled
	InvestigationFields               **kbapi.SecurityDetectionsAPIInvestigationFields
	Filters                           **kbapi.SecurityDetectionsAPIRuleFilterArray
	Threat                            **kbapi.SecurityDetectionsAPIThreatArray
	TimelineId                        **kbapi.SecurityDetectionsAPITimelineTemplateId
	TimelineTitle                     **kbapi.SecurityDetectionsAPITimelineTemplateTitle
}

// CommonUpdateProps holds all the field pointers for setting common update properties
type CommonUpdateProps struct {
	Actions                           **[]kbapi.SecurityDetectionsAPIRuleAction
	ResponseActions                   **[]kbapi.SecurityDetectionsAPIResponseAction
	RuleId                            **kbapi.SecurityDetectionsAPIRuleSignatureId
	Enabled                           **kbapi.SecurityDetectionsAPIIsRuleEnabled
	From                              **kbapi.SecurityDetectionsAPIRuleIntervalFrom
	To                                **kbapi.SecurityDetectionsAPIRuleIntervalTo
	Interval                          **kbapi.SecurityDetectionsAPIRuleInterval
	Index                             **[]string
	Author                            **[]string
	Tags                              **[]string
	FalsePositives                    **[]string
	References                        **[]string
	License                           **kbapi.SecurityDetectionsAPIRuleLicense
	Note                              **kbapi.SecurityDetectionsAPIInvestigationGuide
	Setup                             **kbapi.SecurityDetectionsAPISetupGuide
	MaxSignals                        **kbapi.SecurityDetectionsAPIMaxSignals
	Version                           **kbapi.SecurityDetectionsAPIRuleVersion
	ExceptionsList                    **[]kbapi.SecurityDetectionsAPIRuleExceptionList
	AlertSuppression                  **kbapi.SecurityDetectionsAPIAlertSuppression
	RiskScoreMapping                  **kbapi.SecurityDetectionsAPIRiskScoreMapping
	SeverityMapping                   **kbapi.SecurityDetectionsAPISeverityMapping
	RelatedIntegrations               **kbapi.SecurityDetectionsAPIRelatedIntegrationArray
	RequiredFields                    **[]kbapi.SecurityDetectionsAPIRequiredFieldInput
	BuildingBlockType                 **kbapi.SecurityDetectionsAPIBuildingBlockType
	DataViewId                        **kbapi.SecurityDetectionsAPIDataViewId
	Namespace                         **kbapi.SecurityDetectionsAPIAlertsIndexNamespace
	RuleNameOverride                  **kbapi.SecurityDetectionsAPIRuleNameOverride
	TimestampOverride                 **kbapi.SecurityDetectionsAPITimestampOverride
	TimestampOverrideFallbackDisabled **kbapi.SecurityDetectionsAPITimestampOverrideFallbackDisabled
	InvestigationFields               **kbapi.SecurityDetectionsAPIInvestigationFields
	Filters                           **kbapi.SecurityDetectionsAPIRuleFilterArray
	Threat                            **kbapi.SecurityDetectionsAPIThreatArray
	TimelineId                        **kbapi.SecurityDetectionsAPITimelineTemplateId
	TimelineTitle                     **kbapi.SecurityDetectionsAPITimelineTemplateTitle
}

// Helper function to set common properties across all rule types
func (d SecurityDetectionRuleData) setCommonCreateProps(
	ctx context.Context,
	props *CommonCreateProps,
	diags *diag.Diagnostics,
	client clients.MinVersionEnforceable,
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

	// Set actions
	if props.Actions != nil && utils.IsKnown(d.Actions) {
		actions, actionDiags := d.actionsToApi(ctx)
		diags.Append(actionDiags...)
		if !actionDiags.HasError() && len(actions) > 0 {
			*props.Actions = &actions
		}
	}

	// Set exceptions list
	if props.ExceptionsList != nil && utils.IsKnown(d.ExceptionsList) {
		exceptionsList, exceptionsListDiags := d.exceptionsListToApi(ctx)
		diags.Append(exceptionsListDiags...)
		if !exceptionsListDiags.HasError() && len(exceptionsList) > 0 {
			*props.ExceptionsList = &exceptionsList
		}
	}

	// Set risk score mapping
	if props.RiskScoreMapping != nil && utils.IsKnown(d.RiskScoreMapping) {
		riskScoreMapping, riskScoreMappingDiags := d.riskScoreMappingToApi(ctx)
		diags.Append(riskScoreMappingDiags...)
		if !riskScoreMappingDiags.HasError() && len(riskScoreMapping) > 0 {
			*props.RiskScoreMapping = &riskScoreMapping
		}
	}

	// Set building block type
	if props.BuildingBlockType != nil && utils.IsKnown(d.BuildingBlockType) {
		buildingBlockType := kbapi.SecurityDetectionsAPIBuildingBlockType(d.BuildingBlockType.ValueString())
		*props.BuildingBlockType = &buildingBlockType
	}

	// Set data view ID
	if props.DataViewId != nil && utils.IsKnown(d.DataViewId) {
		dataViewId := kbapi.SecurityDetectionsAPIDataViewId(d.DataViewId.ValueString())
		*props.DataViewId = &dataViewId
	}

	// Set namespace
	if props.Namespace != nil && utils.IsKnown(d.Namespace) {
		namespace := kbapi.SecurityDetectionsAPIAlertsIndexNamespace(d.Namespace.ValueString())
		*props.Namespace = &namespace
	}

	// Set rule name override
	if props.RuleNameOverride != nil && utils.IsKnown(d.RuleNameOverride) {
		ruleNameOverride := kbapi.SecurityDetectionsAPIRuleNameOverride(d.RuleNameOverride.ValueString())
		*props.RuleNameOverride = &ruleNameOverride
	}

	// Set timestamp override
	if props.TimestampOverride != nil && utils.IsKnown(d.TimestampOverride) {
		timestampOverride := kbapi.SecurityDetectionsAPITimestampOverride(d.TimestampOverride.ValueString())
		*props.TimestampOverride = &timestampOverride
	}

	// Set timestamp override fallback disabled
	if props.TimestampOverrideFallbackDisabled != nil && utils.IsKnown(d.TimestampOverrideFallbackDisabled) {
		timestampOverrideFallbackDisabled := kbapi.SecurityDetectionsAPITimestampOverrideFallbackDisabled(d.TimestampOverrideFallbackDisabled.ValueBool())
		*props.TimestampOverrideFallbackDisabled = &timestampOverrideFallbackDisabled
	}

	// Set severity mapping
	if props.SeverityMapping != nil && utils.IsKnown(d.SeverityMapping) {
		severityMapping, severityMappingDiags := d.severityMappingToApi(ctx)
		diags.Append(severityMappingDiags...)
		if !severityMappingDiags.HasError() && severityMapping != nil && len(*severityMapping) > 0 {
			*props.SeverityMapping = severityMapping
		}
	}

	// Set related integrations
	if props.RelatedIntegrations != nil && utils.IsKnown(d.RelatedIntegrations) {
		relatedIntegrations, relatedIntegrationsDiags := d.relatedIntegrationsToApi(ctx)
		diags.Append(relatedIntegrationsDiags...)
		if !relatedIntegrationsDiags.HasError() && relatedIntegrations != nil && len(*relatedIntegrations) > 0 {
			*props.RelatedIntegrations = relatedIntegrations
		}
	}

	// Set required fields
	if props.RequiredFields != nil && utils.IsKnown(d.RequiredFields) {
		requiredFields, requiredFieldsDiags := d.requiredFieldsToApi(ctx)
		diags.Append(requiredFieldsDiags...)
		if !requiredFieldsDiags.HasError() && requiredFields != nil && len(*requiredFields) > 0 {
			*props.RequiredFields = requiredFields
		}
	}

	// Set investigation fields
	if props.InvestigationFields != nil {
		investigationFields, investigationFieldsDiags := d.investigationFieldsToApi(ctx)
		if !investigationFieldsDiags.HasError() && investigationFields != nil {
			*props.InvestigationFields = investigationFields
		}
		diags.Append(investigationFieldsDiags...)
	}

	// Set response actions
	if props.ResponseActions != nil && utils.IsKnown(d.ResponseActions) {
		responseActions, responseActionsDiags := d.responseActionsToApi(ctx, client)
		diags.Append(responseActionsDiags...)
		if !responseActionsDiags.HasError() && len(responseActions) > 0 {
			*props.ResponseActions = &responseActions
		}
	}

	// Set filters
	if props.Filters != nil && utils.IsKnown(d.Filters) {
		filters, filtersDiags := d.filtersToApi(ctx)
		diags.Append(filtersDiags...)
		if !filtersDiags.HasError() && filters != nil {
			*props.Filters = filters
		}
	}

	// Set alert suppression
	if props.AlertSuppression != nil {
		alertSuppression := d.alertSuppressionToApi(ctx, diags)
		if alertSuppression != nil {
			*props.AlertSuppression = alertSuppression
		}
	}

	// Set threat (MITRE ATT&CK framework)
	if props.Threat != nil && utils.IsKnown(d.Threat) {
		threat, threatDiags := d.threatToApi(ctx)
		diags.Append(threatDiags...)
		if !threatDiags.HasError() && len(threat) > 0 {
			*props.Threat = &threat
		}
	}

	// Set timeline ID
	if props.TimelineId != nil && utils.IsKnown(d.TimelineId) {
		timelineId := kbapi.SecurityDetectionsAPITimelineTemplateId(d.TimelineId.ValueString())
		*props.TimelineId = &timelineId
	}

	// Set timeline title
	if props.TimelineTitle != nil && utils.IsKnown(d.TimelineTitle) {
		timelineTitle := kbapi.SecurityDetectionsAPITimelineTemplateTitle(d.TimelineTitle.ValueString())
		*props.TimelineTitle = &timelineTitle
	}
}

// Helper function to set common update properties across all rule types
func (d SecurityDetectionRuleData) setCommonUpdateProps(
	ctx context.Context,
	props *CommonUpdateProps,
	diags *diag.Diagnostics,
	client clients.MinVersionEnforceable,
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

	// Set actions
	if props.Actions != nil && utils.IsKnown(d.Actions) {
		actions, actionDiags := d.actionsToApi(ctx)
		diags.Append(actionDiags...)
		if !actionDiags.HasError() && len(actions) > 0 {
			*props.Actions = &actions
		}
	}

	// Set exceptions list
	if props.ExceptionsList != nil && utils.IsKnown(d.ExceptionsList) {
		exceptionsList, exceptionsListDiags := d.exceptionsListToApi(ctx)
		diags.Append(exceptionsListDiags...)
		if !exceptionsListDiags.HasError() && len(exceptionsList) > 0 {
			*props.ExceptionsList = &exceptionsList
		}
	}

	// Set risk score mapping
	if props.RiskScoreMapping != nil && utils.IsKnown(d.RiskScoreMapping) {
		riskScoreMapping, riskScoreMappingDiags := d.riskScoreMappingToApi(ctx)
		diags.Append(riskScoreMappingDiags...)
		if !riskScoreMappingDiags.HasError() && len(riskScoreMapping) > 0 {
			*props.RiskScoreMapping = &riskScoreMapping
		}
	}

	// Set building block type
	if props.BuildingBlockType != nil && utils.IsKnown(d.BuildingBlockType) {
		buildingBlockType := kbapi.SecurityDetectionsAPIBuildingBlockType(d.BuildingBlockType.ValueString())
		*props.BuildingBlockType = &buildingBlockType
	}

	// Set data view ID
	if props.DataViewId != nil && utils.IsKnown(d.DataViewId) {
		dataViewId := kbapi.SecurityDetectionsAPIDataViewId(d.DataViewId.ValueString())
		*props.DataViewId = &dataViewId
	}

	// Set namespace
	if props.Namespace != nil && utils.IsKnown(d.Namespace) {
		namespace := kbapi.SecurityDetectionsAPIAlertsIndexNamespace(d.Namespace.ValueString())
		*props.Namespace = &namespace
	}

	// Set rule name override
	if props.RuleNameOverride != nil && utils.IsKnown(d.RuleNameOverride) {
		ruleNameOverride := kbapi.SecurityDetectionsAPIRuleNameOverride(d.RuleNameOverride.ValueString())
		*props.RuleNameOverride = &ruleNameOverride
	}

	// Set timestamp override
	if props.TimestampOverride != nil && utils.IsKnown(d.TimestampOverride) {
		timestampOverride := kbapi.SecurityDetectionsAPITimestampOverride(d.TimestampOverride.ValueString())
		*props.TimestampOverride = &timestampOverride
	}

	// Set timestamp override fallback disabled
	if props.TimestampOverrideFallbackDisabled != nil && utils.IsKnown(d.TimestampOverrideFallbackDisabled) {
		timestampOverrideFallbackDisabled := kbapi.SecurityDetectionsAPITimestampOverrideFallbackDisabled(d.TimestampOverrideFallbackDisabled.ValueBool())
		*props.TimestampOverrideFallbackDisabled = &timestampOverrideFallbackDisabled
	}

	// Set severity mapping
	if props.SeverityMapping != nil && utils.IsKnown(d.SeverityMapping) {
		severityMapping, severityMappingDiags := d.severityMappingToApi(ctx)
		diags.Append(severityMappingDiags...)
		if !severityMappingDiags.HasError() && severityMapping != nil && len(*severityMapping) > 0 {
			*props.SeverityMapping = severityMapping
		}
	}

	// Set related integrations
	if props.RelatedIntegrations != nil && utils.IsKnown(d.RelatedIntegrations) {
		relatedIntegrations, relatedIntegrationsDiags := d.relatedIntegrationsToApi(ctx)
		diags.Append(relatedIntegrationsDiags...)
		if !relatedIntegrationsDiags.HasError() && relatedIntegrations != nil && len(*relatedIntegrations) > 0 {
			*props.RelatedIntegrations = relatedIntegrations
		}
	}

	// Set required fields
	if props.RequiredFields != nil && utils.IsKnown(d.RequiredFields) {
		requiredFields, requiredFieldsDiags := d.requiredFieldsToApi(ctx)
		diags.Append(requiredFieldsDiags...)
		if !requiredFieldsDiags.HasError() && requiredFields != nil && len(*requiredFields) > 0 {
			*props.RequiredFields = requiredFields
		}
	}

	// Set investigation fields
	if props.InvestigationFields != nil {
		investigationFields, investigationFieldsDiags := d.investigationFieldsToApi(ctx)
		if !investigationFieldsDiags.HasError() && investigationFields != nil {
			*props.InvestigationFields = investigationFields
		}
		diags.Append(investigationFieldsDiags...)
	}

	// Set response actions
	if props.ResponseActions != nil && utils.IsKnown(d.ResponseActions) {
		responseActions, responseActionsDiags := d.responseActionsToApi(ctx, client)
		diags.Append(responseActionsDiags...)
		if !responseActionsDiags.HasError() && len(responseActions) > 0 {
			*props.ResponseActions = &responseActions
		}
	}

	// Set filters
	if props.Filters != nil && utils.IsKnown(d.Filters) {
		filters, filtersDiags := d.filtersToApi(ctx)
		diags.Append(filtersDiags...)
		if !filtersDiags.HasError() && filters != nil {
			*props.Filters = filters
		}
	}

	// Set alert suppression
	if props.AlertSuppression != nil {
		alertSuppression := d.alertSuppressionToApi(ctx, diags)
		if alertSuppression != nil {
			*props.AlertSuppression = alertSuppression
		}
	}

	// Set threat (MITRE ATT&CK framework)
	if props.Threat != nil && utils.IsKnown(d.Threat) {
		threat, threatDiags := d.threatToApi(ctx)
		diags.Append(threatDiags...)
		if !threatDiags.HasError() && len(threat) > 0 {
			*props.Threat = &threat
		}
	}

	// Set timeline ID
	if props.TimelineId != nil && utils.IsKnown(d.TimelineId) {
		timelineId := kbapi.SecurityDetectionsAPITimelineTemplateId(d.TimelineId.ValueString())
		*props.TimelineId = &timelineId
	}

	// Set timeline title
	if props.TimelineTitle != nil && utils.IsKnown(d.TimelineTitle) {
		timelineTitle := kbapi.SecurityDetectionsAPITimelineTemplateTitle(d.TimelineTitle.ValueString())
		*props.TimelineTitle = &timelineTitle
	}
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

	// Initialize new common fields with proper empty lists
	if !utils.IsKnown(d.RelatedIntegrations) {
		d.RelatedIntegrations = types.ListNull(getRelatedIntegrationElementType())
	}
	if !utils.IsKnown(d.RequiredFields) {
		d.RequiredFields = types.ListNull(getRequiredFieldElementType())
	}
	if !utils.IsKnown(d.SeverityMapping) {
		d.SeverityMapping = types.ListNull(getSeverityMappingElementType())
	}

	// Initialize building block type to null by default
	if !utils.IsKnown(d.BuildingBlockType) {
		d.BuildingBlockType = types.StringNull()
	}

	// Actions field (common across all rule types)
	if !utils.IsKnown(d.Actions) {
		d.Actions = types.ListNull(getActionElementType())
	}

	// Exceptions list field (common across all rule types)
	if !utils.IsKnown(d.ExceptionsList) {
		d.ExceptionsList = types.ListNull(getExceptionsListElementType())
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
		d.ThreatMapping = types.ListNull(getThreatMappingElementType())
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
		d.Threshold = types.ObjectNull(getThresholdType())
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
		d.Threat = types.ListNull(getThreatElementType())
	}
}
