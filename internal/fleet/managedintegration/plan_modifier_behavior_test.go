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

package managedintegration

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

// RequiresReplace modifiers only run on update; they bail out when State or Plan Raw is null.
var planModifierUpdateObjectType = tftypes.Object{AttributeTypes: map[string]tftypes.Type{"_": tftypes.String}}

func nonNullUpdatePlanModifierState(t *testing.T) (tfsdk.State, tfsdk.Plan) {
	t.Helper()
	return tfsdk.State{Raw: tftypes.NewValue(planModifierUpdateObjectType, map[string]tftypes.Value{
			"_": tftypes.NewValue(tftypes.String, "prior"),
		})},
		tfsdk.Plan{Raw: tftypes.NewValue(planModifierUpdateObjectType, map[string]tftypes.Value{
			"_": tftypes.NewValue(tftypes.String, "next"),
		})}
}

// Exercises configured RequiresReplace plan modifiers with differing state/plan
// values (schema wiring is covered separately in schema_test.go).
func TestRequiresReplacePlanModifiers_managedIntegration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := getSchema(ctx)

	t.Run("policy_id", func(t *testing.T) {
		t.Parallel()
		attr := s.Attributes["policy_id"].(schema.StringAttribute)
		requireRequiresReplaceOnStringChange(t, attr.PlanModifiers, "policy-a", "policy-b")
	})

	t.Run("namespace", func(t *testing.T) {
		t.Parallel()
		attr := s.Attributes["namespace"].(schema.StringAttribute)
		requireRequiresReplaceOnStringChange(t, attr.PlanModifiers, "ns-a", "ns-b")
	})

	t.Run("policy_template", func(t *testing.T) {
		t.Parallel()
		attr := s.Attributes[attrPolicyTemplate].(schema.StringAttribute)
		requireRequiresReplaceOnStringChange(t, attr.PlanModifiers, "cspm", "kspm")
	})

	t.Run("package.name", func(t *testing.T) {
		t.Parallel()
		pkg := s.Attributes[attrPackage].(schema.SingleNestedAttribute)
		name := pkg.Attributes["name"].(schema.StringAttribute)
		requireRequiresReplaceOnStringChange(t, name.PlanModifiers, "pkg-a", "pkg-b")
	})

	t.Run("space_ids", func(t *testing.T) {
		t.Parallel()
		attr := s.Attributes["space_ids"].(schema.SetAttribute)
		stateSet, diags := types.SetValueFrom(ctx, types.StringType, []string{"default"})
		require.False(t, diags.HasError())
		planSet, diags := types.SetValueFrom(ctx, types.StringType, []string{"other-space"})
		require.False(t, diags.HasError())
		priorState, proposedPlan := nonNullUpdatePlanModifierState(t)
		req := planmodifier.SetRequest{
			State: priorState, Plan: proposedPlan,
			StateValue: stateSet, PlanValue: planSet, ConfigValue: planSet,
		}
		resp := &planmodifier.SetResponse{PlanValue: planSet}
		for _, m := range attr.PlanModifiers {
			m.PlanModifySet(ctx, req, resp)
		}
		require.False(t, resp.Diagnostics.HasError())
		require.True(t, resp.RequiresReplace)
	})

	t.Run("cloud_connector object", func(t *testing.T) {
		t.Parallel()
		ccAttr := s.Attributes["cloud_connector"].(schema.SingleNestedAttribute)
		stateObj, diags := types.ObjectValueFrom(ctx, cloudConnectorAttrTypes(), cloudConnectorModel{
			Enabled:          types.BoolValue(true),
			CloudConnectorID: types.StringValue("cc-1"),
			Name:             types.StringValue("name-a"),
			TargetCSP:        types.StringValue("aws"),
		})
		require.False(t, diags.HasError())
		planObj, diags := types.ObjectValueFrom(ctx, cloudConnectorAttrTypes(), cloudConnectorModel{
			Enabled:          types.BoolValue(true),
			CloudConnectorID: types.StringValue("cc-1"),
			Name:             types.StringValue("name-b"),
			TargetCSP:        types.StringValue("aws"),
		})
		require.False(t, diags.HasError())
		priorState, proposedPlan := nonNullUpdatePlanModifierState(t)
		req := planmodifier.ObjectRequest{
			State: priorState, Plan: proposedPlan,
			StateValue: stateObj, PlanValue: planObj, ConfigValue: planObj,
		}
		resp := &planmodifier.ObjectResponse{PlanValue: planObj}
		for _, m := range ccAttr.PlanModifiers {
			m.PlanModifyObject(ctx, req, resp)
		}
		require.False(t, resp.Diagnostics.HasError())
		require.True(t, resp.RequiresReplace)
	})

	t.Run("cloud_connector object to null", func(t *testing.T) {
		t.Parallel()
		ccAttr := s.Attributes["cloud_connector"].(schema.SingleNestedAttribute)
		stateObj, diags := types.ObjectValueFrom(ctx, cloudConnectorAttrTypes(), cloudConnectorModel{
			Enabled:          types.BoolValue(true),
			CloudConnectorID: types.StringValue("cc-1"),
			Name:             types.StringValue("name-a"),
			TargetCSP:        types.StringValue("aws"),
		})
		require.False(t, diags.HasError())
		planObj := types.ObjectNull(cloudConnectorAttrTypes())
		priorState, proposedPlan := nonNullUpdatePlanModifierState(t)
		req := planmodifier.ObjectRequest{
			State: priorState, Plan: proposedPlan,
			StateValue: stateObj, PlanValue: planObj, ConfigValue: planObj,
		}
		resp := &planmodifier.ObjectResponse{PlanValue: planObj}
		for _, m := range ccAttr.PlanModifiers {
			m.PlanModifyObject(ctx, req, resp)
		}
		require.False(t, resp.Diagnostics.HasError())
		require.True(t, resp.RequiresReplace)
	})

	t.Run("name unchanged does not require replace", func(t *testing.T) {
		t.Parallel()
		attr := s.Attributes["name"].(schema.StringAttribute)
		requireNoRequiresReplaceOnStringSame(t, attr.PlanModifiers, "same", "same")
	})
}

