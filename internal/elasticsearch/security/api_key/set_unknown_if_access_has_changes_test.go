package api_key

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetUnknownIfAccessHasChanges(t *testing.T) {
	t.Parallel()

	// Define the schema for testing
	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"type":             schema.StringAttribute{},
			"role_descriptors": schema.StringAttribute{},
			"access": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"search": schema.ListNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"names": schema.ListAttribute{ElementType: types.StringType},
							},
						},
					},
					"replication": schema.ListNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"names": schema.ListAttribute{ElementType: types.StringType},
							},
						},
					},
				},
			},
		},
	}

	// Define object type for tftypes
	objectType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"type":             tftypes.String,
			"role_descriptors": tftypes.String,
			"access": tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"search": tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"names": tftypes.List{ElementType: tftypes.String},
							},
						},
					},
					"replication": tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"names": tftypes.List{ElementType: tftypes.String},
							},
						},
					},
				},
			},
		},
	}

	ctx := context.Background()
	modifier := SetUnknownIfAccessHasChanges()

	t.Run("rest API key should not be affected", func(t *testing.T) {
		// Create state and config values for rest API key
		stateValues := map[string]tftypes.Value{
			"type":             tftypes.NewValue(tftypes.String, "rest"),
			"role_descriptors": tftypes.NewValue(tftypes.String, `{"test": "value"}`),
			"access":           tftypes.NewValue(objectType.AttributeTypes["access"], nil),
		}

		configValues := map[string]tftypes.Value{
			"type":             tftypes.NewValue(tftypes.String, "rest"),
			"role_descriptors": tftypes.NewValue(tftypes.String, `{"test": "value"}`),
			"access": tftypes.NewValue(objectType.AttributeTypes["access"], map[string]tftypes.Value{
				"search":      tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"names": tftypes.List{ElementType: tftypes.String}}}}, nil),
				"replication": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"names": tftypes.List{ElementType: tftypes.String}}}}, nil),
			}),
		}

		stateRaw := tftypes.NewValue(objectType, stateValues)
		configRaw := tftypes.NewValue(objectType, configValues)

		state := tfsdk.State{Raw: stateRaw, Schema: testSchema}
		config := tfsdk.Config{Raw: configRaw, Schema: testSchema}

		req := planmodifier.StringRequest{
			Path:        path.Root("role_descriptors"),
			PlanValue:   types.StringValue(`{"test": "value"}`),
			ConfigValue: types.StringValue(`{"test": "value"}`),
			StateValue:  types.StringValue(`{"test": "value"}`),
			Config:      config,
			State:       state,
		}

		resp := &planmodifier.StringResponse{}

		// Call the plan modifier
		modifier.PlanModifyString(ctx, req, resp)

		// Check for errors
		require.False(t, resp.Diagnostics.HasError(), "Plan modifier should not have errors: %v", resp.Diagnostics)

		// For rest type, role_descriptors should not be set to unknown
		assert.False(t, resp.PlanValue.IsUnknown(), "Plan value should not be unknown for rest API key")
	})

	t.Run("cross_cluster with unchanged access should not set unknown", func(t *testing.T) {
		// Create identical access for state and config (no change)
		accessValue := tftypes.NewValue(objectType.AttributeTypes["access"], nil)

		stateValues := map[string]tftypes.Value{
			"type":             tftypes.NewValue(tftypes.String, "cross_cluster"),
			"role_descriptors": tftypes.NewValue(tftypes.String, `{"test": "value"}`),
			"access":           accessValue,
		}

		configValues := map[string]tftypes.Value{
			"type":             tftypes.NewValue(tftypes.String, "cross_cluster"),
			"role_descriptors": tftypes.NewValue(tftypes.String, `{"test": "value"}`),
			"access":           accessValue, // Same as state
		}

		stateRaw := tftypes.NewValue(objectType, stateValues)
		configRaw := tftypes.NewValue(objectType, configValues)

		state := tfsdk.State{Raw: stateRaw, Schema: testSchema}
		config := tfsdk.Config{Raw: configRaw, Schema: testSchema}

		req := planmodifier.StringRequest{
			Path:        path.Root("role_descriptors"),
			PlanValue:   types.StringValue(`{"test": "value"}`),
			ConfigValue: types.StringValue(`{"test": "value"}`),
			StateValue:  types.StringValue(`{"test": "value"}`),
			Config:      config,
			State:       state,
		}

		resp := &planmodifier.StringResponse{}

		// Call the plan modifier
		modifier.PlanModifyString(ctx, req, resp)

		// Check for errors
		require.False(t, resp.Diagnostics.HasError(), "Plan modifier should not have errors: %v", resp.Diagnostics)

		// For unchanged access, role_descriptors should not be set to unknown
		assert.False(t, resp.PlanValue.IsUnknown(), "Plan value should not be unknown when access doesn't change")
	})

	t.Run("cross_cluster with changed access should set unknown", func(t *testing.T) {
		// State has null access
		stateAccessValue := tftypes.NewValue(objectType.AttributeTypes["access"], nil)

		// Config has non-null access with search configuration
		configAccessValue := tftypes.NewValue(objectType.AttributeTypes["access"], map[string]tftypes.Value{
			"search": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"names": tftypes.List{ElementType: tftypes.String}}}}, []tftypes.Value{
				tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"names": tftypes.List{ElementType: tftypes.String}}}, map[string]tftypes.Value{
					"names": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "index-*"),
					}),
				}),
			}),
			"replication": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"names": tftypes.List{ElementType: tftypes.String}}}}, nil),
		})

		stateValues := map[string]tftypes.Value{
			"type":             tftypes.NewValue(tftypes.String, "cross_cluster"),
			"role_descriptors": tftypes.NewValue(tftypes.String, `{"test": "value"}`),
			"access":           stateAccessValue,
		}

		configValues := map[string]tftypes.Value{
			"type":             tftypes.NewValue(tftypes.String, "cross_cluster"),
			"role_descriptors": tftypes.NewValue(tftypes.String, `{"test": "value"}`),
			"access":           configAccessValue, // Different from state
		}

		stateRaw := tftypes.NewValue(objectType, stateValues)
		configRaw := tftypes.NewValue(objectType, configValues)

		state := tfsdk.State{Raw: stateRaw, Schema: testSchema}
		config := tfsdk.Config{Raw: configRaw, Schema: testSchema}

		req := planmodifier.StringRequest{
			Path:        path.Root("role_descriptors"),
			PlanValue:   types.StringValue(`{"test": "value"}`),
			ConfigValue: types.StringValue(`{"test": "value"}`),
			StateValue:  types.StringValue(`{"test": "value"}`),
			Config:      config,
			State:       state,
		}

		resp := &planmodifier.StringResponse{}

		// Call the plan modifier
		modifier.PlanModifyString(ctx, req, resp)

		// Check for errors
		require.False(t, resp.Diagnostics.HasError(), "Plan modifier should not have errors: %v", resp.Diagnostics)

		// For changed access, role_descriptors should be set to unknown
		assert.True(t, resp.PlanValue.IsUnknown(), "Plan value should be unknown when access changes for cross_cluster type")
	})

	t.Run("cross_cluster with different access configurations should set unknown", func(t *testing.T) {
		// State has search configuration
		stateAccessValue := tftypes.NewValue(objectType.AttributeTypes["access"], map[string]tftypes.Value{
			"search": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"names": tftypes.List{ElementType: tftypes.String}}}}, []tftypes.Value{
				tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"names": tftypes.List{ElementType: tftypes.String}}}, map[string]tftypes.Value{
					"names": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "old-index-*"),
					}),
				}),
			}),
			"replication": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"names": tftypes.List{ElementType: tftypes.String}}}}, nil),
		})

		// Config has different search configuration
		configAccessValue := tftypes.NewValue(objectType.AttributeTypes["access"], map[string]tftypes.Value{
			"search": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"names": tftypes.List{ElementType: tftypes.String}}}}, []tftypes.Value{
				tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"names": tftypes.List{ElementType: tftypes.String}}}, map[string]tftypes.Value{
					"names": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "new-index-*"),
					}),
				}),
			}),
			"replication": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"names": tftypes.List{ElementType: tftypes.String}}}}, nil),
		})

		stateValues := map[string]tftypes.Value{
			"type":             tftypes.NewValue(tftypes.String, "cross_cluster"),
			"role_descriptors": tftypes.NewValue(tftypes.String, `{"test": "value"}`),
			"access":           stateAccessValue,
		}

		configValues := map[string]tftypes.Value{
			"type":             tftypes.NewValue(tftypes.String, "cross_cluster"),
			"role_descriptors": tftypes.NewValue(tftypes.String, `{"test": "value"}`),
			"access":           configAccessValue, // Different from state
		}

		stateRaw := tftypes.NewValue(objectType, stateValues)
		configRaw := tftypes.NewValue(objectType, configValues)

		state := tfsdk.State{Raw: stateRaw, Schema: testSchema}
		config := tfsdk.Config{Raw: configRaw, Schema: testSchema}

		req := planmodifier.StringRequest{
			Path:        path.Root("role_descriptors"),
			PlanValue:   types.StringValue(`{"test": "value"}`),
			ConfigValue: types.StringValue(`{"test": "value"}`),
			StateValue:  types.StringValue(`{"test": "value"}`),
			Config:      config,
			State:       state,
		}

		resp := &planmodifier.StringResponse{}

		// Call the plan modifier
		modifier.PlanModifyString(ctx, req, resp)

		// Check for errors
		require.False(t, resp.Diagnostics.HasError(), "Plan modifier should not have errors: %v", resp.Diagnostics)

		// For changed access configuration, role_descriptors should be set to unknown
		assert.True(t, resp.PlanValue.IsUnknown(), "Plan value should be unknown when access configuration changes")
	})

	t.Run("basic functionality tests", func(t *testing.T) {
		// Test that the modifier can be created without errors
		modifier := SetUnknownIfAccessHasChanges()
		assert.NotNil(t, modifier, "Plan modifier should be created successfully")

		// Test the description method
		desc := modifier.Description(ctx)
		assert.NotEmpty(t, desc, "Description should not be empty")

		// Test the markdown description method
		markdownDesc := modifier.MarkdownDescription(ctx)
		assert.NotEmpty(t, markdownDesc, "Markdown description should not be empty")
	})
}
