package security_detection_rule

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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

	// Meta field (common across all rule types) - Metadata object for the rule (gets overwritten when saving changes)
	Meta jsontypes.Normalized `tfsdk:"meta"`

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
	GroupBy               types.List   `tfsdk:"group_by"`
	Duration              types.Object `tfsdk:"duration"`
	MissingFieldsStrategy types.String `tfsdk:"missing_fields_strategy"`
}

type AlertSuppressionDurationModel struct {
	Value types.Int64  `tfsdk:"value"`
	Unit  types.String `tfsdk:"unit"`
}

type CardinalityModel struct {
	Field types.String `tfsdk:"field"`
	Value types.Int64  `tfsdk:"value"`
}

type ActionModel struct {
	ActionTypeId types.String `tfsdk:"action_type_id"`
	Id           types.String `tfsdk:"id"`
	Params       types.Map    `tfsdk:"params"`
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
	Meta                              **kbapi.SecurityDetectionsAPIRuleMetadata
	Filters                           **kbapi.SecurityDetectionsAPIRuleFilterArray
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
	Meta                              **kbapi.SecurityDetectionsAPIRuleMetadata
	Filters                           **kbapi.SecurityDetectionsAPIRuleFilterArray
}

func (d SecurityDetectionRuleData) toCreateProps(ctx context.Context, client clients.MinVersionEnforceable) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	ruleType := d.Type.ValueString()

	switch ruleType {
	case "query":
		return d.toQueryRuleCreateProps(ctx, client)
	case "eql":
		return d.toEqlRuleCreateProps(ctx, client)
	case "esql":
		return d.toEsqlRuleCreateProps(ctx, client)
	case "machine_learning":
		return d.toMachineLearningRuleCreateProps(ctx, client)
	case "new_terms":
		return d.toNewTermsRuleCreateProps(ctx, client)
	case "saved_query":
		return d.toSavedQueryRuleCreateProps(ctx, client)
	case "threat_match":
		return d.toThreatMatchRuleCreateProps(ctx, client)
	case "threshold":
		return d.toThresholdRuleCreateProps(ctx, client)
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
	if !utils.IsKnown(d.Language) {
		return nil
	}
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

	// Set meta
	if props.Meta != nil && utils.IsKnown(d.Meta) {
		meta, metaDiags := d.metaToApi(ctx)
		diags.Append(metaDiags...)
		if !metaDiags.HasError() && meta != nil {
			*props.Meta = meta
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
}

func (d SecurityDetectionRuleData) toUpdateProps(ctx context.Context, client clients.MinVersionEnforceable) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var updateProps kbapi.SecurityDetectionsAPIRuleUpdateProps

	ruleType := d.Type.ValueString()

	switch ruleType {
	case "query":
		return d.toQueryRuleUpdateProps(ctx, client)
	case "eql":
		return d.toEqlRuleUpdateProps(ctx, client)
	case "esql":
		return d.toEsqlRuleUpdateProps(ctx, client)
	case "machine_learning":
		return d.toMachineLearningRuleUpdateProps(ctx, client)
	case "new_terms":
		return d.toNewTermsRuleUpdateProps(ctx, client)
	case "saved_query":
		return d.toSavedQueryRuleUpdateProps(ctx, client)
	case "threat_match":
		return d.toThreatMatchRuleUpdateProps(ctx, client)
	case "threshold":
		return d.toThresholdRuleUpdateProps(ctx, client)
	default:
		diags.AddError(
			"Unsupported rule type",
			fmt.Sprintf("Rule type '%s' is not supported for updates", ruleType),
		)
		return updateProps, diags
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

	// Set meta
	if props.Meta != nil && utils.IsKnown(d.Meta) {
		meta, metaDiags := d.metaToApi(ctx)
		diags.Append(metaDiags...)
		if !metaDiags.HasError() && meta != nil {
			*props.Meta = meta
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

		entriesListValue, diags := types.ListValueFrom(ctx, getThreatMappingEntryElementType(), entries)
		if diags.HasError() {
			return types.ListNull(getThreatMappingElementType()), diags
		}

		threatMappings = append(threatMappings, SecurityDetectionRuleTfDataItem{
			Entries: entriesListValue,
		})
	}

	listValue, diags := types.ListValueFrom(ctx, getThreatMappingElementType(), threatMappings)
	return listValue, diags
}

// updateResponseActionsFromApi updates the ResponseActions field from API response
func (d *SecurityDetectionRuleData) updateResponseActionsFromApi(ctx context.Context, responseActions *[]kbapi.SecurityDetectionsAPIResponseAction) diag.Diagnostics {
	var diags diag.Diagnostics

	if responseActions != nil && len(*responseActions) > 0 {
		responseActionsValue, responseActionsDiags := convertResponseActionsToModel(ctx, responseActions)
		diags.Append(responseActionsDiags...)
		if !responseActionsDiags.HasError() {
			d.ResponseActions = responseActionsValue
		}
	} else {
		d.ResponseActions = types.ListNull(getResponseActionElementType())
	}

	return diags
}

// convertResponseActionsToModel converts kbapi response actions array to the terraform model
func convertResponseActionsToModel(ctx context.Context, apiResponseActions *[]kbapi.SecurityDetectionsAPIResponseAction) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if apiResponseActions == nil || len(*apiResponseActions) == 0 {
		return types.ListNull(getResponseActionElementType()), diags
	}

	var responseActions []ResponseActionModel

	for _, apiResponseAction := range *apiResponseActions {
		var responseAction ResponseActionModel

		// Use ValueByDiscriminator to get the concrete type
		actionValue, err := apiResponseAction.ValueByDiscriminator()
		if err != nil {
			diags.AddError("Failed to get response action discriminator", fmt.Sprintf("Error: %s", err.Error()))
			continue
		}

		switch concreteAction := actionValue.(type) {
		case kbapi.SecurityDetectionsAPIOsqueryResponseAction:
			convertedAction, convertDiags := convertOsqueryResponseActionToModel(ctx, concreteAction)
			diags.Append(convertDiags...)
			if !convertDiags.HasError() {
				responseAction = convertedAction
			}

		case kbapi.SecurityDetectionsAPIEndpointResponseAction:
			convertedAction, convertDiags := convertEndpointResponseActionToModel(ctx, concreteAction)
			diags.Append(convertDiags...)
			if !convertDiags.HasError() {
				responseAction = convertedAction
			}

		default:
			diags.AddError("Unknown response action type", fmt.Sprintf("Unsupported response action type: %T", concreteAction))
			continue
		}

		responseActions = append(responseActions, responseAction)
	}

	listValue, listDiags := types.ListValueFrom(ctx, getResponseActionElementType(), responseActions)
	if listDiags.HasError() {
		diags.Append(listDiags...)
	}

	return listValue, diags
}

// convertOsqueryResponseActionToModel converts an Osquery response action to the terraform model
func convertOsqueryResponseActionToModel(ctx context.Context, osqueryAction kbapi.SecurityDetectionsAPIOsqueryResponseAction) (ResponseActionModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	var responseAction ResponseActionModel

	responseAction.ActionTypeId = types.StringValue(string(osqueryAction.ActionTypeId))

	// Convert osquery params
	paramsModel := ResponseActionParamsModel{}
	paramsModel.Query = types.StringPointerValue(osqueryAction.Params.Query)
	if osqueryAction.Params.PackId != nil {
		paramsModel.PackId = types.StringPointerValue(osqueryAction.Params.PackId)
	} else {
		paramsModel.PackId = types.StringNull()
	}
	if osqueryAction.Params.SavedQueryId != nil {
		paramsModel.SavedQueryId = types.StringPointerValue(osqueryAction.Params.SavedQueryId)
	} else {
		paramsModel.SavedQueryId = types.StringNull()
	}
	if osqueryAction.Params.Timeout != nil {
		paramsModel.Timeout = types.Int64Value(int64(*osqueryAction.Params.Timeout))
	} else {
		paramsModel.Timeout = types.Int64Null()
	}

	// Convert ECS mapping
	if osqueryAction.Params.EcsMapping != nil {
		ecsMappingAttrs := make(map[string]attr.Value)
		for key, value := range *osqueryAction.Params.EcsMapping {
			if value.Field != nil {
				ecsMappingAttrs[key] = types.StringPointerValue(value.Field)
			} else {
				ecsMappingAttrs[key] = types.StringNull()
			}
		}
		ecsMappingValue, ecsDiags := types.MapValue(types.StringType, ecsMappingAttrs)
		if ecsDiags.HasError() {
			diags.Append(ecsDiags...)
		} else {
			paramsModel.EcsMapping = ecsMappingValue
		}
	} else {
		paramsModel.EcsMapping = types.MapNull(types.StringType)
	}

	// Convert queries array
	if osqueryAction.Params.Queries != nil {
		var queries []OsqueryQueryModel
		for _, apiQuery := range *osqueryAction.Params.Queries {
			query := OsqueryQueryModel{
				Id:    types.StringValue(apiQuery.Id),
				Query: types.StringValue(apiQuery.Query),
			}
			if apiQuery.Platform != nil {
				query.Platform = types.StringPointerValue(apiQuery.Platform)
			} else {
				query.Platform = types.StringNull()
			}
			if apiQuery.Version != nil {
				query.Version = types.StringPointerValue(apiQuery.Version)
			} else {
				query.Version = types.StringNull()
			}
			if apiQuery.Removed != nil {
				query.Removed = types.BoolPointerValue(apiQuery.Removed)
			} else {
				query.Removed = types.BoolNull()
			}
			if apiQuery.Snapshot != nil {
				query.Snapshot = types.BoolPointerValue(apiQuery.Snapshot)
			} else {
				query.Snapshot = types.BoolNull()
			}

			// Convert query ECS mapping
			if apiQuery.EcsMapping != nil {
				queryEcsMappingAttrs := make(map[string]attr.Value)
				for key, value := range *apiQuery.EcsMapping {
					if value.Field != nil {
						queryEcsMappingAttrs[key] = types.StringPointerValue(value.Field)
					} else {
						queryEcsMappingAttrs[key] = types.StringNull()
					}
				}
				queryEcsMappingValue, queryEcsDiags := types.MapValue(types.StringType, queryEcsMappingAttrs)
				if queryEcsDiags.HasError() {
					diags.Append(queryEcsDiags...)
				} else {
					query.EcsMapping = queryEcsMappingValue
				}
			} else {
				query.EcsMapping = types.MapNull(types.StringType)
			}

			queries = append(queries, query)
		}

		queriesListValue, queriesDiags := types.ListValueFrom(ctx, getOsqueryQueryElementType(), queries)
		if queriesDiags.HasError() {
			diags.Append(queriesDiags...)
		} else {
			paramsModel.Queries = queriesListValue
		}
	} else {
		paramsModel.Queries = types.ListNull(getOsqueryQueryElementType())
	}

	// Set remaining fields to null since this is osquery
	paramsModel.Command = types.StringNull()
	paramsModel.Comment = types.StringNull()
	paramsModel.Config = types.ObjectNull(getEndpointProcessConfigType())

	paramsObjectValue, paramsDiags := types.ObjectValueFrom(ctx, getResponseActionParamsType(), paramsModel)
	if paramsDiags.HasError() {
		diags.Append(paramsDiags...)
	} else {
		responseAction.Params = paramsObjectValue
	}

	return responseAction, diags
}

// convertEndpointResponseActionToModel converts an Endpoint response action to the terraform model
func convertEndpointResponseActionToModel(ctx context.Context, endpointAction kbapi.SecurityDetectionsAPIEndpointResponseAction) (ResponseActionModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	var responseAction ResponseActionModel

	responseAction.ActionTypeId = types.StringValue(string(endpointAction.ActionTypeId))

	// Convert endpoint params
	paramsModel := ResponseActionParamsModel{}

	commandParams, err := endpointAction.Params.AsSecurityDetectionsAPIDefaultParams()
	if err == nil {
		switch commandParams.Command {
		case "isolate":
			defaultParams, err := endpointAction.Params.AsSecurityDetectionsAPIDefaultParams()
			if err != nil {
				diags.AddError("Failed to parse endpoint default params", fmt.Sprintf("Error: %s", err.Error()))
			} else {
				paramsModel.Command = types.StringValue(string(defaultParams.Command))
				if defaultParams.Comment != nil {
					paramsModel.Comment = types.StringPointerValue(defaultParams.Comment)
				} else {
					paramsModel.Comment = types.StringNull()
				}
				paramsModel.Config = types.ObjectNull(getEndpointProcessConfigType())
			}
		case "kill-process", "suspend-process":
			processesParams, err := endpointAction.Params.AsSecurityDetectionsAPIProcessesParams()
			if err != nil {
				diags.AddError("Failed to parse endpoint processes params", fmt.Sprintf("Error: %s", err.Error()))
			} else {
				paramsModel.Command = types.StringValue(string(processesParams.Command))
				if processesParams.Comment != nil {
					paramsModel.Comment = types.StringPointerValue(processesParams.Comment)
				} else {
					paramsModel.Comment = types.StringNull()
				}

				// Convert config
				configModel := EndpointProcessConfigModel{
					Field: types.StringValue(processesParams.Config.Field),
				}
				if processesParams.Config.Overwrite != nil {
					configModel.Overwrite = types.BoolPointerValue(processesParams.Config.Overwrite)
				} else {
					configModel.Overwrite = types.BoolNull()
				}

				configObjectValue, configDiags := types.ObjectValueFrom(ctx, getEndpointProcessConfigType(), configModel)
				if configDiags.HasError() {
					diags.Append(configDiags...)
				} else {
					paramsModel.Config = configObjectValue
				}
			}
		}
	} else {
		diags.AddError("Unknown endpoint command", fmt.Sprintf("Unsupported endpoint command: %s. Error: %s", commandParams.Command, err.Error()))
	}

	// Set osquery fields to null since this is endpoint
	paramsModel.Query = types.StringNull()
	paramsModel.PackId = types.StringNull()
	paramsModel.SavedQueryId = types.StringNull()
	paramsModel.Timeout = types.Int64Null()
	paramsModel.EcsMapping = types.MapNull(types.StringType)
	paramsModel.Queries = types.ListNull(getOsqueryQueryElementType())

	paramsObjectValue, paramsDiags := types.ObjectValueFrom(ctx, getResponseActionParamsType(), paramsModel)
	if paramsDiags.HasError() {
		diags.Append(paramsDiags...)
	} else {
		responseAction.Params = paramsObjectValue
	}

	return responseAction, diags
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
		cardinalityList = utils.SliceToListType(ctx, *apiThreshold.Cardinality, getCardinalityType(), path.Root("threshold").AtName("cardinality"), &diags,
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
		cardinalityList = types.ListNull(getCardinalityType())
	}

	thresholdModel := ThresholdModel{
		Field:       fieldList,
		Value:       types.Int64Value(int64(apiThreshold.Value)),
		Cardinality: cardinalityList,
	}

	thresholdObject, objDiags := types.ObjectValueFrom(ctx, getThresholdType(), thresholdModel)
	diags.Append(objDiags...)
	return thresholdObject, diags
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

// Helper function to convert alert suppression from TF data to API type
func (d SecurityDetectionRuleData) alertSuppressionToApi(ctx context.Context, diags *diag.Diagnostics) *kbapi.SecurityDetectionsAPIAlertSuppression {
	if !utils.IsKnown(d.AlertSuppression) {
		return nil
	}

	var model AlertSuppressionModel
	objDiags := d.AlertSuppression.As(ctx, &model, basetypes.ObjectAsOptions{})
	diags.Append(objDiags...)
	if diags.HasError() {
		return nil
	}

	suppression := &kbapi.SecurityDetectionsAPIAlertSuppression{}

	// Handle group_by (required)
	if utils.IsKnown(model.GroupBy) {
		groupByList := utils.ListTypeToSlice_String(ctx, model.GroupBy, path.Root("alert_suppression").AtName("group_by"), diags)
		if len(groupByList) > 0 {
			suppression.GroupBy = groupByList
		}
	}

	// Handle duration (optional)
	if utils.IsKnown(model.Duration) {
		var durationModel AlertSuppressionDurationModel
		durationDiags := model.Duration.As(ctx, &durationModel, basetypes.ObjectAsOptions{})
		diags.Append(durationDiags...)
		if !diags.HasError() {
			duration := kbapi.SecurityDetectionsAPIAlertSuppressionDuration{
				Value: int(durationModel.Value.ValueInt64()),
				Unit:  kbapi.SecurityDetectionsAPIAlertSuppressionDurationUnit(durationModel.Unit.ValueString()),
			}
			suppression.Duration = &duration
		}
	}

	// Handle missing_fields_strategy (optional)
	if utils.IsKnown(model.MissingFieldsStrategy) {
		strategy := kbapi.SecurityDetectionsAPIAlertSuppressionMissingFieldsStrategy(model.MissingFieldsStrategy.ValueString())
		suppression.MissingFieldsStrategy = &strategy
	}

	return suppression
}

// Helper function to convert alert suppression from TF data to threshold-specific API type
func (d SecurityDetectionRuleData) alertSuppressionToThresholdApi(ctx context.Context, diags *diag.Diagnostics) *kbapi.SecurityDetectionsAPIThresholdAlertSuppression {
	if !utils.IsKnown(d.AlertSuppression) {
		return nil
	}

	var model AlertSuppressionModel
	objDiags := d.AlertSuppression.As(ctx, &model, basetypes.ObjectAsOptions{})
	diags.Append(objDiags...)
	if diags.HasError() {
		return nil
	}

	suppression := &kbapi.SecurityDetectionsAPIThresholdAlertSuppression{}

	// Handle duration (required for threshold alert suppression)
	if !utils.IsKnown(model.Duration) {
		diags.AddError(
			"Duration required for threshold alert suppression",
			"Threshold alert suppression requires a duration to be specified",
		)
		return nil
	}

	var durationModel AlertSuppressionDurationModel
	durationDiags := model.Duration.As(ctx, &durationModel, basetypes.ObjectAsOptions{})
	diags.Append(durationDiags...)
	if !diags.HasError() {
		duration := kbapi.SecurityDetectionsAPIAlertSuppressionDuration{
			Value: int(durationModel.Value.ValueInt64()),
			Unit:  kbapi.SecurityDetectionsAPIAlertSuppressionDurationUnit(durationModel.Unit.ValueString()),
		}
		suppression.Duration = duration
	}

	// Note: Threshold alert suppression only supports duration field.
	// GroupBy and MissingFieldsStrategy are not supported for threshold rules.

	return suppression
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

// Helper function to process response actions configuration for all rule types
func (d SecurityDetectionRuleData) responseActionsToApi(ctx context.Context, client clients.MinVersionEnforceable) ([]kbapi.SecurityDetectionsAPIResponseAction, diag.Diagnostics) {
	var diags diag.Diagnostics

	if client == nil {
		diags.AddError(
			"Client is not initialized",
			"Response actions require a valid API client",
		)
		return nil, diags
	}

	if !utils.IsKnown(d.ResponseActions) || len(d.ResponseActions.Elements()) == 0 {
		return nil, diags
	}

	// Check version support for response actions
	if supported, versionDiags := client.EnforceMinVersion(ctx, MinVersionResponseActions); versionDiags.HasError() {
		diags.Append(diagutil.FrameworkDiagsFromSDK(versionDiags)...)
		return nil, diags
	} else if !supported {
		// Version is not supported, return nil without error
		diags.AddError("Response actions are unsupported",
			fmt.Sprintf("Response actions require server version %s or higher", MinVersionResponseActions.String()))
		return nil, diags
	}

	apiResponseActions := utils.ListTypeToSlice(ctx, d.ResponseActions, path.Root("response_actions"), &diags,
		func(responseAction ResponseActionModel, meta utils.ListMeta) kbapi.SecurityDetectionsAPIResponseAction {
			if responseAction.ActionTypeId.IsNull() {
				return kbapi.SecurityDetectionsAPIResponseAction{}
			}

			actionTypeId := responseAction.ActionTypeId.ValueString()

			// Extract params using ObjectTypeToStruct
			if responseAction.Params.IsNull() || responseAction.Params.IsUnknown() {
				return kbapi.SecurityDetectionsAPIResponseAction{}
			}

			params := utils.ObjectTypeToStruct(ctx, responseAction.Params, meta.Path.AtName("params"), &diags,
				func(item ResponseActionParamsModel, meta utils.ObjectMeta) ResponseActionParamsModel {
					return item
				})

			if params == nil {
				return kbapi.SecurityDetectionsAPIResponseAction{}
			}

			switch actionTypeId {
			case ".osquery":
				apiAction, actionDiags := d.buildOsqueryResponseAction(ctx, *params)
				diags.Append(actionDiags...)
				return apiAction

			case ".endpoint":
				apiAction, actionDiags := d.buildEndpointResponseAction(ctx, *params)
				diags.Append(actionDiags...)
				return apiAction

			default:
				diags.AddError(
					"Unsupported action_type_id in response actions",
					fmt.Sprintf("action_type_id '%s' is not supported", actionTypeId),
				)
				return kbapi.SecurityDetectionsAPIResponseAction{}
			}
		})

	return apiResponseActions, diags
}

// buildOsqueryResponseAction creates an Osquery response action from the terraform model
func (d SecurityDetectionRuleData) buildOsqueryResponseAction(ctx context.Context, params ResponseActionParamsModel) (kbapi.SecurityDetectionsAPIResponseAction, diag.Diagnostics) {
	var diags diag.Diagnostics

	osqueryAction := kbapi.SecurityDetectionsAPIOsqueryResponseAction{
		ActionTypeId: kbapi.SecurityDetectionsAPIOsqueryResponseActionActionTypeId(".osquery"),
		Params:       kbapi.SecurityDetectionsAPIOsqueryParams{},
	}

	// Set osquery-specific params
	if utils.IsKnown(params.Query) {
		osqueryAction.Params.Query = params.Query.ValueStringPointer()
	}
	if utils.IsKnown(params.PackId) {
		osqueryAction.Params.PackId = params.PackId.ValueStringPointer()
	}
	if utils.IsKnown(params.SavedQueryId) {
		osqueryAction.Params.SavedQueryId = params.SavedQueryId.ValueStringPointer()
	}
	if utils.IsKnown(params.Timeout) {
		timeout := float32(params.Timeout.ValueInt64())
		osqueryAction.Params.Timeout = &timeout
	}
	if utils.IsKnown(params.EcsMapping) && !params.EcsMapping.IsNull() {

		// Convert map to ECS mapping structure
		ecsMappingElems := make(map[string]basetypes.StringValue)
		elemDiags := params.EcsMapping.ElementsAs(ctx, &ecsMappingElems, false)
		if !elemDiags.HasError() {
			ecsMapping := make(kbapi.SecurityDetectionsAPIEcsMapping)
			for key, value := range ecsMappingElems {
				if stringVal := value; utils.IsKnown(value) {
					ecsMapping[key] = struct {
						Field *string                                      `json:"field,omitempty"`
						Value *kbapi.SecurityDetectionsAPIEcsMapping_Value `json:"value,omitempty"`
					}{
						Field: stringVal.ValueStringPointer(),
					}
				}
			}
			osqueryAction.Params.EcsMapping = &ecsMapping
		} else {
			diags.Append(elemDiags...)
		}
	}
	if utils.IsKnown(params.Queries) && !params.Queries.IsNull() {
		queries := make([]OsqueryQueryModel, len(params.Queries.Elements()))
		queriesDiags := params.Queries.ElementsAs(ctx, &queries, false)
		if !queriesDiags.HasError() {
			apiQueries := make([]kbapi.SecurityDetectionsAPIOsqueryQuery, 0)
			for _, query := range queries {
				apiQuery := kbapi.SecurityDetectionsAPIOsqueryQuery{
					Id:    query.Id.ValueString(),
					Query: query.Query.ValueString(),
				}
				if utils.IsKnown(query.Platform) {
					apiQuery.Platform = query.Platform.ValueStringPointer()
				}
				if utils.IsKnown(query.Version) {
					apiQuery.Version = query.Version.ValueStringPointer()
				}
				if utils.IsKnown(query.Removed) {
					apiQuery.Removed = query.Removed.ValueBoolPointer()
				}
				if utils.IsKnown(query.Snapshot) {
					apiQuery.Snapshot = query.Snapshot.ValueBoolPointer()
				}
				if utils.IsKnown(query.EcsMapping) && !query.EcsMapping.IsNull() {
					// Convert map to ECS mapping structure for queries
					queryEcsMappingElems := make(map[string]basetypes.StringValue)
					queryElemDiags := query.EcsMapping.ElementsAs(ctx, &queryEcsMappingElems, false)
					if !queryElemDiags.HasError() {
						queryEcsMapping := make(kbapi.SecurityDetectionsAPIEcsMapping)
						for key, value := range queryEcsMappingElems {
							if stringVal := value; utils.IsKnown(value) {
								queryEcsMapping[key] = struct {
									Field *string                                      `json:"field,omitempty"`
									Value *kbapi.SecurityDetectionsAPIEcsMapping_Value `json:"value,omitempty"`
								}{
									Field: stringVal.ValueStringPointer(),
								}
							}
						}
						apiQuery.EcsMapping = &queryEcsMapping
					}
				}
				apiQueries = append(apiQueries, apiQuery)
			}
			osqueryAction.Params.Queries = &apiQueries
		} else {
			diags = append(diags, queriesDiags...)
		}
	}

	var apiResponseAction kbapi.SecurityDetectionsAPIResponseAction
	err := apiResponseAction.FromSecurityDetectionsAPIOsqueryResponseAction(osqueryAction)
	if err != nil {
		diags.AddError("Error converting osquery response action", err.Error())
	}

	return apiResponseAction, diags
}

// buildEndpointResponseAction creates an Endpoint response action from the terraform model
func (d SecurityDetectionRuleData) buildEndpointResponseAction(ctx context.Context, params ResponseActionParamsModel) (kbapi.SecurityDetectionsAPIResponseAction, diag.Diagnostics) {
	var diags diag.Diagnostics

	endpointAction := kbapi.SecurityDetectionsAPIEndpointResponseAction{
		ActionTypeId: kbapi.SecurityDetectionsAPIEndpointResponseActionActionTypeId(".endpoint"),
	}

	// Determine the type of endpoint action based on the command
	if utils.IsKnown(params.Command) {
		command := params.Command.ValueString()
		switch command {
		case "isolate":
			// Use DefaultParams for isolate command
			defaultParams := kbapi.SecurityDetectionsAPIDefaultParams{
				Command: kbapi.SecurityDetectionsAPIDefaultParamsCommand("isolate"),
			}
			if utils.IsKnown(params.Comment) {
				defaultParams.Comment = params.Comment.ValueStringPointer()
			}
			err := endpointAction.Params.FromSecurityDetectionsAPIDefaultParams(defaultParams)
			if err != nil {
				diags.AddError("Error setting endpoint default params", err.Error())
				return kbapi.SecurityDetectionsAPIResponseAction{}, diags
			}

		case "kill-process", "suspend-process":
			// Use ProcessesParams for process commands
			processesParams := kbapi.SecurityDetectionsAPIProcessesParams{
				Command: kbapi.SecurityDetectionsAPIProcessesParamsCommand(command),
			}
			if utils.IsKnown(params.Comment) {
				processesParams.Comment = params.Comment.ValueStringPointer()
			}

			// Set config if provided
			if !params.Config.IsNull() && !params.Config.IsUnknown() {
				config := utils.ObjectTypeToStruct(ctx, params.Config, path.Root("response_actions").AtName("params").AtName("config"), &diags,
					func(item EndpointProcessConfigModel, meta utils.ObjectMeta) EndpointProcessConfigModel {
						return item
					})

				processesParams.Config = struct {
					Field     string `json:"field"`
					Overwrite *bool  `json:"overwrite,omitempty"`
				}{
					Field: config.Field.ValueString(),
				}
				if utils.IsKnown(config.Overwrite) {
					processesParams.Config.Overwrite = config.Overwrite.ValueBoolPointer()
				}
			}

			err := endpointAction.Params.FromSecurityDetectionsAPIProcessesParams(processesParams)
			if err != nil {
				diags.AddError("Error setting endpoint processes params", err.Error())
				return kbapi.SecurityDetectionsAPIResponseAction{}, diags
			}
		default:
			diags.AddError(
				"Unsupported params type",
				fmt.Sprintf("Params type '%s' is not supported", params.Command.ValueString()),
			)
		}
	}

	var apiResponseAction kbapi.SecurityDetectionsAPIResponseAction
	err := apiResponseAction.FromSecurityDetectionsAPIEndpointResponseAction(endpointAction)
	if err != nil {
		diags.AddError("Error converting endpoint response action", err.Error())
	}

	return apiResponseAction, diags
}

// Helper function to process actions configuration for all rule types
func (d SecurityDetectionRuleData) actionsToApi(ctx context.Context) ([]kbapi.SecurityDetectionsAPIRuleAction, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !utils.IsKnown(d.Actions) || len(d.Actions.Elements()) == 0 {
		return nil, diags
	}

	apiActions := utils.ListTypeToSlice(ctx, d.Actions, path.Root("actions"), &diags,
		func(action ActionModel, meta utils.ListMeta) kbapi.SecurityDetectionsAPIRuleAction {
			if action.ActionTypeId.IsNull() || action.Id.IsNull() {
				return kbapi.SecurityDetectionsAPIRuleAction{}
			}

			apiAction := kbapi.SecurityDetectionsAPIRuleAction{
				ActionTypeId: action.ActionTypeId.ValueString(),
				Id:           kbapi.SecurityDetectionsAPIRuleActionId(action.Id.ValueString()),
			}

			// Convert params map
			if utils.IsKnown(action.Params) {
				paramsStringMap := make(map[string]string)
				paramsDiags := action.Params.ElementsAs(meta.Context, &paramsStringMap, false)
				if !paramsDiags.HasError() {
					paramsMap := make(map[string]interface{})
					for k, v := range paramsStringMap {
						paramsMap[k] = v
					}
					apiAction.Params = kbapi.SecurityDetectionsAPIRuleActionParams(paramsMap)
				}
				meta.Diags.Append(paramsDiags...)
			}

			// Set optional fields
			if utils.IsKnown(action.Group) {
				group := kbapi.SecurityDetectionsAPIRuleActionGroup(action.Group.ValueString())
				apiAction.Group = &group
			}

			if utils.IsKnown(action.Uuid) {
				uuid := kbapi.SecurityDetectionsAPINonEmptyString(action.Uuid.ValueString())
				apiAction.Uuid = &uuid
			}

			if utils.IsKnown(action.AlertsFilter) {
				alertsFilterStringMap := make(map[string]string)
				alertsFilterDiags := action.AlertsFilter.ElementsAs(meta.Context, &alertsFilterStringMap, false)
				if !alertsFilterDiags.HasError() {
					alertsFilterMap := make(map[string]interface{})
					for k, v := range alertsFilterStringMap {
						alertsFilterMap[k] = v
					}
					apiAlertsFilter := kbapi.SecurityDetectionsAPIRuleActionAlertsFilter(alertsFilterMap)
					apiAction.AlertsFilter = &apiAlertsFilter
				}
				meta.Diags.Append(alertsFilterDiags...)
			}

			// Handle frequency using ObjectTypeToStruct
			if utils.IsKnown(action.Frequency) {
				frequency := utils.ObjectTypeToStruct(meta.Context, action.Frequency, meta.Path.AtName("frequency"), meta.Diags,
					func(frequencyModel ActionFrequencyModel, freqMeta utils.ObjectMeta) kbapi.SecurityDetectionsAPIRuleActionFrequency {
						apiFreq := kbapi.SecurityDetectionsAPIRuleActionFrequency{
							NotifyWhen: kbapi.SecurityDetectionsAPIRuleActionNotifyWhen(frequencyModel.NotifyWhen.ValueString()),
							Summary:    frequencyModel.Summary.ValueBool(),
						}

						// Handle throttle - can be string or specific values
						if utils.IsKnown(frequencyModel.Throttle) {
							throttleStr := frequencyModel.Throttle.ValueString()
							var throttle kbapi.SecurityDetectionsAPIRuleActionThrottle
							if throttleStr == "no_actions" || throttleStr == "rule" {
								// Use the enum value
								var throttle0 kbapi.SecurityDetectionsAPIRuleActionThrottle0
								if throttleStr == "no_actions" {
									throttle0 = kbapi.SecurityDetectionsAPIRuleActionThrottle0NoActions
								} else {
									throttle0 = kbapi.SecurityDetectionsAPIRuleActionThrottle0Rule
								}
								err := throttle.FromSecurityDetectionsAPIRuleActionThrottle0(throttle0)
								if err != nil {
									freqMeta.Diags.AddError("Error setting throttle enum", err.Error())
								}
							} else {
								// Use the time interval string
								throttle1 := kbapi.SecurityDetectionsAPIRuleActionThrottle1(throttleStr)
								err := throttle.FromSecurityDetectionsAPIRuleActionThrottle1(throttle1)
								if err != nil {
									freqMeta.Diags.AddError("Error setting throttle interval", err.Error())
								}
							}
							apiFreq.Throttle = throttle
						}

						return apiFreq
					})

				if frequency != nil {
					apiAction.Frequency = frequency
				}
			}

			return apiAction
		})

	// Filter out empty actions (where ActionTypeId or Id was null)
	validActions := make([]kbapi.SecurityDetectionsAPIRuleAction, 0)
	for _, action := range apiActions {
		if action.ActionTypeId != "" && action.Id != "" {
			validActions = append(validActions, action)
		}
	}

	return validActions, diags
}

// convertActionsToModel converts kbapi.SecurityDetectionsAPIRuleAction slice to Terraform model
func convertActionsToModel(ctx context.Context, apiActions []kbapi.SecurityDetectionsAPIRuleAction) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(apiActions) == 0 {
		return types.ListNull(getActionElementType()), diags
	}

	actions := make([]ActionModel, 0)

	for _, apiAction := range apiActions {
		action := ActionModel{
			ActionTypeId: types.StringValue(apiAction.ActionTypeId),
			Id:           types.StringValue(string(apiAction.Id)),
		}

		// Convert params
		if apiAction.Params != nil {
			paramsMap := make(map[string]attr.Value)
			for k, v := range apiAction.Params {
				if v != nil {
					paramsMap[k] = types.StringValue(fmt.Sprintf("%v", v))
				}
			}
			paramsValue, paramsDiags := types.MapValue(types.StringType, paramsMap)
			diags.Append(paramsDiags...)
			action.Params = paramsValue
		} else {
			action.Params = types.MapNull(types.StringType)
		}

		// Set optional fields
		if apiAction.Group != nil {
			action.Group = types.StringValue(string(*apiAction.Group))
		} else {
			action.Group = types.StringNull()
		}

		if apiAction.Uuid != nil {
			action.Uuid = types.StringValue(string(*apiAction.Uuid))
		} else {
			action.Uuid = types.StringNull()
		}

		if apiAction.AlertsFilter != nil {
			alertsFilterMap := make(map[string]attr.Value)
			for k, v := range *apiAction.AlertsFilter {
				if v != nil {
					alertsFilterMap[k] = types.StringValue(fmt.Sprintf("%v", v))
				}
			}
			alertsFilterValue, alertsFilterDiags := types.MapValue(types.StringType, alertsFilterMap)
			diags.Append(alertsFilterDiags...)
			action.AlertsFilter = alertsFilterValue
		} else {
			action.AlertsFilter = types.MapNull(types.StringType)
		}

		// Convert frequency
		if apiAction.Frequency != nil {
			var throttleStr string
			if throttle0, err := apiAction.Frequency.Throttle.AsSecurityDetectionsAPIRuleActionThrottle0(); err == nil {
				throttleStr = string(throttle0)
			} else if throttle1, err := apiAction.Frequency.Throttle.AsSecurityDetectionsAPIRuleActionThrottle1(); err == nil {
				throttleStr = string(throttle1)
			}

			frequencyModel := ActionFrequencyModel{
				NotifyWhen: types.StringValue(string(apiAction.Frequency.NotifyWhen)),
				Summary:    types.BoolValue(apiAction.Frequency.Summary),
				Throttle:   types.StringValue(throttleStr),
			}

			frequencyObj, frequencyDiags := types.ObjectValueFrom(ctx, getActionFrequencyType(), frequencyModel)
			diags.Append(frequencyDiags...)
			action.Frequency = frequencyObj
		} else {
			action.Frequency = types.ObjectNull(getActionFrequencyType())
		}

		actions = append(actions, action)
	}

	listValue, listDiags := types.ListValueFrom(ctx, getActionElementType(), actions)
	diags.Append(listDiags...)
	return listValue, diags
}

// Helper function to update actions from API response
func (d *SecurityDetectionRuleData) updateActionsFromApi(ctx context.Context, actions []kbapi.SecurityDetectionsAPIRuleAction) diag.Diagnostics {
	var diags diag.Diagnostics

	if len(actions) > 0 {
		actionsListValue, actionDiags := convertActionsToModel(ctx, actions)
		diags.Append(actionDiags...)
		if !actionDiags.HasError() {
			d.Actions = actionsListValue
		}
	} else {
		d.Actions = types.ListNull(getActionElementType())
	}

	return diags
}

func (d *SecurityDetectionRuleData) updateAlertSuppressionFromApi(ctx context.Context, apiSuppression *kbapi.SecurityDetectionsAPIAlertSuppression) diag.Diagnostics {
	var diags diag.Diagnostics

	if apiSuppression == nil {
		d.AlertSuppression = types.ObjectNull(getAlertSuppressionType())
		return diags
	}

	model := AlertSuppressionModel{}

	// Convert group_by (required field according to API)
	if len(apiSuppression.GroupBy) > 0 {
		groupByList := make([]attr.Value, len(apiSuppression.GroupBy))
		for i, field := range apiSuppression.GroupBy {
			groupByList[i] = types.StringValue(field)
		}
		model.GroupBy = types.ListValueMust(types.StringType, groupByList)
	} else {
		model.GroupBy = types.ListNull(types.StringType)
	}

	// Convert duration (optional)
	if apiSuppression.Duration != nil {
		durationModel := AlertSuppressionDurationModel{
			Value: types.Int64Value(int64(apiSuppression.Duration.Value)),
			Unit:  types.StringValue(string(apiSuppression.Duration.Unit)),
		}
		durationObj, durationDiags := types.ObjectValueFrom(ctx, getDurationType(), durationModel)
		diags.Append(durationDiags...)
		model.Duration = durationObj
	} else {
		model.Duration = types.ObjectNull(getDurationType())
	}

	// Convert missing_fields_strategy (optional)
	if apiSuppression.MissingFieldsStrategy != nil {
		model.MissingFieldsStrategy = types.StringValue(string(*apiSuppression.MissingFieldsStrategy))
	} else {
		model.MissingFieldsStrategy = types.StringNull()
	}

	alertSuppressionObj, objDiags := types.ObjectValueFrom(ctx, getAlertSuppressionType(), model)
	diags.Append(objDiags...)

	d.AlertSuppression = alertSuppressionObj

	return diags
}

func (d *SecurityDetectionRuleData) updateThresholdAlertSuppressionFromApi(ctx context.Context, apiSuppression *kbapi.SecurityDetectionsAPIThresholdAlertSuppression) diag.Diagnostics {
	var diags diag.Diagnostics

	if apiSuppression == nil {
		d.AlertSuppression = types.ObjectNull(getAlertSuppressionType())
		return diags
	}

	model := AlertSuppressionModel{}

	// Threshold alert suppression only has duration field, so we set group_by and missing_fields_strategy to null
	model.GroupBy = types.ListNull(types.StringType)
	model.MissingFieldsStrategy = types.StringNull()

	// Convert duration (always present in threshold alert suppression)
	durationModel := AlertSuppressionDurationModel{
		Value: types.Int64Value(int64(apiSuppression.Duration.Value)),
		Unit:  types.StringValue(string(apiSuppression.Duration.Unit)),
	}
	durationObj, durationDiags := types.ObjectValueFrom(ctx, getDurationType(), durationModel)
	diags.Append(durationDiags...)
	model.Duration = durationObj

	alertSuppressionObj, objDiags := types.ObjectValueFrom(ctx, getAlertSuppressionType(), model)
	diags.Append(objDiags...)

	d.AlertSuppression = alertSuppressionObj

	return diags
}

// Helper function to process exceptions list configuration for all rule types
func (d SecurityDetectionRuleData) exceptionsListToApi(ctx context.Context) ([]kbapi.SecurityDetectionsAPIRuleExceptionList, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !utils.IsKnown(d.ExceptionsList) || len(d.ExceptionsList.Elements()) == 0 {
		return nil, diags
	}

	apiExceptionsList := utils.ListTypeToSlice(ctx, d.ExceptionsList, path.Root("exceptions_list"), &diags,
		func(exception ExceptionsListModel, meta utils.ListMeta) kbapi.SecurityDetectionsAPIRuleExceptionList {
			if exception.Id.IsNull() || exception.ListId.IsNull() || exception.NamespaceType.IsNull() || exception.Type.IsNull() {
				return kbapi.SecurityDetectionsAPIRuleExceptionList{}
			}

			apiException := kbapi.SecurityDetectionsAPIRuleExceptionList{
				Id:            exception.Id.ValueString(),
				ListId:        exception.ListId.ValueString(),
				NamespaceType: kbapi.SecurityDetectionsAPIRuleExceptionListNamespaceType(exception.NamespaceType.ValueString()),
				Type:          kbapi.SecurityDetectionsAPIExceptionListType(exception.Type.ValueString()),
			}

			return apiException
		})

	// Filter out empty exceptions (where required fields were null)
	validExceptions := make([]kbapi.SecurityDetectionsAPIRuleExceptionList, 0)
	for _, exception := range apiExceptionsList {
		if exception.Id != "" && exception.ListId != "" {
			validExceptions = append(validExceptions, exception)
		}
	}

	return validExceptions, diags
}

// convertExceptionsListToModel converts kbapi.SecurityDetectionsAPIRuleExceptionList slice to Terraform model
func convertExceptionsListToModel(ctx context.Context, apiExceptionsList []kbapi.SecurityDetectionsAPIRuleExceptionList) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(apiExceptionsList) == 0 {
		return types.ListNull(getExceptionsListElementType()), diags
	}

	exceptions := make([]ExceptionsListModel, 0)

	for _, apiException := range apiExceptionsList {
		exception := ExceptionsListModel{
			Id:            types.StringValue(apiException.Id),
			ListId:        types.StringValue(apiException.ListId),
			NamespaceType: types.StringValue(string(apiException.NamespaceType)),
			Type:          types.StringValue(string(apiException.Type)),
		}

		exceptions = append(exceptions, exception)
	}

	listValue, listDiags := types.ListValueFrom(ctx, getExceptionsListElementType(), exceptions)
	diags.Append(listDiags...)
	return listValue, diags
}

// Helper function to update exceptions list from API response
func (d *SecurityDetectionRuleData) updateExceptionsListFromApi(ctx context.Context, exceptionsList []kbapi.SecurityDetectionsAPIRuleExceptionList) diag.Diagnostics {
	var diags diag.Diagnostics

	if len(exceptionsList) > 0 {
		exceptionsListValue, exceptionsListDiags := convertExceptionsListToModel(ctx, exceptionsList)
		diags.Append(exceptionsListDiags...)
		if !exceptionsListDiags.HasError() {
			d.ExceptionsList = exceptionsListValue
		}
	} else {
		d.ExceptionsList = types.ListNull(getExceptionsListElementType())
	}

	return diags
}

// Helper function to process risk score mapping configuration for all rule types
func (d SecurityDetectionRuleData) riskScoreMappingToApi(ctx context.Context) (kbapi.SecurityDetectionsAPIRiskScoreMapping, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !utils.IsKnown(d.RiskScoreMapping) || len(d.RiskScoreMapping.Elements()) == 0 {
		return nil, diags
	}

	apiRiskScoreMapping := utils.ListTypeToSlice(ctx, d.RiskScoreMapping, path.Root("risk_score_mapping"), &diags,
		func(mapping RiskScoreMappingModel, meta utils.ListMeta) struct {
			Field     string                                              `json:"field"`
			Operator  kbapi.SecurityDetectionsAPIRiskScoreMappingOperator `json:"operator"`
			RiskScore *kbapi.SecurityDetectionsAPIRiskScore               `json:"risk_score,omitempty"`
			Value     string                                              `json:"value"`
		} {
			if mapping.Field.IsNull() || mapping.Operator.IsNull() || mapping.Value.IsNull() {
				return struct {
					Field     string                                              `json:"field"`
					Operator  kbapi.SecurityDetectionsAPIRiskScoreMappingOperator `json:"operator"`
					RiskScore *kbapi.SecurityDetectionsAPIRiskScore               `json:"risk_score,omitempty"`
					Value     string                                              `json:"value"`
				}{}
			}

			apiMapping := struct {
				Field     string                                              `json:"field"`
				Operator  kbapi.SecurityDetectionsAPIRiskScoreMappingOperator `json:"operator"`
				RiskScore *kbapi.SecurityDetectionsAPIRiskScore               `json:"risk_score,omitempty"`
				Value     string                                              `json:"value"`
			}{
				Field:    mapping.Field.ValueString(),
				Operator: kbapi.SecurityDetectionsAPIRiskScoreMappingOperator(mapping.Operator.ValueString()),
				Value:    mapping.Value.ValueString(),
			}

			// Set optional risk score if provided
			if utils.IsKnown(mapping.RiskScore) {
				riskScore := kbapi.SecurityDetectionsAPIRiskScore(mapping.RiskScore.ValueInt64())
				apiMapping.RiskScore = &riskScore
			}

			return apiMapping
		})

	// Filter out empty mappings (where required fields were null)
	validMappings := make(kbapi.SecurityDetectionsAPIRiskScoreMapping, 0)
	for _, mapping := range apiRiskScoreMapping {
		if mapping.Field != "" && mapping.Operator != "" && mapping.Value != "" {
			validMappings = append(validMappings, mapping)
		}
	}

	return validMappings, diags
}

// convertRiskScoreMappingToModel converts kbapi.SecurityDetectionsAPIRiskScoreMapping to Terraform model
func convertRiskScoreMappingToModel(ctx context.Context, apiRiskScoreMapping kbapi.SecurityDetectionsAPIRiskScoreMapping) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(apiRiskScoreMapping) == 0 {
		return types.ListNull(getRiskScoreMappingElementType()), diags
	}

	mappings := make([]RiskScoreMappingModel, 0)

	for _, apiMapping := range apiRiskScoreMapping {
		mapping := RiskScoreMappingModel{
			Field:    types.StringValue(apiMapping.Field),
			Operator: types.StringValue(string(apiMapping.Operator)),
			Value:    types.StringValue(apiMapping.Value),
		}

		// Set optional risk score if provided
		if apiMapping.RiskScore != nil {
			mapping.RiskScore = types.Int64Value(int64(*apiMapping.RiskScore))
		} else {
			mapping.RiskScore = types.Int64Null()
		}

		mappings = append(mappings, mapping)
	}

	listValue, listDiags := types.ListValueFrom(ctx, getRiskScoreMappingElementType(), mappings)
	diags.Append(listDiags...)
	return listValue, diags
}

// Helper function to update risk score mapping from API response
func (d *SecurityDetectionRuleData) updateRiskScoreMappingFromApi(ctx context.Context, riskScoreMapping kbapi.SecurityDetectionsAPIRiskScoreMapping) diag.Diagnostics {
	var diags diag.Diagnostics

	if len(riskScoreMapping) > 0 {
		riskScoreMappingValue, riskScoreMappingDiags := convertRiskScoreMappingToModel(ctx, riskScoreMapping)
		diags.Append(riskScoreMappingDiags...)
		if !riskScoreMappingDiags.HasError() {
			d.RiskScoreMapping = riskScoreMappingValue
		}
	} else {
		d.RiskScoreMapping = types.ListNull(getRiskScoreMappingElementType())
	}

	return diags
}

// Helper function to process investigation fields configuration for all rule types
func (d SecurityDetectionRuleData) investigationFieldsToApi(ctx context.Context) (*kbapi.SecurityDetectionsAPIInvestigationFields, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !utils.IsKnown(d.InvestigationFields) || len(d.InvestigationFields.Elements()) == 0 {
		return nil, diags
	}

	fieldNames := make([]string, len(d.InvestigationFields.Elements()))
	fieldDiag := d.InvestigationFields.ElementsAs(ctx, &fieldNames, false)
	if fieldDiag.HasError() {
		diags.Append(fieldDiag...)
		return nil, diags
	}

	// Convert to API type
	apiFieldNames := make([]kbapi.SecurityDetectionsAPINonEmptyString, len(fieldNames))
	for i, field := range fieldNames {
		apiFieldNames[i] = kbapi.SecurityDetectionsAPINonEmptyString(field)
	}

	return &kbapi.SecurityDetectionsAPIInvestigationFields{
		FieldNames: apiFieldNames,
	}, diags
}

// convertInvestigationFieldsToModel converts kbapi.SecurityDetectionsAPIInvestigationFields to Terraform model
func convertInvestigationFieldsToModel(ctx context.Context, apiInvestigationFields *kbapi.SecurityDetectionsAPIInvestigationFields) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if apiInvestigationFields == nil || len(apiInvestigationFields.FieldNames) == 0 {
		return types.ListNull(types.StringType), diags
	}

	fieldNames := make([]string, len(apiInvestigationFields.FieldNames))
	for i, field := range apiInvestigationFields.FieldNames {
		fieldNames[i] = string(field)
	}

	return utils.SliceToListType_String(ctx, fieldNames, path.Root("investigation_fields"), &diags), diags
}

