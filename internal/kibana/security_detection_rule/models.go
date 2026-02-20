package securitydetectionrule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// MinVersionResponseActions defines the minimum server version required for response actions
var MinVersionResponseActions = version.Must(version.NewVersion("8.16.0"))

type Data struct {
	ID       types.String `tfsdk:"id"`
	SpaceID  types.String `tfsdk:"space_id"`
	RuleID   types.String `tfsdk:"rule_id"`
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
	MachineLearningJobID types.List  `tfsdk:"machine_learning_job_id"`

	// New Terms-specific fields
	NewTermsFields     types.List   `tfsdk:"new_terms_fields"`
	HistoryWindowStart types.String `tfsdk:"history_window_start"`

	// Saved Query-specific fields
	SavedID types.String `tfsdk:"saved_id"`

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
	TimelineID    types.String `tfsdk:"timeline_id"`
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
	DataViewID types.String `tfsdk:"data_view_id"`

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
type TfData struct {
	ThreatMapping types.List `tfsdk:"threat_mapping"`
}

type TfDataItem struct {
	Entries types.List `tfsdk:"entries"`
}

type TfDataItemEntry struct {
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
	ActionTypeID types.String `tfsdk:"action_type_id"`
	ID           types.String `tfsdk:"id"`
	Params       types.Map    `tfsdk:"params"`
	Group        types.String `tfsdk:"group"`
	UUID         types.String `tfsdk:"uuid"`
	AlertsFilter types.Map    `tfsdk:"alerts_filter"`
	Frequency    types.Object `tfsdk:"frequency"`
}

type ActionFrequencyModel struct {
	NotifyWhen types.String `tfsdk:"notify_when"`
	Summary    types.Bool   `tfsdk:"summary"`
	Throttle   types.String `tfsdk:"throttle"`
}

type ResponseActionModel struct {
	ActionTypeID types.String `tfsdk:"action_type_id"`
	Params       types.Object `tfsdk:"params"`
}

type ResponseActionParamsModel struct {
	// Osquery params
	Query        types.String `tfsdk:"query"`
	PackID       types.String `tfsdk:"pack_id"`
	SavedQueryID types.String `tfsdk:"saved_query_id"`
	Timeout      types.Int64  `tfsdk:"timeout"`
	EcsMapping   types.Map    `tfsdk:"ecs_mapping"`
	Queries      types.List   `tfsdk:"queries"`

	// Endpoint params
	Command types.String `tfsdk:"command"`
	Comment types.String `tfsdk:"comment"`
	Config  types.Object `tfsdk:"config"`
}

type OsqueryQueryModel struct {
	ID         types.String `tfsdk:"id"`
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
	ID            types.String `tfsdk:"id"`
	ListID        types.String `tfsdk:"list_id"`
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
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Reference types.String `tfsdk:"reference"`
}

type ThreatTechniqueModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Reference    types.String `tfsdk:"reference"`
	Subtechnique types.List   `tfsdk:"subtechnique"`
}

type ThreatSubtechniqueModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Reference types.String `tfsdk:"reference"`
}

// CommonCreateProps holds all the field pointers for setting common create properties
type CommonCreateProps struct {
	Actions                           **[]kbapi.SecurityDetectionsAPIRuleAction
	ResponseActions                   **[]kbapi.SecurityDetectionsAPIResponseAction
	RuleID                            **kbapi.SecurityDetectionsAPIRuleSignatureId
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
	DataViewID                        **kbapi.SecurityDetectionsAPIDataViewId
	Namespace                         **kbapi.SecurityDetectionsAPIAlertsIndexNamespace
	RuleNameOverride                  **kbapi.SecurityDetectionsAPIRuleNameOverride
	TimestampOverride                 **kbapi.SecurityDetectionsAPITimestampOverride
	TimestampOverrideFallbackDisabled **kbapi.SecurityDetectionsAPITimestampOverrideFallbackDisabled
	InvestigationFields               **kbapi.SecurityDetectionsAPIInvestigationFields
	Filters                           **kbapi.SecurityDetectionsAPIRuleFilterArray
	Threat                            **kbapi.SecurityDetectionsAPIThreatArray
	TimelineID                        **kbapi.SecurityDetectionsAPITimelineTemplateId
	TimelineTitle                     **kbapi.SecurityDetectionsAPITimelineTemplateTitle
}

