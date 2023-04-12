# RunConnectorSubactionPushtoserviceSubActionParamsIncident

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AlertId** | Pointer to **string** | The alert identifier for Swimlane connectors. | [optional] 
**CaseId** | Pointer to **string** | The case identifier for the incident for Swimlane connectors. | [optional] 
**CaseName** | Pointer to **string** | The case name for the incident for Swimlane connectors. | [optional] 
**Category** | Pointer to **string** | The category of the incident for ServiceNow ITSM and ServiceNow SecOps connectors. | [optional] 
**CorrelationDisplay** | Pointer to **string** | A descriptive label of the alert for correlation purposes for ServiceNow ITSM and ServiceNow SecOps connectors. | [optional] 
**CorrelationId** | Pointer to **string** | The correlation identifier for the security incident for ServiceNow ITSM and ServiveNow SecOps connectors. Connectors using the same correlation ID are associated with the same ServiceNow incident. This value determines whether a new ServiceNow incident is created or an existing one is updated. Modifying this value is optional; if not modified, the rule ID and alert ID are combined as &#x60;{{ruleID}}:{{alert ID}}&#x60; to form the correlation ID value in ServiceNow. The maximum character length for this value is 100 characters. NOTE: Using the default configuration of &#x60;{{ruleID}}:{{alert ID}}&#x60; ensures that ServiceNow creates a separate incident record for every generated alert that uses a unique alert ID. If the rule generates multiple alerts that use the same alert IDs, ServiceNow creates and continually updates a single incident record for the alert.  | [optional] 
**Description** | Pointer to **string** | The description of the incident for Jira, ServiceNow ITSM, ServiceNow SecOps, and Swimlane connectors. | [optional] 
**DestIp** | Pointer to [**RunConnectorSubactionPushtoserviceSubActionParamsIncidentDestIp**](RunConnectorSubactionPushtoserviceSubActionParamsIncidentDestIp.md) |  | [optional] 
**ExternalId** | Pointer to **string** | The Jira, ServiceNow ITSM, or ServiceNow SecOps issue identifier. If present, the incident is updated. Otherwise, a new incident is created.  | [optional] 
**Impact** | Pointer to **string** | The impact of the incident for ServiceNow ITSM connectors. | [optional] 
**IssueType** | Pointer to **int32** | The type of incident for Jira connectors. For example, 10006. To obtain the list of valid values, set &#x60;subAction&#x60; to &#x60;issueTypes&#x60;. | [optional] 
**Labels** | Pointer to **[]string** | The labels for the incident for Jira connectors. NOTE: Labels cannot contain spaces.  | [optional] 
**MalwareHash** | Pointer to [**RunConnectorSubactionPushtoserviceSubActionParamsIncidentMalwareHash**](RunConnectorSubactionPushtoserviceSubActionParamsIncidentMalwareHash.md) |  | [optional] 
**MalwareUrl** | Pointer to [**RunConnectorSubactionPushtoserviceSubActionParamsIncidentMalwareUrl**](RunConnectorSubactionPushtoserviceSubActionParamsIncidentMalwareUrl.md) |  | [optional] 
**Parent** | Pointer to **string** | The ID or key of the parent issue for Jira connectors. Applies only to &#x60;Sub-task&#x60; types of issues. | [optional] 
**Priority** | Pointer to **string** | The priority of the incident in Jira and ServiceNow SecOps connectors. | [optional] 
**RuleName** | Pointer to **string** | The rule name for Swimlane connectors. | [optional] 
**Severity** | Pointer to **string** | The severity of the incident for ServiceNow ITSM and Swimlane connectors. | [optional] 
**ShortDescription** | Pointer to **string** | A short description of the incident for ServiceNow ITSM and ServiceNow SecOps connectors. It is used for searching the contents of the knowledge base.  | [optional] 
**SourceIp** | Pointer to [**RunConnectorSubactionPushtoserviceSubActionParamsIncidentSourceIp**](RunConnectorSubactionPushtoserviceSubActionParamsIncidentSourceIp.md) |  | [optional] 
**Subcategory** | Pointer to **string** | The subcategory of the incident for ServiceNow ITSM and ServiceNow SecOps connectors. | [optional] 
**Summary** | Pointer to **string** | A summary of the incident for Jira connectors. | [optional] 
**Title** | Pointer to **string** | A title for the incident for Jira connectors. It is used for searching the contents of the knowledge base.  | [optional] 
**Urgency** | Pointer to **string** | The urgency of the incident for ServiceNow ITSM connectors. | [optional] 

