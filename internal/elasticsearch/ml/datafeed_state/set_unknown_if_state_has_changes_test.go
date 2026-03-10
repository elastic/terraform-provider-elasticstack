// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package datafeedstate

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

var testSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"state": schema.StringAttribute{
			Required: true,
		},
		"start": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
	},
}

var testObjectType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"state": tftypes.String,
		"start": tftypes.String,
	},
}

func TestSetUnknownIfStateHasChanges_Create_Stopped(t *testing.T) {
	ctx := context.Background()
	modifier := SetUnknownIfStateHasChanges()

	req := planmodifier.StringRequest{
		Path:        path.Root("start"),
		ConfigValue: types.StringNull(),
		PlanValue:   types.StringUnknown(),
		StateValue:  types.StringNull(),
		State: tfsdk.State{
			Raw:    tftypes.NewValue(testObjectType, nil),
			Schema: testSchema,
		},
		Config: tfsdk.Config{
			Raw: tftypes.NewValue(testObjectType, map[string]tftypes.Value{
				"state": tftypes.NewValue(tftypes.String, "stopped"),
				"start": tftypes.NewValue(tftypes.String, nil),
			}),
			Schema: testSchema,
		},
	}

	resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}
	modifier.PlanModifyString(ctx, req, resp)

	require.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	assert.True(t, resp.PlanValue.IsNull(),
		"expected null plan value for create with stopped state, got unknown=%v null=%v value=%q",
		resp.PlanValue.IsUnknown(), resp.PlanValue.IsNull(), resp.PlanValue.ValueString())
}

func TestSetUnknownIfStateHasChanges_Create_Started(t *testing.T) {
	ctx := context.Background()
	modifier := SetUnknownIfStateHasChanges()

	req := planmodifier.StringRequest{
		Path:        path.Root("start"),
		ConfigValue: types.StringNull(),
		PlanValue:   types.StringUnknown(),
		StateValue:  types.StringNull(),
		State: tfsdk.State{
			Raw:    tftypes.NewValue(testObjectType, nil),
			Schema: testSchema,
		},
		Config: tfsdk.Config{
			Raw: tftypes.NewValue(testObjectType, map[string]tftypes.Value{
				"state": tftypes.NewValue(tftypes.String, "started"),
				"start": tftypes.NewValue(tftypes.String, nil),
			}),
			Schema: testSchema,
		},
	}

	resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}
	modifier.PlanModifyString(ctx, req, resp)

	require.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	assert.True(t, resp.PlanValue.IsUnknown(),
		"expected unknown plan value for create with started state (start will be computed)")
}

func TestSetUnknownIfStateHasChanges_Update_StoppedToStarted(t *testing.T) {
	ctx := context.Background()
	modifier := SetUnknownIfStateHasChanges()

	req := planmodifier.StringRequest{
		Path:        path.Root("start"),
		ConfigValue: types.StringNull(),
		PlanValue:   types.StringNull(),
		StateValue:  types.StringNull(),
		State: tfsdk.State{
			Raw: tftypes.NewValue(testObjectType, map[string]tftypes.Value{
				"state": tftypes.NewValue(tftypes.String, "stopped"),
				"start": tftypes.NewValue(tftypes.String, nil),
			}),
			Schema: testSchema,
		},
		Config: tfsdk.Config{
			Raw: tftypes.NewValue(testObjectType, map[string]tftypes.Value{
				"state": tftypes.NewValue(tftypes.String, "started"),
				"start": tftypes.NewValue(tftypes.String, nil),
			}),
			Schema: testSchema,
		},
	}

	resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}
	modifier.PlanModifyString(ctx, req, resp)

	require.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	assert.True(t, resp.PlanValue.IsUnknown(),
		"expected unknown plan value when state changes from stopped to started, got unknown=%v null=%v",
		resp.PlanValue.IsUnknown(), resp.PlanValue.IsNull())
}

func TestSetUnknownIfStateHasChanges_Update_StartedToStopped(t *testing.T) {
	ctx := context.Background()
	modifier := SetUnknownIfStateHasChanges()

	req := planmodifier.StringRequest{
		Path:        path.Root("start"),
		ConfigValue: types.StringNull(),
		PlanValue:   types.StringValue("2025-01-01T00:00:00Z"),
		StateValue:  types.StringValue("2025-01-01T00:00:00Z"),
		State: tfsdk.State{
			Raw: tftypes.NewValue(testObjectType, map[string]tftypes.Value{
				"state": tftypes.NewValue(tftypes.String, "started"),
				"start": tftypes.NewValue(tftypes.String, "2025-01-01T00:00:00Z"),
			}),
			Schema: testSchema,
		},
		Config: tfsdk.Config{
			Raw: tftypes.NewValue(testObjectType, map[string]tftypes.Value{
				"state": tftypes.NewValue(tftypes.String, "stopped"),
				"start": tftypes.NewValue(tftypes.String, nil),
			}),
			Schema: testSchema,
		},
	}

	resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}
	modifier.PlanModifyString(ctx, req, resp)

	require.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	assert.True(t, resp.PlanValue.IsUnknown(),
		"expected unknown plan value when state changes from started to stopped")
}

func TestSetUnknownIfStateHasChanges_Update_StateUnchanged(t *testing.T) {
	ctx := context.Background()
	modifier := SetUnknownIfStateHasChanges()

	req := planmodifier.StringRequest{
		Path:        path.Root("start"),
		ConfigValue: types.StringNull(),
		PlanValue:   types.StringNull(),
		StateValue:  types.StringNull(),
		State: tfsdk.State{
			Raw: tftypes.NewValue(testObjectType, map[string]tftypes.Value{
				"state": tftypes.NewValue(tftypes.String, "stopped"),
				"start": tftypes.NewValue(tftypes.String, nil),
			}),
			Schema: testSchema,
		},
		Config: tfsdk.Config{
			Raw: tftypes.NewValue(testObjectType, map[string]tftypes.Value{
				"state": tftypes.NewValue(tftypes.String, "stopped"),
				"start": tftypes.NewValue(tftypes.String, nil),
			}),
			Schema: testSchema,
		},
	}

	resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}
	modifier.PlanModifyString(ctx, req, resp)

	require.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	assert.True(t, resp.PlanValue.IsNull(),
		"expected null plan value when state is unchanged (stopped)")
}

func TestSetUnknownIfStateHasChanges_Update_ExplicitStartConfig(t *testing.T) {
	ctx := context.Background()
	modifier := SetUnknownIfStateHasChanges()

	explicitStart := types.StringValue("2025-06-01T00:00:00Z")

	req := planmodifier.StringRequest{
		Path:        path.Root("start"),
		ConfigValue: explicitStart,
		PlanValue:   explicitStart,
		StateValue:  types.StringNull(),
		State: tfsdk.State{
			Raw: tftypes.NewValue(testObjectType, map[string]tftypes.Value{
				"state": tftypes.NewValue(tftypes.String, "stopped"),
				"start": tftypes.NewValue(tftypes.String, nil),
			}),
			Schema: testSchema,
		},
		Config: tfsdk.Config{
			Raw: tftypes.NewValue(testObjectType, map[string]tftypes.Value{
				"state": tftypes.NewValue(tftypes.String, "started"),
				"start": tftypes.NewValue(tftypes.String, "2025-06-01T00:00:00Z"),
			}),
			Schema: testSchema,
		},
	}

	resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}
	modifier.PlanModifyString(ctx, req, resp)

	require.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	assert.Equal(t, "2025-06-01T00:00:00Z", resp.PlanValue.ValueString(),
		"modifier should not touch explicitly configured start value")
}