// CommonUpdateProps holds all the field pointers for setting common update properties
type CommonUpdateProps struct {
	Actions                           **[]kbapi.SecurityDetectionsAPIRuleAction
	ResponseActions                   **[]kbapi.SecurityDetectionsAPIResponseAction
	RuleID                            **kbapi.SecurityDetectionsAPIRuleSignatureId
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
	DataViewID                        **kbapi.SecurityDetectionsAPIDataViewId
	Namespace                         **kbapi.SecurityDetectionsAPIAlertsIndexNamespace
	RuleNameOverride                  **kbapi.SecurityDetectionsAPIRuleNameOverride
	TimestampOverride                 **kbapi.SecurityDetectionsAPITimestampOverride
	TimestampOverrideFallbackDisabled **kbapi.SecurityDetectionsAPITimestampOverrideFallbackDisabled
	InvestigationFields               **kbapi.SecurityDetectionsAPIInvestigationFields
	Filters                           **kbapi.SecurityDetectionsAPIRuleFilterArray
	Threat                            **kbapi.SecurityDetectionsAPIThreatArray
	TimelineID                        **kbapi.SecurityDetectionsAPITimelineTemplateId
	TimelineTitle                     **kbapi.SecurityDetectionsAPITimelineTemplateTitle
}

// Helper function to set common properties across all rule types
func (d Data) setCommonCreateProps(
	ctx context.Context,
	props *CommonCreateProps,
	diags *diag.Diagnostics,
	client clients.MinVersionEnforceable,
) {
	// Set optional rule_id if provided
	if props.RuleID != nil && typeutils.IsKnown(d.RuleID) {
		id := d.RuleID.ValueString()
		*props.RuleID = &id
	}

	// Set enabled status
	if props.Enabled != nil && typeutils.IsKnown(d.Enabled) {
		isEnabled := d.Enabled.ValueBool()
		*props.Enabled = &isEnabled
	}

	// Set time range
	if props.From != nil && typeutils.IsKnown(d.From) {
		fromTime := d.From.ValueString()
		*props.From = &fromTime
	}

	if props.To != nil && typeutils.IsKnown(d.To) {
		toTime := d.To.ValueString()
		*props.To = &toTime
	}

	// Set interval
	if props.Interval != nil && typeutils.IsKnown(d.Interval) {
		intervalTime := d.Interval.ValueString()
		*props.Interval = &intervalTime
	}

	// Set index patterns (if index pointer is provided)
	if props.Index != nil && typeutils.IsKnown(d.Index) {
		indexList := typeutils.ListTypeAs[string](ctx, d.Index, path.Root("index"), diags)
		if !diags.HasError() && len(indexList) > 0 {
			*props.Index = &indexList
		}
	}

	// Set author
	if props.Author != nil && typeutils.IsKnown(d.Author) {
		authorList := typeutils.ListTypeAs[string](ctx, d.Author, path.Root("author"), diags)
		if !diags.HasError() && len(authorList) > 0 {
			*props.Author = &authorList
		}
	}

	// Set tags
	if props.Tags != nil && typeutils.IsKnown(d.Tags) {
		tagsList := typeutils.ListTypeAs[string](ctx, d.Tags, path.Root("tags"), diags)
		if !diags.HasError() && len(tagsList) > 0 {
			*props.Tags = &tagsList
		}
	}

	// Set false positives
	if props.FalsePositives != nil && typeutils.IsKnown(d.FalsePositives) {
		fpList := typeutils.ListTypeAs[string](ctx, d.FalsePositives, path.Root("false_positives"), diags)
		if !diags.HasError() && len(fpList) > 0 {
			*props.FalsePositives = &fpList
		}
	}

	// Set references
	if props.References != nil && typeutils.IsKnown(d.References) {
		refList := typeutils.ListTypeAs[string](ctx, d.References, path.Root("references"), diags)
		if !diags.HasError() && len(refList) > 0 {
			*props.References = &refList
		}
	}

	// Set optional string fields
	if props.License != nil && typeutils.IsKnown(d.License) {
		ruleLicense := d.License.ValueString()
		*props.License = &ruleLicense
	}

	if props.Note != nil && typeutils.IsKnown(d.Note) {
		ruleNote := d.Note.ValueString()
		*props.Note = &ruleNote
	}

	if props.Setup != nil && typeutils.IsKnown(d.Setup) {
		ruleSetup := d.Setup.ValueString()
		*props.Setup = &ruleSetup
	}

	// Set max signals
	if props.MaxSignals != nil && typeutils.IsKnown(d.MaxSignals) {
		maxSig := kbapi.SecurityDetectionsAPIMaxSignals(d.MaxSignals.ValueInt64())
		*props.MaxSignals = &maxSig
	}

	// Set version
	if props.Version != nil && typeutils.IsKnown(d.Version) {
		ruleVersion := kbapi.SecurityDetectionsAPIRuleVersion(d.Version.ValueInt64())
		*props.Version = &ruleVersion
	}

	// Set actions
	if props.Actions != nil && typeutils.IsKnown(d.Actions) {
		actions, actionDiags := d.actionsToAPI(ctx)
		diags.Append(actionDiags...)
		if !actionDiags.HasError() && len(actions) > 0 {
			*props.Actions = &actions
		}
	}

	// Set exceptions list
	if props.ExceptionsList != nil && typeutils.IsKnown(d.ExceptionsList) {
		exceptionsList, exceptionsListDiags := d.exceptionsListToAPI(ctx)
		diags.Append(exceptionsListDiags...)
		if !exceptionsListDiags.HasError() && len(exceptionsList) > 0 {
			*props.ExceptionsList = &exceptionsList
		}
	}

	// Set risk score mapping
	if props.RiskScoreMapping != nil && typeutils.IsKnown(d.RiskScoreMapping) {
		riskScoreMapping, riskScoreMappingDiags := d.riskScoreMappingToAPI(ctx)
		diags.Append(riskScoreMappingDiags...)
		if !riskScoreMappingDiags.HasError() && len(riskScoreMapping) > 0 {
			*props.RiskScoreMapping = &riskScoreMapping
		}
	}

	// Set building block type
	if props.BuildingBlockType != nil && typeutils.IsKnown(d.BuildingBlockType) {
		buildingBlockType := d.BuildingBlockType.ValueString()
		*props.BuildingBlockType = &buildingBlockType
	}

	// Set data view ID
	if props.DataViewID != nil && typeutils.IsKnown(d.DataViewID) {
		dataViewID := d.DataViewID.ValueString()
		*props.DataViewID = &dataViewID
	}

	// Set namespace
	if props.Namespace != nil && typeutils.IsKnown(d.Namespace) {
		namespace := d.Namespace.ValueString()
		*props.Namespace = &namespace
	}

	// Set rule name override
	if props.RuleNameOverride != nil && typeutils.IsKnown(d.RuleNameOverride) {
		ruleNameOverride := d.RuleNameOverride.ValueString()
		*props.RuleNameOverride = &ruleNameOverride
	}

	// Set timestamp override
	if props.TimestampOverride != nil && typeutils.IsKnown(d.TimestampOverride) {
		timestampOverride := d.TimestampOverride.ValueString()
		*props.TimestampOverride = &timestampOverride
	}

	// Set timestamp override fallback disabled
	if props.TimestampOverrideFallbackDisabled != nil && typeutils.IsKnown(d.TimestampOverrideFallbackDisabled) {
		timestampOverrideFallbackDisabled := d.TimestampOverrideFallbackDisabled.ValueBool()
		*props.TimestampOverrideFallbackDisabled = &timestampOverrideFallbackDisabled
	}

	// Set severity mapping
	if props.SeverityMapping != nil && typeutils.IsKnown(d.SeverityMapping) {
		severityMapping, severityMappingDiags := d.severityMappingToAPI(ctx)
		diags.Append(severityMappingDiags...)
		if !severityMappingDiags.HasError() && severityMapping != nil && len(*severityMapping) > 0 {
			*props.SeverityMapping = severityMapping
		}
	}

	// Set related integrations
	if props.RelatedIntegrations != nil && typeutils.IsKnown(d.RelatedIntegrations) {
		relatedIntegrations, relatedIntegrationsDiags := d.relatedIntegrationsToAPI(ctx)
		diags.Append(relatedIntegrationsDiags...)
		if !relatedIntegrationsDiags.HasError() && relatedIntegrations != nil && len(*relatedIntegrations) > 0 {
			*props.RelatedIntegrations = relatedIntegrations
		}
	}

	// Set required fields
	if props.RequiredFields != nil && typeutils.IsKnown(d.RequiredFields) {
		requiredFields, requiredFieldsDiags := d.requiredFieldsToAPI(ctx)
		diags.Append(requiredFieldsDiags...)
		if !requiredFieldsDiags.HasError() && requiredFields != nil && len(*requiredFields) > 0 {
			*props.RequiredFields = requiredFields
		}
	}

	// Set investigation fields
	if props.InvestigationFields != nil {
		investigationFields, investigationFieldsDiags := d.investigationFieldsToAPI(ctx)
		if !investigationFieldsDiags.HasError() && investigationFields != nil {
			*props.InvestigationFields = investigationFields
		}
		diags.Append(investigationFieldsDiags...)
	}

	// Set response actions
	if props.ResponseActions != nil && typeutils.IsKnown(d.ResponseActions) {
		responseActions, responseActionsDiags := d.responseActionsToAPI(ctx, client)
		diags.Append(responseActionsDiags...)
		if !responseActionsDiags.HasError() && len(responseActions) > 0 {
			*props.ResponseActions = &responseActions
		}
	}

	// Set filters
	if props.Filters != nil && typeutils.IsKnown(d.Filters) {
		filters, filtersDiags := d.filtersToAPI(ctx)
		diags.Append(filtersDiags...)
		if !filtersDiags.HasError() && filters != nil {
			*props.Filters = filters
		}
	}

	// Set alert suppression
	if props.AlertSuppression != nil {
		alertSuppression := d.alertSuppressionToAPI(ctx, diags)
		if alertSuppression != nil {
			*props.AlertSuppression = alertSuppression
		}
	}

	// Set threat (MITRE ATT&CK framework)
	if props.Threat != nil && typeutils.IsKnown(d.Threat) {
		threat, threatDiags := d.threatToAPI(ctx)
		diags.Append(threatDiags...)
		if !threatDiags.HasError() && len(threat) > 0 {
			*props.Threat = &threat
		}
	}

	// Set timeline ID
	if props.TimelineID != nil && typeutils.IsKnown(d.TimelineID) {
		timelineID := d.TimelineID.ValueString()
		*props.TimelineID = &timelineID
	}

	// Set timeline title
	if props.TimelineTitle != nil && typeutils.IsKnown(d.TimelineTitle) {
		timelineTitle := d.TimelineTitle.ValueString()
		*props.TimelineTitle = &timelineTitle
	}
}

