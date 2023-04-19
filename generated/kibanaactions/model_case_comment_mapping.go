/*
 * Connectors
 *
 * OpenAPI schema for Connectors endpoints
 *
 * API version: 0.1
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package kibanaactions

// Mapping for the case comments.
type CaseCommentMapping struct {
	// The type of field in Swimlane.
	FieldType string `json:"fieldType"`
	// The identifier for the field in Swimlane.
	Id string `json:"id"`
	// The key for the field in Swimlane.
	Key string `json:"key"`
	// The name of the field in Swimlane.
	Name string `json:"name"`
}
