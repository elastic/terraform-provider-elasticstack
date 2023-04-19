/*
 * Connectors
 *
 * OpenAPI schema for Connectors endpoints
 *
 * API version: 0.1
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package kibanaactions

// The `getIncident` subaction for Jira, ServiceNow ITSM, and ServiceNow SecOps connectors.
type RunConnectorSubactionGetincident struct {
	// The action to test.
	SubAction       string                                           `json:"subAction"`
	SubActionParams *RunConnectorSubactionGetincidentSubActionParams `json:"subActionParams"`
}