// Helper function to set common update properties across all rule types
func (d Data) setCommonUpdateProps(
	ctx context.Context,
	props *CommonUpdateProps,
	diags *diag.Diagnostics,
	client clients.MinVersionEnforceable,
) {
	// Set enabled status
	if props.Enabled != nil && typeutils.IsKnown(d.Enabled) {
		isEnabled := d.Enabled.ValueBool()
		*props.Enabled = &isEnabled
	}

	// Set time range
	if props.From != nil && typeutils.IsKnown(d.From) {
		fromTime := d.From.ValueString()
		*props.From = &fromTime
	}

	if props.To != nil && typeutils.IsKnown(d.To) {
		toTime := d.To.ValueString()
		*props.To = &toTime
	}

	// Set interval
	if props.Interval != nil && typeutils.IsKnown(d.Interval) {
		intervalTime := d.Interval.ValueString()
		*props.Interval = &intervalTime
	}

	// Set index patterns (if index pointer is provided)
	if props.Index != nil && typeutils.IsKnown(d.Index) {
		indexList := typeutils.ListTypeAs[string](ctx, d.Index, path.Root("index"), diags)
		if !diags.HasError() {
			*props.Index = &indexList
		}
	}

	// Set author
	if props.Author != nil && typeutils.IsKnown(d.Author) {
		authorList := typeutils.ListTypeAs[string](ctx, d.Author, path.Root("author"), diags)
		if !diags.HasError() {
			*props.Author = &authorList
		}
	}

	// Set tags
	if props.Tags != nil && typeutils.IsKnown(d.Tags) {
		tagsList := typeutils.ListTypeAs[string](ctx, d.Tags, path.Root("tags"), diags)
		if !diags.HasError() {
			*props.Tags = &tagsList
		}
	}

	// Set false positives
	if props.FalsePositives != nil && typeutils.IsKnown(d.FalsePositives) {
		fpList := typeutils.ListTypeAs[string](ctx, d.FalsePositives, path.Root("false_positives"), diags)
		if !diags.HasError() {
			*props.FalsePositives = &fpList
		}
	}

	// Set references
	if props.References != nil && typeutils.IsKnown(d.References) {
		refList := typeutils.ListTypeAs[string](ctx, d.References, path.Root("references"), diags)
		if !diags.HasError() {
			*props.References = &refList
		}
	}

	// Set optional string fields
	if props.License != nil && typeutils.IsKnown(d.License) {
		ruleLicense := d.License.ValueString()
		*props.License = &ruleLicense
	}

	if props.Note != nil && typeutils.IsKnown(d.Note) {
		ruleNote := d.Note.ValueString()
		*props.Note = &ruleNote
	}

	if props.Setup != nil && typeutils.IsKnown(d.Setup) {
		ruleSetup := d.Setup.ValueString()
		*props.Setup = &ruleSetup
	}

	// Set max signals
	if props.MaxSignals != nil && typeutils.IsKnown(d.MaxSignals) {
		maxSig := kbapi.SecurityDetectionsAPIMaxSignals(d.MaxSignals.ValueInt64())
		*props.MaxSignals = &maxSig
	}

	// Set version
	if props.Version != nil && typeutils.IsKnown(d.Version) {
		ruleVersion := kbapi.SecurityDetectionsAPIRuleVersion(d.Version.ValueInt64())
		*props.Version = &ruleVersion
	}

	// Set actions
	if props.Actions != nil && typeutils.IsKnown(d.Actions) {
		actions, actionDiags := d.actionsToAPI(ctx)
		diags.Append(actionDiags...)
		if !actionDiags.HasError() && len(actions) > 0 {
			*props.Actions = &actions
		}
	}

	// Set exceptions list
	if props.ExceptionsList != nil && typeutils.IsKnown(d.ExceptionsList) {
		exceptionsList, exceptionsListDiags := d.exceptionsListToAPI(ctx)
		diags.Append(exceptionsListDiags...)
		if !exceptionsListDiags.HasError() && len(exceptionsList) > 0 {
			*props.ExceptionsList = &exceptionsList
		}
	}

	// Set risk score mapping
	if props.RiskScoreMapping != nil && typeutils.IsKnown(d.RiskScoreMapping) {
		riskScoreMapping, riskScoreMappingDiags := d.riskScoreMappingToAPI(ctx)
		diags.Append(riskScoreMappingDiags...)
		if !riskScoreMappingDiags.HasError() && len(riskScoreMapping) > 0 {
			*props.RiskScoreMapping = &riskScoreMapping
		}
	}

	// Set building block type
	if props.BuildingBlockType != nil && typeutils.IsKnown(d.BuildingBlockType) {
		buildingBlockType := d.BuildingBlockType.ValueString()
		*props.BuildingBlockType = &buildingBlockType
	}

	// Set data view ID
	if props.DataViewID != nil && typeutils.IsKnown(d.DataViewID) {
		dataViewID := d.DataViewID.ValueString()
		*props.DataViewID = &dataViewID
	}

	// Set namespace
	if props.Namespace != nil && typeutils.IsKnown(d.Namespace) {
		namespace := d.Namespace.ValueString()
		*props.Namespace = &namespace
	}

	// Set rule name override
	if props.RuleNameOverride != nil && typeutils.IsKnown(d.RuleNameOverride) {
		ruleNameOverride := d.RuleNameOverride.ValueString()
		*props.RuleNameOverride = &ruleNameOverride
	}

	// Set timestamp override
	if props.TimestampOverride != nil && typeutils.IsKnown(d.TimestampOverride) {
		timestampOverride := d.TimestampOverride.ValueString()
		*props.TimestampOverride = &timestampOverride
	}

	// Set timestamp override fallback disabled
	if props.TimestampOverrideFallbackDisabled != nil && typeutils.IsKnown(d.TimestampOverrideFallbackDisabled) {
		timestampOverrideFallbackDisabled := d.TimestampOverrideFallbackDisabled.ValueBool()
		*props.TimestampOverrideFallbackDisabled = &timestampOverrideFallbackDisabled
	}

	// Set severity mapping
	if props.SeverityMapping != nil && typeutils.IsKnown(d.SeverityMapping) {
		severityMapping, severityMappingDiags := d.severityMappingToAPI(ctx)
		diags.Append(severityMappingDiags...)
		if !severityMappingDiags.HasError() && severityMapping != nil && len(*severityMapping) > 0 {
			*props.SeverityMapping = severityMapping
		}
	}

	// Set related integrations
	if props.RelatedIntegrations != nil && typeutils.IsKnown(d.RelatedIntegrations) {
		relatedIntegrations, relatedIntegrationsDiags := d.relatedIntegrationsToAPI(ctx)
		diags.Append(relatedIntegrationsDiags...)
		if !relatedIntegrationsDiags.HasError() && relatedIntegrations != nil && len(*relatedIntegrations) > 0 {
			*props.RelatedIntegrations = relatedIntegrations
		}
	}

	// Set required fields
	if props.RequiredFields != nil && typeutils.IsKnown(d.RequiredFields) {
		requiredFields, requiredFieldsDiags := d.requiredFieldsToAPI(ctx)
		diags.Append(requiredFieldsDiags...)
		if !requiredFieldsDiags.HasError() && requiredFields != nil && len(*requiredFields) > 0 {
			*props.RequiredFields = requiredFields
		}
	}

	// Set investigation fields
	if props.InvestigationFields != nil {
		investigationFields, investigationFieldsDiags := d.investigationFieldsToAPI(ctx)
		if !investigationFieldsDiags.HasError() && investigationFields != nil {
			*props.InvestigationFields = investigationFields
		}
		diags.Append(investigationFieldsDiags...)
	}

	// Set response actions
	if props.ResponseActions != nil && typeutils.IsKnown(d.ResponseActions) {
		responseActions, responseActionsDiags := d.responseActionsToAPI(ctx, client)
		diags.Append(responseActionsDiags...)
		if !responseActionsDiags.HasError() && len(responseActions) > 0 {
			*props.ResponseActions = &responseActions
		}
	}

	// Set filters
	if props.Filters != nil && typeutils.IsKnown(d.Filters) {
		filters, filtersDiags := d.filtersToAPI(ctx)
		diags.Append(filtersDiags...)
		if !filtersDiags.HasError() && filters != nil {
			*props.Filters = filters
		}
	}

	// Set alert suppression
	if props.AlertSuppression != nil {
		alertSuppression := d.alertSuppressionToAPI(ctx, diags)
		if alertSuppression != nil {
			*props.AlertSuppression = alertSuppression
		}
	}

	// Set threat (MITRE ATT&CK framework)
	if props.Threat != nil && typeutils.IsKnown(d.Threat) {
		threat, threatDiags := d.threatToAPI(ctx)
		diags.Append(threatDiags...)
		if !threatDiags.HasError() && len(threat) > 0 {
			*props.Threat = &threat
		}
	}

	// Set timeline ID
	if props.TimelineID != nil && typeutils.IsKnown(d.TimelineID) {
		timelineID := d.TimelineID.ValueString()
		*props.TimelineID = &timelineID
	}

	// Set timeline title
	if props.TimelineTitle != nil && typeutils.IsKnown(d.TimelineTitle) {
		timelineTitle := d.TimelineTitle.ValueString()
		*props.TimelineTitle = &timelineTitle
	}
}

