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

package alertingrule

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testActionsList(ctx context.Context, t *testing.T, withFrequency bool) types.List {
	t.Helper()

	var freq types.Object
	if withFrequency {
		fm := frequencyModel{
			Summary:    types.BoolValue(true),
			NotifyWhen: types.StringValue("onActionGroupChange"),
			Throttle:   types.StringValue("10m"),
		}
		o, d := types.ObjectValueFrom(ctx, getFrequencyAttrTypes(), fm)
		require.Empty(t, d)
		freq = o
	} else {
		freq = types.ObjectNull(getFrequencyAttrTypes())
	}

	am := actionModel{
		Group:        types.StringValue("default"),
		ID:           types.StringValue("connector-id"),
		Params:       jsontypes.NewNormalizedValue(`{}`),
		Frequency:    freq,
		AlertsFilter: types.ObjectNull(getAlertsFilterAttrTypes()),
	}
	actionObj, d := types.ObjectValueFrom(ctx, getActionsAttrTypes(), am)
	require.Empty(t, d)

	list, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: getActionsAttrTypes()}, []attr.Value{actionObj})
	require.Empty(t, d)
	return list
}

func TestValidateNotifyWhenThrottleFrequencyExclusivity_notifyWhenAndFrequency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var diags diag.Diagnostics
	data := alertingRuleModel{
		NotifyWhen: types.StringValue("onActiveAlert"),
		Actions:    testActionsList(ctx, t, true),
	}
	validateNotifyWhenThrottleFrequencyExclusivity(ctx, &data, &diags)
	assert.True(t, diags.HasError())
}

func TestValidateNotifyWhenThrottleFrequencyExclusivity_throttleAndFrequency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var diags diag.Diagnostics
	data := alertingRuleModel{
		Throttle: types.StringValue("5m"),
		Actions:  testActionsList(ctx, t, true),
	}
	validateNotifyWhenThrottleFrequencyExclusivity(ctx, &data, &diags)
	assert.True(t, diags.HasError())
}

func TestValidateNotifyWhenThrottleFrequencyExclusivity_frequencyOnly(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var diags diag.Diagnostics
	data := alertingRuleModel{
		NotifyWhen: types.StringNull(),
		Throttle:   types.StringNull(),
		Actions:    testActionsList(ctx, t, true),
	}
	validateNotifyWhenThrottleFrequencyExclusivity(ctx, &data, &diags)
	assert.False(t, diags.HasError())
}

func TestValidateNotifyWhenThrottleFrequencyExclusivity_ruleLevelOnly(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var diags diag.Diagnostics
	data := alertingRuleModel{
		NotifyWhen: types.StringValue("onActiveAlert"),
		Actions:    testActionsList(ctx, t, false),
	}
	validateNotifyWhenThrottleFrequencyExclusivity(ctx, &data, &diags)
	assert.False(t, diags.HasError())
}

func TestValidateNotifyWhenThrottleFrequencyExclusivity_noFalsePositiveWhenFrequencyAbsent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var diags diag.Diagnostics
	data := alertingRuleModel{
		NotifyWhen: types.StringValue("onActiveAlert"),
		Throttle:   types.StringValue("5m"),
		Actions:    testActionsList(ctx, t, false),
	}
	validateNotifyWhenThrottleFrequencyExclusivity(ctx, &data, &diags)
	assert.False(t, diags.HasError())
}

func TestValidateNotifyWhenThrottleFrequencyExclusivity_prefersNotifyWhenDiagnostic(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var diags diag.Diagnostics
	data := alertingRuleModel{
		NotifyWhen: types.StringValue("onActiveAlert"),
		Throttle:   types.StringValue("5m"),
		Actions:    testActionsList(ctx, t, true),
	}
	validateNotifyWhenThrottleFrequencyExclusivity(ctx, &data, &diags)
	require.True(t, diags.HasError())
	if assert.Len(t, diags, 1) {
		assert.Contains(t, diags[0].Summary(), "notify_when")
	}
}

func TestValidateNotifyWhenThrottleFrequencyExclusivity_emptyRuleNotifyWhenIgnored(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var diags diag.Diagnostics
	data := alertingRuleModel{
		NotifyWhen: types.StringValue("   "),
		Throttle:   types.StringNull(),
		Actions:    testActionsList(ctx, t, true),
	}
	validateNotifyWhenThrottleFrequencyExclusivity(ctx, &data, &diags)
	assert.False(t, diags.HasError())
}

func TestPlanNotifyWhenForActionFrequency_unknownWithFrequencyBecomesNull(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var diags diag.Diagnostics
	actions := testActionsList(ctx, t, true)
	out := planNotifyWhenForActionFrequency(ctx, types.StringUnknown(), actions, &diags)
	assert.False(t, diags.HasError())
	assert.True(t, out.IsNull())
}

func TestPlanNotifyWhenForActionFrequency_knownUnchanged(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var diags diag.Diagnostics
	actions := testActionsList(ctx, t, true)
	out := planNotifyWhenForActionFrequency(ctx, types.StringValue("onActiveAlert"), actions, &diags)
	assert.False(t, diags.HasError())
	assert.Equal(t, "onActiveAlert", out.ValueString())
}

func TestPlanNotifyWhenForActionFrequency_unknownWithoutFrequencyUnchanged(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var diags diag.Diagnostics
	actions := testActionsList(ctx, t, false)
	out := planNotifyWhenForActionFrequency(ctx, types.StringUnknown(), actions, &diags)
	assert.False(t, diags.HasError())
	assert.True(t, out.IsUnknown())
}

func TestConfigActionsIncludeKnownFrequencyBlock(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var diags diag.Diagnostics
	assert.True(t, configActionsIncludeKnownFrequencyBlock(ctx, testActionsList(ctx, t, true), &diags))
	assert.False(t, diags.HasError())

	diags = nil
	assert.False(t, configActionsIncludeKnownFrequencyBlock(ctx, testActionsList(ctx, t, false), &diags))
	assert.False(t, diags.HasError())

	diags = nil
	assert.False(t, configActionsIncludeKnownFrequencyBlock(ctx, types.ListUnknown(types.ObjectType{AttrTypes: getActionsAttrTypes()}), &diags))
}
