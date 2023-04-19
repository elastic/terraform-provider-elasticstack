/*
 * Connectors
 *
 * OpenAPI schema for Connectors endpoints
 *
 * API version: 0.1
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package kibanaactions

// The xMatters connector uses the xMatters Workflow for Elastic to send actionable alerts to on-call xMatters resources.
type CreateConnectorRequestXmatters struct {
	Config *ModelMap `json:"config"`
	// The type of connector.
	ConnectorTypeId string `json:"connector_type_id"`
	// The display name for the connector.
	Name    string    `json:"name"`
	Secrets *ModelMap `json:"secrets"`
}
