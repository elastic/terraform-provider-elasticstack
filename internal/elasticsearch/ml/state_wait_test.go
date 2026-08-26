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

package ml

import (
	"context"
	"testing"

	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWaitForResourceState(t *testing.T) {
	ctx := context.Background()

	t.Run("already in desired state settles immediately", func(t *testing.T) {
		calls := 0
		state := "started"
		reached, diags := WaitForResourceState(ctx, "datafeed", "df-1", WaitForResourceStateConfig[string]{
			Get: func(_ context.Context) (*string, fwdiags.Diagnostics) {
				calls++
				return &state, nil
			},
			Desired: "started",
		})
		assert.False(t, diags.HasError())
		assert.True(t, reached)
		assert.Equal(t, 1, calls, "the immediate pre-check should avoid entering the poll loop")
	})

	t.Run("not found is an error by default", func(t *testing.T) {
		reached, diags := WaitForResourceState(ctx, "ml_job", "job-1", WaitForResourceStateConfig[string]{
			Get: func(_ context.Context) (*string, fwdiags.Diagnostics) {
				return nil, nil
			},
			Desired: "closed",
		})
		require.True(t, diags.HasError())
		assert.False(t, reached)
		assert.Contains(t, diags.Errors()[0].Summary(), "ml_job job-1 not found")
	})

	t.Run("not found is treated as desired when configured", func(t *testing.T) {
		reached, diags := WaitForResourceState(ctx, "ml_job", "job-1", WaitForResourceStateConfig[string]{
			Get: func(_ context.Context) (*string, fwdiags.Diagnostics) {
				return nil, nil
			},
			Desired:           "closed",
			NotFoundIsDesired: true,
		})
		assert.False(t, diags.HasError())
		assert.True(t, reached)
	})

	t.Run("terminal mismatch stops polling without an error", func(t *testing.T) {
		state := "stopped"
		reached, diags := WaitForResourceState(ctx, "datafeed", "df-1", WaitForResourceStateConfig[string]{
			Get: func(_ context.Context) (*string, fwdiags.Diagnostics) {
				return &state, nil
			},
			Desired:  "started",
			Terminal: map[string]struct{}{"stopped": {}, "started": {}},
		})
		assert.False(t, diags.HasError())
		assert.False(t, reached)
	})

	t.Run("get error is surfaced as diagnostics", func(t *testing.T) {
		reached, diags := WaitForResourceState(ctx, "ml_job", "job-1", WaitForResourceStateConfig[string]{
			Get: func(_ context.Context) (*string, fwdiags.Diagnostics) {
				var diags fwdiags.Diagnostics
				diags.AddError("Failed to get ml_job stats", "boom")
				return nil, diags
			},
			Desired: "closed",
		})
		require.True(t, diags.HasError())
		assert.False(t, reached)
	})
}
