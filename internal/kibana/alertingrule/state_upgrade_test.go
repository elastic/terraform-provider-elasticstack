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
	"encoding/json"
	"sync"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/require"
)

var (
	testSchemaOnce sync.Once
	testSchema     rschema.Schema
)

func testResourceSchema(t *testing.T) rschema.Schema {
	t.Helper()
	testSchemaOnce.Do(func() {
		ctx := context.Background()
		var resp resource.SchemaResponse
		newResource().Schema(ctx, resource.SchemaRequest{}, &resp)
		require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
		testSchema = resp.Schema
	})
	return testSchema
}

func runMigrateV0ToV1(t *testing.T, raw map[string]any) map[string]any {
	t.Helper()

	resp := runMigrateV0ToV1Resp(t, raw)
	require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	require.NotNil(t, resp.DynamicValue)

	var got map[string]any
	require.NoError(t, json.Unmarshal(resp.DynamicValue.JSON, &got))
	return got
}

func runMigrateV0ToV1Resp(t *testing.T, raw map[string]any) *resource.UpgradeStateResponse {
	t.Helper()

	rawJSON, err := json.Marshal(raw)
	require.NoError(t, err)

	req := resource.UpgradeStateRequest{RawState: &tfprotov6.RawState{JSON: rawJSON}}
	resp := &resource.UpgradeStateResponse{}
	migrateV0ToV1(context.Background(), req, resp)
	return resp
}

func requireUpgradedStateDecodes(t *testing.T, resp *resource.UpgradeStateResponse) {
	t.Helper()
	require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	require.NotNil(t, resp.DynamicValue)
	require.NotNil(t, resp.DynamicValue.JSON)

	ctx := context.Background()
	sch := testResourceSchema(t)
	tfTyp := sch.Type().TerraformType(ctx)
	raw, err := resp.DynamicValue.Unmarshal(tfTyp)
	require.NoError(t, err, "upgraded state must decode under the v1 resource schema's Terraform type")

	state := tfsdk.State{Schema: sch, Raw: raw}
	var model alertingRuleModel
	diags := state.Get(ctx, &model)
	for _, d := range diags.Errors() {
		t.Logf("decode diagnostic: %s: %s", d.Summary(), d.Detail())
	}
	require.False(t, diags.HasError(), "upgraded state must decode into the resource's Go model")
}

func baseAlertingRuleState() map[string]any {
	return map[string]any{
		"id":           clients.DefaultSpaceID + "/rule-id",
		"space_id":     clients.DefaultSpaceID,
		"rule_id":      "rule-id",
		"name":         "rule-name",
		"consumer":     "alerts",
		"rule_type_id": ".index-threshold",
		"interval":     "1m",
		"enabled":      true,
	}
}

