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

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func Test_alertingRuleModel_toAPIModel_flappingVersionGate(t *testing.T) {
	ctx := context.Background()

	flObj, diags := types.ObjectValueFrom(ctx, getFlappingAttrTypes(), flappingModel{
		LookBackWindow:        types.Int64Value(5),
		StatusChangeThreshold: types.Int64Value(2),
		Enabled:               types.BoolNull(),
	})
	require.False(t, diags.HasError())

	m := alertingRuleModel{
		ID:         types.StringValue("default/r1"),
		RuleID:     types.StringValue("r1"),
		SpaceID:    types.StringValue("default"),
		Name:       types.StringValue("n"),
		Consumer:   types.StringValue("alerts"),
		RuleTypeID: types.StringValue(".index-threshold"),
		Interval:   types.StringValue("1m"),
		Params: jsontypes.NewNormalizedValue(
			`{"index":["i"],"threshold":[1],"thresholdComparator":">","timeField":"@timestamp","timeWindowSize":1,"timeWindowUnit":"m"}`,
		),
		NotifyWhen: types.StringValue("onActionGroupChange"),
		Flapping:   flObj,
	}

	_, convDiags := m.toAPIModel(ctx, version.Must(version.NewVersion("8.15.0")))
	require.True(t, convDiags.HasError())
}

func Test_alertingRuleModel_toAPIModel_flappingAllowedAt816(t *testing.T) {
	ctx := context.Background()

	flObj, diags := types.ObjectValueFrom(ctx, getFlappingAttrTypes(), flappingModel{
		LookBackWindow:        types.Int64Value(5),
		StatusChangeThreshold: types.Int64Value(2),
		Enabled:               types.BoolValue(true),
	})
	require.False(t, diags.HasError())

	m := alertingRuleModel{
		ID:         types.StringValue("default/r1"),
		RuleID:     types.StringValue("r1"),
		SpaceID:    types.StringValue("default"),
		Name:       types.StringValue("n"),
		Consumer:   types.StringValue("alerts"),
		RuleTypeID: types.StringValue(".index-threshold"),
		Interval:   types.StringValue("1m"),
		Params: jsontypes.NewNormalizedValue(
			`{"index":["i"],"threshold":[1],"thresholdComparator":">","timeField":"@timestamp","timeWindowSize":1,"timeWindowUnit":"m"}`,
		),
		NotifyWhen: types.StringValue("onActionGroupChange"),
		Flapping:   flObj,
	}

	rule, convDiags := m.toAPIModel(ctx, version.Must(version.NewVersion("8.16.0")))
	require.False(t, convDiags.HasError())
	require.NotNil(t, rule.Flapping)
	require.Equal(t, int64(5), rule.Flapping.LookBackWindow)
	require.Equal(t, int64(2), rule.Flapping.StatusChangeThreshold)
	require.NotNil(t, rule.Flapping.Enabled)
	require.True(t, *rule.Flapping.Enabled)
}