// Helper function to update investigation fields from API response
func (d *SecurityDetectionRuleData) updateInvestigationFieldsFromApi(ctx context.Context, investigationFields *kbapi.SecurityDetectionsAPIInvestigationFields) diag.Diagnostics {
	var diags diag.Diagnostics

	investigationFieldsValue, investigationFieldsDiags := convertInvestigationFieldsToModel(ctx, investigationFields)
	diags.Append(investigationFieldsDiags...)
	if !investigationFieldsDiags.HasError() {
		d.InvestigationFields = investigationFieldsValue
	} else {
		d.InvestigationFields = types.ListNull(types.StringType)
	}

	return diags
}

// Helper function to process related integrations configuration for all rule types
func (d SecurityDetectionRuleData) relatedIntegrationsToApi(ctx context.Context) (*kbapi.SecurityDetectionsAPIRelatedIntegrationArray, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !utils.IsKnown(d.RelatedIntegrations) || len(d.RelatedIntegrations.Elements()) == 0 {
		return nil, diags
	}

	apiRelatedIntegrations := utils.ListTypeToSlice(ctx, d.RelatedIntegrations, path.Root("related_integrations"), &diags,
		func(integration RelatedIntegrationModel, meta utils.ListMeta) kbapi.SecurityDetectionsAPIRelatedIntegration {
			if integration.Package.IsNull() || integration.Version.IsNull() {
				meta.Diags.AddError("Missing required fields", "Package and version are required for related integrations")
				return kbapi.SecurityDetectionsAPIRelatedIntegration{}
			}

			apiIntegration := kbapi.SecurityDetectionsAPIRelatedIntegration{
				Package: kbapi.SecurityDetectionsAPINonEmptyString(integration.Package.ValueString()),
				Version: kbapi.SecurityDetectionsAPINonEmptyString(integration.Version.ValueString()),
			}

			// Set optional integration field if provided
			if utils.IsKnown(integration.Integration) {
				integrationName := kbapi.SecurityDetectionsAPINonEmptyString(integration.Integration.ValueString())
				apiIntegration.Integration = &integrationName
			}

			return apiIntegration
		})

	return &apiRelatedIntegrations, diags
}