func TestMigrateV0ToV1_ParamsNullification(t *testing.T) {
	t.Parallel()

	validParams := `{"test":"value"}`

	cases := []struct {
		name              string
		ruleParamsSet     bool
		ruleParams        any
		actionParams      []any
		wantRuleParams    any
		wantActionParams  []any
		wantDecodeUnderV1 bool
	}{
		{
			name:              "rule params empty string becomes null",
			ruleParamsSet:     true,
			ruleParams:        "",
			wantRuleParams:    nil,
			wantDecodeUnderV1: true,
		},
		{
			name:              "rule params valid JSON unchanged",
			ruleParamsSet:     true,
			ruleParams:        validParams,
			wantRuleParams:    validParams,
			wantDecodeUnderV1: true,
		},
		{
			name:              "rule params null unchanged",
			ruleParamsSet:     true,
			ruleParams:        nil,
			wantRuleParams:    nil,
			wantDecodeUnderV1: true,
		},
		{
			name:              "rule params absent unchanged",
			ruleParamsSet:     false,
			wantDecodeUnderV1: false,
		},
		{
			name:              "action params empty string becomes null",
			actionParams:      []any{""},
			wantActionParams:  []any{nil},
			wantDecodeUnderV1: true,
		},
		{
			name:              "action params valid JSON unchanged",
			actionParams:      []any{validParams},
			wantActionParams:  []any{validParams},
			wantDecodeUnderV1: true,
		},
		{
			name:              "action params null unchanged",
			actionParams:      []any{nil},
			wantActionParams:  []any{nil},
			wantDecodeUnderV1: true,
		},
		{
			name:              "both rule and action params empty strings become null",
			ruleParamsSet:     true,
			ruleParams:        "",
			actionParams:      []any{""},
			wantRuleParams:    nil,
			wantActionParams:  []any{nil},
			wantDecodeUnderV1: true,
		},
		{
			name:              "mixed action params preserves valid JSON and nullifies empty",
			actionParams:      []any{"", validParams, nil},
			wantActionParams:  []any{nil, validParams, nil},
			wantDecodeUnderV1: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			raw := baseAlertingRuleState()
			if tc.ruleParamsSet {
				raw[attrParams] = tc.ruleParams
			}
			if tc.actionParams != nil {
				actions := make([]any, 0, len(tc.actionParams))
				for _, params := range tc.actionParams {
					action := map[string]any{
						"id":              "action-id",
						blockFrequency:    []any{},
						blockAlertsFilter: []any{},
						attrParams:        params,
					}
					actions = append(actions, action)
				}
				raw["actions"] = actions
			}

			resp := runMigrateV0ToV1Resp(t, raw)
			require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)

			var got map[string]any
			require.NoError(t, json.Unmarshal(resp.DynamicValue.JSON, &got))

			if tc.ruleParamsSet {
				require.Equal(t, tc.wantRuleParams, got[attrParams], "rule-level params normalization mismatch")
			} else {
				_, ok := got[attrParams]
				require.False(t, ok, "absent rule params should remain absent")
			}

			if tc.actionParams != nil {
				actions, ok := got["actions"].([]any)
				require.True(t, ok, "actions should remain a list")
				require.Len(t, actions, len(tc.wantActionParams))
				for i, actionAny := range actions {
					action, ok := actionAny.(map[string]any)
					require.True(t, ok, "action %d should be a map", i)
					require.Equal(t, tc.wantActionParams[i], action[attrParams], "action %d params normalization mismatch", i)
				}
			}

			if tc.wantDecodeUnderV1 {
				requireUpgradedStateDecodes(t, resp)
			}
		})
	}
}

func TestMigrateV0ToV1_PreservesExistingTransformations(t *testing.T) {
	t.Parallel()

	raw := baseAlertingRuleState()
	raw[attrParams] = ""
	raw[attrNotifyWhen] = ""
	raw[attrThrottle] = ""
	raw["actions"] = []any{
		map[string]any{
			"id":       "action-1",
			attrParams: `{"foo":"bar"}`,
			blockFrequency: []any{
				map[string]any{
					"summary":      true,
					attrNotifyWhen: "onActiveAlert",
					attrThrottle:   "10m",
				},
			},
			blockAlertsFilter: []any{},
		},
		map[string]any{
			"id":              "action-2",
			attrParams:        "",
			blockFrequency:    []any{},
			blockAlertsFilter: []any{},
		},
	}

	got := runMigrateV0ToV1(t, raw)

	require.Nil(t, got[attrParams], "rule-level params should be nullified")
	require.Nil(t, got[attrNotifyWhen], "notify_when should be nullified")
	require.Nil(t, got[attrThrottle], "throttle should be nullified")

	actions, ok := got["actions"].([]any)
	require.True(t, ok)
	require.Len(t, actions, 2)

	action1, ok := actions[0].(map[string]any)
	require.True(t, ok)
	action1Params, ok := action1[attrParams].(string)
	require.True(t, ok)
	require.JSONEq(t, `{"foo":"bar"}`, action1Params, "non-empty action params should be preserved")
	freq1, ok := action1[blockFrequency].(map[string]any)
	require.True(t, ok)
	require.Equal(t, true, freq1["summary"])
	require.Equal(t, "onActiveAlert", freq1[attrNotifyWhen])
	require.Equal(t, "10m", freq1[attrThrottle])
	require.Nil(t, action1[blockAlertsFilter], "empty alerts_filter list should collapse to null")

	action2, ok := actions[1].(map[string]any)
	require.True(t, ok)
	require.Nil(t, action2[attrParams], "empty action params should be nullified")
	require.Nil(t, action2[blockFrequency], "empty frequency list should collapse to null")
	require.Nil(t, action2[blockAlertsFilter], "empty alerts_filter list should collapse to null")
}

func TestUpgradeState_RegistersV0Upgrader(t *testing.T) {
	t.Parallel()

	r := &Resource{}
	upgraders := r.UpgradeState(context.Background())
	up, ok := upgraders[0]
	require.True(t, ok, "expected a registered v0 state upgrader")
	require.NotNil(t, up.StateUpgrader)
}
