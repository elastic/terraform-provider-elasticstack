# RunConnectorSubactionPushtoserviceSubActionParamsIncident

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AlertId** | **string** | The alert identifier for Swimlane connectors. | [optional] [default to null]
**CaseId** | **string** | The case identifier for the incident for Swimlane connectors. | [optional] [default to null]
**CaseName** | **string** | The case name for the incident for Swimlane connectors. | [optional] [default to null]
**Category** | **string** | The category of the incident for ServiceNow ITSM and ServiceNow SecOps connectors. | [optional] [default to null]
**CorrelationDisplay** | **string** | A descriptive label of the alert for correlation purposes for ServiceNow ITSM and ServiceNow SecOps connectors. | [optional] [default to null]
**CorrelationId** | **string** | The correlation identifier for the security incident for ServiceNow ITSM and ServiveNow SecOps connectors. Connectors using the same correlation ID are associated with the same ServiceNow incident. This value determines whether a new ServiceNow incident is created or an existing one is updated. Modifying this value is optional; if not modified, the rule ID and alert ID are combined as &#x60;{{ruleID}}:{{alert ID}}&#x60; to form the correlation ID value in ServiceNow. The maximum character length for this value is 100 characters. NOTE: Using the default configuration of &#x60;{{ruleID}}:{{alert ID}}&#x60; ensures that ServiceNow creates a separate incident record for every generated alert that uses a unique alert ID. If the rule generates multiple alerts that use the same alert IDs, ServiceNow creates and continually updates a single incident record for the alert.  | [optional] [default to null]
**Description** | **string** | The description of the incident for Jira, ServiceNow ITSM, ServiceNow SecOps, and Swimlane connectors. | [optional] [default to null]
**DestIp** | [***OneOfrunConnectorSubactionPushtoserviceSubActionParamsIncidentDestIp**](OneOfrunConnectorSubactionPushtoserviceSubActionParamsIncidentDestIp.md) | A list of destination IP addresses related to the security incident for ServiceNow SecOps connectors. The IPs are added as observables to the security incident.  | [optional] [default to null]
**ExternalId** | **string** | The Jira, ServiceNow ITSM, or ServiceNow SecOps issue identifier. If present, the incident is updated. Otherwise, a new incident is created.  | [optional] [default to null]
**Impact** | **string** | The impact of the incident for ServiceNow ITSM connectors. | [optional] [default to null]
**IssueType** | **int32** | The type of incident for Jira connectors. For example, 10006. To obtain the list of valid values, set &#x60;subAction&#x60; to &#x60;issueTypes&#x60;. | [optional] [default to null]
**Labels** | **[]string** | The labels for the incident for Jira connectors. NOTE: Labels cannot contain spaces.  | [optional] [default to null]
**MalwareHash** | [***OneOfrunConnectorSubactionPushtoserviceSubActionParamsIncidentMalwareHash**](OneOfrunConnectorSubactionPushtoserviceSubActionParamsIncidentMalwareHash.md) | A list of malware hashes related to the security incident for ServiceNow SecOps connectors. The hashes are added as observables to the security incident. | [optional] [default to null]
**MalwareUrl** | **OneOfrunConnectorSubactionPushtoserviceSubActionParamsIncidentMalwareUrl** | A list of malware URLs related to the security incident for ServiceNow SecOps connectors. The URLs are added as observables to the security incident. | [optional] [default to null]
**Parent** | **string** | The ID or key of the parent issue for Jira connectors. Applies only to &#x60;Sub-task&#x60; types of issues. | [optional] [default to null]
**Priority** | **string** | The priority of the incident in Jira and ServiceNow SecOps connectors. | [optional] [default to null]
**RuleName** | **string** | The rule name for Swimlane connectors. | [optional] [default to null]
**Severity** | **string** | The severity of the incident for ServiceNow ITSM and Swimlane connectors. | [optional] [default to null]
**ShortDescription** | **string** | A short description of the incident for ServiceNow ITSM and ServiceNow SecOps connectors. It is used for searching the contents of the knowledge base.  | [optional] [default to null]
**SourceIp** | [***OneOfrunConnectorSubactionPushtoserviceSubActionParamsIncidentSourceIp**](OneOfrunConnectorSubactionPushtoserviceSubActionParamsIncidentSourceIp.md) | A list of source IP addresses related to the security incident for ServiceNow SecOps connectors. The IPs are added as observables to the security incident. | [optional] [default to null]
**Subcategory** | **string** | The subcategory of the incident for ServiceNow ITSM and ServiceNow SecOps connectors. | [optional] [default to null]
**Summary** | **string** | A summary of the incident for Jira connectors. | [optional] [default to null]
**Title** | **string** | A title for the incident for Jira connectors. It is used for searching the contents of the knowledge base.  | [optional] [default to null]
**Urgency** | **string** | The urgency of the incident for ServiceNow ITSM connectors. | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