// convertRelatedIntegrationsToModel converts kbapi.SecurityDetectionsAPIRelatedIntegrationArray to Terraform model
func convertRelatedIntegrationsToModel(ctx context.Context, apiRelatedIntegrations *kbapi.SecurityDetectionsAPIRelatedIntegrationArray) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if apiRelatedIntegrations == nil || len(*apiRelatedIntegrations) == 0 {
		return types.ListNull(getRelatedIntegrationElementType()), diags
	}

	integrations := make([]RelatedIntegrationModel, 0)

	for _, apiIntegration := range *apiRelatedIntegrations {
		integration := RelatedIntegrationModel{
			Package: types.StringValue(string(apiIntegration.Package)),
			Version: types.StringValue(string(apiIntegration.Version)),
		}

		// Set optional integration field if provided
		if apiIntegration.Integration != nil {
			integration.Integration = types.StringValue(string(*apiIntegration.Integration))
		} else {
			integration.Integration = types.StringNull()
		}

		integrations = append(integrations, integration)
	}

	listValue, listDiags := types.ListValueFrom(ctx, getRelatedIntegrationElementType(), integrations)
	diags.Append(listDiags...)
	return listValue, diags
}

// Helper function to update related integrations from API response
func (d *SecurityDetectionRuleData) updateRelatedIntegrationsFromApi(ctx context.Context, relatedIntegrations *kbapi.SecurityDetectionsAPIRelatedIntegrationArray) diag.Diagnostics {
	var diags diag.Diagnostics

	if relatedIntegrations != nil && len(*relatedIntegrations) > 0 {
		relatedIntegrationsValue, relatedIntegrationsDiags := convertRelatedIntegrationsToModel(ctx, relatedIntegrations)
		diags.Append(relatedIntegrationsDiags...)
		if !relatedIntegrationsDiags.HasError() {
			d.RelatedIntegrations = relatedIntegrationsValue
		}
	} else {
		d.RelatedIntegrations = types.ListNull(getRelatedIntegrationElementType())
	}

	return diags
}

