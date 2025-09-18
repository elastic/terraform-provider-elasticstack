package validators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestStringConditionalRequirement(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		currentValue   types.String
		dependentValue types.String
		expectedError  bool
	}

	testCases := []testCase{
		{
			name:           "valid - current null, dependent any value",
			currentValue:   types.StringNull(),
			dependentValue: types.StringValue("user_pass"),
			expectedError:  false,
		},
		{
			name:           "valid - current unknown, dependent any value",
			currentValue:   types.StringUnknown(),
			dependentValue: types.StringValue("user_pass"),
			expectedError:  false,
		},
		{
			name:           "valid - current set, dependent matches required value",
			currentValue:   types.StringValue("plaintext"),
			dependentValue: types.StringValue("none"),
			expectedError:  false,
		},
		{
			name:           "invalid - current set, dependent doesn't match required value",
			currentValue:   types.StringValue("plaintext"),
			dependentValue: types.StringValue("user_pass"),
			expectedError:  true,
		},
		{
			name:           "invalid - current set, dependent is null",
			currentValue:   types.StringValue("plaintext"),
			dependentValue: types.StringNull(),
			expectedError:  true,
		},
		{
			name:           "invalid - current set, dependent is unknown",
			currentValue:   types.StringValue("plaintext"),
			dependentValue: types.StringUnknown(),
			expectedError:  true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Create a simple schema for testing
			testSchema := schema.Schema{
				Attributes: map[string]schema.Attribute{
					"connection_type": schema.StringAttribute{
						Optional: true,
					},
					"auth_type": schema.StringAttribute{
						Optional: true,
					},
				},
			}

			// Create raw config values
			currentTfValue, err := testCase.currentValue.ToTerraformValue(context.Background())
			if err != nil {
				t.Fatalf("Error converting current value: %v", err)
			}
			dependentTfValue, err := testCase.dependentValue.ToTerraformValue(context.Background())
			if err != nil {
				t.Fatalf("Error converting dependent value: %v", err)
			}

			rawConfigValues := map[string]tftypes.Value{
				"connection_type": currentTfValue,
				"auth_type":       dependentTfValue,
			}

			rawConfig := tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"connection_type": tftypes.String,
						"auth_type":       tftypes.String,
					},
				},
				rawConfigValues,
			)

			config := tfsdk.Config{
				Raw:    rawConfig,
				Schema: testSchema,
			}

			// Create validator
			v := StringConditionalRequirement(
				path.Root("auth_type"),
				[]string{"none"},
			)

			// Create validation request
			request := validator.StringRequest{
				Path:        path.Root("connection_type"),
				ConfigValue: testCase.currentValue,
				Config:      config,
			}

			// Run validation
			response := &validator.StringResponse{}
			v.ValidateString(context.Background(), request, response)

			// Check result
			if testCase.expectedError {
				if !response.Diagnostics.HasError() {
					t.Errorf("Expected validation error but got none")
				}
			} else {
				if response.Diagnostics.HasError() {
					t.Errorf("Expected no validation error but got: %v", response.Diagnostics.Errors())
				}
			}
		})
	}
}

func TestStringConditionalRequirement_Description(t *testing.T) {
	v := StringConditionalRequirement(
		path.Root("auth_type"),
		[]string{"none"},
	)

	description := v.Description(context.Background())
	expected := "value can only be set when auth_type equals \"none\""

	if description != expected {
		t.Errorf("Expected description %q, got %q", expected, description)
	}
}

func TestFloat64ConditionalRequirement(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		currentValue   types.Float64
		dependentValue types.String
		expectedError  bool
	}

	testCases := []testCase{
		{
			name:           "valid - current null, dependent any value",
			currentValue:   types.Float64Null(),
			dependentValue: types.StringValue("none"),
			expectedError:  false,
		},
		{
			name:           "valid - current unknown, dependent any value",
			currentValue:   types.Float64Unknown(),
			dependentValue: types.StringValue("none"),
			expectedError:  false,
		},
		{
			name:           "valid - current set, dependent matches required value",
			currentValue:   types.Float64Value(6.0),
			dependentValue: types.StringValue("gzip"),
			expectedError:  false,
		},
		{
			name:           "invalid - current set, dependent doesn't match required value",
			currentValue:   types.Float64Value(6.0),
			dependentValue: types.StringValue("none"),
			expectedError:  true,
		},
		{
			name:           "invalid - current set, dependent is null",
			currentValue:   types.Float64Value(6.0),
			dependentValue: types.StringNull(),
			expectedError:  true,
		},
		{
			name:           "invalid - current set, dependent is unknown",
			currentValue:   types.Float64Value(6.0),
			dependentValue: types.StringUnknown(),
			expectedError:  true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Create a simple schema for testing
			testSchema := schema.Schema{
				Attributes: map[string]schema.Attribute{
					"compression_level": schema.Float64Attribute{
						Optional: true,
					},
					"compression": schema.StringAttribute{
						Optional: true,
					},
				},
			}

			// Create raw config values
			currentTfValue, err := testCase.currentValue.ToTerraformValue(context.Background())
			if err != nil {
				t.Fatalf("Error converting current value: %v", err)
			}
			dependentTfValue, err := testCase.dependentValue.ToTerraformValue(context.Background())
			if err != nil {
				t.Fatalf("Error converting dependent value: %v", err)
			}

			rawConfigValues := map[string]tftypes.Value{
				"compression_level": currentTfValue,
				"compression":       dependentTfValue,
			}

			rawConfig := tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"compression_level": tftypes.Number,
						"compression":       tftypes.String,
					},
				},
				rawConfigValues,
			)

			config := tfsdk.Config{
				Raw:    rawConfig,
				Schema: testSchema,
			}

			// Create validator
			v := Float64ConditionalRequirement(
				path.Root("compression"),
				[]string{"gzip"},
			)

			// Create validation request
			request := validator.Float64Request{
				Path:        path.Root("compression_level"),
				ConfigValue: testCase.currentValue,
				Config:      config,
			}

			// Run validation
			response := &validator.Float64Response{}
			v.ValidateFloat64(context.Background(), request, response)

			// Check result
			if testCase.expectedError {
				if !response.Diagnostics.HasError() {
					t.Errorf("Expected validation error but got none")
				}
			} else {
				if response.Diagnostics.HasError() {
					t.Errorf("Expected no validation error but got: %v", response.Diagnostics.Errors())
				}
			}
		})
	}
}
