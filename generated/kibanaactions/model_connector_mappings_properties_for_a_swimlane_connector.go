/*
 * Connectors
 *
 * OpenAPI schema for Connectors endpoints
 *
 * API version: 0.1
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package kibanaactions

// The field mapping.
type ConnectorMappingsPropertiesForASwimlaneConnector struct {
	AlertIdConfig     *AlertIdentifierMapping `json:"alertIdConfig,omitempty"`
	CaseIdConfig      *CaseIdentifierMapping  `json:"caseIdConfig,omitempty"`
	CaseNameConfig    *CaseNameMapping        `json:"caseNameConfig,omitempty"`
	CommentsConfig    *CaseCommentMapping     `json:"commentsConfig,omitempty"`
	DescriptionConfig *CaseDescriptionMapping `json:"descriptionConfig,omitempty"`
	RuleNameConfig    *RuleNameMapping        `json:"ruleNameConfig,omitempty"`
	SeverityConfig    *SeverityMapping        `json:"severityConfig,omitempty"`
}