// Helper function to process required fields configuration for all rule types
func (d SecurityDetectionRuleData) requiredFieldsToApi(ctx context.Context) (*[]kbapi.SecurityDetectionsAPIRequiredFieldInput, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !utils.IsKnown(d.RequiredFields) || len(d.RequiredFields.Elements()) == 0 {
		return nil, diags
	}

	apiRequiredFields := utils.ListTypeToSlice(ctx, d.RequiredFields, path.Root("required_fields"), &diags,
		func(field RequiredFieldModel, meta utils.ListMeta) kbapi.SecurityDetectionsAPIRequiredFieldInput {
			if field.Name.IsNull() || field.Type.IsNull() {
				meta.Diags.AddError("Missing required fields", "Name and type are required for required fields")
				return kbapi.SecurityDetectionsAPIRequiredFieldInput{}
			}

			return kbapi.SecurityDetectionsAPIRequiredFieldInput{
				Name: field.Name.ValueString(),
				Type: field.Type.ValueString(),
			}
		})

	return &apiRequiredFields, diags
}

// convertRequiredFieldsToModel converts kbapi.SecurityDetectionsAPIRequiredFieldArray to Terraform model
func convertRequiredFieldsToModel(ctx context.Context, apiRequiredFields *kbapi.SecurityDetectionsAPIRequiredFieldArray) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if apiRequiredFields == nil || len(*apiRequiredFields) == 0 {
		return types.ListNull(getRequiredFieldElementType()), diags
	}

	fields := make([]RequiredFieldModel, 0)

	for _, apiField := range *apiRequiredFields {
		field := RequiredFieldModel{
			Name: types.StringValue(apiField.Name),
			Type: types.StringValue(apiField.Type),
			Ecs:  types.BoolValue(apiField.Ecs),
		}

		fields = append(fields, field)
	}

	listValue, listDiags := types.ListValueFrom(ctx, getRequiredFieldElementType(), fields)
	diags.Append(listDiags...)
	return listValue, diags
}

