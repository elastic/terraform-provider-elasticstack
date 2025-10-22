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
	"github.com/stretchr/testify/require"
)

func TestAllowedIfDependentPathOneOf(t *testing.T) {
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
			v := AllowedIfDependentPathOneOf(
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

func TestAllowedIfDependentPathOneOf_Description(t *testing.T) {
	v := AllowedIfDependentPathOneOf(
		path.Root("auth_type"),
		[]string{"none"},
	)

	description := v.Description(context.Background())
	expected := "value can only be set when auth_type equals \"none\""

	if description != expected {
		t.Errorf("Expected description %q, got %q", expected, description)
	}
}

func TestForbiddenIfDependentPathOneOf(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		currentValue   types.String
		dependentValue types.String
		expectedError  bool
	}

	testCases := []testCase{
		{
			name:           "valid - current null, dependent matches forbidden value",
			currentValue:   types.StringNull(),
			dependentValue: types.StringValue("https"),
			expectedError:  false,
		},
		{
			name:           "valid - current unknown, dependent matches forbidden value",
			currentValue:   types.StringUnknown(),
			dependentValue: types.StringValue("https"),
			expectedError:  false,
		},
		{
			name:           "valid - current set, dependent doesn't match forbidden value",
			currentValue:   types.StringValue("custom_cert"),
			dependentValue: types.StringValue("http"),
			expectedError:  false,
		},
		{
			name:           "invalid - current set, dependent matches forbidden value",
			currentValue:   types.StringValue("custom_cert"),
			dependentValue: types.StringValue("https"),
			expectedError:  true,
		},
		{
			name:           "invalid - current set, dependent matches one of multiple forbidden values",
			currentValue:   types.StringValue("custom_cert"),
			dependentValue: types.StringValue("tls"),
			expectedError:  true,
		},
		{
			name:           "valid - current set, dependent is null",
			currentValue:   types.StringValue("custom_cert"),
			dependentValue: types.StringNull(),
			expectedError:  false,
		},
		{
			name:           "valid - current set, dependent is unknown",
			currentValue:   types.StringValue("custom_cert"),
			dependentValue: types.StringUnknown(),
			expectedError:  false,
		},
		{
			name:           "valid - current null, dependent is null",
			currentValue:   types.StringNull(),
			dependentValue: types.StringNull(),
			expectedError:  false,
		},
		{
			name:           "valid - current null, dependent is unknown",
			currentValue:   types.StringNull(),
			dependentValue: types.StringUnknown(),
			expectedError:  false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Create a simple schema for testing
			testSchema := schema.Schema{
				Attributes: map[string]schema.Attribute{
					"custom_cert": schema.StringAttribute{
						Optional: true,
					},
					"protocol": schema.StringAttribute{
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
				"custom_cert": currentTfValue,
				"protocol":    dependentTfValue,
			}

			rawConfig := tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"custom_cert": tftypes.String,
						"protocol":    tftypes.String,
					},
				},
				rawConfigValues,
			)

			config := tfsdk.Config{
				Raw:    rawConfig,
				Schema: testSchema,
			}

			// Create validator - StringConditionalForbidden forbids the field when dependent matches forbidden values
			v := ForbiddenIfDependentPathOneOf(
				path.Root("protocol"),
				[]string{"https", "tls"},
			)

			// Create validation request
			request := validator.StringRequest{
				Path:        path.Root("custom_cert"),
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

func TestForbiddenIfDependentPathOneOf_Description(t *testing.T) {
	v := ForbiddenIfDependentPathOneOf(
		path.Root("protocol"),
		[]string{"https", "tls"},
	)

	description := v.Description(context.Background())
	// Note: Currently the Description method doesn't differentiate between allowed and forbidden
	// This matches the current implementation behavior
	require.Equal(t, "value cannot be set when protocol is one of [https tls]", description)
}

func TestRequiredIfDependentPathOneOf(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		currentValue   types.String
		dependentValue types.String
		expectedError  bool
	}

	testCases := []testCase{
		{
			name:           "valid - current set, dependent matches required value",
			currentValue:   types.StringValue("some_value"),
			dependentValue: types.StringValue("ssl"),
			expectedError:  false,
		},
		{
			name:           "valid - current null, dependent doesn't match required value",
			currentValue:   types.StringNull(),
			dependentValue: types.StringValue("none"),
			expectedError:  false,
		},
		{
			name:           "valid - current unknown, dependent doesn't match required value",
			currentValue:   types.StringUnknown(),
			dependentValue: types.StringValue("basic"),
			expectedError:  false,
		},
		{
			name:           "valid - current set, dependent matches one of multiple allowed values",
			currentValue:   types.StringValue("certificate_path"),
			dependentValue: types.StringValue("tls"),
			expectedError:  false,
		},
		{
			name:           "invalid - current null, dependent matches required value",
			currentValue:   types.StringNull(),
			dependentValue: types.StringValue("ssl"),
			expectedError:  true,
		},
		{
			name:           "invalid - current unknown, dependent matches required value",
			currentValue:   types.StringUnknown(),
			dependentValue: types.StringValue("tls"),
			expectedError:  true,
		},
		{
			name:           "valid - current null, dependent is null",
			currentValue:   types.StringNull(),
			dependentValue: types.StringNull(),
			expectedError:  false,
		},
		{
			name:           "valid - current null, dependent is unknown",
			currentValue:   types.StringNull(),
			dependentValue: types.StringUnknown(),
			expectedError:  false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Create a simple schema for testing
			testSchema := schema.Schema{
				Attributes: map[string]schema.Attribute{
					"ssl_cert": schema.StringAttribute{
						Optional: true,
					},
					"security_mode": schema.StringAttribute{
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
				"ssl_cert":      currentTfValue,
				"security_mode": dependentTfValue,
			}

			rawConfig := tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"ssl_cert":      tftypes.String,
						"security_mode": tftypes.String,
					},
				},
				rawConfigValues,
			)

			config := tfsdk.Config{
				Raw:    rawConfig,
				Schema: testSchema,
			}

			// Create validator - RequiredIfDependentPathOneOf requires the field when dependent matches allowed values
			v := RequiredIfDependentPathOneOf(
				path.Root("security_mode"),
				[]string{"ssl", "tls"},
			)

			// Create validation request
			request := validator.StringRequest{
				Path:        path.Root("ssl_cert"),
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

func TestRequiredIfDependentPathOneOf_Description(t *testing.T) {
	v := RequiredIfDependentPathOneOf(
		path.Root("security_mode"),
		[]string{"ssl", "tls"},
	)

	description := v.Description(context.Background())
	require.Equal(t, "value required when security_mode is one of [ssl tls]", description)
}

func TestRequiredIfDependentPathEquals_Description(t *testing.T) {
	v := RequiredIfDependentPathEquals(
		path.Root("auth_type"),
		"oauth",
	)

	description := v.Description(context.Background())
	expected := "value required when auth_type equals \"oauth\""

	if description != expected {
		t.Errorf("Expected description %q, got %q", expected, description)
	}
}

func TestDependantPathOneOf(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		currentValue   types.String
		dependentValue types.String
		expectedError  bool
	}

	testCases := []testCase{
		{
			name:           "valid - current null, dependent matches allowed value",
			currentValue:   types.StringNull(),
			dependentValue: types.StringValue("machine_learning"),
			expectedError:  false,
		},
		{
			name:           "valid - current unknown, dependent matches allowed value",
			currentValue:   types.StringUnknown(),
			dependentValue: types.StringValue("esql"),
			expectedError:  false,
		},
		{
			name:           "valid - current set, dependent matches required value",
			currentValue:   types.StringValue("some_value"),
			dependentValue: types.StringValue("machine_learning"),
			expectedError:  false,
		},
		{
			name:           "invalid - current null, dependent doesn't match required value",
			currentValue:   types.StringNull(),
			dependentValue: types.StringValue("other_type"),
			expectedError:  true,
		},
		{
			name:           "invalid - current set, dependent doesn't match required value",
			currentValue:   types.StringValue("some_value"),
			dependentValue: types.StringValue("other_type"),
			expectedError:  true,
		},
		{
			name:           "invalid - current set, dependent is null",
			currentValue:   types.StringValue("some_value"),
			dependentValue: types.StringNull(),
			expectedError:  true,
		},
		{
			name:           "invalid - current set, dependent is unknown",
			currentValue:   types.StringValue("some_value"),
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
					"some_field": schema.StringAttribute{
						Optional: true,
					},
					"type": schema.StringAttribute{
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
				"some_field": currentTfValue,
				"type":       dependentTfValue,
			}

			rawConfig := tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"some_field": tftypes.String,
						"type":       tftypes.String,
					},
				},
				rawConfigValues,
			)

			config := tfsdk.Config{
				Raw:    rawConfig,
				Schema: testSchema,
			}

			// Create validator - StringAssert validates that the dependent field matches allowed values
			v := DependantPathOneOf(
				path.Root("type"),
				[]string{"machine_learning", "esql"},
			)

			// Create validation request
			request := validator.StringRequest{
				Path:        path.Root("some_field"),
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

func TestDependantPathOneOf_Description(t *testing.T) {
	v := DependantPathOneOf(
		path.Root("type"),
		[]string{"machine_learning", "esql"},
	)

	description := v.Description(context.Background())
	expected := "Attribute 'type' is not one of [machine_learning esql]"

	if description != expected {
		t.Errorf("Expected description %q, got %q", expected, description)
	}
}
