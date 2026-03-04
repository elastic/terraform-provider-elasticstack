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

package integration

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestSpaceIDPlanModifier_RequiresReplace(t *testing.T) {
	t.Parallel()

	defaultSpace := "default"
	mod := spaceIDRequiresReplace(defaultSpace)

	type tc struct {
		name        string
		state       types.String
		plan        types.String
		wantReplace bool
	}

	cases := []tc{
		{
			name:        "default to non-default requires replace",
			state:       types.StringValue(defaultSpace),
			plan:        types.StringValue("first"),
			wantReplace: true,
		},
		{
			name:        "non-default to default requires replace",
			state:       types.StringValue("first"),
			plan:        types.StringValue(defaultSpace),
			wantReplace: true,
		},
		{
			name:        "non-default to non-default (different) requires replace",
			state:       types.StringValue("first"),
			plan:        types.StringValue("second"),
			wantReplace: true,
		},
		{
			name:        "non-default to non-default (same) no replace",
			state:       types.StringValue("first"),
			plan:        types.StringValue("first"),
			wantReplace: false,
		},
		{
			name:        "default to null does not require replace",
			state:       types.StringValue(defaultSpace),
			plan:        types.StringNull(),
			wantReplace: false,
		},
		{
			name:        "null to default does not require replace",
			state:       types.StringNull(),
			plan:        types.StringValue(defaultSpace),
			wantReplace: false,
		},
		{
			name:        "default to unknown does not require replace",
			state:       types.StringValue(defaultSpace),
			plan:        types.StringUnknown(),
			wantReplace: false,
		},
		{
			name:        "unknown to default does not require replace",
			state:       types.StringUnknown(),
			plan:        types.StringValue(defaultSpace),
			wantReplace: false,
		},
		{
			name:        "non-default to null requires replace",
			state:       types.StringValue("first"),
			plan:        types.StringNull(),
			wantReplace: true,
		},
		{
			name:        "null to non-default requires replace",
			state:       types.StringNull(),
			plan:        types.StringValue("first"),
			wantReplace: true,
		},
		{
			name:        "non-default to unknown requires replace",
			state:       types.StringValue("first"),
			plan:        types.StringUnknown(),
			wantReplace: true,
		},
		{
			name:        "unknown to non-default requires replace",
			state:       types.StringUnknown(),
			plan:        types.StringValue("first"),
			wantReplace: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			req := planmodifier.StringRequest{
				StateValue: c.state,
				PlanValue:  c.plan,
			}
			var resp planmodifier.StringResponse
			mod.PlanModifyString(context.Background(), req, &resp)

			require.Equal(t, c.wantReplace, resp.RequiresReplace)
		})
	}
}