// Helper function to update required fields from API response
func (d *SecurityDetectionRuleData) updateRequiredFieldsFromApi(ctx context.Context, requiredFields *kbapi.SecurityDetectionsAPIRequiredFieldArray) diag.Diagnostics {
	var diags diag.Diagnostics

	if requiredFields != nil && len(*requiredFields) > 0 {
		requiredFieldsValue, requiredFieldsDiags := convertRequiredFieldsToModel(ctx, requiredFields)
		diags.Append(requiredFieldsDiags...)
		if !requiredFieldsDiags.HasError() {
			d.RequiredFields = requiredFieldsValue
		}
	} else {
		d.RequiredFields = types.ListNull(getRequiredFieldElementType())
	}

	return diags
}

// Helper function to process severity mapping configuration for all rule types
func (d SecurityDetectionRuleData) severityMappingToApi(ctx context.Context) (*kbapi.SecurityDetectionsAPISeverityMapping, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !utils.IsKnown(d.SeverityMapping) || len(d.SeverityMapping.Elements()) == 0 {
		return nil, diags
	}

	apiSeverityMapping := utils.ListTypeToSlice(ctx, d.SeverityMapping, path.Root("severity_mapping"), &diags,
		func(mapping SeverityMappingModel, meta utils.ListMeta) struct {
			Field    string                                             `json:"field"`
			Operator kbapi.SecurityDetectionsAPISeverityMappingOperator `json:"operator"`
			Severity kbapi.SecurityDetectionsAPISeverity                `json:"severity"`
			Value    string                                             `json:"value"`
		} {
			if mapping.Field.IsNull() || mapping.Operator.IsNull() || mapping.Value.IsNull() || mapping.Severity.IsNull() {
				meta.Diags.AddError("Missing required fields", "Field, operator, value, and severity are required for severity mapping")
				return struct {
					Field    string                                             `json:"field"`
					Operator kbapi.SecurityDetectionsAPISeverityMappingOperator `json:"operator"`
					Severity kbapi.SecurityDetectionsAPISeverity                `json:"severity"`
					Value    string                                             `json:"value"`
				}{}
			}

			return struct {
				Field    string                                             `json:"field"`
				Operator kbapi.SecurityDetectionsAPISeverityMappingOperator `json:"operator"`
				Severity kbapi.SecurityDetectionsAPISeverity                `json:"severity"`
				Value    string                                             `json:"value"`
			}{
				Field:    mapping.Field.ValueString(),
				Operator: kbapi.SecurityDetectionsAPISeverityMappingOperator(mapping.Operator.ValueString()),
				Severity: kbapi.SecurityDetectionsAPISeverity(mapping.Severity.ValueString()),
				Value:    mapping.Value.ValueString(),
			}
		})

	// Convert to the expected slice type
	severityMappingSlice := make(kbapi.SecurityDetectionsAPISeverityMapping, len(apiSeverityMapping))
	copy(severityMappingSlice, apiSeverityMapping)

	return &severityMappingSlice, diags
}

