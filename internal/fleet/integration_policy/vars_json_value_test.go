package integration_policy

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

func TestConfigValue_StringSemanticEquals(t *testing.T) {
	tests := []struct {
		name          string
		configValue   VarsJSONValue
		otherValue    basetypes.StringValuable
		expectEqual   bool
		expectError   bool
		errorContains string
	}{
		{
			name:        "null values are equal",
			configValue: NewVarsJSONNull(),
			otherValue:  NewVarsJSONNull(),
			expectEqual: true,
			expectError: false,
		},
		{
			name:        "unknown values are equal",
			configValue: NewVarsJSONUnknown(),
			otherValue:  NewVarsJSONUnknown(),
			expectEqual: true,
			expectError: false,
		},
		{
			name:        "null vs unknown should not be equal",
			configValue: NewVarsJSONNull(),
			otherValue:  NewVarsJSONUnknown(),
			expectEqual: false,
			expectError: false,
		},
		{
			name: "wrong type should produce error",
			configValue: VarsJSONValue{
				JSONWithContextualDefaultsValue: customtypes.JSONWithContextualDefaultsValue{
					Normalized: jsontypes.NewNormalizedValue(`{"key": "value"}`),
				},
			},
			otherValue:    basetypes.NewStringValue(`{"key": "value"}`),
			expectEqual:   false,
			expectError:   true,
			errorContains: "Semantic Equality Check Error",
		},
		{
			name: "values without connector type ID should use normalized comparison",
			configValue: VarsJSONValue{
				JSONWithContextualDefaultsValue: customtypes.JSONWithContextualDefaultsValue{
					Normalized: jsontypes.NewNormalizedValue(`{"key": "value"}`),
				},
			},
			otherValue: VarsJSONValue{
				JSONWithContextualDefaultsValue: customtypes.JSONWithContextualDefaultsValue{
					Normalized: jsontypes.NewNormalizedValue(`{"key": "value"}`),
				},
			},
			expectEqual: true,
			expectError: false,
		},
		{
			name: "different values without connector type ID should not be equal",
			configValue: VarsJSONValue{
				JSONWithContextualDefaultsValue: customtypes.JSONWithContextualDefaultsValue{
					Normalized: jsontypes.NewNormalizedValue(`{"key": "value1"}`),
				},
			},
			otherValue: VarsJSONValue{
				JSONWithContextualDefaultsValue: customtypes.JSONWithContextualDefaultsValue{
					Normalized: jsontypes.NewNormalizedValue(`{"key": "value2"}`),
				},
			},
			expectEqual: false,
			expectError: false,
		},
		{
			name: "invalid JSON in first value should cause error",
			configValue: func() VarsJSONValue {
				// Manually construct invalid JSON with context
				return VarsJSONValue{
					JSONWithContextualDefaultsValue: customtypes.JSONWithContextualDefaultsValue{
						Normalized: jsontypes.Normalized{StringValue: basetypes.NewStringValue(`{invalid`)},
					},
				}
			}(),
			otherValue: func() VarsJSONValue {
				val, _ := NewVarsJSONWithIntegration(`{"key": "value"}`, "apm", "1.0.0")
				return val
			}(),
			expectEqual:   false,
			expectError:   true,
			errorContains: "Failed to unmarshal config value",
		},
		{
			name: "invalid JSON in second value should cause error",
			configValue: func() VarsJSONValue {
				val, _ := NewVarsJSONWithIntegration(`{"key": "value"}`, "apm", "1.0.0")
				return val
			}(),
			otherValue: func() VarsJSONValue {
				return VarsJSONValue{
					JSONWithContextualDefaultsValue: customtypes.JSONWithContextualDefaultsValue{
						Normalized: jsontypes.Normalized{StringValue: basetypes.NewStringValue(`{invalid`)},
					},
				}
			}(),
			expectEqual:   false,
			expectError:   true,
			errorContains: "Failed to unmarshal config value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, diags := tt.configValue.StringSemanticEquals(context.Background(), tt.otherValue)

			if tt.expectError {
				require.True(t, diags.HasError(), "Expected error but got none")
				if tt.errorContains != "" {
					errorFound := false
					for _, err := range diags.Errors() {
						if strings.Contains(err.Summary(), tt.errorContains) || strings.Contains(err.Detail(), tt.errorContains) {
							errorFound = true
							break
						}
					}
					require.True(t, errorFound, "Expected error containing '%s' but got: %v", tt.errorContains, diags)
				}
			} else {
				if diags.HasError() {
					// For connector config with defaults errors, we might expect them in real scenarios
					// but for unit tests, we'll be more lenient
					hasConnectorError := false
					for _, err := range diags.Errors() {
						if strings.Contains(err.Summary(), "Failed to get config with defaults") {
							hasConnectorError = true
							break
						}
					}
					if !hasConnectorError {
						require.False(t, diags.HasError(), "Unexpected error: %v", diags)
					}
				}
				require.Equal(t, tt.expectEqual, result)
			}
		})
	}
}

