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

package pfresource

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeAPI struct {
	getCalls  int
	lastID    string
	lastSpace string
	result    *string
	diags     diag.Diagnostics
}

func (a *fakeAPI) Create(context.Context, *kibanaoapi.Client, string, any) (string, diag.Diagnostics) {
	return "", nil
}
func (a *fakeAPI) Get(_ context.Context, _ *kibanaoapi.Client, spaceID string, resourceID string) (*string, diag.Diagnostics) {
	a.getCalls++
	a.lastSpace = spaceID
	a.lastID = resourceID
	return a.result, a.diags
}
func (a *fakeAPI) Update(context.Context, *kibanaoapi.Client, string, string, any) diag.Diagnostics {
	return nil
}
func (a *fakeAPI) Delete(context.Context, *kibanaoapi.Client, string, string) diag.Diagnostics {
	return nil
}

func TestComposeAndParseCompositeID(t *testing.T) {
	id := ComposeCompositeID("observability", "tool-123")
	assert.Equal(t, "observability/tool-123", id)

	parsed, diags := ParseCompositeID(id)
	require.False(t, diags.HasError())
	assert.Equal(t, "observability", parsed.ClusterID)
	assert.Equal(t, "tool-123", parsed.ResourceID)
}

func TestEffectiveSpaceID_DefaultsToDefault(t *testing.T) {
	assert.Equal(t, DefaultSpaceID, EffectiveSpaceID(types.StringNull()))
	assert.Equal(t, DefaultSpaceID, EffectiveSpaceID(types.StringUnknown()))
	assert.Equal(t, DefaultSpaceID, EffectiveSpaceID(types.StringValue("")))
	assert.Equal(t, "security", EffectiveSpaceID(types.StringValue("security")))
}

func TestEnforceVersion_WithNilMinimumVersionIsNoop(t *testing.T) {
	diags := EnforceVersion(context.Background(), nil, VersionRequirement{})
	assert.False(t, diags.HasError())
}

func TestReadAfterWrite_UsesFollowUpRead(t *testing.T) {
	api := &fakeAPI{result: pointer("created-id")}
	result, diags := ReadAfterWrite[*string](context.Background(), api, nil, "default", "created-id")

	require.False(t, diags.HasError())
	require.NotNil(t, result)
	assert.Equal(t, 1, api.getCalls)
	assert.Equal(t, "default", api.lastSpace)
	assert.Equal(t, "created-id", api.lastID)
	assert.Equal(t, "created-id", *result)
}

func TestReadRemote_NotFoundReturnsPresentFalse(t *testing.T) {
	api := &fakeAPI{result: nil}
	result, present, diags := ReadRemote[*string](context.Background(), api, nil, "default", "missing")

	require.False(t, diags.HasError())
	assert.False(t, present)
	assert.Nil(t, result)
	assert.Equal(t, 1, api.getCalls)
}

func TestReadRemote_PropagatesDiagnostics(t *testing.T) {
	api := &fakeAPI{diags: diag.Diagnostics{diag.NewErrorDiagnostic("boom", "detail")}}
	result, present, diags := ReadRemote[*string](context.Background(), api, nil, "default", "broken")

	require.True(t, diags.HasError())
	assert.False(t, present)
	assert.Nil(t, result)
	assert.Equal(t, "boom", diags[0].Summary())
}

func TestVersionRequirement_StructureCarriesConfiguredMessage(t *testing.T) {
	req := VersionRequirement{
		MinimumVersion: version.Must(version.NewVersion("9.3.0")),
		ErrorSummary:   "Unsupported server version",
		ErrorDetail:    "Agent Builder agents require Elastic Stack v9.3.0 or later.",
	}

	assert.Equal(t, "9.3.0", req.MinimumVersion.String())
	assert.Equal(t, "Unsupported server version", req.ErrorSummary)
	assert.Equal(t, "Agent Builder agents require Elastic Stack v9.3.0 or later.", req.ErrorDetail)
}

func pointer[T any](v T) *T { return &v }

var _ ResourceAPI[any, any, *string] = (*fakeAPI)(nil)
var _ = clients.CompositeID{}
