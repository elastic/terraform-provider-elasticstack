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

package entitycore

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveUpdateSpaceID(t *testing.T) {
	t.Run("Prior set returns Prior's space, ignoring req.SpaceID", func(t *testing.T) {
		prior := simpleWriteTestModel{ID: "abc", SpaceID: "prior-space"}

		got := ResolveUpdateSpaceID(KibanaWriteRequest[simpleWriteTestModel]{
			Plan:    simpleWriteTestModel{ID: "abc", SpaceID: "plan-space"},
			Prior:   &prior,
			SpaceID: "plan-space",
		})

		assert.Equal(t, "prior-space", got)
	})

	t.Run("Prior nil falls back to req.SpaceID", func(t *testing.T) {
		got := ResolveUpdateSpaceID(KibanaWriteRequest[simpleWriteTestModel]{
			Plan:    simpleWriteTestModel{ID: "abc"},
			SpaceID: "plan-space",
		})

		assert.Equal(t, "plan-space", got)
	})
}

func TestSimpleFleetCreate(t *testing.T) {
	t.Run("happy path calls toBody then apiCreate then populate", func(t *testing.T) {
		var gotSpaceID string
		var gotBody string
		resp := "created-body"

		writeFn := SimpleFleetCreate[simpleWriteTestModel, string, string](
			func(plan simpleWriteTestModel, _ context.Context) (string, diag.Diagnostics) {
				return "body-for-" + plan.ID, nil
			},
			func(_ context.Context, _ *fleetclient.Client, spaceID string, body string) (*string, diag.Diagnostics) {
				gotSpaceID = spaceID
				gotBody = body
				return &resp, nil
			},
			(*simpleWriteTestModel).setSpaceID,
		)

		result, diags := writeFn(context.Background(), &clients.KibanaScopedClient{}, KibanaWriteRequest[simpleWriteTestModel]{
			Plan:    simpleWriteTestModel{ID: "abc"},
			SpaceID: "default",
		})

		require.False(t, diags.HasError())
		assert.Equal(t, "default", gotSpaceID)
		assert.Equal(t, "body-for-abc", gotBody)
		assert.Equal(t, "default", result.Model.SpaceID)
	})

	t.Run("toBody error short-circuits before apiCreate", func(t *testing.T) {
		apiCreateCalled := false

		writeFn := SimpleFleetCreate[simpleWriteTestModel, string, string](
			func(_ simpleWriteTestModel, _ context.Context) (string, diag.Diagnostics) {
				var diags diag.Diagnostics
				diags.AddError("bad plan", "cannot convert")
				return "", diags
			},
			func(_ context.Context, _ *fleetclient.Client, _ string, _ string) (*string, diag.Diagnostics) {
				apiCreateCalled = true
				resp := "unused"
				return &resp, nil
			},
			(*simpleWriteTestModel).setSpaceID,
		)

		_, diags := writeFn(context.Background(), &clients.KibanaScopedClient{}, KibanaWriteRequest[simpleWriteTestModel]{
			Plan: simpleWriteTestModel{ID: "abc"},
		})

		require.True(t, diags.HasError())
		assert.False(t, apiCreateCalled)
	})

	t.Run("apiCreate error short-circuits before populate", func(t *testing.T) {
		populateCalled := false

		writeFn := SimpleFleetCreate[simpleWriteTestModel, string, string](
			func(_ simpleWriteTestModel, _ context.Context) (string, diag.Diagnostics) {
				return "body", nil
			},
			func(_ context.Context, _ *fleetclient.Client, _ string, _ string) (*string, diag.Diagnostics) {
				var diags diag.Diagnostics
				diags.AddError("api failed", "boom")
				return nil, diags
			},
			func(_ *simpleWriteTestModel, _ context.Context, _ string, _ *string) diag.Diagnostics {
				populateCalled = true
				return nil
			},
		)

		_, diags := writeFn(context.Background(), &clients.KibanaScopedClient{}, KibanaWriteRequest[simpleWriteTestModel]{
			Plan: simpleWriteTestModel{ID: "abc"},
		})

		require.True(t, diags.HasError())
		assert.False(t, populateCalled)
	})
}

func TestSimpleFleetUpdate(t *testing.T) {
	t.Run("happy path forwards writeID and Prior's space to apiUpdate then populate", func(t *testing.T) {
		var gotSpaceID, gotWriteID string
		resp := "updated-body"
		prior := simpleWriteTestModel{ID: "abc", SpaceID: "prior-space"}

		writeFn := SimpleFleetUpdate[simpleWriteTestModel, string, string](
			func(plan simpleWriteTestModel, _ context.Context) (string, diag.Diagnostics) {
				return "body-for-" + plan.ID, nil
			},
			func(_ context.Context, _ *fleetclient.Client, writeID, spaceID string, _ string) (*string, diag.Diagnostics) {
				gotWriteID = writeID
				gotSpaceID = spaceID
				return &resp, nil
			},
			(*simpleWriteTestModel).setSpaceID,
		)

		result, diags := writeFn(context.Background(), &clients.KibanaScopedClient{}, KibanaWriteRequest[simpleWriteTestModel]{
			Plan:    simpleWriteTestModel{ID: "abc", SpaceID: "plan-space"},
			Prior:   &prior,
			SpaceID: "plan-space",
			WriteID: "abc-123",
		})

		require.False(t, diags.HasError())
		assert.Equal(t, "abc-123", gotWriteID)
		assert.Equal(t, "prior-space", gotSpaceID)
		assert.Equal(t, "prior-space", result.Model.SpaceID)
	})

	t.Run("Prior nil falls back to req.SpaceID for apiUpdate and populate", func(t *testing.T) {
		var gotSpaceID string
		resp := "updated-body"

		writeFn := SimpleFleetUpdate[simpleWriteTestModel, string, string](
			func(_ simpleWriteTestModel, _ context.Context) (string, diag.Diagnostics) {
				return "body", nil
			},
			func(_ context.Context, _ *fleetclient.Client, _, spaceID string, _ string) (*string, diag.Diagnostics) {
				gotSpaceID = spaceID
				return &resp, nil
			},
			(*simpleWriteTestModel).setSpaceID,
		)

		_, diags := writeFn(context.Background(), &clients.KibanaScopedClient{}, KibanaWriteRequest[simpleWriteTestModel]{
			Plan:    simpleWriteTestModel{ID: "abc"},
			SpaceID: "plan-space",
			WriteID: "abc-123",
		})

		require.False(t, diags.HasError())
		assert.Equal(t, "plan-space", gotSpaceID)
	})

	t.Run("apiUpdate error short-circuits before populate", func(t *testing.T) {
		populateCalled := false

		writeFn := SimpleFleetUpdate[simpleWriteTestModel, string, string](
			func(_ simpleWriteTestModel, _ context.Context) (string, diag.Diagnostics) {
				return "body", nil
			},
			func(_ context.Context, _ *fleetclient.Client, _, _ string, _ string) (*string, diag.Diagnostics) {
				var diags diag.Diagnostics
				diags.AddError("api failed", "boom")
				return nil, diags
			},
			func(_ *simpleWriteTestModel, _ context.Context, _ string, _ *string) diag.Diagnostics {
				populateCalled = true
				return nil
			},
		)

		_, diags := writeFn(context.Background(), &clients.KibanaScopedClient{}, KibanaWriteRequest[simpleWriteTestModel]{
			Plan:    simpleWriteTestModel{ID: "abc"},
			WriteID: "abc-123",
		})

		require.True(t, diags.HasError())
		assert.False(t, populateCalled)
	})
}