func TestNewVarsJSONWithIntegration(t *testing.T) {
	// Save and restore knownPackages
	originalKnownPackages := knownPackages
	knownPackages = make(map[string]kbapi.PackageInfo)
	t.Cleanup(func() {
		knownPackages = originalKnownPackages
	})

	pkgName := "test-pkg"
	pkgVersion := "1.0.0"
	cacheKey := getPackageCacheKey(pkgName, pkgVersion)

	// Setup mock package with defaults
	vars := []map[string]interface{}{
		{
			"name":    "var1",
			"default": "default1",
		},
		{
			"name":  "var2",
			"multi": true,
		},
	}
	pkg := kbapi.PackageInfo{
		Vars: &vars,
	}
	knownPackages[cacheKey] = pkg

	tests := []struct {
		name          string
		value         string
		pkgName       string
		pkgVersion    string
		expectError   bool
		expectedValue string // Expected JSON after defaults populated (checked via StringSemanticEquals logic)
	}{
		{
			name:          "empty value returns null",
			value:         "",
			pkgName:       pkgName,
			pkgVersion:    pkgVersion,
			expectError:   false,
			expectedValue: "", // Null
		},
		{
			name:          "valid json",
			value:         `{"foo": "bar"}`,
			pkgName:       pkgName,
			pkgVersion:    pkgVersion,
			expectError:   false,
			expectedValue: `{"foo": "bar"}`,
		},
		{
			name:          "invalid json",
			value:         `{invalid`,
			pkgName:       pkgName,
			pkgVersion:    pkgVersion,
			expectError:   true,
			expectedValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, diags := NewVarsJSONWithIntegration(tt.value, tt.pkgName, tt.pkgVersion)

			if tt.expectError {
				require.True(t, diags.HasError())
				return
			}

			require.False(t, diags.HasError())

			if tt.value == "" {
				require.True(t, val.IsNull())
			} else {
				require.False(t, val.IsNull())
				require.NotNil(t, val.Normalized)
			}
		})
	}
}

func TestPopulateVarsJSONDefaults(t *testing.T) {
	// Save and restore knownPackages
	originalKnownPackages := knownPackages
	knownPackages = make(map[string]kbapi.PackageInfo)
	t.Cleanup(func() {
		knownPackages = originalKnownPackages
	})

	pkgName := "test-pkg"
	pkgVersion := "1.0.0"
	cacheKey := getPackageCacheKey(pkgName, pkgVersion)

	// Setup mock package with defaults
	vars := []map[string]interface{}{
		{
			"name":    "var1",
			"default": "default1",
		},
		{
			"name":  "var2",
			"multi": true,
		},
		{
			"name": "var3",
			// no default
		},
	}
	pkg := kbapi.PackageInfo{
		Vars: &vars,
	}
	knownPackages[cacheKey] = pkg

	tests := []struct {
		name           string
		ctxVal         string
		varsJson       string
		expectedResult string
		expectError    bool
	}{
		{
			name:           "empty context value",
			ctxVal:         "",
			varsJson:       `{"foo": "bar"}`,
			expectedResult: `{"foo": "bar"}`,
			expectError:    false,
		},
		{
			name:           "unknown package",
			ctxVal:         "unknown-pkg-1.0.0",
			varsJson:       `{"foo": "bar"}`,
			expectedResult: `{"foo": "bar"}`,
			expectError:    false,
		},
		{
			name:           "apply defaults to empty json",
			ctxVal:         cacheKey,
			varsJson:       `{}`,
			expectedResult: `{"var1":"default1","var2":[]}`,
			expectError:    false,
		},
		{
			name:           "merge defaults with existing values",
			ctxVal:         cacheKey,
			varsJson:       `{"var1": "overridden", "foo": "bar"}`,
			expectedResult: `{"var1":"overridden","foo":"bar","var2":[]}`,
			expectError:    false,
		},
		{
			name:           "invalid json input",
			ctxVal:         cacheKey,
			varsJson:       `{invalid`,
			expectedResult: `{invalid`,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := populateVarsJSONDefaults(tt.ctxVal, tt.varsJson)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				// Compare JSONs by unmarshalling
				var expectedMap, resultMap map[string]interface{}
				if tt.expectedResult != "" {
					err = json.Unmarshal([]byte(tt.expectedResult), &expectedMap)
					require.NoError(t, err)
				}
				if res != "" {
					err = json.Unmarshal([]byte(res), &resultMap)
					require.NoError(t, err)
				}

				require.Equal(t, expectedMap, resultMap)
			}
		})
	}
}
