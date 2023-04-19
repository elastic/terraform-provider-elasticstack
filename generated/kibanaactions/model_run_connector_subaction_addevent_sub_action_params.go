/*
 * Connectors
 *
 * OpenAPI schema for Connectors endpoints
 *
 * API version: 0.1
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package kibanaactions

// The set of configuration properties for the action.
type RunConnectorSubactionAddeventSubActionParams struct {
	// Additional information about the event.
	AdditionalInfo string `json:"additional_info,omitempty"`
	// The details about the event.
	Description string `json:"description,omitempty"`
	// A specific instance of the source.
	EventClass string `json:"event_class,omitempty"`
	// All actions sharing this key are associated with the same ServiceNow alert. The default value is `<rule ID>:<alert instance ID>`.
	MessageKey string `json:"message_key,omitempty"`
	// The name of the metric.
	MetricName string `json:"metric_name,omitempty"`
	// The host that the event was triggered for.
	Node string `json:"node,omitempty"`
	// The name of the resource.
	Resource string `json:"resource,omitempty"`
	// The severity of the event.
	Severity string `json:"severity,omitempty"`
	// The name of the event source type.
	Source string `json:"source,omitempty"`
	// The time of the event.
	TimeOfEvent string `json:"time_of_event,omitempty"`
	// The type of event.
	Type_ string `json:"type,omitempty"`
}
