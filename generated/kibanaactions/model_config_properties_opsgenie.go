/*
 * Connectors
 *
 * OpenAPI schema for Connectors endpoints
 *
 * API version: 0.1
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package kibanaactions

// Defines properties for connectors when type is `.opsgenie`.
type ConfigPropertiesOpsgenie struct {
	// The Opsgenie URL. For example, `https://api.opsgenie.com` or `https://api.eu.opsgenie.com`. If you are using the `xpack.actions.allowedHosts` setting, add the hostname to the allowed hosts.
	ApiUrl string `json:"apiUrl"`
}
