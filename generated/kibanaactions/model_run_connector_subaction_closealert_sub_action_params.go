/*
 * Connectors
 *
 * OpenAPI schema for Connectors endpoints
 *
 * API version: 0.1
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package kibanaactions

type RunConnectorSubactionClosealertSubActionParams struct {
	// The unique identifier used for alert deduplication in Opsgenie. The alias must match the value used when creating the alert.
	Alias string `json:"alias"`
	// Additional information for the alert.
	Note string `json:"note,omitempty"`
	// The display name for the source of the alert.
	Source string `json:"source,omitempty"`
	// The display name for the owner.
	User string `json:"user,omitempty"`
}