## Methods

### NewRunConnectorSubactionPushtoserviceSubActionParamsIncident

`func NewRunConnectorSubactionPushtoserviceSubActionParamsIncident() *RunConnectorSubactionPushtoserviceSubActionParamsIncident`

NewRunConnectorSubactionPushtoserviceSubActionParamsIncident instantiates a new RunConnectorSubactionPushtoserviceSubActionParamsIncident object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRunConnectorSubactionPushtoserviceSubActionParamsIncidentWithDefaults

`func NewRunConnectorSubactionPushtoserviceSubActionParamsIncidentWithDefaults() *RunConnectorSubactionPushtoserviceSubActionParamsIncident`

NewRunConnectorSubactionPushtoserviceSubActionParamsIncidentWithDefaults instantiates a new RunConnectorSubactionPushtoserviceSubActionParamsIncident object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAlertId

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetAlertId() string`

GetAlertId returns the AlertId field if non-nil, zero value otherwise.

### GetAlertIdOk

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetAlertIdOk() (*string, bool)`

GetAlertIdOk returns a tuple with the AlertId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAlertId

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetAlertId(v string)`

SetAlertId sets AlertId field to given value.

### HasAlertId

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasAlertId() bool`

HasAlertId returns a boolean if a field has been set.

### GetCaseId

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetCaseId() string`

GetCaseId returns the CaseId field if non-nil, zero value otherwise.

### GetCaseIdOk

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetCaseIdOk() (*string, bool)`

GetCaseIdOk returns a tuple with the CaseId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCaseId

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetCaseId(v string)`

SetCaseId sets CaseId field to given value.

### HasCaseId

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasCaseId() bool`

HasCaseId returns a boolean if a field has been set.

### GetCaseName

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetCaseName() string`

GetCaseName returns the CaseName field if non-nil, zero value otherwise.

### GetCaseNameOk

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetCaseNameOk() (*string, bool)`

GetCaseNameOk returns a tuple with the CaseName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCaseName

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetCaseName(v string)`

SetCaseName sets CaseName field to given value.

### HasCaseName

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasCaseName() bool`

HasCaseName returns a boolean if a field has been set.

### GetCategory

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetCategory() string`

GetCategory returns the Category field if non-nil, zero value otherwise.

### GetCategoryOk

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetCategoryOk() (*string, bool)`

GetCategoryOk returns a tuple with the Category field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCategory

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetCategory(v string)`

SetCategory sets Category field to given value.

### HasCategory

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasCategory() bool`

HasCategory returns a boolean if a field has been set.

### GetCorrelationDisplay

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetCorrelationDisplay() string`

GetCorrelationDisplay returns the CorrelationDisplay field if non-nil, zero value otherwise.

### GetCorrelationDisplayOk

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetCorrelationDisplayOk() (*string, bool)`

GetCorrelationDisplayOk returns a tuple with the CorrelationDisplay field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCorrelationDisplay

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetCorrelationDisplay(v string)`

SetCorrelationDisplay sets CorrelationDisplay field to given value.

### HasCorrelationDisplay

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasCorrelationDisplay() bool`

HasCorrelationDisplay returns a boolean if a field has been set.

### GetCorrelationId

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetCorrelationId() string`

GetCorrelationId returns the CorrelationId field if non-nil, zero value otherwise.

### GetCorrelationIdOk

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetCorrelationIdOk() (*string, bool)`

GetCorrelationIdOk returns a tuple with the CorrelationId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCorrelationId

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetCorrelationId(v string)`

SetCorrelationId sets CorrelationId field to given value.

### HasCorrelationId

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasCorrelationId() bool`