// Helper function to initialize fields that should be set to default values for all rule types
func (d *Data) initializeAllFieldsToDefaults() {

	// Initialize fields that should be empty lists for all rule types initially
	if !typeutils.IsKnown(d.Author) {
		d.Author = types.ListNull(types.StringType)
	}
	if !typeutils.IsKnown(d.Tags) {
		d.Tags = types.ListNull(types.StringType)
	}
	if !typeutils.IsKnown(d.FalsePositives) {
		d.FalsePositives = types.ListNull(types.StringType)
	}
	if !typeutils.IsKnown(d.References) {
		d.References = types.ListNull(types.StringType)
	}

	// Initialize new common fields with proper empty lists
	if !typeutils.IsKnown(d.RelatedIntegrations) {
		d.RelatedIntegrations = types.ListNull(getRelatedIntegrationElementType())
	}
	if !typeutils.IsKnown(d.RequiredFields) {
		d.RequiredFields = types.ListNull(getRequiredFieldElementType())
	}
	if !typeutils.IsKnown(d.SeverityMapping) {
		d.SeverityMapping = types.ListNull(getSeverityMappingElementType())
	}

	// Initialize building block type to null by default
	if !typeutils.IsKnown(d.BuildingBlockType) {
		d.BuildingBlockType = types.StringNull()
	}

	// Actions field (common across all rule types)
	if !typeutils.IsKnown(d.Actions) {
		d.Actions = types.ListNull(getActionElementType())
	}

	// Exceptions list field (common across all rule types)
	if !typeutils.IsKnown(d.ExceptionsList) {
		d.ExceptionsList = types.ListNull(getExceptionsListElementType())
	}

	// Initialize all type-specific fields to null/empty by default
	d.initializeTypeSpecificFieldsToDefaults()
}