// metaToApi converts the Terraform meta field to the API type
func (d SecurityDetectionRuleData) metaToApi(ctx context.Context) (*kbapi.SecurityDetectionsAPIRuleMetadata, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !utils.IsKnown(d.Meta) {
		return nil, diags
	}

	// Unmarshal the JSON string to map[string]interface{}
	var metadata kbapi.SecurityDetectionsAPIRuleMetadata
	unmarshalDiags := d.Meta.Unmarshal(&metadata)
	diags.Append(unmarshalDiags...)

	if diags.HasError() {
		return nil, diags
	}

	return &metadata, diags
}

// filtersToApi converts the Terraform filters field to the API type
func (d SecurityDetectionRuleData) filtersToApi(ctx context.Context) (*kbapi.SecurityDetectionsAPIRuleFilterArray, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !utils.IsKnown(d.Filters) {
		return nil, diags
	}

	// Unmarshal the JSON string to []interface{}
	var filters kbapi.SecurityDetectionsAPIRuleFilterArray
	unmarshalDiags := d.Filters.Unmarshal(&filters)
	diags.Append(unmarshalDiags...)

	if diags.HasError() {
		return nil, diags
	}

	return &filters, diags
}

// convertMetaFromApi converts the API meta field back to the Terraform type
func (d *SecurityDetectionRuleData) updateMetaFromApi(ctx context.Context, apiMeta *kbapi.SecurityDetectionsAPIRuleMetadata) diag.Diagnostics {
	var diags diag.Diagnostics

	if apiMeta == nil || len(*apiMeta) == 0 {
		d.Meta = jsontypes.NewNormalizedNull()
		return diags
	}

	// Marshal the map[string]interface{} to JSON string
	jsonBytes, err := json.Marshal(*apiMeta)
	if err != nil {
		diags.AddError("Failed to marshal metadata", err.Error())
		return diags
	}

	// Create a NormalizedValue from the JSON string
	d.Meta = jsontypes.NewNormalizedValue(string(jsonBytes))
	return diags
}

