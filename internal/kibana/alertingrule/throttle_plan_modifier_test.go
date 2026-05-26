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

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestPlanThrottleForActionFrequency_introducingFrequencyResetsToUnknown(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var diags diag.Diagnostics
	stateActions := testActionsList(ctx, t, false) // prior state had no frequency
	configActions := testActionsList(ctx, t, true) // new config introduces frequency
	out := planThrottleForActionFrequency(ctx, types.StringValue("5m"), stateActions, configActions, &diags)
	assert.False(t, diags.HasError())
	assert.True(t, out.IsUnknown(), "USFU-restored throttle should be reset to unknown when actions.frequency is newly introduced")
}

func TestPlanThrottleForActionFrequency_alreadyInFrequencyModeIsIdempotent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var diags diag.Diagnostics
	stateActions := testActionsList(ctx, t, true) // prior state already has frequency
	configActions := testActionsList(ctx, t, true)
	out := planThrottleForActionFrequency(ctx, types.StringValue("5m"), stateActions, configActions, &diags)
	assert.False(t, diags.HasError())
	assert.Equal(t, "5m", out.ValueString(), "throttle plan should remain at the USFU-restored value once the rule is already in frequency mode")
}

func TestPlanThrottleForActionFrequency_unknownWithFrequencyIntroductionStaysUnknown(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var diags diag.Diagnostics
	stateActions := testActionsList(ctx, t, false)
	configActions := testActionsList(ctx, t, true)
	out := planThrottleForActionFrequency(ctx, types.StringUnknown(), stateActions, configActions, &diags)
	assert.False(t, diags.HasError())
	assert.True(t, out.IsUnknown())
}

func TestPlanThrottleForActionFrequency_knownWithoutConfigFrequencyUnchanged(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var diags diag.Diagnostics
	stateActions := testActionsList(ctx, t, false)
	configActions := testActionsList(ctx, t, false)
	out := planThrottleForActionFrequency(ctx, types.StringValue("5m"), stateActions, configActions, &diags)
	assert.False(t, diags.HasError())
	assert.Equal(t, "5m", out.ValueString(), "throttle should be preserved when actions.frequency is absent from config")
}

func TestPlanThrottleForActionFrequency_nullWithoutConfigFrequencyUnchanged(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var diags diag.Diagnostics
	stateActions := testActionsList(ctx, t, false)
	configActions := testActionsList(ctx, t, false)
	out := planThrottleForActionFrequency(ctx, types.StringNull(), stateActions, configActions, &diags)
	assert.False(t, diags.HasError())
	assert.True(t, out.IsNull())
}

func TestPlanThrottleForActionFrequency_nullConfigActionsLeavesPlanUnchanged(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var diags diag.Diagnostics
	out := planThrottleForActionFrequency(
		ctx,
		types.StringValue("5m"),
		types.ListNull(actionsListElementType()),
		types.ListNull(actionsListElementType()),
		&diags,
	)
	assert.False(t, diags.HasError())
	assert.Equal(t, "5m", out.ValueString())
}

func TestPlanThrottleForActionFrequency_unknownConfigActionsLeavesPlanUnchanged(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var diags diag.Diagnostics
	out := planThrottleForActionFrequency(
		ctx,
		types.StringValue("5m"),
		types.ListNull(actionsListElementType()),
		types.ListUnknown(actionsListElementType()),
		&diags,
	)
	assert.False(t, diags.HasError())
	assert.Equal(t, "5m", out.ValueString())
}

func TestSetUnknownIfActionsFrequencyConfigured_descriptions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	m := SetUnknownIfActionsFrequencyConfigured()
	assert.NotEmpty(t, m.Description(ctx))
	assert.NotEmpty(t, m.MarkdownDescription(ctx))
}

// actionsListElementType returns the element type used for the `actions` list,
// matching the alerting rule schema. Kept separate from testActionsList so
// tests can build empty/null lists without round-tripping action objects.
func actionsListElementType() types.ObjectType {
	return types.ObjectType{AttrTypes: getActionsAttrTypes()}
}