// Helper function to initialize type-specific fields to default/null values
func (d *Data) initializeTypeSpecificFieldsToDefaults() {
	// EQL-specific fields
	if !typeutils.IsKnown(d.TiebreakerField) {
		d.TiebreakerField = types.StringNull()
	}

	// Machine Learning-specific fields
	if !typeutils.IsKnown(d.AnomalyThreshold) {
		d.AnomalyThreshold = types.Int64Null()
	}
	if !typeutils.IsKnown(d.MachineLearningJobID) {
		d.MachineLearningJobID = types.ListNull(types.StringType)
	}

	// New Terms-specific fields
	if !typeutils.IsKnown(d.NewTermsFields) {
		d.NewTermsFields = types.ListNull(types.StringType)
	}
	if !typeutils.IsKnown(d.HistoryWindowStart) {
		d.HistoryWindowStart = types.StringNull()
	}

	// Saved Query-specific fields
	if !typeutils.IsKnown(d.SavedID) {
		d.SavedID = types.StringNull()
	}

	// Threat Match-specific fields
	if !typeutils.IsKnown(d.ThreatIndex) {
		d.ThreatIndex = types.ListNull(types.StringType)
	}
	if !typeutils.IsKnown(d.ThreatQuery) {
		d.ThreatQuery = types.StringNull()
	}
	if !typeutils.IsKnown(d.ThreatMapping) {
		d.ThreatMapping = types.ListNull(getThreatMappingElementType())
	}
	if !typeutils.IsKnown(d.ThreatFilters) {
		d.ThreatFilters = types.ListNull(types.StringType)
	}
	if !typeutils.IsKnown(d.ThreatIndicatorPath) {
		d.ThreatIndicatorPath = types.StringNull()
	}
	if !typeutils.IsKnown(d.ConcurrentSearches) {
		d.ConcurrentSearches = types.Int64Null()
	}
	if !typeutils.IsKnown(d.ItemsPerSearch) {
		d.ItemsPerSearch = types.Int64Null()
	}

	// Threshold-specific fields
	if !typeutils.IsKnown(d.Threshold) {
		d.Threshold = types.ObjectNull(getThresholdType())
	}

	// Timeline fields (common across multiple rule types)
	if !typeutils.IsKnown(d.TimelineID) {
		d.TimelineID = types.StringNull()
	}
	if !typeutils.IsKnown(d.TimelineTitle) {
		d.TimelineTitle = types.StringNull()
	}

	// Threat field (common across multiple rule types) - MITRE ATT&CK framework
	if !typeutils.IsKnown(d.Threat) {
		d.Threat = types.ListNull(getThreatElementType())
	}
}