HasCorrelationId returns a boolean if a field has been set.

### GetDescription

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetDestIp

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetDestIp() RunConnectorSubactionPushtoserviceSubActionParamsIncidentDestIp`

GetDestIp returns the DestIp field if non-nil, zero value otherwise.

### GetDestIpOk

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetDestIpOk() (*RunConnectorSubactionPushtoserviceSubActionParamsIncidentDestIp, bool)`

GetDestIpOk returns a tuple with the DestIp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDestIp

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetDestIp(v RunConnectorSubactionPushtoserviceSubActionParamsIncidentDestIp)`

SetDestIp sets DestIp field to given value.

### HasDestIp

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasDestIp() bool`

HasDestIp returns a boolean if a field has been set.

### GetExternalId

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetExternalId() string`

GetExternalId returns the ExternalId field if non-nil, zero value otherwise.

### GetExternalIdOk

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetExternalIdOk() (*string, bool)`

GetExternalIdOk returns a tuple with the ExternalId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExternalId

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetExternalId(v string)`

SetExternalId sets ExternalId field to given value.

### HasExternalId

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasExternalId() bool`

HasExternalId returns a boolean if a field has been set.

### GetImpact

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetImpact() string`

GetImpact returns the Impact field if non-nil, zero value otherwise.

### GetImpactOk

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetImpactOk() (*string, bool)`

GetImpactOk returns a tuple with the Impact field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImpact

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetImpact(v string)`

SetImpact sets Impact field to given value.

### HasImpact

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasImpact() bool`

HasImpact returns a boolean if a field has been set.

### GetIssueType

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetIssueType() int32`

GetIssueType returns the IssueType field if non-nil, zero value otherwise.

### GetIssueTypeOk

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetIssueTypeOk() (*int32, bool)`

GetIssueTypeOk returns a tuple with the IssueType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIssueType

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetIssueType(v int32)`

SetIssueType sets IssueType field to given value.

### HasIssueType

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasIssueType() bool`

HasIssueType returns a boolean if a field has been set.

### GetLabels

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetLabels() []string`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetLabelsOk() (*[]string, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetLabels(v []string)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### GetMalwareHash

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetMalwareHash() RunConnectorSubactionPushtoserviceSubActionParamsIncidentMalwareHash`

GetMalwareHash returns the MalwareHash field if non-nil, zero value otherwise.

### GetMalwareHashOk

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetMalwareHashOk() (*RunConnectorSubactionPushtoserviceSubActionParamsIncidentMalwareHash, bool)`

GetMalwareHashOk returns a tuple with the MalwareHash field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMalwareHash

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetMalwareHash(v RunConnectorSubactionPushtoserviceSubActionParamsIncidentMalwareHash)`

SetMalwareHash sets MalwareHash field to given value.

### HasMalwareHash

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasMalwareHash() bool`

HasMalwareHash returns a boolean if a field has been set.

### GetMalwareUrl

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetMalwareUrl() RunConnectorSubactionPushtoserviceSubActionParamsIncidentMalwareUrl`

GetMalwareUrl returns the MalwareUrl field if non-nil, zero value otherwise.

### GetMalwareUrlOk

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetMalwareUrlOk() (*RunConnectorSubactionPushtoserviceSubActionParamsIncidentMalwareUrl, bool)`

GetMalwareUrlOk returns a tuple with the MalwareUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMalwareUrl

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetMalwareUrl(v RunConnectorSubactionPushtoserviceSubActionParamsIncidentMalwareUrl)`

SetMalwareUrl sets MalwareUrl field to given value.

### HasMalwareUrl

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasMalwareUrl() bool`

HasMalwareUrl returns a boolean if a field has been set.

### GetParent

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetParent() string`

GetParent returns the Parent field if non-nil, zero value otherwise.

### GetParentOk

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetParentOk() (*string, bool)`

GetParentOk returns a tuple with the Parent field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParent

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetParent(v string)`

SetParent sets Parent field to given value.

### HasParent

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasParent() bool`

HasParent returns a boolean if a field has been set.

### GetPriority

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetPriority() string`

