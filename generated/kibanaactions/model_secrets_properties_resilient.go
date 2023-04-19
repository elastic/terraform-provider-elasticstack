/*
 * Connectors
 *
 * OpenAPI schema for Connectors endpoints
 *
 * API version: 0.1
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package kibanaactions

// Defines secrets for connectors when type is `.resilient`.
type SecretsPropertiesResilient struct {
	// The authentication key ID for HTTP Basic authentication.
	ApiKeyId string `json:"apiKeyId"`
	// The authentication key secret for HTTP Basic authentication.
	ApiKeySecret string `json:"apiKeySecret"`
}