// convertFiltersFromApi converts the API filters field back to the Terraform type
func (d *SecurityDetectionRuleData) updateFiltersFromApi(ctx context.Context, apiFilters *kbapi.SecurityDetectionsAPIRuleFilterArray) diag.Diagnostics {
	var diags diag.Diagnostics

	if apiFilters == nil || len(*apiFilters) == 0 {
		d.Filters = jsontypes.NewNormalizedNull()
		return diags
	}

	// Marshal the []interface{} to JSON string
	jsonBytes, err := json.Marshal(*apiFilters)
	if err != nil {
		diags.AddError("Failed to marshal filters", err.Error())
		return diags
	}

	// Create a NormalizedValue from the JSON string
	d.Filters = jsontypes.NewNormalizedValue(string(jsonBytes))
	return diags
} // convertSeverityMappingToModel converts kbapi.SecurityDetectionsAPISeverityMapping to Terraform model
func convertSeverityMappingToModel(ctx context.Context, apiSeverityMapping *kbapi.SecurityDetectionsAPISeverityMapping) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if apiSeverityMapping == nil || len(*apiSeverityMapping) == 0 {
		return types.ListNull(getSeverityMappingElementType()), diags
	}

	mappings := make([]SeverityMappingModel, 0)

	for _, apiMapping := range *apiSeverityMapping {
		mapping := SeverityMappingModel{
			Field:    types.StringValue(apiMapping.Field),
			Operator: types.StringValue(string(apiMapping.Operator)),
			Value:    types.StringValue(apiMapping.Value),
			Severity: types.StringValue(string(apiMapping.Severity)),
		}

		mappings = append(mappings, mapping)
	}

	listValue, listDiags := types.ListValueFrom(ctx, getSeverityMappingElementType(), mappings)
	diags.Append(listDiags...)
	return listValue, diags
}

