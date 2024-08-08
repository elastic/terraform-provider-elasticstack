package index

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

func mapToJsonStringValue(t *testing.T, m map[string]interface{}) basetypes.StringValue {
	mBytes, err := json.Marshal(m)
	require.NoError(t, err)

	return types.StringValue(string(mBytes))
}

func Test_PlanModifyString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                    string
		stateMappings           basetypes.StringValue
		configMappings          basetypes.StringValue
		expectedPlanMappings    basetypes.StringValue
		expectedDiags           diag.Diagnostics
		expectedRequiresReplace bool
	}{
		{
			name:           "should do nothing if the state value is unknown",
			stateMappings:  basetypes.NewStringUnknown(),
			configMappings: basetypes.NewStringValue("{}"),
		},
		{
			name:           "should do nothing if the state value is null",
			stateMappings:  basetypes.NewStringNull(),
			configMappings: basetypes.NewStringValue("{}"),
		},
		{
			name:           "should do nothing if the config value is unknown",
			configMappings: basetypes.NewStringUnknown(),
			stateMappings:  basetypes.NewStringValue("{}"),
		},
		{
			name:           "should do nothing if the config value is null",
			configMappings: basetypes.NewStringNull(),
			stateMappings:  basetypes.NewStringValue("{}"),
		},
		{
			name: "should do nothing if the state mappings do not define any properties",
			stateMappings: mapToJsonStringValue(t, map[string]interface{}{
				"not_properties": map[string]interface{}{
					"hello": "world",
				},
			}),
			configMappings: basetypes.NewStringValue("{}"),
		},
		{
			name: "requires replace if state mappings define properties but the config value does not",
			stateMappings: mapToJsonStringValue(t, map[string]interface{}{
				"properties": map[string]interface{}{
					"hello": "world",
				},
			}),
			configMappings:          basetypes.NewStringValue("{}"),
			expectedRequiresReplace: true,
		},
		{
			name: "should not alter the final plan when a new field is added",
			stateMappings: mapToJsonStringValue(t, map[string]interface{}{
				"properties": map[string]interface{}{
					"field1": map[string]interface{}{
						"type": "string",
					},
				},
			}),
			configMappings: mapToJsonStringValue(t, map[string]interface{}{
				"properties": map[string]interface{}{
					"field1": map[string]interface{}{
						"type": "string",
					},
					"field2": map[string]interface{}{
						"type": "string",
					},
				},
			}),
			expectedPlanMappings: mapToJsonStringValue(t, map[string]interface{}{
				"properties": map[string]interface{}{
					"field1": map[string]interface{}{
						"type": "string",
					},
					"field2": map[string]interface{}{
						"type": "string",
					},
				},
			}),
		},
		{
			name: "requires replace when the type of an existing field is changed",
			stateMappings: mapToJsonStringValue(t, map[string]interface{}{
				"properties": map[string]interface{}{
					"field1": map[string]interface{}{
						"type": "string",
					},
				},
			}),
			configMappings: mapToJsonStringValue(t, map[string]interface{}{
				"properties": map[string]interface{}{
					"field1": map[string]interface{}{
						"type": "int",
					},
				},
			}),
			expectedPlanMappings: mapToJsonStringValue(t, map[string]interface{}{
				"properties": map[string]interface{}{
					"field1": map[string]interface{}{
						"type": "int",
					},
				},
			}),
			expectedRequiresReplace: true,
		},
		{
			name: "should add the removed field to the plan and include a warning when a field is removed from config",
			stateMappings: mapToJsonStringValue(t, map[string]interface{}{
				"properties": map[string]interface{}{
					"field1": map[string]interface{}{
						"type": "string",
					},
					"field2": map[string]interface{}{
						"type": "string",
					},
				},
			}),
			configMappings: mapToJsonStringValue(t, map[string]interface{}{
				"properties": map[string]interface{}{
					"field1": map[string]interface{}{
						"type": "string",
					},
				},
			}),
			expectedPlanMappings: mapToJsonStringValue(t, map[string]interface{}{
				"properties": map[string]interface{}{
					"field1": map[string]interface{}{
						"type": "string",
					},
					"field2": map[string]interface{}{
						"type": "string",
					},
				},
			}),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeWarningDiagnostic(
					path.Root("mappings"),
					`removing field [mappings["properties"]["field2"]] in mappings is ignored.`,
					"Elasticsearch will maintain the current field in it's mapping. Re-index to remove the field completely",
				),
			},
		},
		{
			name: "should add the removed field to the plan and include a warning when a sub-field is removed from config",
			stateMappings: mapToJsonStringValue(t, map[string]interface{}{
				"properties": map[string]interface{}{
					"field1": map[string]interface{}{
						"properties": map[string]interface{}{
							"field2": map[string]interface{}{
								"type": "string",
							},
						},
					},
				},
			}),
			configMappings: mapToJsonStringValue(t, map[string]interface{}{
				"properties": map[string]interface{}{
					"field1": map[string]interface{}{
						"properties": map[string]interface{}{
							"field3": map[string]interface{}{
								"type": "string",
							},
						},
					},
				},
			}),
			expectedPlanMappings: mapToJsonStringValue(t, map[string]interface{}{
				"properties": map[string]interface{}{
					"field1": map[string]interface{}{
						"properties": map[string]interface{}{
							"field2": map[string]interface{}{
								"type": "string",
							},
							"field3": map[string]interface{}{
								"type": "string",
							},
						},
					},
				},
			}),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeWarningDiagnostic(
					path.Root("mappings"),
					`removing field [mappings["properties"]["field1"]["properties"]["field2"]] in mappings is ignored.`,
					"Elasticsearch will maintain the current field in it's mapping. Re-index to remove the field completely",
				),
			},
		},
		{
			name: "requires replace when a sub-fields type is changed",
			stateMappings: mapToJsonStringValue(t, map[string]interface{}{
				"properties": map[string]interface{}{
					"field1": map[string]interface{}{
						"properties": map[string]interface{}{
							"field2": map[string]interface{}{
								"type": "string",
							},
						},
					},
				},
			}),
			configMappings: mapToJsonStringValue(t, map[string]interface{}{
				"properties": map[string]interface{}{
					"field1": map[string]interface{}{
						"properties": map[string]interface{}{
							"field2": map[string]interface{}{
								"type": "int",
							},
						},
					},
				},
			}),
			expectedPlanMappings: mapToJsonStringValue(t, map[string]interface{}{
				"properties": map[string]interface{}{
					"field1": map[string]interface{}{
						"properties": map[string]interface{}{
							"field2": map[string]interface{}{
								"type": "int",
							},
						},
					},
				},
			}),
			expectedRequiresReplace: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modifier := mappingsPlanModifier{}
			resp := planmodifier.StringResponse{}
			modifier.PlanModifyString(context.Background(), planmodifier.StringRequest{
				ConfigValue: tt.configMappings,
				StateValue:  tt.stateMappings,
			}, &resp)

			require.Equal(t, tt.expectedDiags, resp.Diagnostics)
			require.Equal(t, tt.expectedPlanMappings, resp.PlanValue)
			require.Equal(t, tt.expectedRequiresReplace, resp.RequiresReplace)
		})
	}
}