func requireRequiresReplaceOnStringChange(t *testing.T, mods []planmodifier.String, state, plan string) {
	t.Helper()
	ctx := context.Background()
	priorState, proposedPlan := nonNullUpdatePlanModifierState(t)
	stateVal := types.StringValue(state)
	planVal := types.StringValue(plan)
	req := planmodifier.StringRequest{
		State:       priorState,
		Plan:        proposedPlan,
		StateValue:  stateVal,
		PlanValue:   planVal,
		ConfigValue: planVal,
	}
	resp := &planmodifier.StringResponse{PlanValue: planVal}
	for _, m := range mods {
		m.PlanModifyString(ctx, req, resp)
	}
	require.False(t, resp.Diagnostics.HasError())
	require.True(t, resp.RequiresReplace, "expected RequiresReplace when value changes from %q to %q", state, plan)
}

func requireNoRequiresReplaceOnStringSame(t *testing.T, mods []planmodifier.String, state, plan string) {
	t.Helper()
	ctx := context.Background()
	priorState, proposedPlan := nonNullUpdatePlanModifierState(t)
	req := planmodifier.StringRequest{
		State:       priorState,
		Plan:        proposedPlan,
		StateValue:  types.StringValue(state),
		PlanValue:   types.StringValue(plan),
		ConfigValue: types.StringValue(plan),
	}
	resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}
	for _, m := range mods {
		m.PlanModifyString(ctx, req, resp)
	}
	require.False(t, resp.Diagnostics.HasError())
	require.False(t, resp.RequiresReplace)
}