GetPriority returns the Priority field if non-nil, zero value otherwise.

### GetPriorityOk

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetPriorityOk() (*string, bool)`

GetPriorityOk returns a tuple with the Priority field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriority

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetPriority(v string)`

SetPriority sets Priority field to given value.

### HasPriority

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasPriority() bool`

HasPriority returns a boolean if a field has been set.

### GetRuleName

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetRuleName() string`

GetRuleName returns the RuleName field if non-nil, zero value otherwise.

### GetRuleNameOk

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetRuleNameOk() (*string, bool)`

GetRuleNameOk returns a tuple with the RuleName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRuleName

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetRuleName(v string)`

SetRuleName sets RuleName field to given value.

### HasRuleName

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasRuleName() bool`

HasRuleName returns a boolean if a field has been set.

### GetSeverity

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetSeverity() string`

GetSeverity returns the Severity field if non-nil, zero value otherwise.

### GetSeverityOk

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetSeverityOk() (*string, bool)`

GetSeverityOk returns a tuple with the Severity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSeverity

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetSeverity(v string)`

SetSeverity sets Severity field to given value.

### HasSeverity

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasSeverity() bool`

HasSeverity returns a boolean if a field has been set.

### GetShortDescription

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetShortDescription() string`

GetShortDescription returns the ShortDescription field if non-nil, zero value otherwise.

### GetShortDescriptionOk

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetShortDescriptionOk() (*string, bool)`

GetShortDescriptionOk returns a tuple with the ShortDescription field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetShortDescription

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetShortDescription(v string)`

SetShortDescription sets ShortDescription field to given value.

### HasShortDescription

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasShortDescription() bool`

HasShortDescription returns a boolean if a field has been set.

### GetSourceIp

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetSourceIp() RunConnectorSubactionPushtoserviceSubActionParamsIncidentSourceIp`

GetSourceIp returns the SourceIp field if non-nil, zero value otherwise.

### GetSourceIpOk

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetSourceIpOk() (*RunConnectorSubactionPushtoserviceSubActionParamsIncidentSourceIp, bool)`

GetSourceIpOk returns a tuple with the SourceIp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceIp

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetSourceIp(v RunConnectorSubactionPushtoserviceSubActionParamsIncidentSourceIp)`

SetSourceIp sets SourceIp field to given value.

### HasSourceIp

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasSourceIp() bool`

HasSourceIp returns a boolean if a field has been set.

### GetSubcategory

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetSubcategory() string`

GetSubcategory returns the Subcategory field if non-nil, zero value otherwise.

### GetSubcategoryOk

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetSubcategoryOk() (*string, bool)`

GetSubcategoryOk returns a tuple with the Subcategory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubcategory

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetSubcategory(v string)`

SetSubcategory sets Subcategory field to given value.

### HasSubcategory

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasSubcategory() bool`

HasSubcategory returns a boolean if a field has been set.

### GetSummary

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetSummary() string`

GetSummary returns the Summary field if non-nil, zero value otherwise.

### GetSummaryOk

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetSummaryOk() (*string, bool)`

GetSummaryOk returns a tuple with the Summary field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSummary

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetSummary(v string)`

SetSummary sets Summary field to given value.

### HasSummary

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasSummary() bool`

HasSummary returns a boolean if a field has been set.

### GetTitle

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetTitle() string`

GetTitle returns the Title field if non-nil, zero value otherwise.

### GetTitleOk

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetTitleOk() (*string, bool)`

GetTitleOk returns a tuple with the Title field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTitle

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetTitle(v string)`

SetTitle sets Title field to given value.

### HasTitle

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasTitle() bool`

HasTitle returns a boolean if a field has been set.

### GetUrgency

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetUrgency() string`

GetUrgency returns the Urgency field if non-nil, zero value otherwise.

### GetUrgencyOk

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetUrgencyOk() (*string, bool)`

GetUrgencyOk returns a tuple with the Urgency field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrgency

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetUrgency(v string)`

SetUrgency sets Urgency field to given value.

### HasUrgency

`func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasUrgency() bool`

HasUrgency returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


