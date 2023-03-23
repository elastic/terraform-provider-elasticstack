/*
Connectors

OpenAPI schema for Connectors endpoints

API version: 0.1
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package kibanaactions

import (
	"encoding/json"
)

// checks if the RunConnectorSubactionPushtoserviceSubActionParamsIncident type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &RunConnectorSubactionPushtoserviceSubActionParamsIncident{}

// RunConnectorSubactionPushtoserviceSubActionParamsIncident Information necessary to create or update a Jira, ServiceNow ITSM, ServiveNow SecOps, or Swimlane incident.
type RunConnectorSubactionPushtoserviceSubActionParamsIncident struct {
	// The alert identifier for Swimlane connectors.
	AlertId *string `json:"alertId,omitempty"`
	// The case identifier for the incident for Swimlane connectors.
	CaseId *string `json:"caseId,omitempty"`
	// The case name for the incident for Swimlane connectors.
	CaseName *string `json:"caseName,omitempty"`
	// The category of the incident for ServiceNow ITSM and ServiceNow SecOps connectors.
	Category *string `json:"category,omitempty"`
	// A descriptive label of the alert for correlation purposes for ServiceNow ITSM and ServiceNow SecOps connectors.
	CorrelationDisplay *string `json:"correlation_display,omitempty"`
	// The correlation identifier for the security incident for ServiceNow ITSM and ServiveNow SecOps connectors. Connectors using the same correlation ID are associated with the same ServiceNow incident. This value determines whether a new ServiceNow incident is created or an existing one is updated. Modifying this value is optional; if not modified, the rule ID and alert ID are combined as `{{ruleID}}:{{alert ID}}` to form the correlation ID value in ServiceNow. The maximum character length for this value is 100 characters. NOTE: Using the default configuration of `{{ruleID}}:{{alert ID}}` ensures that ServiceNow creates a separate incident record for every generated alert that uses a unique alert ID. If the rule generates multiple alerts that use the same alert IDs, ServiceNow creates and continually updates a single incident record for the alert.
	CorrelationId *string `json:"correlation_id,omitempty"`
	// The description of the incident for Jira, ServiceNow ITSM, ServiceNow SecOps, and Swimlane connectors.
	Description *string                                                          `json:"description,omitempty"`
	DestIp      *RunConnectorSubactionPushtoserviceSubActionParamsIncidentDestIp `json:"dest_ip,omitempty"`
	// The Jira, ServiceNow ITSM, or ServiceNow SecOps issue identifier. If present, the incident is updated. Otherwise, a new incident is created.
	ExternalId *string `json:"externalId,omitempty"`
	// The impact of the incident for ServiceNow ITSM connectors.
	Impact *string `json:"impact,omitempty"`
	// The type of incident for Jira connectors. For example, 10006. To obtain the list of valid values, set `subAction` to `issueTypes`.
	IssueType *int32 `json:"issueType,omitempty"`
	// The labels for the incident for Jira connectors. NOTE: Labels cannot contain spaces.
	Labels      []string                                                              `json:"labels,omitempty"`
	MalwareHash *RunConnectorSubactionPushtoserviceSubActionParamsIncidentMalwareHash `json:"malware_hash,omitempty"`
	MalwareUrl  *RunConnectorSubactionPushtoserviceSubActionParamsIncidentMalwareUrl  `json:"malware_url,omitempty"`
	// The ID or key of the parent issue for Jira connectors. Applies only to `Sub-task` types of issues.
	Parent *string `json:"parent,omitempty"`
	// The priority of the incident in Jira and ServiceNow SecOps connectors.
	Priority *string `json:"priority,omitempty"`
	// The rule name for Swimlane connectors.
	RuleName *string `json:"ruleName,omitempty"`
	// The severity of the incident for ServiceNow ITSM and Swimlane connectors.
	Severity *string `json:"severity,omitempty"`
	// A short description of the incident for ServiceNow ITSM and ServiceNow SecOps connectors. It is used for searching the contents of the knowledge base.
	ShortDescription *string                                                            `json:"short_description,omitempty"`
	SourceIp         *RunConnectorSubactionPushtoserviceSubActionParamsIncidentSourceIp `json:"source_ip,omitempty"`
	// The subcategory of the incident for ServiceNow ITSM and ServiceNow SecOps connectors.
	Subcategory *string `json:"subcategory,omitempty"`
	// A summary of the incident for Jira connectors.
	Summary *string `json:"summary,omitempty"`
	// A title for the incident for Jira connectors. It is used for searching the contents of the knowledge base.
	Title *string `json:"title,omitempty"`
	// The urgency of the incident for ServiceNow ITSM connectors.
	Urgency *string `json:"urgency,omitempty"`
}

// NewRunConnectorSubactionPushtoserviceSubActionParamsIncident instantiates a new RunConnectorSubactionPushtoserviceSubActionParamsIncident object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewRunConnectorSubactionPushtoserviceSubActionParamsIncident() *RunConnectorSubactionPushtoserviceSubActionParamsIncident {
	this := RunConnectorSubactionPushtoserviceSubActionParamsIncident{}
	return &this
}

// NewRunConnectorSubactionPushtoserviceSubActionParamsIncidentWithDefaults instantiates a new RunConnectorSubactionPushtoserviceSubActionParamsIncident object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewRunConnectorSubactionPushtoserviceSubActionParamsIncidentWithDefaults() *RunConnectorSubactionPushtoserviceSubActionParamsIncident {
	this := RunConnectorSubactionPushtoserviceSubActionParamsIncident{}
	return &this
}

// GetAlertId returns the AlertId field value if set, zero value otherwise.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetAlertId() string {
	if o == nil || IsNil(o.AlertId) {
		var ret string
		return ret
	}
	return *o.AlertId
}

// GetAlertIdOk returns a tuple with the AlertId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetAlertIdOk() (*string, bool) {
	if o == nil || IsNil(o.AlertId) {
		return nil, false
	}
	return o.AlertId, true
}

// HasAlertId returns a boolean if a field has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasAlertId() bool {
	if o != nil && !IsNil(o.AlertId) {
		return true
	}

	return false
}

// SetAlertId gets a reference to the given string and assigns it to the AlertId field.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetAlertId(v string) {
	o.AlertId = &v
}

// GetCaseId returns the CaseId field value if set, zero value otherwise.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetCaseId() string {
	if o == nil || IsNil(o.CaseId) {
		var ret string
		return ret
	}
	return *o.CaseId
}

// GetCaseIdOk returns a tuple with the CaseId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetCaseIdOk() (*string, bool) {
	if o == nil || IsNil(o.CaseId) {
		return nil, false
	}
	return o.CaseId, true
}

// HasCaseId returns a boolean if a field has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasCaseId() bool {
	if o != nil && !IsNil(o.CaseId) {
		return true
	}

	return false
}

// SetCaseId gets a reference to the given string and assigns it to the CaseId field.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetCaseId(v string) {
	o.CaseId = &v
}

// GetCaseName returns the CaseName field value if set, zero value otherwise.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetCaseName() string {
	if o == nil || IsNil(o.CaseName) {
		var ret string
		return ret
	}
	return *o.CaseName
}

// GetCaseNameOk returns a tuple with the CaseName field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetCaseNameOk() (*string, bool) {
	if o == nil || IsNil(o.CaseName) {
		return nil, false
	}
	return o.CaseName, true
}

// HasCaseName returns a boolean if a field has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasCaseName() bool {
	if o != nil && !IsNil(o.CaseName) {
		return true
	}

	return false
}

// SetCaseName gets a reference to the given string and assigns it to the CaseName field.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetCaseName(v string) {
	o.CaseName = &v
}

// GetCategory returns the Category field value if set, zero value otherwise.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetCategory() string {
	if o == nil || IsNil(o.Category) {
		var ret string
		return ret
	}
	return *o.Category
}

// GetCategoryOk returns a tuple with the Category field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetCategoryOk() (*string, bool) {
	if o == nil || IsNil(o.Category) {
		return nil, false
	}
	return o.Category, true
}

// HasCategory returns a boolean if a field has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasCategory() bool {
	if o != nil && !IsNil(o.Category) {
		return true
	}

	return false
}

// SetCategory gets a reference to the given string and assigns it to the Category field.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetCategory(v string) {
	o.Category = &v
}

// GetCorrelationDisplay returns the CorrelationDisplay field value if set, zero value otherwise.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetCorrelationDisplay() string {
	if o == nil || IsNil(o.CorrelationDisplay) {
		var ret string
		return ret
	}
	return *o.CorrelationDisplay
}

// GetCorrelationDisplayOk returns a tuple with the CorrelationDisplay field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetCorrelationDisplayOk() (*string, bool) {
	if o == nil || IsNil(o.CorrelationDisplay) {
		return nil, false
	}
	return o.CorrelationDisplay, true
}

// HasCorrelationDisplay returns a boolean if a field has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasCorrelationDisplay() bool {
	if o != nil && !IsNil(o.CorrelationDisplay) {
		return true
	}

	return false
}

// SetCorrelationDisplay gets a reference to the given string and assigns it to the CorrelationDisplay field.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetCorrelationDisplay(v string) {
	o.CorrelationDisplay = &v
}

// GetCorrelationId returns the CorrelationId field value if set, zero value otherwise.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetCorrelationId() string {
	if o == nil || IsNil(o.CorrelationId) {
		var ret string
		return ret
	}
	return *o.CorrelationId
}

// GetCorrelationIdOk returns a tuple with the CorrelationId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetCorrelationIdOk() (*string, bool) {
	if o == nil || IsNil(o.CorrelationId) {
		return nil, false
	}
	return o.CorrelationId, true
}

// HasCorrelationId returns a boolean if a field has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasCorrelationId() bool {
	if o != nil && !IsNil(o.CorrelationId) {
		return true
	}

	return false
}

// SetCorrelationId gets a reference to the given string and assigns it to the CorrelationId field.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetCorrelationId(v string) {
	o.CorrelationId = &v
}

// GetDescription returns the Description field value if set, zero value otherwise.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetDescription() string {
	if o == nil || IsNil(o.Description) {
		var ret string
		return ret
	}
	return *o.Description
}

// GetDescriptionOk returns a tuple with the Description field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetDescriptionOk() (*string, bool) {
	if o == nil || IsNil(o.Description) {
		return nil, false
	}
	return o.Description, true
}

// HasDescription returns a boolean if a field has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasDescription() bool {
	if o != nil && !IsNil(o.Description) {
		return true
	}

	return false
}

// SetDescription gets a reference to the given string and assigns it to the Description field.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetDescription(v string) {
	o.Description = &v
}

// GetDestIp returns the DestIp field value if set, zero value otherwise.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetDestIp() RunConnectorSubactionPushtoserviceSubActionParamsIncidentDestIp {
	if o == nil || IsNil(o.DestIp) {
		var ret RunConnectorSubactionPushtoserviceSubActionParamsIncidentDestIp
		return ret
	}
	return *o.DestIp
}

// GetDestIpOk returns a tuple with the DestIp field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetDestIpOk() (*RunConnectorSubactionPushtoserviceSubActionParamsIncidentDestIp, bool) {
	if o == nil || IsNil(o.DestIp) {
		return nil, false
	}
	return o.DestIp, true
}

// HasDestIp returns a boolean if a field has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasDestIp() bool {
	if o != nil && !IsNil(o.DestIp) {
		return true
	}

	return false
}

// SetDestIp gets a reference to the given RunConnectorSubactionPushtoserviceSubActionParamsIncidentDestIp and assigns it to the DestIp field.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetDestIp(v RunConnectorSubactionPushtoserviceSubActionParamsIncidentDestIp) {
	o.DestIp = &v
}

// GetExternalId returns the ExternalId field value if set, zero value otherwise.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetExternalId() string {
	if o == nil || IsNil(o.ExternalId) {
		var ret string
		return ret
	}
	return *o.ExternalId
}

// GetExternalIdOk returns a tuple with the ExternalId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetExternalIdOk() (*string, bool) {
	if o == nil || IsNil(o.ExternalId) {
		return nil, false
	}
	return o.ExternalId, true
}

// HasExternalId returns a boolean if a field has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasExternalId() bool {
	if o != nil && !IsNil(o.ExternalId) {
		return true
	}

	return false
}

// SetExternalId gets a reference to the given string and assigns it to the ExternalId field.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetExternalId(v string) {
	o.ExternalId = &v
}

// GetImpact returns the Impact field value if set, zero value otherwise.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetImpact() string {
	if o == nil || IsNil(o.Impact) {
		var ret string
		return ret
	}
	return *o.Impact
}

// GetImpactOk returns a tuple with the Impact field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetImpactOk() (*string, bool) {
	if o == nil || IsNil(o.Impact) {
		return nil, false
	}
	return o.Impact, true
}

// HasImpact returns a boolean if a field has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasImpact() bool {
	if o != nil && !IsNil(o.Impact) {
		return true
	}

	return false
}

// SetImpact gets a reference to the given string and assigns it to the Impact field.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetImpact(v string) {
	o.Impact = &v
}

// GetIssueType returns the IssueType field value if set, zero value otherwise.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetIssueType() int32 {
	if o == nil || IsNil(o.IssueType) {
		var ret int32
		return ret
	}
	return *o.IssueType
}

// GetIssueTypeOk returns a tuple with the IssueType field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetIssueTypeOk() (*int32, bool) {
	if o == nil || IsNil(o.IssueType) {
		return nil, false
	}
	return o.IssueType, true
}

// HasIssueType returns a boolean if a field has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasIssueType() bool {
	if o != nil && !IsNil(o.IssueType) {
		return true
	}

	return false
}

// SetIssueType gets a reference to the given int32 and assigns it to the IssueType field.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetIssueType(v int32) {
	o.IssueType = &v
}

// GetLabels returns the Labels field value if set, zero value otherwise.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetLabels() []string {
	if o == nil || IsNil(o.Labels) {
		var ret []string
		return ret
	}
	return o.Labels
}

// GetLabelsOk returns a tuple with the Labels field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetLabelsOk() ([]string, bool) {
	if o == nil || IsNil(o.Labels) {
		return nil, false
	}
	return o.Labels, true
}

// HasLabels returns a boolean if a field has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasLabels() bool {
	if o != nil && !IsNil(o.Labels) {
		return true
	}

	return false
}

// SetLabels gets a reference to the given []string and assigns it to the Labels field.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetLabels(v []string) {
	o.Labels = v
}

// GetMalwareHash returns the MalwareHash field value if set, zero value otherwise.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetMalwareHash() RunConnectorSubactionPushtoserviceSubActionParamsIncidentMalwareHash {
	if o == nil || IsNil(o.MalwareHash) {
		var ret RunConnectorSubactionPushtoserviceSubActionParamsIncidentMalwareHash
		return ret
	}
	return *o.MalwareHash
}

// GetMalwareHashOk returns a tuple with the MalwareHash field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetMalwareHashOk() (*RunConnectorSubactionPushtoserviceSubActionParamsIncidentMalwareHash, bool) {
	if o == nil || IsNil(o.MalwareHash) {
		return nil, false
	}
	return o.MalwareHash, true
}

// HasMalwareHash returns a boolean if a field has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasMalwareHash() bool {
	if o != nil && !IsNil(o.MalwareHash) {
		return true
	}

	return false
}

// SetMalwareHash gets a reference to the given RunConnectorSubactionPushtoserviceSubActionParamsIncidentMalwareHash and assigns it to the MalwareHash field.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetMalwareHash(v RunConnectorSubactionPushtoserviceSubActionParamsIncidentMalwareHash) {
	o.MalwareHash = &v
}

// GetMalwareUrl returns the MalwareUrl field value if set, zero value otherwise.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetMalwareUrl() RunConnectorSubactionPushtoserviceSubActionParamsIncidentMalwareUrl {
	if o == nil || IsNil(o.MalwareUrl) {
		var ret RunConnectorSubactionPushtoserviceSubActionParamsIncidentMalwareUrl
		return ret
	}
	return *o.MalwareUrl
}

// GetMalwareUrlOk returns a tuple with the MalwareUrl field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetMalwareUrlOk() (*RunConnectorSubactionPushtoserviceSubActionParamsIncidentMalwareUrl, bool) {
	if o == nil || IsNil(o.MalwareUrl) {
		return nil, false
	}
	return o.MalwareUrl, true
}

// HasMalwareUrl returns a boolean if a field has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasMalwareUrl() bool {
	if o != nil && !IsNil(o.MalwareUrl) {
		return true
	}

	return false
}

// SetMalwareUrl gets a reference to the given RunConnectorSubactionPushtoserviceSubActionParamsIncidentMalwareUrl and assigns it to the MalwareUrl field.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetMalwareUrl(v RunConnectorSubactionPushtoserviceSubActionParamsIncidentMalwareUrl) {
	o.MalwareUrl = &v
}

// GetParent returns the Parent field value if set, zero value otherwise.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetParent() string {
	if o == nil || IsNil(o.Parent) {
		var ret string
		return ret
	}
	return *o.Parent
}

// GetParentOk returns a tuple with the Parent field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetParentOk() (*string, bool) {
	if o == nil || IsNil(o.Parent) {
		return nil, false
	}
	return o.Parent, true
}

// HasParent returns a boolean if a field has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasParent() bool {
	if o != nil && !IsNil(o.Parent) {
		return true
	}

	return false
}

// SetParent gets a reference to the given string and assigns it to the Parent field.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetParent(v string) {
	o.Parent = &v
}

// GetPriority returns the Priority field value if set, zero value otherwise.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetPriority() string {
	if o == nil || IsNil(o.Priority) {
		var ret string
		return ret
	}
	return *o.Priority
}

// GetPriorityOk returns a tuple with the Priority field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetPriorityOk() (*string, bool) {
	if o == nil || IsNil(o.Priority) {
		return nil, false
	}
	return o.Priority, true
}

// HasPriority returns a boolean if a field has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasPriority() bool {
	if o != nil && !IsNil(o.Priority) {
		return true
	}

	return false
}

// SetPriority gets a reference to the given string and assigns it to the Priority field.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetPriority(v string) {
	o.Priority = &v
}

// GetRuleName returns the RuleName field value if set, zero value otherwise.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetRuleName() string {
	if o == nil || IsNil(o.RuleName) {
		var ret string
		return ret
	}
	return *o.RuleName
}

// GetRuleNameOk returns a tuple with the RuleName field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetRuleNameOk() (*string, bool) {
	if o == nil || IsNil(o.RuleName) {
		return nil, false
	}
	return o.RuleName, true
}

// HasRuleName returns a boolean if a field has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasRuleName() bool {
	if o != nil && !IsNil(o.RuleName) {
		return true
	}

	return false
}

// SetRuleName gets a reference to the given string and assigns it to the RuleName field.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetRuleName(v string) {
	o.RuleName = &v
}

// GetSeverity returns the Severity field value if set, zero value otherwise.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetSeverity() string {
	if o == nil || IsNil(o.Severity) {
		var ret string
		return ret
	}
	return *o.Severity
}

// GetSeverityOk returns a tuple with the Severity field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetSeverityOk() (*string, bool) {
	if o == nil || IsNil(o.Severity) {
		return nil, false
	}
	return o.Severity, true
}

// HasSeverity returns a boolean if a field has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasSeverity() bool {
	if o != nil && !IsNil(o.Severity) {
		return true
	}

	return false
}

// SetSeverity gets a reference to the given string and assigns it to the Severity field.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetSeverity(v string) {
	o.Severity = &v
}

// GetShortDescription returns the ShortDescription field value if set, zero value otherwise.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetShortDescription() string {
	if o == nil || IsNil(o.ShortDescription) {
		var ret string
		return ret
	}
	return *o.ShortDescription
}

// GetShortDescriptionOk returns a tuple with the ShortDescription field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetShortDescriptionOk() (*string, bool) {
	if o == nil || IsNil(o.ShortDescription) {
		return nil, false
	}
	return o.ShortDescription, true
}

// HasShortDescription returns a boolean if a field has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasShortDescription() bool {
	if o != nil && !IsNil(o.ShortDescription) {
		return true
	}

	return false
}

// SetShortDescription gets a reference to the given string and assigns it to the ShortDescription field.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetShortDescription(v string) {
	o.ShortDescription = &v
}

// GetSourceIp returns the SourceIp field value if set, zero value otherwise.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetSourceIp() RunConnectorSubactionPushtoserviceSubActionParamsIncidentSourceIp {
	if o == nil || IsNil(o.SourceIp) {
		var ret RunConnectorSubactionPushtoserviceSubActionParamsIncidentSourceIp
		return ret
	}
	return *o.SourceIp
}

// GetSourceIpOk returns a tuple with the SourceIp field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetSourceIpOk() (*RunConnectorSubactionPushtoserviceSubActionParamsIncidentSourceIp, bool) {
	if o == nil || IsNil(o.SourceIp) {
		return nil, false
	}
	return o.SourceIp, true
}

// HasSourceIp returns a boolean if a field has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasSourceIp() bool {
	if o != nil && !IsNil(o.SourceIp) {
		return true
	}

	return false
}

// SetSourceIp gets a reference to the given RunConnectorSubactionPushtoserviceSubActionParamsIncidentSourceIp and assigns it to the SourceIp field.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetSourceIp(v RunConnectorSubactionPushtoserviceSubActionParamsIncidentSourceIp) {
	o.SourceIp = &v
}

// GetSubcategory returns the Subcategory field value if set, zero value otherwise.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetSubcategory() string {
	if o == nil || IsNil(o.Subcategory) {
		var ret string
		return ret
	}
	return *o.Subcategory
}

// GetSubcategoryOk returns a tuple with the Subcategory field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetSubcategoryOk() (*string, bool) {
	if o == nil || IsNil(o.Subcategory) {
		return nil, false
	}
	return o.Subcategory, true
}

// HasSubcategory returns a boolean if a field has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasSubcategory() bool {
	if o != nil && !IsNil(o.Subcategory) {
		return true
	}

	return false
}

// SetSubcategory gets a reference to the given string and assigns it to the Subcategory field.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetSubcategory(v string) {
	o.Subcategory = &v
}

// GetSummary returns the Summary field value if set, zero value otherwise.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetSummary() string {
	if o == nil || IsNil(o.Summary) {
		var ret string
		return ret
	}
	return *o.Summary
}

// GetSummaryOk returns a tuple with the Summary field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetSummaryOk() (*string, bool) {
	if o == nil || IsNil(o.Summary) {
		return nil, false
	}
	return o.Summary, true
}

// HasSummary returns a boolean if a field has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasSummary() bool {
	if o != nil && !IsNil(o.Summary) {
		return true
	}

	return false
}

// SetSummary gets a reference to the given string and assigns it to the Summary field.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetSummary(v string) {
	o.Summary = &v
}

// GetTitle returns the Title field value if set, zero value otherwise.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetTitle() string {
	if o == nil || IsNil(o.Title) {
		var ret string
		return ret
	}
	return *o.Title
}

// GetTitleOk returns a tuple with the Title field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetTitleOk() (*string, bool) {
	if o == nil || IsNil(o.Title) {
		return nil, false
	}
	return o.Title, true
}

// HasTitle returns a boolean if a field has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasTitle() bool {
	if o != nil && !IsNil(o.Title) {
		return true
	}

	return false
}

// SetTitle gets a reference to the given string and assigns it to the Title field.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetTitle(v string) {
	o.Title = &v
}

// GetUrgency returns the Urgency field value if set, zero value otherwise.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetUrgency() string {
	if o == nil || IsNil(o.Urgency) {
		var ret string
		return ret
	}
	return *o.Urgency
}

// GetUrgencyOk returns a tuple with the Urgency field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) GetUrgencyOk() (*string, bool) {
	if o == nil || IsNil(o.Urgency) {
		return nil, false
	}
	return o.Urgency, true
}

// HasUrgency returns a boolean if a field has been set.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) HasUrgency() bool {
	if o != nil && !IsNil(o.Urgency) {
		return true
	}

	return false
}

// SetUrgency gets a reference to the given string and assigns it to the Urgency field.
func (o *RunConnectorSubactionPushtoserviceSubActionParamsIncident) SetUrgency(v string) {
	o.Urgency = &v
}

func (o RunConnectorSubactionPushtoserviceSubActionParamsIncident) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o RunConnectorSubactionPushtoserviceSubActionParamsIncident) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.AlertId) {
		toSerialize["alertId"] = o.AlertId
	}
	if !IsNil(o.CaseId) {
		toSerialize["caseId"] = o.CaseId
	}
	if !IsNil(o.CaseName) {
		toSerialize["caseName"] = o.CaseName
	}
	if !IsNil(o.Category) {
		toSerialize["category"] = o.Category
	}
	if !IsNil(o.CorrelationDisplay) {
		toSerialize["correlation_display"] = o.CorrelationDisplay
	}
	if !IsNil(o.CorrelationId) {
		toSerialize["correlation_id"] = o.CorrelationId
	}
	if !IsNil(o.Description) {
		toSerialize["description"] = o.Description
	}
	if !IsNil(o.DestIp) {
		toSerialize["dest_ip"] = o.DestIp
	}
	if !IsNil(o.ExternalId) {
		toSerialize["externalId"] = o.ExternalId
	}
	if !IsNil(o.Impact) {
		toSerialize["impact"] = o.Impact
	}
	if !IsNil(o.IssueType) {
		toSerialize["issueType"] = o.IssueType
	}
	if !IsNil(o.Labels) {
		toSerialize["labels"] = o.Labels
	}
	if !IsNil(o.MalwareHash) {
		toSerialize["malware_hash"] = o.MalwareHash
	}
	if !IsNil(o.MalwareUrl) {
		toSerialize["malware_url"] = o.MalwareUrl
	}
	if !IsNil(o.Parent) {
		toSerialize["parent"] = o.Parent
	}
	if !IsNil(o.Priority) {
		toSerialize["priority"] = o.Priority
	}
	if !IsNil(o.RuleName) {
		toSerialize["ruleName"] = o.RuleName
	}
	if !IsNil(o.Severity) {
		toSerialize["severity"] = o.Severity
	}
	if !IsNil(o.ShortDescription) {
		toSerialize["short_description"] = o.ShortDescription
	}
	if !IsNil(o.SourceIp) {
		toSerialize["source_ip"] = o.SourceIp
	}
	if !IsNil(o.Subcategory) {
		toSerialize["subcategory"] = o.Subcategory
	}
	if !IsNil(o.Summary) {
		toSerialize["summary"] = o.Summary
	}
	if !IsNil(o.Title) {
		toSerialize["title"] = o.Title
	}
	if !IsNil(o.Urgency) {
		toSerialize["urgency"] = o.Urgency
	}
	return toSerialize, nil
}

type NullableRunConnectorSubactionPushtoserviceSubActionParamsIncident struct {
	value *RunConnectorSubactionPushtoserviceSubActionParamsIncident
	isSet bool
}

func (v NullableRunConnectorSubactionPushtoserviceSubActionParamsIncident) Get() *RunConnectorSubactionPushtoserviceSubActionParamsIncident {
	return v.value
}

func (v *NullableRunConnectorSubactionPushtoserviceSubActionParamsIncident) Set(val *RunConnectorSubactionPushtoserviceSubActionParamsIncident) {
	v.value = val
	v.isSet = true
}

func (v NullableRunConnectorSubactionPushtoserviceSubActionParamsIncident) IsSet() bool {
	return v.isSet
}

func (v *NullableRunConnectorSubactionPushtoserviceSubActionParamsIncident) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableRunConnectorSubactionPushtoserviceSubActionParamsIncident(val *RunConnectorSubactionPushtoserviceSubActionParamsIncident) *NullableRunConnectorSubactionPushtoserviceSubActionParamsIncident {
	return &NullableRunConnectorSubactionPushtoserviceSubActionParamsIncident{value: val, isSet: true}
}

func (v NullableRunConnectorSubactionPushtoserviceSubActionParamsIncident) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableRunConnectorSubactionPushtoserviceSubActionParamsIncident) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
