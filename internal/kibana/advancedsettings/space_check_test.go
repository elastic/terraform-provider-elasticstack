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

package advancedsettings

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnsureSpaceExists_defaultSpaceIsNeverLookedUp(t *testing.T) {
	t.Parallel()
	lookups := 0
	getSpace := func(context.Context, string) (*kbapi.SpaceResponse, diag.Diagnostics) {
		lookups++
		return nil, nil
	}

	diags := ensureSpaceExists(t.Context(), "default", getSpace)

	require.False(t, diags.HasError())
	assert.Zero(t, lookups)
}

func TestEnsureSpaceExists_missingSpaceIsReportedAsMustExist(t *testing.T) {
	t.Parallel()
	getSpace := func(context.Context, string) (*kbapi.SpaceResponse, diag.Diagnostics) {
		return nil, nil
	}

	diags := ensureSpaceExists(t.Context(), "ops", getSpace)

	require.Len(t, diags.Errors(), 1)
	assert.Contains(t, diags.Errors()[0].Detail(), `"ops"`)
	assert.Contains(t, diags.Errors()[0].Detail(), "must exist")
}

func TestEnsureSpaceExists_spacesAPIFailureIsSurfacedAndNotTreatedAsMissing(t *testing.T) {
	t.Parallel()
	getSpace := func(context.Context, string) (*kbapi.SpaceResponse, diag.Diagnostics) {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Unexpected status code from server: got HTTP 403", "forbidden")}
	}

	diags := ensureSpaceExists(t.Context(), "ops", getSpace)

	require.Len(t, diags.Errors(), 1)
	assert.Equal(t, "Unexpected status code from server: got HTTP 403", diags.Errors()[0].Summary())
	assert.NotContains(t, diags.Errors()[0].Detail(), "must exist")
}

func TestEnsureSpaceExists_existingSpaceIsLookedUpByIDAndAccepted(t *testing.T) {
	t.Parallel()
	var lookedUp []string
	getSpace := func(_ context.Context, spaceID string) (*kbapi.SpaceResponse, diag.Diagnostics) {
		lookedUp = append(lookedUp, spaceID)
		return &kbapi.SpaceResponse{Id: spaceID}, nil
	}

	diags := ensureSpaceExists(t.Context(), "ops", getSpace)

	require.False(t, diags.HasError())
	assert.Equal(t, []string{"ops"}, lookedUp)
}

func TestSpaceIsMissing_defaultSpaceIsNeverLookedUpNorMissing(t *testing.T) {
	t.Parallel()
	lookups := 0
	getSpace := func(context.Context, string) (*kbapi.SpaceResponse, diag.Diagnostics) {
		lookups++
		return nil, nil
	}

	missing, diags := spaceIsMissing(t.Context(), "default", getSpace)

	require.False(t, diags.HasError())
	assert.False(t, missing)
	assert.Zero(t, lookups)
}

func TestSpaceIsMissing_aSpaceTheSpacesAPIDoesNotKnowIsMissing(t *testing.T) {
	t.Parallel()
	getSpace := func(context.Context, string) (*kbapi.SpaceResponse, diag.Diagnostics) {
		return nil, nil
	}

	missing, diags := spaceIsMissing(t.Context(), "ops", getSpace)

	require.False(t, diags.HasError())
	assert.True(t, missing)
}

func TestSpaceIsMissing_aSpacesAPIFailureIsSurfacedAndNeverReportedAsMissing(t *testing.T) {
	t.Parallel()
	getSpace := func(context.Context, string) (*kbapi.SpaceResponse, diag.Diagnostics) {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Unexpected status code from server: got HTTP 403", "forbidden")}
	}

	missing, diags := spaceIsMissing(t.Context(), "ops", getSpace)

	require.Len(t, diags.Errors(), 1)
	assert.Equal(t, "Unexpected status code from server: got HTTP 403", diags.Errors()[0].Summary())
	assert.False(t, missing)
}
