package alerting_rule

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestValidJSONValidator(t *testing.T) {
	tests := []struct {
		name        string
		value       types.String
		expectError bool
	}{
		{
			name:        "valid JSON object",
			value:       types.StringValue(`{"key": "value"}`),
			expectError: false,
		},
		{
			name:        "valid JSON array",
			value:       types.StringValue(`[1, 2, 3]`),
			expectError: false,
		},
		{
			name:        "valid JSON with nested objects",
			value:       types.StringValue(`{"nested": {"key": "value"}, "array": [1, 2]}`),
			expectError: false,
		},
		{
			name:        "valid empty JSON object",
			value:       types.StringValue(`{}`),
			expectError: false,
		},
		{
			name:        "valid JSON string",
			value:       types.StringValue(`"just a string"`),
			expectError: false,
		},
		{
			name:        "valid JSON number",
			value:       types.StringValue(`42`),
			expectError: false,
		},
		{
			name:        "invalid JSON - missing closing brace",
			value:       types.StringValue(`{"key": "value"`),
			expectError: true,
		},
		{
			name:        "invalid JSON - unquoted key",
			value:       types.StringValue(`{key: "value"}`),
			expectError: true,
		},
		{
			name:        "invalid JSON - trailing comma",
			value:       types.StringValue(`{"key": "value",}`),
			expectError: true,
		},
		{
			name:        "invalid JSON - random text",
			value:       types.StringValue(`not json at all`),
			expectError: true,
		},
		{
			name:        "null value - should pass",
			value:       types.StringNull(),
			expectError: false,
		},
		{
			name:        "unknown value - should pass",
			value:       types.StringUnknown(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := ValidJSON()
			req := validator.StringRequest{
				Path:        path.Root("test"),
				ConfigValue: tt.value,
			}
			resp := &validator.StringResponse{}

			v.ValidateString(context.Background(), req, resp)

			if tt.expectError && !resp.Diagnostics.HasError() {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && resp.Diagnostics.HasError() {
				t.Errorf("unexpected error: %v", resp.Diagnostics)
			}
		})
	}
}

func TestValidateParamsForRuleType(t *testing.T) {
	tests := []struct {
		name        string
		ruleTypeID  string
		paramsJSON  string
		expectError bool
		errorMsg    string
	}{
		// Index threshold rule tests
		{
			name:       "valid index-threshold params",
			ruleTypeID: ".index-threshold",
			paramsJSON: `{
				"index": ["test-index"],
				"threshold": [100],
				"thresholdComparator": ">=",
				"timeField": "@timestamp",
				"timeWindowSize": 5,
				"timeWindowUnit": "m"
			}`,
			expectError: false,
		},
		{
			name:       "index-threshold missing index",
			ruleTypeID: ".index-threshold",
			paramsJSON: `{
				"threshold": [100],
				"thresholdComparator": ">=",
				"timeField": "@timestamp",
				"timeWindowSize": 5,
				"timeWindowUnit": "m"
			}`,
			expectError: true,
			errorMsg:    "index",
		},
		{
			name:       "index-threshold missing threshold",
			ruleTypeID: ".index-threshold",
			paramsJSON: `{
				"index": ["test-index"],
				"thresholdComparator": ">=",
				"timeField": "@timestamp",
				"timeWindowSize": 5,
				"timeWindowUnit": "m"
			}`,
			expectError: true,
			errorMsg:    "threshold",
		},
		{
			name:        "index-threshold missing all required",
			ruleTypeID:  ".index-threshold",
			paramsJSON:  `{"aggType": "count"}`,
			expectError: true,
			errorMsg:    "missing required fields",
		},
		// ES Query DSL rule tests
		{
			name:       "valid es-query DSL params",
			ruleTypeID: ".es-query",
			paramsJSON: `{
				"searchType": "esQuery",
				"esQuery": "{\"query\": {\"match_all\": {}}}",
				"index": ["test-index"],
				"threshold": [10],
				"thresholdComparator": ">",
				"timeField": "@timestamp",
				"timeWindowSize": 5,
				"timeWindowUnit": "m"
			}`,
			expectError: false,
		},
		{
			name:       "es-query DSL missing esQuery",
			ruleTypeID: ".es-query",
			paramsJSON: `{
				"searchType": "esQuery",
				"index": ["test-index"],
				"threshold": [10],
				"thresholdComparator": ">",
				"timeField": "@timestamp",
				"timeWindowSize": 5,
				"timeWindowUnit": "m"
			}`,
			expectError: true,
			errorMsg:    "esQuery",
		},
		// ES Query ES|QL rule tests
		{
			name:       "valid es-query ESQL params",
			ruleTypeID: ".es-query",
			paramsJSON: `{
				"searchType": "esqlQuery",
				"esqlQuery": {"esql": "FROM logs | LIMIT 10"},
				"threshold": [0],
				"thresholdComparator": ">",
				"size": 100,
				"timeWindowSize": 5,
				"timeWindowUnit": "m"
			}`,
			expectError: false,
		},
		{
			name:       "es-query ESQL missing esqlQuery",
			ruleTypeID: ".es-query",
			paramsJSON: `{
				"searchType": "esqlQuery",
				"threshold": [0],
				"thresholdComparator": ">",
				"size": 100,
				"timeWindowSize": 5,
				"timeWindowUnit": "m"
			}`,
			expectError: true,
			errorMsg:    "esqlQuery",
		},
		// ES Query KQL rule tests
		{
			name:       "valid es-query KQL params",
			ruleTypeID: ".es-query",
			paramsJSON: `{
				"searchType": "searchSource",
				"threshold": [10],
				"thresholdComparator": ">",
				"size": 100,
				"timeWindowSize": 5,
				"timeWindowUnit": "m"
			}`,
			expectError: false,
		},
		// APM anomaly rule tests
		{
			name:       "valid apm.anomaly params",
			ruleTypeID: "apm.anomaly",
			paramsJSON: `{
				"windowSize": 6,
				"windowUnit": "h",
				"environment": "production",
				"anomalySeverityType": "critical"
			}`,
			expectError: false,
		},
		{
			name:       "apm.anomaly missing environment",
			ruleTypeID: "apm.anomaly",
			paramsJSON: `{
				"windowSize": 6,
				"windowUnit": "h",
				"anomalySeverityType": "critical"
			}`,
			expectError: true,
			errorMsg:    "environment",
		},
		// Unknown rule type should pass (no validation)
		{
			name:        "unknown rule type passes",
			ruleTypeID:  "custom.unknown.rule",
			paramsJSON:  `{"anything": "goes"}`,
			expectError: false,
		},
		// Invalid JSON
		{
			name:        "invalid JSON fails",
			ruleTypeID:  ".index-threshold",
			paramsJSON:  `{invalid`,
			expectError: true,
			errorMsg:    "failed to parse",
		},
		// Unknown fields tests
		{
			name:       "index-threshold with unknown field",
			ruleTypeID: ".index-threshold",
			paramsJSON: `{
				"index": ["test-index"],
				"threshold": [100],
				"thresholdComparator": ">=",
				"timeField": "@timestamp",
				"timeWindowSize": 5,
				"timeWindowUnit": "m",
				"unknownField": "someValue"
			}`,
			expectError: true,
			errorMsg:    "unknown fields",
		},
		{
			name:       "index-threshold with multiple unknown fields",
			ruleTypeID: ".index-threshold",
			paramsJSON: `{
				"index": ["test-index"],
				"threshold": [100],
				"thresholdComparator": ">=",
				"timeField": "@timestamp",
				"timeWindowSize": 5,
				"timeWindowUnit": "m",
				"saddfs": "saddsf",
				"anotherBadField": 123
			}`,
			expectError: true,
			errorMsg:    "unknown fields",
		},
		{
			name:       "es-query DSL with unknown field",
			ruleTypeID: ".es-query",
			paramsJSON: `{
				"searchType": "esQuery",
				"esQuery": "{\"query\": {\"match_all\": {}}}",
				"index": ["test-index"],
				"threshold": [10],
				"thresholdComparator": ">",
				"timeField": "@timestamp",
				"timeWindowSize": 5,
				"timeWindowUnit": "m",
				"notAValidField": true
			}`,
			expectError: true,
			errorMsg:    "unknown fields",
		},
		{
			name:       "apm.anomaly with unknown field",
			ruleTypeID: "apm.anomaly",
			paramsJSON: `{
				"windowSize": 6,
				"windowUnit": "h",
				"environment": "production",
				"anomalySeverityType": "critical",
				"extraParam": "value"
			}`,
			expectError: true,
			errorMsg:    "unknown fields",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateParamsForRuleType(tt.ruleTypeID, tt.paramsJSON)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
