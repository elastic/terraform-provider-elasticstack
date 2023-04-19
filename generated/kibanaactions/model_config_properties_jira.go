/*
 * Connectors
 *
 * OpenAPI schema for Connectors endpoints
 *
 * API version: 0.1
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package kibanaactions

// Defines properties for connectors when type is `.jira`.
type ConfigPropertiesJira struct {
	// The Jira instance URL.
	ApiUrl string `json:"apiUrl"`
	// The Jira project key.
	ProjectKey string `json:"projectKey"`
}
