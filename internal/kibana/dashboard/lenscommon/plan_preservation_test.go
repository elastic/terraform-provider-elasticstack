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

package lenscommon

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestPreserveKnownTfValueIfStateNull_Bool(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		plan       types.Bool
		state      types.Bool
		wantState  types.Bool
		wantChange bool
	}{
		{"known plan, null state -> copy", types.BoolValue(true), types.BoolNull(), types.BoolValue(true), true},
		{"known plan, unknown state -> copy", types.BoolValue(false), types.BoolUnknown(), types.BoolValue(false), true},
		{"known plan, known state -> keep state", types.BoolValue(true), types.BoolValue(false), types.BoolValue(false), false},
		{"null plan, null state -> keep state", types.BoolNull(), types.BoolNull(), types.BoolNull(), false},
		{"unknown plan, null state -> keep state", types.BoolUnknown(), types.BoolNull(), types.BoolNull(), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			state := tc.state
			PreserveKnownTfValueIfStateNull(tc.plan, &state)
			assert.Equal(t, tc.wantState, state)
		})
	}
}

func TestPreserveKnownTfValueIfStateNull_Float64(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		plan      types.Float64
		state     types.Float64
		wantState types.Float64
	}{
		{"known plan, null state -> copy", types.Float64Value(1.5), types.Float64Null(), types.Float64Value(1.5)},
		{"known plan, unknown state -> copy", types.Float64Value(2.5), types.Float64Unknown(), types.Float64Value(2.5)},
		{"known plan, known state -> keep state", types.Float64Value(1.5), types.Float64Value(9.9), types.Float64Value(9.9)},
		{"null plan, null state -> keep state", types.Float64Null(), types.Float64Null(), types.Float64Null()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			state := tc.state
			PreserveKnownTfValueIfStateNull(tc.plan, &state)
			assert.Equal(t, tc.wantState, state)
		})
	}
}

func TestPreserveKnownTfValueIfStateNull_String(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		plan      types.String
		state     types.String
		wantState types.String
	}{
		{"known plan, null state -> copy", types.StringValue("a"), types.StringNull(), types.StringValue("a")},
		{"known plan, unknown state -> copy", types.StringValue("b"), types.StringUnknown(), types.StringValue("b")},
		{"known plan, known state -> keep state", types.StringValue("a"), types.StringValue("z"), types.StringValue("z")},
		{"null plan, null state -> keep state", types.StringNull(), types.StringNull(), types.StringNull()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			state := tc.state
			PreserveKnownTfValueIfStateNull(tc.plan, &state)
			assert.Equal(t, tc.wantState, state)
		})
	}
}

func TestPreserveKnownTfValueIfStateNull_Int64(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		plan      types.Int64
		state     types.Int64
		wantState types.Int64
	}{
		{"known plan, null state -> copy", types.Int64Value(7), types.Int64Null(), types.Int64Value(7)},
		{"known plan, unknown state -> copy", types.Int64Value(8), types.Int64Unknown(), types.Int64Value(8)},
		{"known plan, known state -> keep state", types.Int64Value(7), types.Int64Value(99), types.Int64Value(99)},
		{"null plan, null state -> keep state", types.Int64Null(), types.Int64Null(), types.Int64Null()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			state := tc.state
			PreserveKnownTfValueIfStateNull(tc.plan, &state)
			assert.Equal(t, tc.wantState, state)
		})
	}
}

func TestPreserveKnownTfValueIfStateNull_List(t *testing.T) {
	t.Parallel()

	knownList := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("x")})
	otherKnown := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("y")})

	cases := []struct {
		name      string
		plan      types.List
		state     types.List
		wantState types.List
	}{
		{"known plan, null state -> copy", knownList, types.ListNull(types.StringType), knownList},
		{"known plan, unknown state -> copy", knownList, types.ListUnknown(types.StringType), knownList},
		{"known plan, known state -> keep state", knownList, otherKnown, otherKnown},
		{"null plan, null state -> keep state", types.ListNull(types.StringType), types.ListNull(types.StringType), types.ListNull(types.StringType)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			state := tc.state
			PreserveKnownTfValueIfStateNull(tc.plan, &state)
			assert.Equal(t, tc.wantState, state)
		})
	}
}
