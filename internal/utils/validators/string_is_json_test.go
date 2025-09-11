package validators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestStringIsJSON_ValidateString(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input       types.String
		expectError bool
	}{
		"valid json object": {
			input:       types.StringValue(`{"key": "value"}`),
			expectError: false,
		},
		"valid json array": {
			input:       types.StringValue(`[1, 2, 3]`),
			expectError: false,
		},
		"valid json string": {
			input:       types.StringValue(`"string"`),
			expectError: false,
		},
		"valid json number": {
			input:       types.StringValue(`42`),
			expectError: false,
		},
		"invalid json": {
			input:       types.StringValue(`{invalid json}`),
			expectError: true,
		},
		"empty string": {
			input:       types.StringValue(""),
			expectError: true,
		},
		"null value": {
			input:       types.StringNull(),
			expectError: false,
		},
		"unknown value": {
			input:       types.StringUnknown(),
			expectError: false,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			request := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    testCase.input,
			}
			response := &validator.StringResponse{}

			StringIsJSON{}.ValidateString(context.Background(), request, response)

			if !response.Diagnostics.HasError() && testCase.expectError {
				t.Fatal("expected error, got no error")
			}

			if response.Diagnostics.HasError() && !testCase.expectError {
				t.Fatalf("expected no error, got: %s", response.Diagnostics)
			}
		})
	}
}