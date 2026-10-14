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

package casetemplate

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestSuppressSemanticYAMLDiff(t *testing.T) {
	const state = "name: triage\nseverity: low\nfields:\n  - name: a\n    type: keyword\n"

	tests := []struct {
		name        string
		state       types.String
		plan        types.String
		wantPlanned types.String
	}{
		{
			name:        "whitespace and comment only change keeps state value",
			state:       types.StringValue(state),
			plan:        types.StringValue("# comment\nname:   triage\nseverity: low\n\nfields:\n    - name: a\n      type: keyword\n"),
			wantPlanned: types.StringValue(state),
		},
		{
			name:        "key order only change keeps state value",
			state:       types.StringValue(state),
			plan:        types.StringValue("severity: low\nname: triage\nfields:\n  - type: keyword\n    name: a\n"),
			wantPlanned: types.StringValue(state),
		},
		{
			name:        "different value keeps planned value",
			state:       types.StringValue(state),
			plan:        types.StringValue("name: triage\nseverity: high\nfields:\n  - name: a\n    type: keyword\n"),
			wantPlanned: types.StringValue("name: triage\nseverity: high\nfields:\n  - name: a\n    type: keyword\n"),
		},
		{
			name:        "list order change keeps planned value",
			state:       types.StringValue("fields: [a, b]\n"),
			plan:        types.StringValue("fields: [b, a]\n"),
			wantPlanned: types.StringValue("fields: [b, a]\n"),
		},
		{
			name:        "null state keeps planned value on create",
			state:       types.StringNull(),
			plan:        types.StringValue(state),
			wantPlanned: types.StringValue(state),
		},
		{
			name:        "unknown plan is left unchanged",
			state:       types.StringValue(state),
			plan:        types.StringUnknown(),
			wantPlanned: types.StringUnknown(),
		},
		{
			name:        "invalid YAML falls back to the planned value",
			state:       types.StringValue(state),
			plan:        types.StringValue("a: [unterminated"),
			wantPlanned: types.StringValue("a: [unterminated"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := planmodifier.StringRequest{StateValue: tc.state, PlanValue: tc.plan, ConfigValue: tc.plan}
			resp := &planmodifier.StringResponse{PlanValue: tc.plan}

			suppressSemanticYAMLDiff().PlanModifyString(context.Background(), req, resp)

			assert.False(t, resp.Diagnostics.HasError())
			assert.True(t, tc.wantPlanned.Equal(resp.PlanValue), "got %s", resp.PlanValue)
		})
	}
}