// Helper function to update severity mapping from API response
func (d *SecurityDetectionRuleData) updateSeverityMappingFromApi(ctx context.Context, severityMapping *kbapi.SecurityDetectionsAPISeverityMapping) diag.Diagnostics {
	var diags diag.Diagnostics

	if severityMapping != nil && len(*severityMapping) > 0 {
		severityMappingValue, severityMappingDiags := convertSeverityMappingToModel(ctx, severityMapping)
		diags.Append(severityMappingDiags...)
		if !severityMappingDiags.HasError() {
			d.SeverityMapping = severityMappingValue
		}
	} else {
		d.SeverityMapping = types.ListNull(getSeverityMappingElementType())
	}

	return diags
}

// Helper function to update index patterns from API response
func (d *SecurityDetectionRuleData) updateIndexFromApi(ctx context.Context, index *[]string) diag.Diagnostics {
	var diags diag.Diagnostics

	if index != nil && len(*index) > 0 {
		d.Index = utils.ListValueFrom(ctx, *index, types.StringType, path.Root("index"), &diags)
	} else {
		d.Index = types.ListValueMust(types.StringType, []attr.Value{})
	}

	return diags
}

// Helper function to update author from API response
func (d *SecurityDetectionRuleData) updateAuthorFromApi(ctx context.Context, author []string) diag.Diagnostics {
	var diags diag.Diagnostics

	if len(author) > 0 {
		d.Author = utils.ListValueFrom(ctx, author, types.StringType, path.Root("author"), &diags)
	} else {
		d.Author = types.ListValueMust(types.StringType, []attr.Value{})
	}

	return diags
}

// Helper function to update tags from API response
func (d *SecurityDetectionRuleData) updateTagsFromApi(ctx context.Context, tags []string) diag.Diagnostics {
	var diags diag.Diagnostics

	if len(tags) > 0 {
		d.Tags = utils.ListValueFrom(ctx, tags, types.StringType, path.Root("tags"), &diags)
	} else {
		d.Tags = types.ListValueMust(types.StringType, []attr.Value{})
	}

	return diags
}

// Helper function to update false positives from API response
func (d *SecurityDetectionRuleData) updateFalsePositivesFromApi(ctx context.Context, falsePositives []string) diag.Diagnostics {
	var diags diag.Diagnostics

	if len(falsePositives) > 0 {
		d.FalsePositives = utils.ListValueFrom(ctx, falsePositives, types.StringType, path.Root("false_positives"), &diags)
	} else {
		d.FalsePositives = types.ListValueMust(types.StringType, []attr.Value{})
	}

	return diags
}

// Helper function to update references from API response
func (d *SecurityDetectionRuleData) updateReferencesFromApi(ctx context.Context, references []string) diag.Diagnostics {
	var diags diag.Diagnostics

	if len(references) > 0 {
		d.References = utils.ListValueFrom(ctx, references, types.StringType, path.Root("references"), &diags)
	} else {
		d.References = types.ListValueMust(types.StringType, []attr.Value{})
	}

	return diags
}

// Helper function to update data view ID from API response
func (d *SecurityDetectionRuleData) updateDataViewIdFromApi(ctx context.Context, dataViewId *kbapi.SecurityDetectionsAPIDataViewId) diag.Diagnostics {
	var diags diag.Diagnostics

	if dataViewId != nil {
		d.DataViewId = types.StringValue(string(*dataViewId))
	} else {
		d.DataViewId = types.StringNull()
	}

	return diags
}

// Helper function to update namespace from API response
func (d *SecurityDetectionRuleData) updateNamespaceFromApi(ctx context.Context, namespace *kbapi.SecurityDetectionsAPIAlertsIndexNamespace) diag.Diagnostics {
	var diags diag.Diagnostics

	if namespace != nil {
		d.Namespace = types.StringValue(string(*namespace))
	} else {
		d.Namespace = types.StringNull()
	}

	return diags
}

// Helper function to update rule name override from API response
func (d *SecurityDetectionRuleData) updateRuleNameOverrideFromApi(ctx context.Context, ruleNameOverride *kbapi.SecurityDetectionsAPIRuleNameOverride) diag.Diagnostics {
	var diags diag.Diagnostics

	if ruleNameOverride != nil {
		d.RuleNameOverride = types.StringValue(string(*ruleNameOverride))
	} else {
		d.RuleNameOverride = types.StringNull()
	}

	return diags
}

// Helper function to update timestamp override from API response
func (d *SecurityDetectionRuleData) updateTimestampOverrideFromApi(ctx context.Context, timestampOverride *kbapi.SecurityDetectionsAPITimestampOverride) diag.Diagnostics {
	var diags diag.Diagnostics

	if timestampOverride != nil {
		d.TimestampOverride = types.StringValue(string(*timestampOverride))
	} else {
		d.TimestampOverride = types.StringNull()
	}

	return diags
}

// Helper function to update timestamp override fallback disabled from API response
func (d *SecurityDetectionRuleData) updateTimestampOverrideFallbackDisabledFromApi(ctx context.Context, timestampOverrideFallbackDisabled *kbapi.SecurityDetectionsAPITimestampOverrideFallbackDisabled) diag.Diagnostics {
	var diags diag.Diagnostics

	if timestampOverrideFallbackDisabled != nil {
		d.TimestampOverrideFallbackDisabled = types.BoolValue(bool(*timestampOverrideFallbackDisabled))
	} else {
		d.TimestampOverrideFallbackDisabled = types.BoolNull()
	}

	return diags
}

// Helper function to update building block type from API response
func (d *SecurityDetectionRuleData) updateBuildingBlockTypeFromApi(ctx context.Context, buildingBlockType *kbapi.SecurityDetectionsAPIBuildingBlockType) diag.Diagnostics {
	var diags diag.Diagnostics

	if buildingBlockType != nil {
		d.BuildingBlockType = types.StringValue(string(*buildingBlockType))
	} else {
		d.BuildingBlockType = types.StringNull()
	}

	return diags
}

// Helper function to update license from API response
func (d *SecurityDetectionRuleData) updateLicenseFromApi(ctx context.Context, license *kbapi.SecurityDetectionsAPIRuleLicense) diag.Diagnostics {
	var diags diag.Diagnostics

	if license != nil {
		d.License = types.StringValue(string(*license))
	} else {
		d.License = types.StringNull()
	}

	return diags
}

// Helper function to update note from API response
func (d *SecurityDetectionRuleData) updateNoteFromApi(ctx context.Context, note *kbapi.SecurityDetectionsAPIInvestigationGuide) diag.Diagnostics {
	var diags diag.Diagnostics

	if note != nil {
		d.Note = types.StringValue(string(*note))
	} else {
		d.Note = types.StringNull()
	}

	return diags
}

// Helper function to update setup from API response
func (d *SecurityDetectionRuleData) updateSetupFromApi(ctx context.Context, setup kbapi.SecurityDetectionsAPISetupGuide) diag.Diagnostics {
	var diags diag.Diagnostics

	// Handle setup field - if empty, set to null to maintain consistency with optional schema
	if string(setup) != "" {
		d.Setup = types.StringValue(string(setup))
	} else {
		d.Setup = types.StringNull()
	}

	return diags
}
