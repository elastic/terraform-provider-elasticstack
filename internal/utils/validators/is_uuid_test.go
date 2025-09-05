package validators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestIsUUID_ValidateString(t *testing.T) {
	tests := []struct {
		name        string
		value       types.String
		expectError bool
		errorText   string
	}{
		{
			name:        "null value should not validate",
			value:       types.StringNull(),
			expectError: false,
		},
		{
			name:        "unknown value should not validate",
			value:       types.StringUnknown(),
			expectError: false,
		},
		{
			name:        "valid UUID v4 should pass",
			value:       types.StringValue("550e8400-e29b-41d4-a716-446655440000"),
			expectError: false,
		},
		{
			name:        "valid UUID v1 should pass",
			value:       types.StringValue("6ba7b810-9dad-11d1-80b4-00c04fd430c8"),
			expectError: false,
		},
		{
			name:        "valid UUID v4 with uppercase should pass",
			value:       types.StringValue("550E8400-E29B-41D4-A716-446655440000"),
			expectError: false,
		},
		{
			name:        "valid UUID v4 with mixed case should pass",
			value:       types.StringValue("550e8400-E29B-41d4-A716-446655440000"),
			expectError: false,
		},
		{
			name:        "nil UUID should pass",
			value:       types.StringValue("00000000-0000-0000-0000-000000000000"),
			expectError: false,
		},
		{
			name:        "empty string should fail",
			value:       types.StringValue(""),
			expectError: true,
			errorText:   "Invalid UUID",
		},
		{
			name:        "string without hyphens should fail",
			value:       types.StringValue("550e8400e29b41d4a716446655440000"),
			expectError: true,
			errorText:   "Invalid UUID",
		},
		{
			name:        "UUID with wrong number of hyphens should fail",
			value:       types.StringValue("550e8400-e29b-41d4-a716-44665544-0000"),
			expectError: true,
			errorText:   "Invalid UUID",
		},
		{
			name:        "UUID with too many characters should fail",
			value:       types.StringValue("550e8400-e29b-41d4-a716-4466554400001"),
			expectError: true,
			errorText:   "Invalid UUID",
		},
		{
			name:        "UUID with too few characters should fail",
			value:       types.StringValue("550e8400-e29b-41d4-a716-44665544000"),
			expectError: true,
			errorText:   "Invalid UUID",
		},
		{
			name:        "UUID with invalid characters should fail",
			value:       types.StringValue("550e8400-e29b-41d4-a716-44665544000g"),
			expectError: true,
			errorText:   "Invalid UUID",
		},
		{
			name:        "UUID with spaces should fail",
			value:       types.StringValue("550e8400-e29b-41d4-a716-446655440000 "),
			expectError: true,
			errorText:   "Invalid UUID",
		},
		{
			name:        "UUID with leading spaces should fail",
			value:       types.StringValue(" 550e8400-e29b-41d4-a716-446655440000"),
			expectError: true,
			errorText:   "Invalid UUID",
		},
		{
			name:        "completely invalid string should fail",
			value:       types.StringValue("not-a-uuid-at-all"),
			expectError: true,
			errorText:   "Invalid UUID",
		},
		{
			name:        "numeric string should fail",
			value:       types.StringValue("123456789012345678901234567890123456"),
			expectError: true,
			errorText:   "Invalid UUID",
		},
		{
			name:        "UUID with wrong hyphen positions should fail",
			value:       types.StringValue("550e84-00e2-9b41d4-a716-446655440000"),
			expectError: true,
			errorText:   "Invalid UUID",
		},
		{
			name:        "UUID with missing segments should fail",
			value:       types.StringValue("550e8400--41d4-a716-446655440000"),
			expectError: true,
			errorText:   "Invalid UUID",
		},
		{
			name:        "valid UUID with curly braces should fail (RFC format required)",
			value:       types.StringValue("{550e8400-e29b-41d4-a716-446655440000}"),
			expectError: true,
			errorText:   "Invalid UUID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := IsUUID()
			req := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    tt.value,
			}
			resp := &validator.StringResponse{}

			v.ValidateString(context.Background(), req, resp)

			if tt.expectError {
				require.True(t, resp.Diagnostics.HasError(), "Expected validation error but got none")
				require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), tt.errorText)
				require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), tt.value.ValueString())
			} else {
				require.False(t, resp.Diagnostics.HasError(), "Unexpected validation error: %v", resp.Diagnostics)
			}
		})
	}
}
